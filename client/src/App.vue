<template>
    <vk-offcanvas-content>
        <vk-sticky :media="960">
            <navbar @menu="offcanvas = true"/>
        </vk-sticky>

        <router-view/>

        <vk-offcanvas :show.sync="offcanvas" :overlay="true" transition="push">
            <div class="uk-panel">
                <guide-nav class="uk-margin-bottom"/>
            </div>
        </vk-offcanvas>
    </vk-offcanvas-content>
</template>

<script>
import Navbar from "@/components/navbar";
import GuideNav from "@/components/guide-nav";

export default {
    components: {Navbar, GuideNav},
    data: () => ({
        offcanvas: false
    })
};
</script>
