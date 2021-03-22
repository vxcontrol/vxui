<template>
    <div>
        <h2 class="uk-text-lead inline-block">{{ $t('Agents.Events.Title.Events') }} "{{ agent.description }}"</h2>
        <hr class="uk-divider-icon">
        <events-table :agentEvents="agentEventsAPI" :agentModules="agentModulesAPI"></events-table>
    </div>
</template>

<script>
import AgentEvents from "@/api/agentEventsClass";
import AgentModules from "@/api/agentModulesClass";
import eventsTable from "@/components/events-table.vue";

export default {
    components: {
        eventsTable
    },
    name: "agent-events",
    data() {
        return {
            agent: {},
            agentDescription: "",
            agentInfo: "",
            initAgentDesc: "",
            agentHash: "",
            isEditing: false,
            editIconVisible: false,
            errors: [],
            agentEventsAPI: new AgentEvents({
                agentHash: this.$route.params.hash,
                http: this.$http
            }),
            agentModulesAPI: new AgentModules({
                agentHash: this.$route.params.hash,
                http: this.$http
            })
        };
    },
    created() {
        this.agentHash = this.$route.params.hash;
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
        this.loadAgentInfo();
        this.$root.$options.sidebarStore.dispatch("search", menu);
    },

    methods: {
        loadAgentInfo() {
            this.$http
                .get("/api/v1/agents/" + this.$route.params.hash)
                .then(r => {
                    if (r.data.status == "success") {
                        this.agent = {...r.data.data.agent, ...r.data.data.details};
                        this.initAgentDesc = this.agent.description;
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
