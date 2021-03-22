<template>
    <div v-if="loading">
        <loader></loader>
    </div>
    <div v-else>
        <el-tabs v-model="activeModuleTab" class="module-view">
            <el-tab-pane name="view" :label="$t('Module.Info.TabTitle.Management')">
                <component
                    :is="subview"
                    :hash="agentHash"
                    :module="module"
                    :components="components"
                    :protoAPI="vxapi"
                    :eventsAPI="agentEventsAPI"
                    :modulesAPI="agentModulesAPI"
                ></component>
            </el-tab-pane>
            <el-tab-pane name="about" :label="$t('Module.Info.TabTitle.AboutModule')">
                <el-tabs tab-position="left" v-model="activeAboutTab">
                    <el-tab-pane name="info" :label="$t('Module.Info.TabTitle.Info')">
                        <table class="uk-table uk-table-striped uk-margin-small-top uk-margin-small-left">
                            <tbody>
                            <tr>
                                <td>{{ $t("Module.Info.Label.Name") }}</td>
                                <td>{{ parsedModuleInfo.title }}</td>
                            </tr>
                            <tr>
                                <td>{{ $t("Module.Info.Label.Description") }}</td>
                                <td>{{ parsedModuleInfo.description }}</td>
                            </tr>
                            <tr>
                                <td>{{ $t("Module.Info.Label.InstalledVersion") }}</td>
                                <td>{{ parsedModuleInfo.version }}</td>
                            </tr>
                            <tr>
                                <td>{{ $t("Module.Info.Label.SupportedOS") }}</td>
                                <td>{{ parsedModuleInfo.supportedOS }}</td>
                            </tr>
                            <tr>
                                <td>{{ $t("Module.Info.Label.Developer") }}</td>
                                <td>{{ parsedModuleInfo.developer }}</td>
                            </tr>
                            </tbody>
                        </table>
                    </el-tab-pane>
                    <el-tab-pane name="changelog" :label="$t('Module.Info.TabTitle.Versions')">
                        <listDtDd :content="parsedVersions" class="uk-margin-small-top uk-margin-small-left"></listDtDd>
                    </el-tab-pane>
                </el-tabs>
            </el-tab-pane>
        </el-tabs>
    </div>
</template>

<script>
import loader from "@/components/loader.vue";
import listDtDd from "@/components/list-dt-dd";
import VXAPI from "@/api/proto.js";
import eventsTable from "@/components/events-table.vue";
import agentModuleConfig from "@/components/agent-module-config.vue";
import AgentEvents from "@/api/agentEventsClass";
import AgentModules from "@/api/agentModulesClass";
import semverSort from "semver-sort";
import httpVueLoader from 'http-vue-loader'
import * as monaco from "monaco-editor";

export default {
    components: {
        listDtDd,
        loader
    },
    data() {
        return {
            subview: loader,
            module: {
                "info": {},
                "locale": {},
                "changelog": {}
            },
            loading: true,
            activeAboutTab: "info",
            activeModuleTab: "view",
            vxapi: {},
            agentEventsAPI: new AgentEvents({
                agentHash: this.$route.params.hash,
                http: this.$http
            }),
            agentModulesAPI: new AgentModules({
                agentHash: this.$route.params.hash,
                http: this.$http
            }),
            messages: [],
            agentHash: this.$route.params.hash,
            moduleName: this.$route.params.module,
            components: {
                agentModuleConfig,
                eventsTable,
                monaco
            }
        };
    },

    computed: {
        parsedModuleInfo() {
            let os = "";
            let title = "";
            let description = "";
            if ("os" in this.module["info"]) {
                os = Object.keys(this.module["info"]["os"]).join(", ");
            }
            if ("module" in this.module["locale"] && this.$i18n.locale in this.module["locale"]["module"]) {
                title = this.module["locale"]["module"][this.$i18n.locale]["title"];
            }
            if ("module" in this.module["locale"] && this.$i18n.locale in this.module["locale"]["module"]) {
                description = this.module["locale"]["module"][this.$i18n.locale]["description"];
            }
            return {
                title: title,
                description: description,
                version: this.module["info"]["version"],
                supportedOS: os,
                developer: "VXDev Team (support@vxcontrol.app)"
            };
        },
        parsedVersions() {
            let data = [];
            let title = "";
            let description = "";
            let versions = semverSort.desc(Object.keys(this.module["changelog"]));
            for (var i in versions) {
                title =
                    versions[i] +
                    " - " +
                    this.module["changelog"][versions[i]][this.$i18n.locale].date +
                    " - " +
                    this.module["changelog"][versions[i]][this.$i18n.locale].title;
                description = this.module["changelog"][versions[i]][this.$i18n.locale].description;
                data.push({title: title, description: description});
            }
            return data;
        }
    },
    beforeMount() {
        let vxProto = localStorage.getItem("vx_server_proto");
        let vxHost = localStorage.getItem("vx_server_host");
        let vxPort = localStorage.getItem("vx_server_port");
        let vxHostPort = "";
        if (vxProto) {
            vxHostPort = vxProto + "://" + vxHost + ":" + vxPort;
        } else if (vxPort == 443) {
            vxHostPort = "wss://" + vxHost + ":" + vxPort;
        } else {
            vxHostPort = "ws://" + vxHost + ":" + vxPort;
        }
        this.vxapi = new VXAPI({
            agentHash: this.agentHash,
            moduleName: this.moduleName,
            hostPort: vxHostPort
        });
        this.subview = httpVueLoader("/api/v1/agents/" +
            this.agentHash +
            "/modules/" +
            this.moduleName +
            "/bmodule.vue")
    },
    beforeDestroy() {
        this.vxapi.close();
    },
    unmounted() {
        console.log("Unmounted");
    },
    mounted() {
        console.log("AGENT MODULE: MOUNTED");
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
        this.loadModuleInfo();
        this.$root.$options.sidebarStore.dispatch("search", menu);
    },

    methods: {
        loadModuleInfo() {
            this.loading = true;
            this.$http
                .get(
                    "/api/v1/agents/" +
                    this.$route.params.hash +
                    "/modules/" +
                    this.$route.params.module
                )
                .then(r => {
                    if (r.data.status == "success") {
                        this.module = r.data.data;
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
        loadAgentInfo() {
            this.$http
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
