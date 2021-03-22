export default class API {
    constructor(params) {
        this._agentHash = params.agentHash
        this._moduleName = params.moduleName
        this._http = params.http
    }

    getEvents(params) {
        let self = this
        return new Promise(function (resolve, reject) {
            console.log(
                `API: getEvents ${self._agentHash} ${self._moduleName} ${JSON.stringify(
                    params
                )}`
            )
            self._http
                .get('/api/v1/')
                .then(
                    response => {
                        console.log('API RESPONSE: ', response.data)
                        resolve(response.data)
                    },
                    error => {
                        reject(error)
                    }
                )
        })
    }
}
