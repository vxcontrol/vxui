<template>
    <div>
        <div
            class="uk-card uk-card-hover uk-card-body uk-transition-toggle bordered"
            :class="moduleActive"
        >
            <div class="uk-card-badge uk-label" v-if="active == true">{{ module["info"]["version"] }}</div>
            <div class="uk-card-badge uk-label" v-else>{{ $t("Module.Card.Text.Inactive") }}</div>
            <h3 class="uk-card-title">{{ moduleInfo.title }}</h3>
            <p>{{ moduleInfo.description }}</p>
            <p>{{ $t("Module.Card.Text.Events", { today: details["today"], total: details["total"]}) }}</p>
            <span class="uk-label uk-label-success" v-for="tag in tags" :key="tag">#{{ tag }}</span>
            <div
                class="uk-transition-slide-bottom uk-position-bottom uk-overlay uk-overlay-default"
            >
                <router-link
                    :to="{ name: 'agent_module_view', params: {hash: agentHash, module: module['info']['name']} }"
                    v-if="active == true"
                >
                    <el-button
                        type="primary" size="medium"
                    >{{ $t("Common.Pseudo.Button.Open") }}
                    </el-button>
                </router-link>
                <el-button
                    type="primary" size="medium"
                    :loading="inProgressDeactivate"
                    @click="deactivate"
                    v-if="active == true"
                >{{ $t("Common.Pseudo.Button.Disable") }}
                </el-button>
                <el-button
                    type="primary" size="medium"
                    :loading="inProgressActivate"
                    @click="activate"
                    v-if="active == false"
                >{{ $t("Common.Pseudo.Button.Enable") }}
                </el-button>
                <el-button
                    type="primary" size="medium"
                    :loading="inProgressUpdate"
                    @click="update"
                    v-if="details['update'] == true"
                >{{ $t("Common.Pseudo.Button.Update") }}
                </el-button>
            </div>
        </div>
    </div>
</template>
<script>
export default {
    props: ["agentHash", "module", "details", "messages"],
    data() {
        return {
            active: this.details["active"],
            inProgressActivate: false,
            inProgressDeactivate: false,
            inProgressUpdate: false
        };
    },
    computed: {
        moduleInfo() {
            return this.module["locale"]["module"][this.$i18n.locale]; //this.$i18n.locale
        },
        moduleActive() {
            return {
                "uk-card-secondary": this.active == false
            };
        },
        tags() {
            let tags = [];
            for (var key in this.module["locale"]["tags"]) {
                if (this.module["locale"]["tags"].hasOwnProperty(key)) {
                    tags.push(this.module["locale"]["tags"][key][this.$i18n.locale]["title"]);
                }
            }
            return tags;
        }
    },

    methods: {
        update() {
            this.inProgressUpdate = true;
            this.$http
                .post(
                    '/api/v1/agents/' + this.$route.params.hash +
                    '/modules/' + this.module["info"]["name"],
                    {action: "update"}
                )
                .then(r => {
                    if (r.data.status == "success") {
                        this.messages.push({
                            message: this.$t('Module.Notifications.Text.UpdateSuccess'),
                            type: "success"
                        });
                        this.$router.push({
                            name: "agent_module_view",
                            params: {hash: this.agentHash, module: this.module["info"]["name"]}
                        });
                    } else {
                        throw new Error("response format error");
                    }
                    this.inProgressUpdate = false;
                })
                .catch(e => {
                    console.log(e);
                    this.messages.push({
                        message: this.$t('Module.Notifications.Text.UpdateError'),
                        type: "danger"
                    });
                    this.inProgressUpdate = false;
                });
        },
        activate() {
            this.inProgressActivate = true;
            this.$http
                .post(
                    '/api/v1/agents/' + this.$route.params.hash +
                    '/modules/' + this.module["info"]["name"],
                    {action: "activate"}
                )
                .then(r => {
                    if (r.data.status == "success") {
                        this.messages.push({
                            message: this.$t('Module.Notifications.Text.ActivateSuccess'),
                            type: "success"
                        });
                        this.$router.push({
                            name: "agent_module_view",
                            params: {hash: this.agentHash, module: this.module["info"]["name"]}
                        });
                    } else {
                        throw new Error("response format error");
                    }
                    this.inProgressActivate = false;
                    this.active = true;
                })
                .catch(e => {
                    console.log(e);
                    this.messages.push({
                        message: this.$t('Module.Notifications.Text.ActivateError'),
                        type: "danger"
                    });
                    this.inProgressActivate = false;
                });
        },
        deactivate() {
            this.inProgressDeactivate = true;
            this.$http
                .post(
                    '/api/v1/agents/' + this.$route.params.hash +
                    '/modules/' + this.module["info"]["name"],
                    {action: "deactivate"}
                )
                .then(r => {
                    if (r.data.status == "success") {
                        this.messages.push({
                            message: this.$t('Module.Notifications.Text.DeactivateSuccess'),
                            type: "success"
                        });
                    } else {
                        throw new Error("response format error");
                    }
                    this.inProgressDeactivate = false;
                    this.active = false;
                })
                .catch(e => {
                    console.log(e);
                    this.messages.push({
                        message: this.$t('Module.Notifications.Text.DeactivateError'),
                        type: "danger"
                    });
                    this.inProgressDeactivate = false;
                });
        }
    }
};
</script>
