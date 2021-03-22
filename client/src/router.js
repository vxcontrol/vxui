import Vue from 'vue'
import Router from 'vue-router'
import NProgress from 'nprogress'
import Application from './views/Application'
import SigninPage from './views/guide/Signin'
import SignupPage from './views/guide/Signup'
import AgentsPage from './views/guide/Agents'
import AgentDashboard from './views/guide/AgentDash'
import AgentModules from './views/guide/AgentModules'
import AgentModule from './views/guide/AgentModule'
import AgentEvents from './views/guide/AgentEvents'
import Account from './views/guide/Account'
import SystemModules from './views/guide/SystemModules'
import SystemModule from './views/guide/SystemModule'
import SystemModuleEdit from './views/guide/SystemModuleEdit'

import axios from 'axios'

Vue.use(Router)
const router = new Router({
    mode: 'history',
    linkActiveClass: 'uk-active',
    linkExactActiveClass: 'uk-active',
    routes: [
        {
            path: '/',
            redirect: '/app/'
        },
        {
            path: '/app',
            redirect: '/app/agents',
            component: Application,
            children: [
                {
                    path: 'signin',
                    name: 'signin',
                    component: SigninPage,
                    meta: {
                        public: true
                    }
                },
                {
                    path: 'signup',
                    name: 'signup',
                    component: SignupPage,
                    meta: {
                        public: true
                    }
                },
                {
                    path: 'agents',
                    name: 'agents',
                    component: AgentsPage
                },
                {
                    name: 'agent_dashboard',
                    path: 'agents/:hash/dash',
                    component: AgentDashboard
                },
                {
                    name: 'agent_events',
                    path: 'agents/:hash/events',
                    component: AgentEvents
                },
                {
                    name: 'agent_modules',
                    path: 'agents/:hash/modules',
                    component: AgentModules
                },
                {
                    name: 'agent_module_view',
                    path: 'agents/:hash/module/:module/view',
                    component: AgentModule
                },
                {
                    name: 'system_modules',
                    path: 'modules',
                    component: SystemModules
                },
                {
                    name: 'system_module_view',
                    path: 'modules/:module/view',
                    component: SystemModule
                },
                {
                    name: 'system_module_edit',
                    path: 'modules/:module/edit',
                    component: SystemModuleEdit
                },
                {
                    name: 'account',
                    path: 'account',
                    component: Account
                }
            ]
        }
    ]
})
router.beforeEach((to, from, next) => {
    if (to.matched.some(record => record.meta.public)) {
        next();
    } else {
        axios
            .get('/api/v1/info')
            .then(r => {
                localStorage.setItem('recaptcha_html_key', r.data.data.recaptcha_html_key);
                if (r.data.data.type === 'user') {
                    localStorage.setItem('vx_server_proto', r.data.data.server.proto);
                    localStorage.setItem('vx_server_host', r.data.data.server.host);
                    localStorage.setItem('vx_server_port', r.data.data.server.port);
                    localStorage.setItem('user_group', r.data.data.group.name);
                    next()
                } else {
                    next({name: 'signin', params: {nextUrl: to.fullPath}});
                }
            })
            .catch(e => {
                console.log(e);
            });
    }
})
router.beforeResolve((to, from, next) => {
    // If this isn't an initial page load.
    if (to.name) {
        // Start the route progress bar.
        NProgress.start();
    }
    next();
})
router.afterEach((to, from) => {
    // Complete the animation of the route progress bar.
    NProgress.done();
})
export default router
