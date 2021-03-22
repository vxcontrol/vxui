<template>
    <div>
        <div style="margin-top: 15px;">
            <demo-block>
                <template slot="source">
                    <el-input
                        :placeholder="$t('Agent.Events.InputPlaceholder.SearchEvent')"
                        v-model="searchInput"
                        class="input-with-select"
                        @change="doFilter"
                    >
                        <el-button
                            slot="append" size="medium" icon="el-icon-search"
                            @click="doFilter"
                        ></el-button>
                    </el-input>
                </template>
                <div>
                    <el-select
                        v-model="filters[1].value"
                        multiple
                        :placeholder="$t('Agent.Events.SelectPlaceholder.Module')"
                        class="uk-margin longest"
                        v-if="this.agentModules && !this.moduleName"
                    >
                        <el-option
                            v-for="item in modulesNames"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                        ></el-option>
                    </el-select>
                </div>
            </demo-block>
        </div>

        <data-tables-server
            @query-change="loadData"
            :loading="loading"
            :data="localizedData"
            :pagination-props="{ pageSizes: [10, 30, 50], total: total }"
            :filters="filters"
            :table-props="tableProps"
        >
            <el-table-column
                v-for="title in titles"
                :prop="title.prop"
                :label="title.label"
                :key="title.label"
                sortable="custom"
            ></el-table-column>
        </data-tables-server>
    </div>
</template>
<script>
import filter from "lodash/filter";
import Handlebars from "handlebars";

export default {
    props: ["agentEvents", "agentModules", "moduleName"],
    data() {
        return {
            modulesInfoLoaded: false,
            eventsInfoLoaded: false,
            data: [],
            searchInput: "",
            searchField: "data",
            loading: false,
            total: 0,
            test: [],
            titles: [
                {
                    prop: "localizedModuleName",
                    label: this.$t('Agent.Events.TableColumnTitle.Module'),
                },
                {
                    prop: "localizedEventName",
                    label: this.$t('Agent.Events.TableColumnTitle.Name'),
                    search: true
                },
                {
                    prop: "localizedEventDescription",
                    label: this.$t('Agent.Events.TableColumnTitle.Description'),
                    search: true
                },
                {
                    prop: "localizedDate",
                    label: this.$t('Agent.Events.TableColumnTitle.Date'),
                    search: true
                }
            ],
            filters: [
                {
                    value: "",
                    field: "data" // define search_prop for backend usage.
                },
                {
                    value: [],
                    field: "name" // define search_prop for backend usage.
                }
            ],
            tableProps: {
                defaultSort: {
                    prop: "localizedDate",
                    order: "descending"
                }
            },
            modulesLocales: [],
            modulesNames: []
        };
    },
    created() {
        console.log("EVENTS TABLE CREATED");
    },
    mounted() {
        console.log(`EVENTS TABLE MOUNTED (${this.moduleName})`);
        if (this.moduleName !== undefined) {
            this.filters[1].value = [this.moduleName];
        }
    },
    computed: {
        searchableTitles() {
            return filter(this.titles, ["search", true]);
        },
        localizedData() {
            if (this.data.length > 0) {
                for (var i in this.data) {
                    const module_id = this.data[i].module_id.toString();
                    // module name
                    this.data[i].localizedModuleName = this.modulesLocales[
                        module_id
                        ]["module"][this.$i18n.locale]["title"];

                    // event name
                    this.data[i].localizedEventName = this.modulesLocales[
                        module_id
                        ]["events"][this.data[i].info.name][this.$i18n.locale]["title"];

                    // description
                    var template = Handlebars.compile(
                        this.modulesLocales[module_id]["events"][
                            this.data[i].info.name
                            ][this.$i18n.locale]["description"]
                    );

                    // date
                    this.data[i].localizedDate = new Date(this.data[i].date).toLocaleString();

                    var html = template(this.data[i].info.data);
                    this.data[i].localizedEventDescription = html;
                }
            }
            return this.data;
        }
    },
    methods: {
        doFilter() {
            let filters = [...this.filters];
            filters[0].field = this.searchField;
            filters[0].value = this.searchInput;
            this.filters = filters;
        },
        async loadData(queryInfo) {
            this.loading = true;
            if (Object.keys(this.modulesLocales).length < 1) {
                await this.loadModulesInfo();
            }
            var queryInfoCopy = JSON.parse(JSON.stringify(queryInfo));

            if (this.agentEvents.getAgentHash()) {
                if (queryInfoCopy) {
                    if (queryInfoCopy.filters) {
                        queryInfoCopy.filters.push({
                            field: "hash",
                            value: [this.agentEvents.getAgentHash()]
                        });
                    }
                }
            }
            if (queryInfoCopy.page === null) {
                queryInfoCopy.page = 1;
            }
            if (queryInfoCopy.pageSize === null) {
                queryInfoCopy.pageSize = 10;
            }
            queryInfoCopy.lang = this.$i18n.locale;
            let resp = await this.agentEvents.get(queryInfoCopy);
            if (resp.data.status == "success") {
                let {events, total} = resp.data.data;
                this.data = events;
                this.total = total;
            }
            this.loading = false;
        },

        async loadModulesInfo() {
            console.log("Loading modules info...");
            return new Promise((resolve, reject) => {
                if (this.agentModules) {
                    this.agentModules.get({})
                        .then(r => {
                            if (r.data.status == "success") {
                                for (var i in r.data.data.modules) {
                                    const module = r.data.data.modules[i];
                                    this.modulesLocales[module["id"]] = module["locale"];
                                    this.modulesNames.push({
                                        label: this.modulesLocales[module["id"]]["module"][
                                            this.$i18n.locale
                                            ].title,
                                        value: module["info"]["name"]
                                    });
                                }
                            } else {
                                throw new Error("response format error");
                            }
                            resolve(r);
                        })
                        .catch(e => {
                            console.log(e);
                            reject();
                        });
                } else {
                    console.log("INVALID agentModules", this.agentModules);
                }
            });
        }
    }
};
</script>
<style scoped>
.input-with-select .el-input-group__prepend {
    background-color: #fff;
}

.longest {
    width: 100%;
}
</style>
