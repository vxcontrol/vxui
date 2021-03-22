<template>
    <div>
        <div v-if="loading">
            <loader></loader>
        </div>
        <div v-else>
            <h2 class="uk-text-lead">{{ $t("AgentsModules.Page.Header.AgentsModules") }} "{{ agent.description }}"</h2>
            <hr class="uk-divider-icon">
            <div class="uk-card uk-card-default uk-width-12@m" v-if="agent">
                <div class="uk-child-width-1-1@s uk-grid-match" uk-grid>
                    <agent-module-card
                        v-for="module in modules"
                        :agentHash="agentHash"
                        :module="module"
                        :messages="messages"
                        :details="getDetails(module.info.name)"
                        :key="module.info.name"
                    ></agent-module-card>
                </div>
            </div>
        </div>

        <vk-notification :messages.sync="messages"></vk-notification>
    </div>
</template>

<script>
import agentModuleCard from "@/components/agent-module-card.vue";
import loader from "@/components/loader.vue";

export default {
    components: {
        agentModuleCard,
        loader
    },
    data() {
        return {
            loading: false,
            agent: {},
            agentHash: "",
            messages: [],
            details: [],
            modules: []
        };
    },
    beforeMount() {
        this.agentHash = this.$route.params.hash;
        this.loadInfo();
    },
    mounted() {
        const menu = {
            Agent: {
                Overview: this.$router.resolve({
                    name: "agent_dashboard",
                    params: {hash: this.$route.params.hash}
                }).route.path,
                Modules: this.$router.resolve({
                    name: "agent_modules",
                    params: {hash: this.$route.params.hash}
                }).route.path,
                Events: this.$router.resolve({
                    name: "agent_events",
                    params: {hash: this.$route.params.hash}
                }).route.path
            }
        };
        this.$root.$options.sidebarStore.dispatch("search", menu);
    },

    methods: {
        getDetails(name) {
            return this.details.find(d => d.name === name);
        },
        async loadInfo() {
            this.loading = true;
            await this.$http
                .get("/api/v1/agents/" + this.$route.params.hash + "/modules")
                .then(r => {
                    if (r.data.status == "success") {
                        this.details = r.data.data.details;
                        this.modules = r.data.data.modules;
                    } else {
                        throw new Error("response format error");
                    }
                    this.loading = false;
                })
                .catch(e => {
                    console.log(e);
                    this.loading = false;
                });
            await this.$http
                .get("/api/v1/agents/" + this.$route.params.hash)
                .then(r => {
                    if (r.data.status == "success") {
                        this.agent = {...r.data.data.agent, ...r.data.data.details};
                    } else {
                        throw new Error("response format error");
                    }
                })
                .catch(e => {
                    console.log(e);
                });
        }
    }
};
</script>
