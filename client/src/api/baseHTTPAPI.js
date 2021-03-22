export class BaseAPI {
    constructor(params) {
        if (params) {
            if (params.agentHash) this._agentHash = params.agentHash
            if (params.moduleName) this._moduleName = params.moduleName
            if (params.http) this._http = params.http
        }
    }

    getAgentHash() {
        return this._agentHash
    }

    get(endpoint, params) {
        let self = this
        if (params.params === undefined) params.params = {}

        // check params
        return new Promise(function (resolve, reject) {
            self._http.get(endpoint, {params: params.params}).then(
                response => {
                    resolve(response)
                },
                error => {
                    reject(error)
                }
            )
        })
    }

    post(endpoint, params) {
        let self = this
        return new Promise(function (resolve, reject) {
            self._http
                .post(endpoint, params, {
                    headers: {
                        'Content-Type': 'application/json'
                    }
                })
                .then(
                    response => {
                        if (
                            response.data.status !== undefined &&
                            response.data.status === 'error' &&
                            response.data.messages !== undefined &&
                            response.data.messages.length > 0
                        ) {
                            reject(new Error(response.data.messages[0]))
                        } else {
                            resolve(response)
                        }
                    },
                    error => {
                        reject(error)
                    }
                )
        })
    }
}
