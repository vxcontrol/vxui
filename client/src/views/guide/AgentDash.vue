<template>
    <div>
        <form v-if="isEditing === true">
            <input
                class="uk-input uk-form-width-medium uk-form-large"
                type="text"
                :placeholder="$t('Agents.Dashboard.InputPlaceholder.AgentName')"
                v-model="agent.description"
            >
            <vk-button-group>
                <vk-button size="large" @click="clickOut">{{ $t('Common.Pseudo.Button.Save') }}</vk-button>
                <vk-button size="large" @click="cancelEdit">{{ $t('Common.Pseudo.Button.Abort') }}</vk-button>
            </vk-button-group>
        </form>
        <div>
            <h2 class="uk-text-lead inline-block" @click="isEditing = true" v-if="isEditing === false">
                {{ $t('Agents.Dashboard.Title.Agent', { description: agent.description }) }}
                <sup>
                    <vk-icon icon="cog" ratio="0.5"></vk-icon>
                </sup>
            </h2>
        </div>
        <hr class="uk-divider-icon">
        <div>
            <vk-tabs-vertical align="left">
                <vk-tabs-item :title="$t('Agents.Dashboard.TabTitle.Info')">
                    <div class="uk-card uk-card-default uk-width-12@m" v-if="agent">
                        <vk-table :data="data" hoverable>
                            <vk-table-column :title="$t('Agents.Dashboard.TableColumnTitle.Name')">
                                <p slot-scope="{ row }">{{ $tck("Agents.Dashboard.Label.", row.key) }}</p>
                            </vk-table-column>
                            <vk-table-column :title="$t('Agents.Dashboard.TableColumnTitle.Value')"
                                             cell="value"></vk-table-column>
                        </vk-table>
                    </div>
                </vk-tabs-item>
                <vk-tabs-item :title="$t('Agents.Dashboard.TabTitle.Statistics')">---</vk-tabs-item>
                <vk-tabs-item :title="$t('Agents.Dashboard.TabTitle.Settings')">---</vk-tabs-item>
            </vk-tabs-vertical>
        </div>

        <vk-notification :messages.sync="messages"></vk-notification>
    </div>
</template>

<script>
import osToImage from "@/components/os-to-image.vue";

export default {
    components: {
        osToImage
    },

    data() {
        return {
            agent: {},
            initAgentDesc: "",
            isEditing: false,
            editIconVisible: false,
            messages: [],
            data: []
        };
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
                        this.data.push(
                            {
                                key: "hash",
                                value: this.agent.hash
                            },
                            {
                                key: "status",
                                value: this.agent.status
                            },
                            {
                                key: "os_type",
                                value: this.agent.info.os.type
                            },
                            {
                                key: "os_arch",
                                value: this.agent.info.os.arch
                            },
                            {
                                key: "os_name",
                                value: this.agent.info.os.name
                            },
                            {
                                key: "ip",
                                value: this.agent.ip
                            }
                        );
                    } else {
                        throw new Error("response format error");
                    }
                })
                .catch(e => {
                    console.log(e);
                });
        },
        clickOut() {
            if (this.agent.description.length < 1) {
                this.agent.description =
                    "Agent-" +
                    Math.random()
                        .toString(36)
                        .replace(/[^a-z]+/g, "")
                        .substr(0, 5);
            }

            this.$http
                .post("/api/v1/agents/" + this.$route.params.hash, this.agent)
                .then(r => {
                    if (r.data.status == "success") {
                        this.isEditing = false;
                        this.initAgentDesc = this.agent.description;
                    } else {
                        throw new Error("response format error");
                    }
                })
                .catch(e => {
                    console.log(e);
                    this.messages.push({
                        message: this.$t("SignIn.Notifications.Text.InvalidCredentials"),
                        status: "danger"
                    });
                });
        },
        cancelEdit() {
            this.agent.description = this.initAgentDesc;
            this.isEditing = false;
        }
    }
};
</script>
