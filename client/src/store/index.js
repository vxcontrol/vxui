import Vue from "vue";
import Vuex from "vuex";

Vue.use(Vuex);

const sidebarStore = new Vuex.Store({
    state: {
        results: {},
        hidden: false
    },
    getters: {
        results(state) {
            return state.results;
        },
        hidden(state) {
            return state.hidden;
        }

    },
    mutations: {
        set(state, {type, items}) {
            state[type] = items;
        }
    },
    actions: {
        search({commit}, data) {
            commit("set", {type: "results", items: data});
            commit("set", {type: "hidden", items: false});
        },
        hide({commit}, data) {
            commit("set", {type: "hidden", items: true});
        },
        show({commit}, data) {
            commit("set", {type: "hidden", items: false});
        }
    }
});
export default sidebarStore;
