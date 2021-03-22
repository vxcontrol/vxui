<template>
    <div class="uk-margin-small-left uk-margin-small-top">
        <el-tabs v-model="activeTab">
            <el-tab-pane
                v-for="name in tabs"
                :key="name"
                :name="name"
                :label="$tck('SystemModule.FilesTab.TabTitle.', name)"
            >
                <system-module-files-tree
                    :module="module"
                    :messages="messages"
                    :type="name"
                    :files="files"
                    @onBusy="setBusyState"
                    title="code"
                ></system-module-files-tree>
                <system-module-files-tree
                    v-if="name != 'bmodule'"
                    :module="module"
                    :messages="messages"
                    :type="name"
                    :files="files"
                    @onBusy="setBusyState"
                    title="data"
                ></system-module-files-tree>
                <system-module-files-tree
                    v-if="name != 'bmodule'"
                    :module="module"
                    :messages="messages"
                    :type="name"
                    :files="files"
                    @onBusy="setBusyState"
                    title="clibs"
                ></system-module-files-tree>
            </el-tab-pane>
        </el-tabs>
    </div>
</template>
<script>
import systemModuleFilesTree from "@/components/system-module-files-tree.vue";

export default {
    props: ["module", "messages", "files"],
    components: {
        systemModuleFilesTree
    },
    data() {
        return {
            activeTab: "cmodule",
            tabs: ["cmodule", "smodule", "bmodule"],
        };
    },
    methods: {
        setBusyState(state) {
            this.$emit('onBusy', state);
        },
    }
};
</script>
