import pb from 'protobufjs/light'
import AgentProto from '@/proto/agent.proto'
import ProtocolProto from '@/proto/protocol.proto'
import EventEmitter from 'events'

const DEBUG = false;
let _agentProto = pb.Root.fromJSON(AgentProto);
let _protocolProto = pb.Root.fromJSON(ProtocolProto);
let protocolPacket = _protocolProto.lookupType('protocol.Packet');
let protocolPacketContent = _protocolProto.lookupType('protocol.Packet.Content');

class Events extends EventEmitter {
};

let authenticationResponse = _agentProto.lookupType(
    'agent.AuthenticationResponse'
);

let debug = (...args) => {
    if (DEBUG) console.log(...args);
}

class PublicAPI {
    constructor(vxapi) {
        this.vxapi = vxapi;
    }

    subscribe(f, type = '*') {
        debug(`SUBSCRIBE type= ${type}`);
        switch (type) {
            case '*':
                this.vxapi._userHandlers = {text: f, data: f};
                break
            case 'text':
                this.vxapi._userHandlers = {text: f};
                break
            case 'data':
                this.vxapi._userHandlers = {data: f};
                break
        }
    }

    getState() {
        return this.vxapi._state;
    }

    sendData(data) {
        debug(`sendData: ${data}`);
        let m = {
            type: 'data',
            data: data
        };
        this.vxapi._send(m);
    }

    sendText(text) {
        debug(`sendText: ${text}`);
        let m = {
            type: 'text',
            data: text
        };
        this.vxapi._send(m);
    }
}

export default class VXAPI {
    constructor(params) {
        this._module = params.moduleName;
        this._srcToken = '';
        this._dstToken = '';

        this.events = new Events();
        this._queue = [];

        this.pb = pb;
        this.endpoint = params.hostPort + '/api/v1/vxpws/browser/' + params.agentHash + '/';
        this._socket = null;
        this._state = 0; // init state of socket connection
        this.publicAPI = new PublicAPI(this);
        this.events.on('stateChanged', () => this.changedState());
    }

    changedState() {
        let self = this;
        if (this._state >= 2) {
            if (this._queue.length > 0) {
                debug('queue before: ', this._queue.length);
                for (let i = 0; i < self._queue.length; i++) {
                    this._socket.send(this._queue.pop());
                }
                debug('queue after: ', this._queue.length);
            }
        }
    }

    connect() {
        let self = this;
        return new Promise(function (resolve, reject) {
            let socket = new WebSocket(self.endpoint); // get endpoint from private field
            socket.onopen = function () {
                self._socket = socket;
                self._handshake();
                const pingDelay = 5000;
                const pingBuffer = new Uint8Array([80, 73, 78, 71]);
                ;
                let sendPing = () => {
                    if ([0, 1].includes(socket.readyState)) {
                        if (self._state == 2) socket.send(pingBuffer.buffer);
                        debug(`_send: ping`);
                        setTimeout(sendPing, pingDelay);
                    } else {
                        debug("ping sender on connection was closed", socket.readyState);
                    }
                };
                setTimeout(sendPing, pingDelay);
                resolve(self.publicAPI);
            }
            socket.onclose = function (event) {
                if (event.wasClean || self._state == 3) {
                    debug('Connection was closed correctly: ', event.wasClean, self._state);
                    self._state = 4; // closed state
                } else {
                    debug('Dirty connection closed: ', self._state);
                    debug('Return code: ' + event.code + ' reason: ' + event.reason)
                    setTimeout(() => {
                        self.connect();
                    }, 1000);
                }
            }
            socket.onerror = function (error) {
                reject(error);
            }
            socket.onmessage = function (e) {
                self._msgHandler(e);
            }
        })
    }

    close() {
        debug("run close function on socket");
        if (this._socket) {
            this._state = 3; // closing state
            this._socket.close();
        }
    }

    _send(msg) {
        let _packet = {
            module: this._module,
            source: this._srcToken,
            destination: this._dstToken,
            timestamp: Math.round(new Date().getTime() / 1000)
        };
        let _content = {};
        switch (msg.type) {
            case 'data':
                debug(`Pack content: ${(msg.type, msg.data)}`);
                _content = {
                    type: protocolPacketContent.Type.DATA,
                    data: new TextEncoder('utf-8').encode(msg.data)
                };
                break
            case 'text':
                _content = {
                    type: protocolPacketContent.Type.TEXT,
                    data: new TextEncoder('utf-8').encode(msg.data)
                };
                break
        }

        let err = protocolPacketContent.verify(_content);
        if (err) throw Error(err); // check error
        _packet.content = protocolPacketContent.create(_content);

        err = protocolPacket.verify(_packet);
        if (err) throw Error(err); // check error

        let packet = protocolPacket.create(_packet);
        let buffer = protocolPacket.encode(packet).finish();
        if (DEBUG) {
            debug(`_send: packet = ${Array.prototype.toString.call(buffer)}`);
        }
        if (this._state < 2) {
            this._queue.push(buffer); // add message to queue
        } else {
            this._socket.send(buffer);
        }
    }

    _msgHandler(msg) {
        let data = msg.data;
        let self = this;

        switch (this._state) {
            case 0: // waiting for auth request...
                break
            case 1: // waiting for auth response...
                try {
                    let fileReader = new FileReader();
                    fileReader.onload = function (event) {
                        let authenticationResponseMsg = authenticationResponse.decode(
                            new Uint8Array(event.target.result)
                        );
                        self._srcToken = authenticationResponseMsg.atoken;
                        self._dstToken = authenticationResponseMsg.stoken;
                        debug(`RECV HS RESPONSE ${JSON.stringify(authenticationResponseMsg)}`);
                        self._state = 2; // connected
                        self.events.emit('stateChanged');
                    }

                    fileReader.onerror = function (e) {
                        console.log(e);
                        self._state = 0;
                    }
                    fileReader.readAsArrayBuffer(data);
                } catch (e) {
                    if (e instanceof pb.util.ProtocolError) {
                        console.log(e);
                    } else {
                        debug('invalid format', e);
                    }
                }
                break
            case 2: // auth ok, recv packet
                let fileReader = new FileReader();
                fileReader.onload = function (event) {
                    let decodedMessage = protocolPacket.decode(
                        new Uint8Array(event.target.result)
                    );
                    let obj = protocolPacket.toObject(decodedMessage);
                    if (self._userHandlers !== undefined) {
                        switch (obj.content.type) {
                            case protocolPacketContent.Type.TEXT:
                                if (self._userHandlers.text !== undefined) {
                                    self._userHandlers.text(obj);
                                }
                                break
                            case protocolPacketContent.Type.DATA:
                                if (self._userHandlers.data !== undefined) {
                                    self._userHandlers.data(obj);
                                }
                                break
                        }
                    }
                }

                fileReader.onerror = function (e) {
                    console.log(e);
                }
                fileReader.readAsArrayBuffer(data);
                break
            default:
                debug('invalid state', this._state);
        }
    }

    _handshake() {
        let authenticationRequest = _agentProto.lookupType(
            'agent.AuthenticationRequest'
        );
        let ts = Math.round(new Date().getTime() / 1000);
        let message = authenticationRequest.create({
            timestamp: ts,
            atoken: 'empty'
        });
        let buffer = authenticationRequest.encode(message).finish();
        this._state = 1;
        this._socket.send(buffer);
    }
}
