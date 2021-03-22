import Vue from 'vue'
import App from './App.vue'
import router from './router'
import Vuikit from 'vuikit'
import VuikitIcons from '@vuikit/icons'
import '@/styles/index.less'
import sidebarStore from './store/index.js'
import axios from 'axios'

import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'

import vueNcform from '@vxcontrol/ncform'
import ncformStdComps from '@vxcontrol/ncform-theme-elementui'

import VueDataTables from 'vue-data-tables'
import demoBlock from '@/components/expander.vue'
import { elementUiMessages, getLocale, i18n } from '@/localization';

// Init plugin
Vue.use(Vuikit);
Vue.use(VuikitIcons);
Vue.use(VueDataTables);
Vue.component('demo-block', demoBlock);

Vue.prototype.$http = axios;
Vue.config.productionTip = false;

const locale = getLocale();

Vue.use(ElementUI, { locale: elementUiMessages[locale] })
Vue.use(vueNcform, { extComponents: ncformStdComps, lang: locale });

new Vue({
    i18n,
    router,
    sidebarStore,
    render: h => h(App)
}).$mount('#app')
