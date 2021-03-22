<template>
    <div>
        <h2 class="uk-h3 tm-heading-fragment">{{ $t("SignIn.Page.Header.SignInTitle") }}</h2>
        <el-form
            :model="login"
            :rules="rules"
            ref="login"
            label-width="150px"
            size="medium"
            style="width: 520px;"
            @submit.native.prevent="submit('login')"
        >
            <el-form-item :label='$t("SignIn.Form.Label.Email")' prop="mail">
                <el-input v-model="login.mail" autofocus=true></el-input>
            </el-form-item>
            <el-form-item :label='$t("SignIn.Form.Label.Password")' prop="password">
                <el-input v-model="login.password" type="password" show-password></el-input>
            </el-form-item>
            <el-form-item>
                <vue-recaptcha
                    v-if="sitekey"
                    ref="recaptcha"
                    size="invisible"
                    :sitekey="sitekey"
                    @verify="signin"
                    @expired="onCaptchaExpired"
                />
            </el-form-item>
            <el-form-item>
                <el-button
                    name="submit" type="primary" size="medium" native-type="submit"
                >{{ $t("Common.Pseudo.Button.Login") }}
                </el-button>
            </el-form-item>
            <router-link :to="{ name: 'signup' }">{{ $t("SignIn.Form.Link.SignUp") }}</router-link>
        </el-form>
        <vk-notification :messages.sync="messages"></vk-notification>
    </div>
</template>
<script>
import sidebarStore from "../../store/index.js";
import VueRecaptcha from 'vue-recaptcha';

export default {
    components: {
        VueRecaptcha
    },
    data() {
        return {
            login: {
                mail: "",
                password: "",
                token: ""
            },
            rules: {
                mail: [
                    {
                        required: true,
                        pattern: /^(vxadmin)|([A-Za-z0-9]+([._\\-]*[A-Za-z0-9])*@([A-Za-z0-9]+[-A-Za-z0-9]*[A-Za-z0-9]+.){1,63}[A-Za-z0-9]+)$/,
                        max: 50,
                        message: this.$t("SignIn.Form.ValidationText.EmailError"),
                        trigger: 'blur'
                    }
                ],
                password: [
                    {
                        min: 5,
                        max: 100,
                        required: true,
                        message: this.$t("SignIn.Form.ValidationText.PasswordError"),
                        trigger: 'blur'
                    }
                ]
            },
            sitekey: localStorage.getItem("recaptcha_html_key"),
            messages: []
        };
    },
    created() {
        document.cookie = 'auth=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        sidebarStore.dispatch("hide");
    },
    mounted() {
        this.reset();
    },
    methods: {
        reset() {
            if (this.sitekey) {
                this.$refs.recaptcha.reset();
            }
        },
        submit(name) {
            this.$refs[name].validate((valid) => {
                if (!this.sitekey) {
                    this.signin("");
                } else if (valid) {
                    this.$refs.recaptcha.execute();
                } else {
                    return false;
                }
            });
        },
        signin(token) {
            this.login.token = token;
            this.$http
                .post("/api/v1/signin", this.login)
                .then(r => {
                    if (r.data.status == "success") {
                        if (this.$route.params.nextUrl != null) {
                            this.$router.push(this.$route.params.nextUrl);
                        } else {
                            this.$router.push({name: "agents"});
                        }
                    } else {
                        throw new Error("response format error")
                    }
                })
                .catch(e => {
                    if (e.response.data.msg === "user is inactive") {
                        this.messages.push({
                            message: this.$t("SignIn.Notifications.Text.UserInactive"),
                            status: "danger"
                        });
                    } else {
                        this.messages.push({
                            message: this.$t("SignIn.Notifications.Text.InvalidCredentials"),
                            status: "danger"
                        });
                    }
                });
            this.reset();
        },
        onCaptchaExpired() {
            this.reset();
        }
    }
};
</script>
