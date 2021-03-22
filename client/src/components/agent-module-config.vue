<template>
    <div class="uk-margin-small-top uk-margin-small-left">
        <el-collapse accordion>
            <el-collapse-item title="Module" name="module">
                <ncform
                    name="from-edit-module-current-config"
                    :form-schema="editModuleCurrentConfigSchema"
                    form-name="edit-module-current-config"
                    v-model="moduleCurrentConfigModel"
                    @change="throttledSetEditedState"
                ></ncform>
            </el-collapse-item>
            <el-collapse-item title="Events" name="events">
                <ncform
                    name="from-edit-module-events-current-config"
                    :form-schema="editModuleCurrentEventConfigSchema"
                    form-name="edit-module-events-current-config"
                    v-model="moduleCurrentEventConfigModel"
                    @change="throttledSetEditedState"
                ></ncform>
            </el-collapse-item>
        </el-collapse>
        <el-button
            v-if="edited"
            type="primary" size="medium"
            style="margin: 13px auto auto 13px"
            :loading="inProgressSave"
            @click="saveModule"
        >{{ $t("Common.Pseudo.Button.Save") }}
        </el-button>

        <vk-notification :messages.sync="messages"></vk-notification>
    </div>
</template>
<script>
import definitionsEventsSchema from "@/schemas/edit_module_definitions.json";

export default {
    props: ["module"],
    data() {
        return {
            messages: [],
            edited: false,
            inProgressSave: false,
            moduleCurrentData: JSON.parse(JSON.stringify(this.module)),
            moduleCurrentConfigModel: JSON.parse(JSON.stringify(this.module["current_config"])),
            moduleCurrentEventConfigModel: JSON.parse(JSON.stringify(this.module["current_event_config"])),
            editModuleCurrentConfigSchema: this.module["config_schema"],
            editModuleCurrentEventConfigSchema: this.getEventConfigSchema(),
            throttledSetEditedState: this.throttle(this.setEditedState, 1000)
        }
    },
    methods: {
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
            return Object.keys(x).sort().reduce((o, k) => ({...o, [k]: this.sortKeys(x[k])}), {});
        },
        compareObjs(obj1, obj2) {
            return JSON.stringify(this.sortKeys(obj1)) === JSON.stringify(this.sortKeys(obj2));
        },
        setEditedState() {
            const cc_model = JSON.parse(JSON.stringify(this.moduleCurrentConfigModel));
            const cec_model = JSON.parse(JSON.stringify(this.moduleCurrentEventConfigModel));

            const changes = [
                this.compareObjs(this.moduleCurrentData["current_config"], cc_model),
                this.compareObjs(this.moduleCurrentData["current_event_config"], cec_model)
            ];

            this.edited = changes.some(e => !e);
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
                return [[], [], {}]
            }
            if (diff1.length === diff2.length) {
                return [[], [], Object.assign(...diff1.map((k, i) => ({[k]: diff2[i]})))]
            }
            return [diff2, diff1, {}]
        },
        // in: result, from, to
        // out: result
        makeChangesObj(obj, arr1, arr2, valCb = null) {
            const [add, remove, rename] = this.getChangesArrays(arr1, arr2);
            add.forEach(k => obj[k] = valCb ? valCb(k) : {});
            remove.forEach(k => delete obj[k]);
            Object.entries(rename).forEach(([key, val]) => {
                const get_val = () => {
                    let resVal = undefined;
                    if (obj[key] !== undefined) return obj[key];
                    if (obj[val] !== undefined) return obj[val];
                    if (valCb && (resVal = valCb(val)) !== undefined) return resVal;
                    return {};
                }
                delete Object.assign(obj, {[val]: get_val()})[key];
            });
            return obj;
        },
        getMessageErrorSave(fails) {
            const h = this.$createElement;
            return h('p', null, [
                h('span', null, this.$t("SystemModule.Alert.Text.ErrorSave")),
                h('br', null, null), h('br', null, null),
                h('span', null, this.$t("SystemModule.Alert.Text.ErrorSaveTextErrorsList")),
                h('br', null, null),
                h('p', null, Array.from(fails,
                    path => h('span', null,
                        [
                            h('br', null, null),
                            h('i', {style: 'color: gray'}, (p => p.replace(/data./i, ''))(path))
                        ])
                    )
                )
            ]);
        },
        saveModule() {
            let promises = [
                this.$ncformValidate('edit-module-current-config'),
                this.$ncformValidate('edit-module-events-current-config')
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
                            this.$alert('', '', {
                                title: this.$t("SystemModule.Alert.Title.ErrorSave"),
                                message: this.getMessageErrorSave(fails),
                                confirmButtonText: 'OK'
                            });
                            throw new Error("validation error");
                        } else {
                            // validation was successed
                        }
                    });

                    const cc_model = JSON.parse(JSON.stringify(this.moduleCurrentConfigModel));
                    const cec_model = JSON.parse(JSON.stringify(this.moduleCurrentEventConfigModel));

                    let module = {
                        current_config: cc_model,
                        current_event_config: cec_model
                    };
                    this.makeChangesObj(module, Object.keys(module), Object.keys(this.moduleCurrentData), k => this.moduleCurrentData[k]);

                    this.$http
                        .post("/api/v1/agents/" + this.$route.params.hash + "/modules/" + this.$route.params.module, {
                            action: "store",
                            module
                        }, {
                            headers: {
                                'Content-Type': 'application/json'
                            }
                        })
                        .then(r => {
                            if (r.data.status == "success") {
                                this.messages.push({
                                    message: this.$t("SystemModule.Notifications.Text.SaveSuccess"),
                                    status: "success"
                                });
                            } else {
                                throw new Error("response format error");
                            }
                            this.inProgressSave = false;
                            this.moduleCurrentData = module;
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
        getEventConfigSchema() {
            let schema = {}, event_ids = [], event_data_keys = [];
            event_ids = Object.keys(this.module["event_config_schema"]["properties"]);
            event_data_keys = Object.keys(this.module["event_data_schema"]["properties"]);
            schema = JSON.parse(JSON.stringify(this.module["event_config_schema"]));
            let definitions = JSON.parse(JSON.stringify(definitionsEventsSchema));
            definitions["events.ids"]["enum"] = event_ids;
            definitions["events.ids"]["ui"]["widgetConfig"]["enumSource"] = event_ids.map(id => ({value: id}));
            definitions["events.keys"]["enum"] = event_data_keys;
            definitions["events.keys"]["ui"]["widgetConfig"]["enumSource"] = event_data_keys.map(key => ({value: key}));
            schema["definitions"] = definitions;
            schema["ui"] = Object.assign(schema["ui"] || {}, {"widgetConfig": {"itemCollapse": true}});
            return schema;
        }
    }
};
</script>
