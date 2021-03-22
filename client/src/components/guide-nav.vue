<template>
    <div>
        <vk-nav class="tm-nav main-nav">
            <ul
                class="uk-nav uk-nav-default tm-nav"
                :class="{ 'uk-margin-top': index }"
                :key="index"
                v-for="(subitems, category, index) in mainNavMenu"
            >
                <vk-nav-item-header :title="translate(category)"/>
                <router-link tag="li" exact :to="p" :key="p" v-for="(p, label) in subitems">
                    <a>{{ translate(label) }}</a>
                </router-link>
            </ul>
        </vk-nav>
        <br>
        <vk-nav class="tm-nav">
            <ul
                class="uk-nav uk-nav-default tm-nav"
                :class="{ 'uk-margin-top': index }"
                :key="index"
                v-for="(subitems, category, index) in results"
            >
                <vk-nav-item-header :title="translate(category)"/>
                <router-link tag="li" exact :to="p" :key="p" v-for="(p, label) in subitems">
                    <a>{{ translate(label) }}</a>
                </router-link>
            </ul>
        </vk-nav>
        <div class="uk-margin">
            <div class="uk-text-center">
                <el-select v-model="locale" placeholder="Language">
                    <el-option
                        v-for="item in options"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value"
                    >
                        <span style="float: left">{{ item.label }}</span>
                        <span style="float: right; color: #8492a6; font-size: 13px">
              <img :src="'/images/lang/' + item.value + '.png'">
            </span>
                    </el-option>
                </el-select>
            </div>
        </div>
    </div>
</template>

<script>
import sidebarStore from "@/store/index.js";
import { availableLocales, setLocale } from '@/localization';

export default {
    data() {
        return {
            mainNavMenu: [],
            locale: this.$i18n.locale,
            options: [
                {
                    value: "ru",
                    label: this.$t('App.Menu.SelectItem.Russian')
                },
                {
                    value: "en",
                    label: this.$t('App.Menu.SelectItem.English')
                }
            ],
            value: ""
        };
    },
    watch: {
        locale(val) {
            if (availableLocales.includes(val)) {
                setLocale(val);
            }
        }
    },
    mounted() {
        this.mainNavMenu = {
            Account: {
                Info: this.$router.resolve({
                    name: "account"
                }).route.path,
                Logout: this.$router.resolve({
                    name: "signin"
                }).route.path
            },
            Modules: {
                List: this.$router.resolve({
                    name: "system_modules"
                }).route.path
            },
            Agents: {
                List: this.$router.resolve({
                    name: "agents"
                }).route.path
            }
        };
    },
    beforeCreate() {
        sidebarStore.subscribe((mutation, state) => {
            if (mutation.type === "set") {
                this.hidden = mutation.payload.type === 'hidden' &&
                    mutation.payload.items === true;
            }
        });
    },
    computed: {
        results() {
            return sidebarStore.getters.results;
        }
    },
    methods: {
        translate(val) {
            return this.$tck('App.Menu.Item.', val);
        },
        hidden() {
            return sidebarStore.getters.hidden;
        }
    }
};
</script>
