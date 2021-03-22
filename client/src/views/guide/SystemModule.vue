<template>
    <div v-if="loading">
        <loader></loader>
    </div>
    <div v-else>
        <vk-tabs align="justify">
            <vk-tabs-item :title="$t('Module.Info.TabTitle.AboutModule')">
                <vk-tabs-vertical align="left">
                    <vk-tabs-item :title="$t('Module.Info.TabTitle.Info')">
                        <table class="uk-table uk-table-striped">
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
                    </vk-tabs-item>
                    <vk-tabs-item :title="$t('Module.Info.TabTitle.Versions')">
                        <listDtDd :content="parsedVersions"></listDtDd>
                    </vk-tabs-item>
                </vk-tabs-vertical>
            </vk-tabs-item>
        </vk-tabs>
    </div>
</template>

<script>
import loader from "@/components/loader.vue";
import listDtDd from "@/components/list-dt-dd";
import semverSort from "semver-sort";

export default {
    components: {
        listDtDd,
        loader
    },
    data() {
        return {
            module: {
                "info": {},
                "locale": {},
                "changelog": {}
            },
            loading: false,
            versions: {}
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
            if (this.versions != null) {
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
        }
    },
    mounted() {
        this.loadModuleInfo();
    },

    methods: {
        mountMenu() {
            let menu = {
                Module: {
                    Info: this.$router.resolve({
                        name: "system_module_view",
                        params: {hash: this.$route.params.hash}
                    }).route.path
                }
            };
            if (localStorage.getItem("user_group") === "Admin" && this.module["info"]["system"] !== true) {
                menu.Module.Edit = this.$router.resolve({
                    name: "system_module_edit",
                    params: {hash: this.$route.params.hash}
                }).route.path
            }
            this.$root.$options.sidebarStore.dispatch("search", menu);
        },
        loadModuleInfo() {
            this.loading = true;
            this.$http
                .get("/api/v1/modules/" + this.$route.params.module)
                .then(r => {
                    if (r.data.status == "success") {
                        this.module = r.data.data;
                    } else {
                        throw new Error("response format error");
                    }
                    this.mountMenu();
                    this.loading = false;
                })
                .catch(e => {
                    console.log(e);
                    this.loading = false;
                });
        }
    }
};
</script>
