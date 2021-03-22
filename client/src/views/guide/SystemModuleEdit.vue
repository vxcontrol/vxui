<template>
    <div v-if="loading">
        <loader></loader>
    </div>
    <div v-else>
        <h2 class="uk-text-lead">{{ $t("SystemModule.Page.Header.SystemModuleEditor") }}</h2>
        <el-tabs tab-position="left" v-model="activeTab">
            <el-tab-pane name="general" :label="$t('SystemModule.Page.TabTitle.General')">
                <ncform
                    name="from-edit-module-general"
                    :form-schema="editModuleGeneralSchema"
                    form-name="edit-module-general"
                    v-model="moduleGeneralModel"
                    @change="onChangeGeneralModel"
                ></ncform>
                <el-button
                    v-if="edited"
                    type="primary" size="medium"
                    style="margin-left: 143px"
                    :loading="inProgressSave"
                    @click="saveModule"
                >{{ $t("Common.Pseudo.Button.Save") }}
                </el-button>
            </el-tab-pane>
            <el-tab-pane name="config" :label="$t('SystemModule.Page.TabTitle.Config')">
                <ncform
                    name="from-edit-module-config"
                    :form-schema="editModuleConfigSchema"
                    form-name="edit-module-config"
                    v-model="moduleConfigSchemaModel"
                    @change="onChangeConfigModel"
                ></ncform>
                <ncform
                    name="from-edit-module-default-config"
                    :form-schema="editModuleDefaultConfigSchema"
                    form-name="edit-module-default-config"
                    v-model="moduleDefaultConfigModel"
                    @change="onChangeDefaultConfigModel"
                ></ncform>
                <el-button
                    v-if="edited"
                    type="primary" size="medium"
                    style="margin-left: 13px"
                    :loading="inProgressSave"
                    @click="saveModule"
                >{{ $t("Common.Pseudo.Button.Save") }}
                </el-button>
            </el-tab-pane>
            <el-tab-pane name="events" :label="$t('SystemModule.Page.TabTitle.Events')">
                <ncform
                    name="from-edit-module-events"
                    :form-schema="editModuleEventsSchema"
                    form-name="edit-module-events"
                    v-model="moduleEventsModel"
                    @change="onChangeEventsModel"
                ></ncform>
                <ncform
                    name="from-edit-module-events-default-config"
                    :form-schema="editModuleDefaultEventConfigSchema"
                    form-name="edit-module-events-default-config"
                    v-model="moduleDefaultEventConfigModel"
                    @change="onChangeDefaultEventConfigModel"
                ></ncform>
                <el-button
                    v-if="edited"
                    type="primary" size="medium"
                    style="margin-left: 13px"
                    :loading="inProgressSave"
                    @click="saveModule"
                >{{ $t("Common.Pseudo.Button.Save") }}
                </el-button>
            </el-tab-pane>
            <el-tab-pane name="locale" :label="$t('SystemModule.Page.TabTitle.Locale')">
                <ncform
                    name="from-edit-module-locale"
                    :form-schema="editModuleLocaleSchema"
                    form-name="edit-module-locale"
                    v-model="moduleLocaleModel"
                    @change="onChangeLocaleModel"
                ></ncform>
                <el-button
                    v-if="edited"
                    type="primary" size="medium"
                    style="margin-left: 13px"
                    :loading="inProgressSave"
                    @click="saveModule"
                >{{ $t("Common.Pseudo.Button.Save") }}
                </el-button>
            </el-tab-pane>
            <el-tab-pane name="files-new" :label="$t('SystemModule.Page.TabTitle.Files')">
                <system-module-files
                    v-if="isLoaded"
                    :module="module"
                    :messages="messages"
                    :files="files"
                    @onBusy="setBusyState"
                ></system-module-files>
                <el-button
                    v-if="edited"
                    type="primary" size="medium"
                    style="margin-left: 13px; margin-top: 25px"
                    :loading="inProgressSave"
                    @click="saveModule"
                >{{ $t("Common.Pseudo.Button.Save") }}
                </el-button>
            </el-tab-pane>
            <el-tab-pane name="changelog" :label="$t('SystemModule.Page.TabTitle.Changelog')">
                <ncform
                    name="from-edit-module-changelog"
                    :form-schema="editModuleChangelogSchema"
                    form-name="edit-module-changelog"
                    v-model="moduleChangelogModel"
                    @change="onChangeChangelogModel"
                ></ncform>
                <el-button
                    v-if="edited"
                    type="primary" size="medium"
                    style="margin-left: 13px"
                    :loading="inProgressSave"
                    @click="saveModule"
                >{{ $t("Common.Pseudo.Button.Save") }}
                </el-button>
            </el-tab-pane>
        </el-tabs>

        <vk-notification :messages.sync="messages"></vk-notification>
    </div>
</template>
<script>
import editModuleGeneralSchema from "@/schemas/edit_module_general.json";
import editModuleConfigSchema from "@/schemas/edit_module_config.json";
import editModuleEventsSchema from "@/schemas/edit_module_events.json";
import editModuleChangelogSchema from "@/schemas/edit_module_changelog.json";
import editModuleLocaleSchema from "@/schemas/edit_module_locale.json";
import editModuleDefinitionsEventsSchema from "@/schemas/edit_module_definitions.json";
import systemModuleFiles from "@/components/system-module-files.vue";
import loader from "@/components/loader.vue";
import semverSort from "semver-sort";
import moment from "moment";

export default {
    components: {
        systemModuleFiles,
        loader
    },
    data() {
        const empty_schema = {
            type: "object",
            properties: {},
            ui: {},
            rules: {}
        };
        return {
            module: {},
            files: [],
            edited: false,
            loading: false,
            isBusy: false,
            isLoaded: false,
            inProgressSave: false,
            activeTab: "general",
            messages: [],
            moduleGeneralModel: {},
            moduleConfigSchemaModel: {},
            moduleDefaultConfigModel: {},
            moduleDefaultEventConfigModel: {},
            moduleEventsModel: {},
            moduleChangelogModel: {},
            moduleLocaleModel: {},
            editModuleDefaultConfigSchema: empty_schema,
            editModuleDefaultEventConfigSchema: empty_schema,
            editModuleGeneralSchema,
            editModuleConfigSchema,
            editModuleEventsSchema,
            editModuleChangelogSchema,
            editModuleLocaleSchema,
            editModuleDefinitionsEventsSchema,
            throttledSetModuleDefaultConfigSchema:
                this.throttle(this.setModuleDefaultConfigSchema, 3000),
            throttledSetModuleDefaultEventConfigSchema:
                this.throttle(this.setModuleDefaultEventConfigSchema, 3000),
            throttledSetModuleLocaleSchema:
                this.throttle(this.setModuleLocaleSchema, 3000),
            throttledSetEditedState: this.throttle(this.setEditedState, 3000),
            forms: [
                { id: "general", tab: "general" },
                { id: "config", tab: "config" },
                { id: "default-config", tab: "config" },
                { id: "events", tab: "events" },
                { id: "events-default-config", tab: "events" },
                { id: "locale", tab: "locale" },
                { id: "changelog", tab: "changelog" },
            ]
        }
    },
    mounted() {
        let menu = {
            Module: {
                Info: this.$router.resolve({
                    name: "system_module_view",
                    params: { hash: this.$route.params.hash }
                }).route.path
            }
        };
        if (localStorage.getItem("user_group") === "Admin") {
            menu.Module.Edit = this.$router.resolve({
                name: "system_module_edit",
                params: { hash: this.$route.params.hash }
            }).route.path
        }
        this.$nextTick(function () {
            this.loading = true;
            let promises = [
                this.$http.get("/api/v1/modules/" + this.$route.params.module + '/files'),
                this.$http.get("/api/v1/modules/" + this.$route.params.module)
            ];
            Promise.all(promises)
                .then(results => {
                    results.forEach(r => {
                        if (r.data.status != "success") {
                            throw new Error("response format error");
                        }
                    });
                    this.files = results[0].data.data;
                    this.module = results[1].data.data;
                    this.setModelsData();
                    setTimeout(() => {
                        this.isLoaded = true;
                    }, 2000);
                    this.loading = false;
                })
                .catch(e => {
                    console.log(e);
                    this.loading = false;
                });
        });
        this.$root.$options.sidebarStore.dispatch("search", menu);
    },
    beforeRouteLeave(to, from, next) {
        if (!this.isModuleChanged() && !this.isBusy) {
            next();
            return;
        }
        this.$confirm(this.$t("SystemModule.Confirmation.Text.LeavePage"),
            this.$t("SystemModule.Confirmation.Title.LeavePage"), {
                confirmButtonText: this.$t("Common.Pseudo.Button.Yes"),
                cancelButtonText: this.$t("Common.Pseudo.Button.Cancel"),
                type: 'warning'
            }).then(() => {
            next();
        }).catch(() => {
            next(false);
        });
    },
    methods: {
        setBusyState(state) {
            this.isBusy = state;
        },
        throttle(cb, ms = 50, context = window) {
            let to, final = null, wait = false;
            return (...args) => {
                let later = () => {
                    cb.apply(context, args);
                };
                if (!wait) {
                    later();
                    wait = true;
                    to = setTimeout(() => {
                        if (final !== null) {
                            final();
                        }
                        final = null;
                        wait = false;
                    }, ms);
                } else {
                    final = later;
                }
            };
        },
        sortKeys(x) {
            if (typeof x !== 'object' || !x)
                return x;
            if (Array.isArray(x))
                return x.map(this.sortKeys).sort();
            return Object.keys(x).sort().reduce((o, k) => ({ ...o, [k]: this.sortKeys(x[k]) }), {});
        },
        compareObjs(obj1, obj2) {
            return JSON.stringify(this.sortKeys(obj1)) === JSON.stringify(this.sortKeys(obj2));
        },
        // in: from, to
        // out: add, remove, rename
        getChangesArrays(arr1, arr2) {
            arr1.sort();
            arr2.sort();
            let inter = arr1.filter(x => arr2.includes(x));
            let diff1 = arr1.filter(x => !arr2.includes(x));
            let diff2 = arr2.filter(x => !arr1.includes(x));
            let diff = [].concat(diff1).concat(diff2);
            if (diff.length === 0) {
                return [ [], [], {} ]
            }
            if (diff1.length === diff2.length) {
                return [ [], [], Object.assign(...diff1.map((k, i) => ({ [k]: diff2[i] }))) ]
            }
            return [ diff2, diff1, {} ]
        },
        // in: result, from, to
        // out: result
        makeChangesObj(obj, arr1, arr2, valCb = null) {
            const [ add, remove, rename ] = this.getChangesArrays(arr1, arr2);
            add.forEach(k => obj[k] = valCb ? valCb(k) : {});
            remove.forEach(k => delete obj[k]);
            Object.entries(rename).forEach(([ key, val ]) => {
                const get_val = () => {
                    let resVal = undefined;
                    if (obj[key] !== undefined) return obj[key];
                    if (obj[val] !== undefined) return obj[val];
                    if (valCb && (resVal = valCb(val)) !== undefined) return resVal;
                    return {};
                }
                delete Object.assign(obj, { [val]: get_val() })[key];
            });
            return obj;
        },
        getMessageErrorSave(id, fails) {
            const h = this.$createElement;
            const form = this.$tck('SystemModule.Alert.Text.', id);
            return h('p', null, [
                h('span', null, this.$t('SystemModule.Alert.Text.ErrorSave')),
                h('br', null, null), h('br', null, null),
                h('span', null, this.$t('SystemModule.Alert.Text.ErrorSaveTextForm')),
                h('span', { style: 'font-style: italic' }, form),
                h('br', null, null),
                h('span', null, this.$t('SystemModule.Alert.Text.ErrorSaveTextErrorsList')),
                h('br', null, null),
                h('p', null, Array.from(fails,
                    path => h('span', null,
                        [
                            h('br', null, null),
                            h('i', { style: 'color: gray' }, (p => p.replace(/data./i, ''))(path))
                        ])
                    )
                )
            ]);
        },
        setEditedState() {
            if (this.isLoaded) {
                this.edited = this.isModuleChanged();
            }
        },
        isModuleChanged() {
            const g_model = this.parseGeneralModel(this.moduleGeneralModel);
            const cl_model = this.parseChangelogModel(this.moduleChangelogModel);
            const cs_model = this.parseJSONSchemaModel(this.moduleConfigSchemaModel["config_schema"], false);
            const dc_model = JSON.parse(JSON.stringify(this.moduleDefaultConfigModel["default_config"]));
            const e_model = JSON.parse(JSON.stringify(this.moduleEventsModel));
            const eds_model = this.parseJSONSchemaModel(e_model["event_data_schema"], true);
            const ecs_model = this.parseEventConfigSchemaModel(e_model["event_config_schema"], false);
            const dec_model = JSON.parse(JSON.stringify(this.moduleDefaultEventConfigModel["default_event_config"]));
            const loc_model = JSON.parse(JSON.stringify(this.moduleLocaleModel));

            const changes = [
                this.compareObjs(this.module["info"], g_model),
                this.compareObjs(this.module["config_schema"], cs_model),
                this.compareObjs(this.module["default_config"], dc_model),
                this.compareObjs(this.module["event_data_schema"], eds_model),
                this.compareObjs(this.module["event_config_schema"], ecs_model),
                this.compareObjs(this.module["default_event_config"], dec_model),
                this.compareObjs(this.module["locale"], loc_model),
                this.compareObjs(this.module["changelog"], cl_model)
            ];

            return changes.some(e => !e);
        },
        saveModule() {
            let promises = [
                this.$ncformValidate('edit-module-general'),
                this.$ncformValidate('edit-module-config'),
                this.$ncformValidate('edit-module-default-config'),
                this.$ncformValidate('edit-module-events'),
                this.$ncformValidate('edit-module-events-default-config'),
                this.$ncformValidate('edit-module-locale'),
                this.$ncformValidate('edit-module-changelog')
            ];
            this.inProgressSave = true;
            Promise.all(promises)
                .then(data => {
                    data.forEach((d, i) => {
                        if (!d) {
                            console.log("form doesn't exist");
                        } else if (!d.result) {
                            let fails = [];
                            d.detail.forEach((e, i) => {
                                if (!e.result.result) fails.push(e.__dataPath);
                            });
                            this.activeTab = this.forms[i].tab;
                            this.$alert('', '', {
                                title: this.$t("SystemModule.Alert.Title.ErrorSave"),
                                message: this.getMessageErrorSave(this.forms[i].id, fails),
                                confirmButtonText: 'OK'
                            });
                            throw new Error("validation error");
                        } else {
                            // validation was successed
                        }
                    });

                    const g_model = this.parseGeneralModel(this.moduleGeneralModel);
                    const cl_model = this.parseChangelogModel(this.moduleChangelogModel);
                    const cs_model = this.parseJSONSchemaModel(this.moduleConfigSchemaModel["config_schema"], false);
                    const dc_model = JSON.parse(JSON.stringify(this.moduleDefaultConfigModel["default_config"]));
                    const e_model = JSON.parse(JSON.stringify(this.moduleEventsModel));
                    const eds_model = this.parseJSONSchemaModel(e_model["event_data_schema"], true);
                    const ecs_model = this.parseEventConfigSchemaModel(e_model["event_config_schema"], false);
                    const dec_model = JSON.parse(JSON.stringify(this.moduleDefaultEventConfigModel["default_event_config"]));
                    const loc_model = JSON.parse(JSON.stringify(this.moduleLocaleModel));

                    let module = {
                        info: g_model,
                        changelog: cl_model,
                        config_schema: cs_model,
                        default_config: dc_model,
                        event_data_schema: eds_model,
                        event_config_schema: ecs_model,
                        default_event_config: dec_model,
                        locale: loc_model
                    };
                    this.makeChangesObj(module, Object.keys(module), Object.keys(this.module), k => this.module[k]);

                    this.$http
                        .post("/api/v1/modules/" + this.$route.params.module, module, {
                            headers: {
                                'Content-Type': 'application/json'
                            }
                        })
                        .then(r => {
                            if (r.data.status == "success") {
                                this.module = module;
                                this.messages.push({
                                    message: this.$t("SystemModule.Notifications.Text.SaveSuccess"),
                                    status: "success"
                                });
                            } else {
                                throw new Error("response format error");
                            }
                            this.inProgressSave = false;
                            this.setEditedState();
                        })
                        .catch(e => {
                            console.log(e);
                            this.messages.push({
                                message: this.$t("SystemModule.Notifications.Text.SaveError"),
                                status: "danger"
                            });
                            this.inProgressSave = false;
                        });
                })
                .catch(e => {
                    console.log(e);
                    this.inProgressSave = false;
                });
        },
        makeJSONSchemaModel(schema) {
            let model = [];
            let make = (obj, base) => {
                if (typeof (obj) !== "object" || Array.isArray(obj)) {
                    return;
                }
                if (typeof (obj["properties"]) === "object" && !Array.isArray(obj["properties"])) {
                    const reqs = Array.isArray(obj["required"]) ? obj["required"] : [];
                    for (const key in obj["properties"]) {
                        model.push({
                            "required": reqs.indexOf(key) !== -1,
                            "name": base === "" ? key : base + "." + key,
                            "type": obj["properties"][key]["type"],
                            "fields": JSON.stringify(Object.keys(obj["properties"][key])
                                .filter(k => k !== "type" && k !== "properties" && k !== "required")
                                .reduce((res, k) => (res[k] = obj["properties"][key][k], res), {}), undefined, 2)
                        });
                        if (obj["properties"][key]["type"] === "object") {
                            make(obj["properties"][key], key);
                        }
                    }
                }
            };
            make(schema, "");
            return model;
        },
        parseJSONSchemaModel(model, ap) {
            let schema = {
                "type": "object",
                "properties": {},
                "required": []
            };
            if (ap !== undefined) {
                schema["additionalProperties"] = ap;
            }
            if (model === undefined) {
                return schema;
            }
            model.forEach(data => {
                let obj = schema;
                let parent = schema;
                const path = (data["name"] === undefined ? "" : data["name"]).split(".");
                const field = path.slice(-1)[0];
                path.slice(0, -1).forEach(d => {
                    obj = parent["properties"][d];
                    if (typeof (obj) !== "object" || Array.isArray(obj) || obj["type"] != "object") {
                        obj = {
                            "type": "object",
                            "properties": {},
                            "required": []
                        }
                        parent["properties"][d] = obj;
                    }
                    parent = obj;
                });
                if (data["required"]) {
                    obj["required"].push(field);
                }
                let fobj = obj["properties"][field];
                if (fobj === undefined) {
                    obj["properties"][field] = fobj = {};
                }
                if (fobj["type"] !== "none") {
                    Object.assign(fobj, { "type": data["type"] });
                }
                if (fobj["type"] === "object") {
                    if (typeof (fobj["properties"]) !== "object" || Array.isArray([ "properties" ])) {
                        fobj["properties"] = {};
                    }
                    if (typeof (fobj["required"]) !== "object" || !Array.isArray([ "required" ])) {
                        fobj["required"] = [];
                    }
                }
                try {
                    Object.assign(fobj, JSON.parse(data["fields"]));
                } catch (e) {
                }
            });
            return schema;
        },
        makeEventConfigSchemaModel(schema) {
            let model = [];
            if (typeof (schema) !== "object" || Array.isArray(schema)) {
                return model;
            }
            if (typeof (schema["properties"]) !== "object" || Array.isArray(schema["properties"])) {
                return model;
            }
            Object.keys(schema["properties"]).forEach(eventID => {
                const eventType = schema["properties"][eventID]["allOf"][0];
                const eventProp = schema["properties"][eventID]["allOf"][1];
                model.push({
                    "id": eventID,
                    "type": JSON.stringify(eventType),
                    "fields": JSON.stringify(Object.keys(eventProp)
                        .filter(k => k !== "type" && k !== "properties" && k !== "required")
                        .reduce((res, k) => (res[k] = eventProp[k], res), {}), undefined, 2),
                    "keys": this.makeJSONSchemaModel({
                        ...eventProp, ...{
                            "properties": {
                                ...Object.keys(eventProp["properties"])
                                    .filter(k => k !== "type" && k !== "actions")
                                    .reduce((res, k) => (res[k] = eventProp["properties"][k], res), {})
                            }
                        }
                    })
                });
            });
            return model;
        },
        parseEventConfigSchemaModel(model, ap) {
            let schema = {
                "type": "object",
                "properties": {},
                "required": []
            };
            if (ap !== undefined) {
                schema["additionalProperties"] = ap;
            }
            model.forEach(data => {
                schema["required"].push(data["id"]);
                schema["properties"][data["id"]] = {
                    "allOf": [
                        JSON.parse(data["type"] || "{}"),
                        { ...this.parseJSONSchemaModel(data["keys"]), ...JSON.parse(data["fields"] || "{}") }
                    ]
                }
            });
            return schema;
        },
        makeChangelogModel(cl) {
            return {
                current: this.module["info"]["version"],
                changelog: semverSort.desc(Object.keys(cl)).map(ver => ({
                    ver,
                    date: String(new Date(cl[ver]["en"]["date"]).getTime()),
                    desc: Object.keys(cl[ver]).reduce((res, k) => (res[k] = {
                        title: cl[ver][k]["title"],
                        description: cl[ver][k]["description"],
                    }, res), {})
                }))
            };
        },
        parseChangelogModel(model) {
            const getDateEn = date => moment(parseInt(date)).format("MM-DD-YYYY");
            const getDateRu = date => moment(parseInt(date)).format("DD.MM.YYYY");
            return model["changelog"].reduce((res, k) => (res[k["ver"]] = Object.keys(k["desc"] || {})
                .reduce((desc, lng) => (desc[lng] = {
                    date: lng === "en" ? getDateEn(k["date"]) : getDateRu(k["date"]),
                    title: k["desc"][lng]["title"],
                    description: k["desc"][lng]["description"]
                }, desc), {}),
                res), {});
        },
        makeGeneralModel(info) {
            let model = JSON.parse(JSON.stringify(info));
            model["os"] = Object.keys(model["os"])
                .map(key => (model["os"][key].map(value => `${key}.${value}`))).flat(1);
            return model;
        },
        parseGeneralModel(model) {
            let info = JSON.parse(JSON.stringify(model));
            info["os"] = info["os"].reduce(function (acc, cur) {
                const item = cur.split(".");
                acc[item[0]] = (acc[item[0]] || []);
                acc[item[0]].push(item[1]);
                return acc;
            }, {});
            return info;
        },
        onChangeGeneralModel({ paths, itemValue, formValue, itemOldValue }) {
            if (!this.isLoaded) return;
            (async () => {
                this.throttledSetModuleLocaleSchema();
                this.throttledSetEditedState();
            })();
        },
        onChangeConfigModel({ paths, itemValue, formValue, itemOldValue }) {
            if (!this.isLoaded) return;
            (async () => {
                this.throttledSetModuleDefaultConfigSchema(formValue);
                this.throttledSetModuleLocaleSchema();
                this.throttledSetEditedState();
            })();
        },
        onChangeDefaultConfigModel({ paths, itemValue, formValue, itemOldValue }) {
            if (!this.isLoaded) return;
            (async () => {
                this.throttledSetEditedState();
            })();
        },
        onChangeEventsModel({ paths, itemValue, formValue, itemOldValue }) {
            if (!this.isLoaded) return;
            (async () => {
                let moduleGeneralModel = JSON.parse(JSON.stringify(this.moduleGeneralModel));
                moduleGeneralModel["events"] = formValue["event_config_schema"].map(({ id }) => id);
                this.moduleGeneralModel = moduleGeneralModel;
                this.throttledSetModuleDefaultEventConfigSchema(formValue);
                this.throttledSetModuleLocaleSchema();
                this.throttledSetEditedState();
            })();
        },
        onChangeDefaultEventConfigModel({ paths, itemValue, formValue, itemOldValue }) {
            if (!this.isLoaded) return;
            (async () => {
                this.throttledSetEditedState();
            })();
        },
        onChangeLocaleModel({ paths, itemValue, formValue, itemOldValue }) {
            if (!this.isLoaded) return;
            (async () => {
                this.throttledSetEditedState();
            })();
        },
        onChangeChangelogModel({ paths, itemValue, formValue, itemOldValue }) {
            if (!this.isLoaded) return;
            (async () => {
                this.throttledSetEditedState();
            })();
        },
        setModuleDefaultConfigSchema(model = null) {
            const curSchema = (this.editModuleDefaultConfigSchema["properties"] || {})["default_config"] || {};
            let schema = {};
            if (model !== null) {
                schema = this.parseJSONSchemaModel(model["config_schema"], false);
            } else {
                schema = this.module["config_schema"];
            }
            this.editModuleDefaultConfigSchema = JSON.parse(JSON.stringify({
                "type": "object",
                "properties": {
                    "default_config": {
                        ...schema, ...{
                            "ui": {
                                "label": "Default",
                                "description": "\u200B",
                                "legend": "Default config",
                                "widgetConfig": {
                                    "collapsed": this.compareObjs(schema, this.module["config_schema"])
                                }
                            }
                        }
                    }
                },
                "ui": {
                    "showLabel": false,
                    "showLegend": false,
                    "widgetConfig": {}
                }
            }));

            const get_def_value = k => {
                switch (schema["properties"][k]["type"]) {
                    case "object":
                        return {};
                    case "array":
                        return [];
                    case "boolean":
                        return false;
                    case "integer":
                    case "number":
                        return 0;
                    default:
                        return "";
                }
            };
            let dc_model = JSON.parse(JSON.stringify(this.moduleDefaultConfigModel));
            let def_config = dc_model["default_config"] || {};
            dc_model["default_config"] = this.makeChangesObj(def_config, Object.keys(def_config),
                Object.keys(schema["properties"]), get_def_value);
            def_config = dc_model["default_config"];
            Object.keys(def_config).forEach(k => {
                const curType = ((curSchema["properties"] || {})[k] || {})["type"];
                if (curType && schema["properties"][k]["type"] !== curType)
                    def_config[k] = get_def_value(k);
            });
            if (!this.compareObjs(this.moduleDefaultConfigModel, dc_model)) {
                this.moduleDefaultConfigModel = dc_model;
            }
        },
        setModuleDefaultEventConfigSchema(model = null) {
            const curSchema = (this.editModuleDefaultEventConfigSchema["properties"] || {})["default_event_config"] || {};
            let schema = {}, event_ids = [], event_data_keys = [];
            if (model !== null) {
                event_ids = model["event_config_schema"].map(({ id }) => id);
                event_data_keys = model["event_data_schema"].map(({ name }) => name);
                schema = this.parseEventConfigSchemaModel(model["event_config_schema"], false);
            } else {
                event_ids = Object.keys(this.module["event_config_schema"]["properties"]);
                event_data_keys = Object.keys(this.module["event_data_schema"]["properties"]);
                schema = this.module["event_config_schema"];
            }
            let definitions = JSON.parse(JSON.stringify(this.editModuleDefinitionsEventsSchema));
            definitions["events.ids"]["enum"] = event_ids;
            definitions["events.ids"]["ui"]["widgetConfig"]["enumSource"] = event_ids.map(id => ({ value: id }));
            definitions["events.keys"]["enum"] = event_data_keys;
            definitions["events.keys"]["ui"]["widgetConfig"]["enumSource"] = event_data_keys.map(key => ({ value: key }));
            this.editModuleDefaultEventConfigSchema = JSON.parse(JSON.stringify({
                "type": "object",
                "definitions": definitions,
                "properties": {
                    "default_event_config": {
                        ...schema, ...{
                            "ui": {
                                "label": "Default config",
                                "description": "\u200B",
                                "legend": "Default event config",
                                "widgetConfig": {
                                    "collapsed": this.compareObjs(schema, this.module["event_config_schema"]),
                                    "itemCollapse": true
                                }
                            }
                        }
                    }
                },
                "ui": {
                    "showLabel": false,
                    "showLegend": false,
                    "widgetConfig": {}
                }
            }));

            const get_def_value = t => {
                switch (t) {
                    case "object":
                        return {};
                    case "array":
                        return [];
                    case "boolean":
                        return false;
                    case "integer":
                    case "number":
                        return 0;
                    default:
                        return "";
                }
            };
            const get_event_props = k => {
                const ev_props = schema["properties"][k]["allOf"][1]["properties"];
                return Object.keys(ev_props)
                    .reduce((o, k) => ({
                        ...o,
                        [k]: ev_props[k]["default"] || get_def_value(ev_props[k]["type"])
                    }), {});
            };
            const get_event_type = k => {
                const ev_type = schema["properties"][k]["allOf"][0]["$ref"];
                switch (true) {
                    case /atomic/.test(ev_type):
                        return "atomic";
                    case /correlation/.test(ev_type):
                        return "correlation";
                    case /aggregation/.test(ev_type):
                        return "aggregation";
                    default:
                        return "";
                }
            };
            const get_def_event = k => {
                const ev_complex = {
                    seq: [],
                    group_by: [],
                    max_count: 0,
                    max_time: 0,
                    actions: []
                };
                switch (get_event_type(k)) {
                    case "atomic":
                        return { type: "atomic", actions: [] }
                    case "correlation":
                        return Object.assign({ type: "correlation" }, ev_complex);
                    case "aggregation":
                        return Object.assign({ type: "aggregation" }, ev_complex);
                    default:
                        return {};
                }
            };
            let dec_model = JSON.parse(JSON.stringify(this.moduleDefaultEventConfigModel));
            let def_config = dec_model["default_event_config"] || {};
            dec_model["default_event_config"] = this.makeChangesObj(def_config, Object.keys(def_config),
                Object.keys(schema["properties"]), get_def_event);
            def_config = dec_model["default_event_config"];
            Object.keys(def_config).forEach(k => {
                const ev_empty = Object.assign(get_def_event(k), get_event_props(k));
                def_config[k] = this.makeChangesObj(def_config[k], Object.keys(def_config[k]),
                    Object.keys(ev_empty), ko => ev_empty[ko]);
                def_config[k]["type"] = get_event_type(k);
            });
            if (!this.compareObjs(this.moduleDefaultEventConfigModel, dec_model)) {
                this.moduleDefaultEventConfigModel = dec_model;
            }
        },
        setModuleLocaleSchema() {
            // Make locale schema
            const make_loc_object = keys => keys.sort()
                .reduce((o, k) => ({
                    ...o, [k]: {
                        "$ref": "#/definitions/locale",
                        "ui": {
                            "showLabel": false,
                            "noLabelSpace": true,
                            "legend": k
                        }
                    }
                }), {});
            const fix_ui_legend = obj => Object.keys(obj).forEach(k => (obj[k]["ui"] || {})["legend"] = k);
            let editModuleLocaleSchema = JSON.parse(JSON.stringify(this.editModuleLocaleSchema));
            let properties = editModuleLocaleSchema["properties"];

            const g_model = this.moduleGeneralModel;
            if (g_model["__dataSchema"] === undefined) {
                const p_tags = make_loc_object(g_model["tags"]);
                let s_tags = properties["tags"]["properties"] || {};
                properties["tags"]["properties"] = this.makeChangesObj(s_tags, Object.keys(s_tags),
                    Object.keys(p_tags), k => p_tags[k]);
                fix_ui_legend(properties["tags"]["properties"]);
            }

            const cs_model = this.moduleConfigSchemaModel["config_schema"];
            if (cs_model.length > 0 && cs_model[0]["__dataSchema"] === undefined) {
                const p_config = make_loc_object(Array.from(cs_model, k => k["name"]))
                let s_config = properties["config"]["properties"] || {};
                properties["config"]["properties"] = this.makeChangesObj(s_config, Object.keys(s_config),
                    Object.keys(p_config), k => p_config[k]);
                fix_ui_legend(properties["config"]["properties"]);
            }

            const ecs_model = this.moduleEventsModel["event_config_schema"];
            if (ecs_model.length > 0 && ecs_model[0]["__dataSchema"] === undefined) {
                const p_events = make_loc_object(Array.from(ecs_model, k => k["id"]));
                let s_events = properties["events"]["properties"] || {};
                properties["events"]["properties"] = this.makeChangesObj(s_events, Object.keys(s_events),
                    Object.keys(p_events), k => p_events[k]);
                fix_ui_legend(properties["events"]["properties"]);

                const get_event_options = (id) => ecs_model
                    .reduce((a, ev) => ev["id"] === id ? a.concat(Array.from(ev["keys"], k => k["name"])) : a, []);
                const p_event_config = ecs_model
                    .reduce((a, k) => k["type"].indexOf("atomic") !== -1 ? a.concat([ k["id"] ]) : a, [])
                    .reduce((o, k) => ({
                        ...o, [k]: {
                            "$ref": "#/definitions/group",
                            "properties": make_loc_object(get_event_options(k)),
                            "ui": {
                                "showLabel": false,
                                "noLabelSpace": true,
                                "legend": k
                            }
                        }
                    }), {});
                let s_event_config = properties["event_config"]["properties"] || {};
                properties["event_config"]["properties"] = this.makeChangesObj(s_event_config,
                    Object.keys(s_event_config), Object.keys(p_event_config), k => p_event_config[k]);
                fix_ui_legend(properties["event_config"]["properties"]);
                s_event_config = properties["event_config"]["properties"];
                Object.keys(s_event_config).forEach(k => {
                    const p_event_config_props = p_event_config[k]["properties"];
                    let s_event_config_props = s_event_config[k]["properties"] || {};
                    s_event_config[k]["properties"] = this.makeChangesObj(s_event_config_props,
                        Object.keys(s_event_config_props), Object.keys(p_event_config_props),
                        ko => p_event_config_props[ko]);
                    fix_ui_legend(s_event_config[k]["properties"]);
                });
            }

            const eds_model = this.moduleEventsModel["event_data_schema"];
            if (eds_model.length > 0 && eds_model[0]["__dataSchema"] === undefined) {
                const p_event_data = make_loc_object(Array.from(eds_model, k => k["name"]));
                let s_event_data = properties["event_data"]["properties"] || {};
                properties["event_data"]["properties"] = this.makeChangesObj(s_event_data,
                    Object.keys(s_event_data), Object.keys(p_event_data), k => p_event_data[k]);
                fix_ui_legend(properties["event_data"]["properties"]);
            }

            if (!this.compareObjs(this.editModuleLocaleSchema, editModuleLocaleSchema)) {
                this.editModuleLocaleSchema = editModuleLocaleSchema;
            }

            // Make locale model
            const l_model = JSON.parse(JSON.stringify(this.moduleLocaleModel));
            const empty_loc_struct = () => ({
                "ru": {
                    "title": "",
                    "description": ""
                },
                "en": {
                    "title": "",
                    "description": ""
                }
            });

            let m_tags = l_model["tags"] || {};
            const tags_props = properties["tags"]["properties"];
            l_model["tags"] = this.makeChangesObj(m_tags, Object.keys(m_tags),
                Object.keys(tags_props), empty_loc_struct);

            let m_config = l_model["config"] || {};
            const config_props = properties["config"]["properties"];
            l_model["config"] = this.makeChangesObj(m_config, Object.keys(m_config),
                Object.keys(config_props), empty_loc_struct);

            let m_events = l_model["events"] || {};
            const events_props = properties["events"]["properties"];
            l_model["events"] = this.makeChangesObj(m_events, Object.keys(m_events),
                Object.keys(events_props), empty_loc_struct);

            let m_event_config = l_model["event_config"] || {};
            const event_config_props = properties["event_config"]["properties"];
            l_model["event_config"] = this.makeChangesObj(m_event_config,
                Object.keys(m_event_config), Object.keys(event_config_props));
            m_event_config = l_model["event_config"];
            Object.keys(m_event_config).forEach(k => {
                m_event_config[k] = this.makeChangesObj(m_event_config[k], Object.keys(m_event_config[k]),
                    Object.keys(event_config_props[k]["properties"]), empty_loc_struct);
            });

            let m_event_data = l_model["event_data"] || {};
            const event_data_props = properties["event_data"]["properties"];
            l_model["event_data"] = this.makeChangesObj(m_event_data, Object.keys(m_event_data),
                Object.keys(event_data_props), empty_loc_struct);

            if (!this.compareObjs(this.moduleLocaleModel, l_model)) {
                this.moduleLocaleModel = l_model;
            }
        },
        setModelsData(module) {
            this.moduleGeneralModel = this.makeGeneralModel(this.module["info"]);

            this.moduleConfigSchemaModel = {
                "config_schema": this.makeJSONSchemaModel(this.module["config_schema"])
            };
            this.moduleDefaultConfigModel = {
                "default_config": JSON.parse(JSON.stringify(this.module["default_config"]))
            };
            this.setModuleDefaultConfigSchema();

            this.moduleEventsModel = {
                "event_data_schema": this.makeJSONSchemaModel(this.module["event_data_schema"]),
                "event_config_schema": this.makeEventConfigSchemaModel(this.module["event_config_schema"])
            };
            this.moduleDefaultEventConfigModel = {
                "default_event_config": JSON.parse(JSON.stringify(this.module["default_event_config"]))
            };
            this.setModuleDefaultEventConfigSchema();

            this.moduleLocaleModel = JSON.parse(JSON.stringify(this.module["locale"]));
            this.setModuleLocaleSchema();

            this.moduleChangelogModel = this.makeChangelogModel(this.module["changelog"]);
        }
    }
};
</script>
