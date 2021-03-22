<template>
    <div>
        <div class="tm-sidebar-left uk-visible@m">
            <h3>{{ $t("App.Menu.Title.Menu") }}</h3>
            <guide-nav class="uk-margin-bottom"/>
        </div>

        <div class="tm-main uk-section uk-section-default">
            <div class="uk-container uk-container-xsmall uk-position-relative">
                <transition name="slide">
                    <router-view></router-view>
                </transition>
            </div>
        </div>
    </div>
</template>

<script>
import GuideNav from "@/components/guide-nav";
import PageNav from "@/components/guide-page-nav";

export default {
    components: {
        GuideNav,
        PageNav
    },
    data: () => ({}),
    mounted() {
        this.$el.addEventListener("click", e => {
            const isAnchor = e.target.nodeName === "A";
            const isDemo = e.target.getAttribute("href") === "#";

            if (isAnchor && isDemo) {
                e.preventDefault();
            }
        });
    }
};
</script>
<style>
div.__array-form-item.__ncform-control > div.list-item > div.list-item-label > label {
    padding-right: 10px;
    font-weight: bold;
}

div.__array-form-item.__ncform-control > div.list-item > div.list-item-label > div {
    padding-top: 0px;
    float: right;
}

div.__array-form-item.__ncform-control > div.el-button-group {
    padding-top: 6px;
    padding-left: 10px;
}

div.__array-table-form-item.__ncform-control .table thead th {
    font-size: 14px;
    font-weight: inherit;
}

div.__object-form-item > legend,
div.__array-tabs-form-item > legend,
div.__array-form-item > legend {
    display: flex;
    background-color: azure;
}

.el-button--primary {
    background-color: #1e87e0 !important;
}

.el-collapse-item__header {
    font-size: 14px;
    font-weight: 500;
}

.module-view .el-tabs__nav.is-top {
    left: 50%;
}
</style>
