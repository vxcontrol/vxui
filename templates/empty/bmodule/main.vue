<template>
  <div>
    <el-tabs tab-position="left" v-model="leftTab">
      <el-tab-pane name="events" :label="$t('BrowserModule.Page.TabTitle.Events')">
        <component
          :is="components['eventsTable']"
          :module-name="module.info.name"
          :agent-events="eventsAPI"
          :agent-modules="modulesAPI"
        ></component>
      </el-tab-pane>
      <el-tab-pane name="config" :label="$t('BrowserModule.Page.TabTitle.Config')">
        <component
          :is="components['agentModuleConfig']"
          :module="module"
        ></component>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
const name = "empty";

module.exports = {
  name,
  props: ["protoAPI", "hash", "module", "eventsAPI", "modulesAPI", "components"],
  data: () => ({
    leftTab: "events"
  })
};
</script>
