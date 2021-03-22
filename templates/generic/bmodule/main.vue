<template>
  <div>
    <el-tabs tab-position="left" v-model="leftTab">
      <el-tab-pane name="api" :label="locale[$i18n.locale]['api']">
        <p class="uk-margin buttons">
          <el-button @click="submitData"
          >{{ locale[$i18n.locale]['buttonData'] }}</el-button>
          <el-button @click="submitText"
          >{{ locale[$i18n.locale]['buttonText'] }}</el-button>
        </p>
        <ul>
          <li :key="line" v-for="line in lines">{{line}}</li>
        </ul>
      </el-tab-pane>
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

    <vk-notification status="primary" :messages.sync="messages"></vk-notification>
  </div>
</template>

<script>
const name = "generic";

module.exports = {
  name,
  props: ["protoAPI", "hash", "module", "eventsAPI", "modulesAPI", "components"],
  data: () => ({
    leftTab: "api",
    connection: {},
    lines: [],
    messages: [],
    locale: {
      ru: {
        api: "VX API",
        buttonData: "Отправить data",
        buttonText: "Отправить text",
        connected: "подключен",
        recvError: "Ошибка при выполнении"
      },
      en: {
        api: "VX API",
        buttonData: "Send data",
        buttonText: "Send text",
        connected: "connected",
        recvError: "Error on execute"
      }
    }
  }),
  created() {
    this.protoAPI.connect().then(
      connection => {
        const date = new Date().toLocaleTimeString();
        this.connection = connection;
        this.connection.subscribe(this.recvData, "data");
        this.messages.push({
          message: `${date} ${this.locale[this.$i18n.locale]['connected']}`,
          status: "success"
        });
      },
      error => console.log(error)
    );
  },
  methods: {
    recvData(msg) {
      const date = new Date();
      const date_ms = date.toLocaleTimeString() + `.${date.getMilliseconds()}`;
      this.lines.push(
        `${date_ms} RECV DATA: ${new TextDecoder(
          "utf-8"
        ).decode(msg.content.data)}`
      );
    },
    submitData() {
      const date = new Date();
      const date_ms = date.toLocaleTimeString() + `.${date.getMilliseconds()}`;
      let data = JSON.stringify({ type: "hs_browser", data: "test test test" });
      this.lines.push(
        `${date_ms} SEND DATA: ${data}`
      );
      this.connection.sendData(data);
    },
    submitText() {
      const date = new Date();
      const date_ms = date.toLocaleTimeString() + `.${date.getMilliseconds()}`;
      let text = "simple request";
      this.lines.push(
        `${date_ms} SEND TEXT: ${text}`
      );
      this.connection.sendText(text);
    }
  }
};
</script>
