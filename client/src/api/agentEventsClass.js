import {BaseAPI} from './baseHTTPAPI'

export default class Events extends BaseAPI {
    constructor(p) {
        super(p)
        this.endpoint = '/api/v1/events/' // super.getAgentHash()
    }

    get(params) {
        return super.get(this.endpoint, {params: params})
    }

    update(params) {
        return super.post(this.endpoint, {params: params})
    }
}
