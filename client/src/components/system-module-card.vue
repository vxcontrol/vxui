<template>
    <div>
        <div
            class="uk-card uk-card-hover uk-card-body uk-transition-toggle bordered"
        >
            <div class="uk-card-badge uk-label">{{ module['info']['version'] }}</div>
            <h3 class="uk-card-title">{{ moduleInfo.title }}</h3>
            <p>{{ moduleInfo.description }}</p>
            <span class="uk-label uk-label-success" v-for="tag in tags" :key="tag">#{{ tag }}</span>
            <div
                class="uk-transition-slide-bottom uk-position-bottom uk-overlay uk-overlay-default"
            >
                <router-link
                    :to="{ name: 'system_module_view', params: {module: module['info']['name']} }"
                >
                    <el-button
                        type="primary" size="medium"
                    >{{ $t("Module.Card.Button.Info") }}
                    </el-button>
                </router-link>
                <router-link
                    :to="{ name: 'system_module_edit', params: {module: module['info']['name']} }"
                    v-if="isEditable"
                >
                    <el-button
                        type="primary" size="medium"
                    >{{ $t("Common.Pseudo.Button.Edit") }}
                    </el-button>
                </router-link>
                <el-button
                    v-if="isRemovable"
                    type="primary" size="medium"
                    :loading="inProgressRemove"
                    @click="remove"
                >{{ $t("Common.Pseudo.Button.Remove") }}
                </el-button>
            </div>
        </div>
    </div>
</template>
<script>
export default {
    props: ["module", "loader", "messages"],
    data() {
        return {
            inProgressRemove: false,
            userGroup: localStorage.getItem("user_group")
        };
    },
    computed: {
        moduleInfo() {
            return this.module["locale"]["module"][this.$i18n.locale];
        },
        isEditable() {
            return this.userGroup === "Admin" && this.module["info"]["system"] !== true;
        },
        isRemovable() {
            return this.userGroup === "Admin" && this.module["info"]["system"] !== true;
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
        remove() {
            this.$confirm(this.$t('Module.RemoveDialog.Text.Remove'), this.$t('Common.Pseudo.Text.Warning'), {
                confirmButtonText: this.$t('Common.Pseudo.Button.Ok'),
                cancelButtonText: this.$t('Common.Pseudo.Button.Cancel'),
                type: 'warning'
            }).then(() => {
                this.inProgressRemove = true;
                this.$http
                    .delete("/api/v1/modules/" + this.module["info"]["name"])
                    .then(r => {
                        if (r.data.status === "success") {
                            this.messages.push({
                                message: this.$t('Module.Notifications.Text.RemoveSuccess'),
                                type: "success"
                            });
                            this.loader();
                        } else {
                            throw new Error("response format error");
                        }
                        this.inProgressRemove = false;
                    })
                    .catch(e => {
                        console.log(e);
                        this.messages.push({
                            message: this.$t('Module.Notifications.Text.RemoveError'),
                            type: "danger"
                        });
                        this.inProgressRemove = false;
                    });
            });
        }
    }
};
</script>
