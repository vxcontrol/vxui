import {BaseAPI} from './baseHTTPAPI'

export default class Events extends BaseAPI {
    constructor(p) {
        super(p)
        this.endpoint = '/api/v1/agents/' + super.getAgentHash() + '/modules'
    }

    getHumanModulesNames() {
        const modules = []
        this.get({})
            .then(
                resp => {
                    if (resp.data.length > 0)
                        resp.data.forEach(e => {
                            modules.push({label: e.name, value: e.name})
                        })
                }
            )
            .catch(e => {
                console.log(e)
            })
        return modules
    }

    activateModule(name) {
        let params = {action: 'activate', data: {name: name}}
        return super.post(this.endpoint, {params: params})
    }

    deactivateModule(name) {
        let params = {action: 'deactivate', data: {name: name}}
        return super.post(this.endpoint, {params: params})
    }

    updateModule(name) {
        let params = {action: 'update', data: {name: name}}
        return super.post(this.endpoint, {params: params})
    }

    get(params) {
        return super.get(this.endpoint, {params: params})
    }

    update(params) {
        return super.post(this.endpoint, {params: params})
    }
}
