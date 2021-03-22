<template>
    <div>
        <div v-if="loading">
            <loader></loader>
        </div>
        <div v-else>
            <h2 class="uk-text-lead">{{ $t("Agents.Page.Header.AgentsList") }}</h2>
            <div class="uk-child-width-expand@s uk-margin-small-top uk-flex-middle uk-grid" uk-grid>
                <div class="uk-width-expand uk-first-column">
                    <div class="uk-child-width-expand@s uk-grid uk-grid-stack" uk-grid>
                        <div class="uk-width-expand uk-first-column">
                            <form class="uk-search uk-search-default uk-width-5-6" v-on:submit.prevent>
                <span uk-search-icon class="uk-search-icon uk-icon">
                  <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="20"
                      height="20"
                      viewBox="0 0 20 20"
                      icon="search-icon"
                      ratio="1"
                  >
                    <circle fill="none" stroke="#000" stroke-width="1.1" cx="9" cy="9" r="7"></circle>
                    <path fill="none" stroke="#000" stroke-width="1.1" d="M14,14 L18,18 L14,14 Z"></path>
                  </svg>
                </span>
                                <input
                                    class="uk-search-input uk-flex-wrap-stretch"
                                    id="search"
                                    type="search"
                                    :placeholder="$t('Agents.Page.InputPlaceholder.Search')"
                                    v-model="query"
                                    autocomplete="off"
                                >
                            </form>
                        </div>
                    </div>
                </div>
                <div class="uk-width-auto">
                    <el-button
                        type="primary" size="medium" icon="el-icon-plus"
                        @click="addAgentLayout = true"
                    >{{ $t("Agents.Page.Button.AddNewAgent") }}
                    </el-button>
                </div>
            </div>
            <hr class="uk-divider-icon">
            <div class="uk-card uk-card-default uk-width-12@m">
                <div
                    class="uk-child-width-1-1@s uk-grid-match"
                    uk-grid
                    v-for="agent in filteredAgents"
                    v-bind:key="agent.hash"
                >
                    <div>
                        <div
                            class="uk-card uk-card-default uk-card-hover uk-card-body bordered uk-transition-toggle"
                        >
                            <vk-grid gutter="small" class="uk-flex-middle">
                                <div class="uk-width-auto">
                                    <os-to-image :os="agent.info.os.type"></os-to-image>
                                </div>
                                <div class="uk-width-expand">
                                    <router-link
                                        :to="{name: 'agent_dashboard', params: {hash: agent.hash}}"
                                        class="uk-link-heading"
                                    >
                                        <vk-card-title class="uk-margin-remove-bottom">
                                            {{ agent.description }}
                                        </vk-card-title>
                                    </router-link>
                                    <p class="uk-text-meta uk-margin-remove-top">
                                        <time
                                            datetime="2016-04-01T19:00"
                                        >
                                            <status-to-image :status="agent.status">
                                            </status-to-image>
                                            {{ agent.info.os.type }} ({{ agent.info.os.arch }}) [{{ agent.ip }}]
                                        </time>
                                    </p>
                                </div>
                                <vk-table :data="agentData(agent)" hoverable>
                                    <vk-table-column cell="key">
                                        <p slot-scope="{ row }">{{ $tck(`Agents.AgentItem.Label.`, row.key) }}</p>
                                    </vk-table-column>
                                    <vk-table-column cell="value"></vk-table-column>
                                </vk-table>
                            </vk-grid>
                            <div
                                class="uk-transition-slide-bottom uk-position-bottom uk-overlay uk-overlay-default"
                            >
                                <router-link
                                    :to="{name: 'agent_dashboard', params: {hash: agent.hash}}"
                                    class="uk-link-heading"
                                >
                                    <el-button
                                        type="primary" size="medium"
                                    >{{ $t("Agents.AgentItem.Button.Manage") }}
                                    </el-button>
                                </router-link>
                                <router-link
                                    :to="{name: 'agent_modules', params: {hash: agent.hash}}"
                                    class="uk-link-heading"
                                >
                                    <el-button
                                        type="primary" size="medium"
                                    >{{ $t("Agents.AgentItem.Button.Modules") }}
                                    </el-button>
                                </router-link>
                                <router-link
                                    :to="{name: 'agent_events', params: {hash: agent.hash}}"
                                    class="uk-link-heading"
                                >
                                    <el-button
                                        type="primary" size="medium"
                                    >{{ $t("Agents.AgentItem.Button.Events") }}
                                    </el-button>
                                </router-link>
                                <el-button
                                    type="primary" size="medium"
                                    :loading="inProgressRemove"
                                    @click="deleteAgent(agent.hash)"
                                >{{ $t("Common.Pseudo.Button.Remove") }}
                                </el-button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <el-dialog
            :visible.sync="addAgentLayout"
            :show-close="false"
            :close-on-click-modal="false"
            center
            fullscreen
        >
            <vk-modal-full-close large @click="addAgentLayout = false"></vk-modal-full-close>
            <div class="uk-container uk-container-xsmall uk-position-center">
                <div class="uk-width-2xlarge">
                    <h1>{{ $t("Agents.CreateForm.Title.AddNewAgent") }}</h1>
                    <pre>{{ $t("Agents.CreateForm.Text.AddNewAgent") }}</pre>
                    <ncform
                        :form-schema="newAgentSchema"
                        form-name="new-agent"
                        v-model="newAgentModel"
                        @submit="createAgent()"
                    ></ncform>
                    <el-row style="margin: 0px auto; display: table;">
                        <el-button
                            size="medium"
                            :loading="inProgressDownload"
                            @click="downloadAgent"
                        >{{ $t("Common.Pseudo.Button.Download") }}
                        </el-button>
                        <el-button
                            type="primary" size="medium"
                            :autofocus="true"
                            @click="createAgent"
                        >{{ $t("Common.Pseudo.Button.Create") }}
                        </el-button>
                    </el-row>
                </div>
            </div>
        </el-dialog>

        <vk-notification :messages.sync="messages"></vk-notification>
    </div>
</template>

<script>
import newAgentSchema from "@/schemas/new_agent.json";
import osToImage from "@/components/os-to-image.vue";
import statusToImage from "@/components/status-to-image.vue";
import loader from "@/components/loader.vue";

export default {
    name: "agents",
    components: {
        osToImage,
        statusToImage,
        loader
    },
    data() {
        return {
            addAgentLayout: false,
            inProgressRemove: false,
            inProgressDownload: false,
            loading: false,
            agents: [],
            query: "",
            messages: [],
            newAgentModel: {},
            newAgentSchema
        };
    },
    mounted() {
        this.$root.$options.sidebarStore.dispatch("search", {});
        this.loadAgents();
    },
    methods: {
        agentData(agent) {
            return [
                {key: "active_modules", value: agent["active_modules"]},
                {key: "events_per_last_day", value: agent["events_per_last_day"]}
            ]
        },
        randID(length) {
            var result = '';
            var characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
            var charactersLength = characters.length;
            for (var i = 0; i < length; i++) {
                result += characters.charAt(Math.floor(Math.random() * charactersLength));
            }
            return result;
        },
        loadAgents() {
            this.loading = true;
            this.$http
                .get("/api/v1/agents/")
                .then(r => {
                    if (r.data.status == "success") {
                        this.agents = [];
                        const details = r.data.data.details;
                        const getOpts = (hash) => details.find(d => d.hash === hash);
                        r.data.data.agents.forEach(agent => {
                            this.agents.push({...agent, ...getOpts(agent.hash)});
                        });
                    } else {
                        throw new Error("response format error");
                    }
                    this.loading = false;
                })
                .catch(e => {
                    console.log(e);
                    this.loading = false;
                });
        },
        downloadAgent() {
            const ts = this.newAgentModel.os.split(".");
            const url = "/api/v1/downloads/vxagent/" + ts[0] + "/" + ts[1] + "?rand=" + this.randID(10);
            this.inProgressDownload = true;
            this.$http
                .get(url, {responseType: "blob"})
                .then(r => {
                    const name = r.headers["content-disposition"].split("filename=")[1].split("\"")[1];
                    const blob = new Blob([r.data], {type: "octet/stream"});
                    const link = document.createElement('a');
                    link.href = window.URL.createObjectURL(blob);
                    link.download = name;
                    link.style = "display: none";
                    document.body.appendChild(link);
                    link.click();
                    document.body.removeChild(link);
                    window.URL.revokeObjectURL(link.href);
                    this.inProgressDownload = false;
                })
                .catch(e => {
                    console.log(e);
                    this.inProgressDownload = false;
                    this.messages.push({
                        message: this.$t("Agents.Notifications.Text.DownloadError"),
                        status: "danger"
                    });
                });
        },
        createAgent() {
            this.$ncformValidate('new-agent').then(data => {
                if (data.result) {
                    let model = JSON.parse(JSON.stringify(this.newAgentModel));
                    const ts = model.os.split(".");
                    model.os = ts[0];
                    model.arch = ts[1];
                    this.loading = true;
                    this.addAgentLayout = false;
                    this.$http
                        .put("/api/v1/agents/", model, {
                            headers: {
                                'Content-Type': 'application/json'
                            }
                        })
                        .then(r => {
                            if (r.data.status == "success") {
                                this.newAgentModel = {};
                                this.$ncformReset('new-agent', {});
                                this.messages.push({
                                    message: this.$t("Agents.Notifications.Text.CreateSuccess"),
                                    status: "success"
                                });
                                this.loadAgents();
                            } else {
                                throw new Error("response format error");
                            }
                        })
                        .catch(e => {
                            console.log(e);
                            this.messages.push({
                                message: this.$t("Agents.Notifications.Text.CreateError"),
                                status: "danger"
                            });
                            this.loading = false;
                            this.addAgentLayout = true;
                        });
                }
            });
        },
        deleteAgent(hash) {
            this.$confirm(this.$t('Agents.RemoveDialog.Text.Remove'), this.$t('Common.Pseudo.Text.Warning'), {
                confirmButtonText: this.$t('Common.Pseudo.Button.Ok'),
                cancelButtonText: this.$t('Common.Pseudo.Button.Cancel'),
                type: 'warning'
            }).then(() => {
                this.inProgressRemove = true;
                this.$http
                    .delete("/api/v1/agents/" + hash)
                    .then(r => {
                        if (r.data.status === "success") {
                            this.messages.push({
                                message: this.$t("Agents.Notifications.Text.RemoveSuccess"),
                                type: "success"
                            });
                            this.loadAgents();
                        } else {
                            throw new Error("response format error");
                        }
                        this.inProgressRemove = false;
                    })
                    .catch(e => {
                        console.log(e);
                        this.messages.push({
                            message: this.$t("Agents.Notifications.Text.RemoveError"),
                            status: "danger"
                        });
                        this.inProgressRemove = false;
                    });
            });
        }
    },
    computed: {
        filteredAgents() {
            return this.agents.filter(agent => {
                const isContains = (val) => {
                    return val.toLowerCase().includes(this.query.toLowerCase());
                };
                return ["description", "ip", "hash", "info.os.type", "info.os.arch"].filter(field => {
                    const opts = field.split(".");
                    if (opts.length == 3) {
                        return isContains(agent[opts[0]][opts[1]][opts[2]]);
                    } else if (opts.length == 2) {
                        return isContains(agent[opts[0]][opts[1]]);
                    } else {
                        return isContains(agent[field]);
                    }
                }).length !== 0;
            });
        }
    }
};
</script>
<style scoped>
pre {
    background: #22313f;
    border-radius: 0.3em;
    box-sizing: border-box;
    color: #f8f8f2;
    margin: 1.6em 0;
    overflow: auto;
    padding: 10px;
    width: 100%;
    white-space: pre-wrap;
}
</style>
