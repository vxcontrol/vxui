<template>
    <div>
        <h2 class="uk-h3 tm-heading-fragment">{{ $t("SignUp.Page.Header.SignUp") }}</h2>
        <el-form
            :model="reg"
            :rules="rules"
            ref="reg"
            label-width="150px"
            size="medium"
            style="width: 520px;"
            @submit.native.prevent="submit('reg')"
        >
            <el-form-item :label='$t("SignUp.Form.Label.Email")' prop="mail">
                <el-input v-model="reg.mail" type="email" autofocus=true></el-input>
            </el-form-item>
            <el-form-item :label='$t("SignUp.Form.Label.Password")' prop="password">
                <el-input v-model="reg.password" type="password" show-password></el-input>
            </el-form-item>
            <el-form-item :label='$t("SignUp.Form.Label.RepeatPassword")' prop="confirm_password">
                <el-input v-model="reg.confirm_password" type="password" show-password></el-input>
            </el-form-item>
            <el-form-item>
                <vue-recaptcha
                    v-if="sitekey"
                    ref="recaptcha"
                    size="invisible"
                    :sitekey="sitekey"
                    @verify="signup"
                    @expired="onCaptchaExpired"
                />
            </el-form-item>
            <el-form-item>
                <el-button
                    name="submit" type="primary" size="medium" native-type="submit"
                >{{ $t("Common.Pseudo.Button.Register") }}
                </el-button>
            </el-form-item>
            <span>{{ $t("SignUp.Form.HintText.SignIn") }}</span><br>
            <router-link :to="{ name: 'signin' }">{{ $t("SignUp.Form.Link.SignIn") }}</router-link>
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
        var vConfirmPassword = (rule, value, callback) => {
            if (value !== this.reg.password) {
                callback(new Error(this.$t("SignInConfirmPasswordError")));
            } else {
                callback();
            }
        };
        return {
            reg: {
                mail: "",
                password: "",
                confirm_password: "",
                token: ""
            },
            rules: {
                mail: [
                    {
                        type: 'email',
                        required: true,
                        max: 50,
                        message: this.$t("SignUp.Form.ValidationText.EmailError"),
                        trigger: 'blur'
                    }
                ],
                password: [
                    {
                        min: 8,
                        max: 100,
                        required: true,
                        message: this.$t("SignUp.Form.ValidationText.PasswordError"),
                        trigger: 'blur'
                    },
                    {
                        pattern: "[0-9]",
                        message: this.$t("SignUp.Form.ValidationText.PasswordError"),
                        trigger: 'blur'
                    },
                    {
                        pattern: "[a-z]",
                        message: this.$t("SignUp.Form.ValidationText.PasswordError"),
                        trigger: 'blur'
                    },
                    {
                        pattern: "[A-Z]",
                        message: this.$t("SignUp.Form.ValidationText.PasswordError"),
                        trigger: 'blur'
                    },
                    {
                        pattern: "[!@#$&*]",
                        message: this.$t("SignUp.Form.ValidationText.PasswordError"),
                        trigger: 'blur'
                    }
                ],
                confirm_password: [
                    {
                        required: true,
                        message: this.$t("SignUp.Form.ValidationText.ConfirmPasswordError"),
                        trigger: 'blur'
                    },
                    {
                        validator: vConfirmPassword,
                        trigger: 'blur'
                    }
                ]
            },
            sitekey: localStorage.getItem("recaptcha_html_key"),
            messages: []
        };
    },
    created() {
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
                    this.signup("");
                } else if (valid) {
                    this.$refs.recaptcha.execute();
                } else {
                    return false;
                }
            });
        },
        signup(token) {
            this.reg.token = token;
            this.$http
                .post("/api/v1/signup", this.reg)
                .then(r => {
                    if (r.data.status == "success") {
                        this.messages.push({
                            message: this.$t("SignUp.Notifications.Text.Success"),
                            status: "success"
                        });
                        let that = this;
                        setTimeout(function () {
                            that.$router.push({name: "signin"});
                        }, 2000);
                    } else {
                        throw new Error("response format error");
                    }
                })
                .catch(e => {
                    this.messages.push({
                        message: this.$t("SignUp.Notifications.Text.AlreadyExists"),
                        status: "danger"
                    });
                });
            this.reset();
        },
        onCaptchaExpired() {
            this.reset();
        }
    }
};
</script>
