<template>
    <div>
        <div v-if="loading">
            <loader></loader>
        </div>
        <div v-else>
            <h2 class="uk-text-lead">{{ $t("Account.Page.Header.Account") }}</h2>
            <hr class="uk-divider-icon">
            <table class="uk-table uk-table-divider">
                <tbody>
                <tr>
                    <td>{{ $t("Account.Page.Label.Name") }}</td>
                    <td>{{ accountInfo.name }}</td>
                </tr>
                <tr>
                    <td>{{ $t("Account.Page.Label.Group") }}</td>
                    <td>{{ accountInfo.group.name }}</td>
                </tr>
                <tr>
                    <td>{{ $t("Account.Page.Label.Email") }}</td>
                    <td>{{ accountInfo.mail }}</td>
                </tr>
                <tr>
                    <td>{{ $t("Account.Page.Label.Password") }}</td>
                    <td>
                        <el-button
                            type="primary" size="medium"
                            @click="changePasswordLayout = true"
                        >{{ $t("Common.Pseudo.Button.Change") }}
                        </el-button>
                    </td>
                </tr>
                </tbody>
            </table>
        </div>

        <el-dialog
            :visible.sync="changePasswordLayout"
            :show-close="false"
            :close-on-click-modal="false"
            center
            fullscreen
        >
            <vk-modal-full-close large @click="changePasswordLayout = false"></vk-modal-full-close>
            <div class="uk-container uk-container-xsmall uk-position-center">
                <div class="uk-width-2xlarge">
                    <h1>{{ $t("Account.Form.Title.ChangePassword") }}</h1>
                    <ncform
                        :form-schema="changePasswordSchema"
                        form-name="change-password"
                        v-model="changePasswordModel"
                        @submit="changePassword()"
                    ></ncform>
                    <el-row style="margin: 0px auto; display: table;">
                        <el-button
                            type="primary" size="medium"
                            :autofocus="true"
                            @click="changePassword"
                        >{{ $t("Common.Pseudo.Button.Change") }}
                        </el-button>
                    </el-row>
                </div>
            </div>
        </el-dialog>

        <vk-notification :messages.sync="messages"></vk-notification>
    </div>
</template>

<script>
import changePasswordSchema from "@/schemas/change_password.json";
import loader from "@/components/loader.vue";

export default {
    name: "account",
    components: {
        loader
    },
    data() {
        return {
            changePasswordLayout: false,
            changePasswordModel: {},
            loading: false,
            messages: [],
            accountInfo: {
                group: {}
            },
            changePasswordSchema
        };
    },
    beforeMount() {
        this.loadAccountInfo();
    },
    mounted() {
        this.$root.$options.sidebarStore.dispatch("search", {});
    },
    methods: {
        changePassword() {
            this.$ncformValidate('change-password').then(data => {
                if (data.result) {
                    let model = JSON.parse(JSON.stringify(this.changePasswordModel));
                    this.loading = true;
                    this.changePasswordLayout = false;
                    this.$http
                        .post("/api/v1/users/current/password", model, {
                            headers: {
                                'Content-Type': 'application/json'
                            }
                        })
                        .then(r => {
                            if (r.data.status == "success") {
                                this.changePasswordModel = {};
                                this.$ncformReset('change-password', {});
                                this.messages.push({
                                    message: this.$t("Account.Notifications.Text.PasswordChangeSuccess"),
                                    status: "success"
                                });
                            } else {
                                throw new Error("response format error");
                            }
                            this.loading = false;
                        })
                        .catch(e => {
                            console.log(e);
                            this.messages.push({
                                message: this.$t("Account.Notifications.Text.PasswordChangeError"),
                                status: "danger"
                            });
                            this.loading = false;
                            this.changePasswordLayout = true;
                        });
                }
            });
        },
        loadAccountInfo() {
            this.$http
                .get("/api/v1/users/current")
                .then(r => {
                    if (r.data.status == "success") {
                        this.accountInfo = r.data.data;
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
