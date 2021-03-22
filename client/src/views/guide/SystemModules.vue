<template>
    <div>
        <div v-if="loading">
            <loader></loader>
        </div>
        <div v-else>
            <h2 class="uk-text-lead">{{ $t("SystemModules.Page.Header.SystemModules") }}</h2>
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
                                    v-model="query"
                                    autocomplete="off"
                                    :placeholder="$t('SystemModule.Page.InputPlaceholder.Search')"
                                >
                            </form>
                        </div>
                    </div>
                </div>
                <div class="uk-width-auto">
                    <el-button
                        type="primary" size="medium" icon="el-icon-plus"
                        @click="addModuleLayout = true"
                    >{{ $t("SystemModules.Page.Button.AddNewModule") }}
                    </el-button>
                </div>
            </div>
            <hr class="uk-divider-icon">
            <div class="uk-card uk-card-default uk-width-12@m" v-if="filteredModules.length !== 0">
                <div class="uk-child-width-1-1@s uk-grid-match" uk-grid>
                    <system-module-card
                        v-for="module in filteredModules"
                        :key="module.info.name"
                        :module="module"
                        :messages="messages"
                        :loader="loadModules"
                    ></system-module-card>
                </div>
            </div>
        </div>

        <el-dialog
            :visible.sync="addModuleLayout"
            :show-close="false"
            :close-on-click-modal="false"
            center
            fullscreen
        >
            <vk-modal-full-close large @click="addModuleLayout = false"></vk-modal-full-close>
            <div class="uk-container uk-container-xsmall uk-position-center">
                <div class="uk-width-2xlarge">
                    <h1>{{ $t("SystemModules.CreateForm.Title.AddNewModule") }}</h1>
                    <pre>{{ $t("SystemModules.CreateForm.Text.AddNewModule") }}</pre>
                    <ncform
                        :form-schema="newModuleSchema"
                        form-name="new-module"
                        v-model="newModuleModel"
                        @submit="createModule()"
                    ></ncform>
                    <el-button
                        type="primary" size="medium" style="margin: 0px auto; display: block;"
                        @click="createModule()"
                    >{{ $t("Common.Pseudo.Button.Create") }}
                    </el-button>
                </div>
            </div>
        </el-dialog>

        <vk-notification :messages.sync="messages"></vk-notification>
    </div>
</template>

<script>
import newModuleSchema from "@/schemas/new_module.json";
import systemModuleCard from "@/components/system-module-card.vue";
import loader from "@/components/loader.vue";
import utils from "@/utils.js";

export default {
    components: {
        systemModuleCard,
        loader
    },
    data() {
        return {
            addModuleLayout: false,
            loading: false,
            modules: [],
            query: "",
            messages: [],
            newModuleModel: {},
            newModuleSchema
        };
    },
    beforeMount() {
        this.loadModules();
    },
    mounted() {
        this.$root.$options.sidebarStore.dispatch("search", {});
    },
    methods: {
        async loadModules() {
            this.loading = true;
            await this.$http
                .get("/api/v1/modules/")
                .then(r => {
                    if (r.data.status == "success") {
                        this.modules = r.data.data;
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
        async storeModule(info) {
            this.loading = true;
            await this.$http
                .put("/api/v1/modules/", info, {
                    headers: {
                        'Content-Type': 'application/json'
                    }
                })
                .then(r => {
                    if (r.data.status == "success") {
                        this.newModuleModel = {};
                        this.$ncformReset('new-module', {});
                        this.messages.push({
                            message: this.$t("SystemModule.Notifications.Text.StoreSuccess"),
                            status: "success"
                        });
                        this.loadModules();
                    } else {
                        throw new Error("response format error");
                    }
                })
                .catch(e => {
                    console.log(e);
                    this.messages.push({
                        message: this.$t("SystemModule.Notifications.Text.StoreError"),
                        status: "danger"
                    });
                    this.loading = false;
                    this.addModuleLayout = true;
                });
        },
        createModule() {
            this.$ncformValidate('new-module').then(data => {
                if (data.result) {
                    let model = JSON.parse(JSON.stringify(this.newModuleModel));
                    model.os = model.os.reduce(function (acc, cur) {
                        const item = cur.split(".");
                        acc[item[0]] = (acc[item[0]] || []);
                        acc[item[0]].push(item[1]);
                        return acc;
                    }, {});
                    this.addModuleLayout = false;
                    this.storeModule(model);
                }
            });
        }
    },
    computed: {
        filteredModules() {
            const fields = ["info.name", "info.version", "locale"]
            return this.modules.filter(utils.getFilterObjs(fields, this.query))
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
