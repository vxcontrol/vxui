<template>
    <vk-navbar-full expanded class="tm-navbar-container">
        <vk-navbar-item>
            <vk-navbar-logo>
                <router-link to="/" class="uk-link-reset">
                    <img src="/images/vuikit-logo.svg">VX
                </router-link>
            </vk-navbar-logo>
        </vk-navbar-item>

        <vk-navbar-item slot="right">
            <vk-navbar-nav class="uk-visible@m">
                <router-link to="/app" tag="li">
                    <a>{{ $t("App.Navbar.Item.Dashboard") }}</a>
                </router-link>
                <router-link :to="{name: 'account'}" tag="li">
                    <a>{{ $t("App.Navbar.Item.Account") }}</a>
                </router-link>
            </vk-navbar-nav>

            <vk-navbar-toggle class="uk-hidden@m" @click="$emit('menu')"/>
        </vk-navbar-item>
    </vk-navbar-full>
</template>
