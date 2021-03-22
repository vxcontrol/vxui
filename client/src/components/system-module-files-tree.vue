<template>
    <div>
        <div class="uk-child-width-expand@s uk-margin-small-top uk-margin-small-bottom uk-flex-middle uk-grid" uk-grid>
            <div class="uk-width-expand uk-first-column">
                <div class="uk-child-width-expand@s uk-grid uk-grid-stack" uk-grid>
                    <div class="uk-width-expand uk-first-column">
                        <span style="padding-right: 15px;"><strong>{{ title }}</strong></span>
                        <el-input
                            class="uk-search-input uk-flex-wrap-stretch tree-search"
                            id="search"
                            type="search"
                            autocomplete="off"
                            v-model="treeQuery"
                            size="small"
                            :placeholder="$t('SystemModule.FilesTab.InputPlaceholder.File')"
                        >
                            <i slot="prefix" class="el-input__icon el-icon-search"></i>
                        </el-input>
                    </div>
                </div>
            </div>
            <div class="uk-width-auto">
                <div class="uk-grid uk-flex-middle">
                    <el-upload
                        class="uk-margin-small-right"
                        ref="upload"
                        :action="filesURI"
                        :data="uploadData"
                        :auto-upload="true"
                        :multiple="false"
                        :show-file-list="false"
                        :on-error="uploadError"
                        :on-success="uploadSuccess"
                        :before-upload="uploadHook"
                        :http-request="uploadRequest"
                    >
                        <el-button
                            type="primary"
                            size="small"
                            icon="el-icon-upload"
                        >{{ $t("Common.Pseudo.Button.Upload") }}
                        </el-button>
                    </el-upload>
                    <el-button
                        class="uk-margin-small-left"
                        style="padding-left: 15px"
                        type="primary"
                        size="small"
                        icon="el-icon-document"
                        @click="create"
                    >{{ $t("Common.Pseudo.Button.Create") }}
                    </el-button>
                </div>
            </div>
        </div>
        <div class="tree-block">
            <el-tree
                ref="tree"
                node-key="id"
                default-expand-all
                :data="treeData"
                :expand-on-click-node="false"
                :filter-node-method="filterNode"
            >
          <span class="tree-node" slot-scope="{ node, data }">
            <span>{{ node.label }}</span>
            <span v-if="data.can_download || data.can_edit || data.can_move || data.can_remove">
              <el-tooltip
                  v-for="name in context.list"
                  :key="name"
                  :open-delay="800"
                  :content="$tck('SystemModule.FilesTab.ButtonTooltip.', name)"
                  placement="top"
                  effect="light"
              >
                <el-button
                    type="text"
                    size="mini"
                    :loading="data.loading[name]"
                    :disabled="!data['can_'+name]"
                    :icon="context.ctx[name].icon"
                    @click="() => context.ctx[name].func(node, data)">
                </el-button>
              </el-tooltip>
            </span>
          </span>
            </el-tree>
            <br>
        </div>

        <el-dialog
            :visible.sync="editModuleFileLayout"
            :show-close="false"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            center
            fullscreen
        >
            <vk-modal-full-close large @click="closeFile"></vk-modal-full-close>
            <div class="uk-container uk-container-xlarge uk-position-center">
                <div class="uk-width-2xlarge">
                    <h1>{{ $t("SystemModule.FilesTab.Title.ModuleFileTitle") }}</h1>
                    <pre>{{ editModuleFile }}</pre>
                    <div class="uk-margin editor" v-if="editModuleFile">
                        <div v-if="editModuleFileLoading">
                            <loader></loader>
                        </div>
                        <div
                            id="editor-container"
                            class="editor"
                            :class="{ hidden: editModuleFileLoading }"
                        ></div>
                    </div>
                    <div
                        class="uk-margin-medium-top uk-position-relative uk-position-center"
                        :class="{ hiddenr: editModuleFileLoading }"
                    >
                        <el-button
                            type="primary" size="medium"
                            @click="saveFile"
                        >{{ $t("Common.Pseudo.Button.Save") }}
                        </el-button>
                        <el-button
                            size="medium"
                            @click="loadFile"
                        >{{ $t("Common.Pseudo.Button.Reload") }}
                        </el-button>
                    </div>
                </div>
            </div>
        </el-dialog>
    </div>
</template>
<script>
import * as monaco from "monaco-editor";
import base64 from "base64-js"
import loader from "@/components/loader.vue";

export default {
    components: {
        loader
    },
    props: ["module", "messages", "files", "type", "title"],
    data() {
        this.files.sort();
        return {
            prefix: [
                this.module["info"]["name"],
                this.module["info"]["version"],
                this.type
            ].join("/"),
            languages: {
                bat: "bat",
                css: "css",
                css3: "css",
                html: "html",
                ini: "ini",
                js: "javascript",
                json: "json",
                less: "less",
                lua: "lua",
                php: "php",
                pl: "perl",
                ps1: "powershell",
                psd1: "powershell",
                psm1: "powershell",
                py: "python",
                rb: "ruby",
                scss: "scss",
                sh: "shell",
                sql: "sql",
                tcl: "tcl",
                ts: "typescript",
                txt: "plain",
                vb: "vb",
                vba: "vb",
                vbs: "vb",
                vue: "html",
                xml: "xml",
                yaml: "yaml",
                yml: "yaml",
            },
            filesURI: "/api/v1/modules/" + this.module["info"]["name"] + "/files/file",
            treeData: null,
            treeQuery: "",
            uploadData: {},
            uploadCallback: () => "",
            editor: null,
            editModuleFile: "",
            editModuleFileLayout: false,
            editModuleFileLoading: false,
            context: {
                list: ["download", "edit", "move", "remove"],
                ctx: {
                    download: {
                        func: this.download,
                        icon: "el-icon-download"
                    },
                    edit: {
                        func: this.edit,
                        icon: "el-icon-edit"
                    },
                    move: {
                        func: this.move,
                        icon: "el-icon-files"
                    },
                    remove: {
                        func: this.remove,
                        icon: "el-icon-delete"
                    }
                }
            }
        };
    },
    beforeMount() {
        this.treeData = this.treeify();
    },
    beforeDestroy() {
        this.editor && this.editor.getModel().dispose() && this.editor.dispose();
    },
    watch: {
        treeQuery(val) {
            this.$refs.tree.filter(val);
        },
        files(val) {
            this.treeData = this.treeify();
        },
        editModuleFileLayout(val) {
            this.$emit('onBusy', val);
        }
    },
    methods: {
        getPatternToCreate() {
            let pattern = /((([\w.][\w-])|([\w]))[\w\-.]*\/)*((([\w.][\w-])|([\w]))[\w\-.]*){1}$/;
            if (["data", "clibs"].indexOf(this.title) != -1) {
                pattern = new RegExp("^/" + pattern.source);
            } else {
                pattern = new RegExp("^/(?!(data|clibs)/)(?!(data|clibs)$)" + pattern.source);
            }
            return pattern;
        },
        getPatternToUpload() {
            let pattern = /((([\w.][\w-])|([\w]))[\w\-.]*\/)*((([\w.][\w-])|([\w]))[\w\-.]*){0,1}$/;
            if (["data", "clibs"].indexOf(this.title) != -1) {
                pattern = new RegExp("^/" + pattern.source);
            } else {
                pattern = new RegExp("^/(?!(data|clibs)/)(?!(data|clibs)$)" + pattern.source);
            }
            return pattern;
        },
        getFullPrefix() {
            let prefix = this.prefix;
            if (["data", "clibs"].indexOf(this.title) != -1) {
                prefix = [prefix, this.title].join("/");
            }
            return prefix;
        },
        getCurrentRelativeDir() {
            const prefix = this.getFullPrefix();
            const node = this.$refs.tree.getCurrentNode();
            const value = node
                ? (node.id.endsWith(node.label) ? node.id.split(node.label)[0] : node.id)
                : prefix + "/";
            return value.split(prefix)[1];
        },
        getNodeRelativeDir(node) {
            const prefix = this.getFullPrefix();
            const value = node
                ? (node.id.endsWith(node.label) ? node.id.split(node.label)[0] : node.id)
                : prefix + "/";
            return value.split(prefix)[1];
        },
        getMessageToCreate() {
            const h = this.$createElement;
            const prefix = this.getFullPrefix();
            return h('p', null, [
                h('span', null, this.$t("SystemModule.FilesTab.Text.CreateMsgText")),
                h('br', null, null), h('span', null, '('),
                h('span', {style: 'font-style: italic'},
                    this.$t("SystemModule.FilesTab.Text.CreateMsgBase")),
                h('span', null, ': '), h('i', {style: 'color: gray'}, prefix),
                h('span', null, ')')
            ]);
        },
        getMessageToUpload(filename) {
            const h = this.$createElement;
            const prefix = this.getFullPrefix();
            return h('p', null, [
                h('span', null, this.$t("SystemModule.FilesTab.Text.UploadMsgText")),
                h('br', null, null), h('span', null, '('),
                h('span', {style: 'font-style: italic'},
                    this.$t("SystemModule.FilesTab.Text.UploadMsgBase")),
                h('span', null, ': '), h('i', {style: 'color: gray'}, prefix),
                h('span', null, ')'),
                h('br', null, null), h('span', null, '('),
                h('span', {style: 'font-style: italic'},
                    this.$t("SystemModule.FilesTab.Text.UploadMsgName")),
                h('span', null, ': '), h('i', {style: 'color: gray'}, filename),
                h('span', null, ')')
            ]);
        },
        getMessageToMove(oldpath) {
            const h = this.$createElement;
            const prefix = this.getFullPrefix();
            return h('p', null, [
                h('span', null, this.$t("SystemModule.FilesTab.Text.MoveMsgText")),
                h('br', null, null), h('span', null, '('),
                h('span', {style: 'font-style: italic'},
                    this.$t("SystemModule.FilesTab.Text.MoveMsgBase")),
                h('span', null, ': '), h('i', {style: 'color: gray'}, prefix),
                h('span', null, ')'),
                h('br', null, null), h('span', null, '('),
                h('span', {style: 'font-style: italic'},
                    this.$t("SystemModule.FilesTab.Text.MoveMsgPath")),
                h('span', null, ': '), h('i', {style: 'color: gray'}, oldpath),
                h('span', null, ')')
            ]);
        },
        getMessageToRemove(curpath) {
            const h = this.$createElement;
            const prefix = this.getFullPrefix();
            return h('p', null, [
                h('span', null, this.$t("SystemModule.FilesTab.Text.RemoveMsgText")),
                h('br', null, null), h('span', null, '('),
                h('span', {style: 'font-style: italic'},
                    this.$t("SystemModule.FilesTab.Text.RemoveMsgPath")),
                h('span', null, ': '), h('i', {style: 'color: gray'}, curpath),
                h('span', null, ')')
            ]);
        },
        closeFile() {
            this.editModuleFile = "";
            this.editModuleFileLayout = false;
        },
        loadFile() {
            const path = this.editModuleFile;
            const arrpath = path.split(".");
            const language = this.languages[arrpath[arrpath.length - 1]] || "";
            this.editModuleFileLoading = true;
            this.$http
                .get(this.filesURI, {
                    params: {
                        path
                    }
                })
                .then(r => {
                    if (this.editor != null) {
                        this.editor.setValue("");
                        this.editor.getModel().dispose();
                        this.editor.dispose();
                    }

                    const value = new TextDecoder("utf-8").decode(base64.toByteArray(r.data.data.data));
                    if (r.data.status == "success") {
                        const cntr = document.getElementById("editor-container");
                        this.editor = monaco.editor.create(cntr, {value, language});
                        const KM = monaco.KeyMod;
                        const KC = monaco.KeyCode;
                        this.editor.addCommand(KM.CtrlCmd | KC.KEY_S, this.saveFile);
                        this.editor.addCommand(KM.CtrlCmd | KC.KEY_O, this.loadFile);
                    } else {
                        throw new Error("response format error");
                    }
                    this.editModuleFileLoading = false;
                })
                .catch(e => {
                    console.log(e);
                    this.messages.push({
                        message: this.$t("SystemModule.Notifications.Text.LoadError"),
                        status: "danger"
                    });
                    this.editModuleFileLoading = false;
                });
        },
        saveFile() {
            const path = this.editModuleFile;
            const value = new TextEncoder("utf-8").encode(this.editor.getModel().getValue());
            const data = base64.fromByteArray(value);
            this.editModuleFileLoading = true;
            this.$http
                .post(this.filesURI, {
                    action: "save",
                    path,
                    data
                }, {
                    headers: {
                        'Content-Type': 'application/json'
                    }
                })
                .then(r => {
                    if (r.data.status == "success") {
                    } else {
                        throw new Error("response format error");
                    }
                    this.messages.push({
                        message: this.$t("SystemModule.Notifications.Text.FilesSaveSuccess"),
                        status: "success"
                    });
                    this.editModuleFileLoading = false;
                })
                .catch(e => {
                    console.log(e);
                    this.messages.push({
                        message: this.$t("SystemModule.Notifications.Text.FilesSaveError"),
                        status: "danger"
                    });
                    this.editModuleFileLoading = false;
                });
        },
        create() {
            this.$prompt('', '', {
                title: this.$t("SystemModule.FilesTab.Title.CreateMsg"),
                message: this.getMessageToCreate(),
                closeOnClickModal: false,
                showCancelButton: true,
                confirmButtonText: this.$t("Common.Pseudo.Button.Create"),
                cancelButtonText: this.$t("Common.Pseudo.Button.Abort"),
                inputValue: this.getCurrentRelativeDir() + "new_file.ext",
                inputPattern: this.getPatternToCreate(),
                inputErrorMessage: this.$t("SystemModule.FilesTab.Text.CreateMsgError"),
                inputPlaceholder: '/new_file.ext',
                beforeClose: (action, instance, done) => {
                    if (action === 'confirm') {
                        instance.confirmButtonLoading = true;
                        instance.confirmButtonText = this.$t("SystemModule.FilesTab.Button.Creating");
                        const path = this.getFullPrefix() + instance.inputValue;
                        const closePrompt = () => {
                            instance.confirmButtonText = this.$t("Common.Pseudo.Button.Create");
                            instance.confirmButtonLoading = false;
                            done();
                        };
                        const index = this.files.indexOf(path);
                        if (index !== -1) {
                            this.messages.push({
                                message: this.$t("SystemModule.Notifications.Text.FileExists"),
                                status: "danger"
                            });
                            closePrompt();
                        } else {
                            this.$http
                                .post(this.filesURI, {
                                    action: "save",
                                    data: "",
                                    path
                                }, {
                                    headers: {
                                        'Content-Type': 'application/json'
                                    }
                                })
                                .then(r => {
                                    if (r.data.status == "success") {
                                        this.messages.push({
                                            message: this.$t("SystemModule.Notifications.Text.FileCreateSuccess"),
                                            status: "success"
                                        });
                                        this.files.push(path);
                                        this.files.sort();
                                        closePrompt();
                                    } else {
                                        throw new Error("response format error");
                                    }
                                })
                                .catch(e => {
                                    console.log(e);
                                    this.messages.push({
                                        message: this.$t("SystemModule.Notifications.Text.FileCreateError"),
                                        status: "danger"
                                    });
                                    closePrompt();
                                });
                        }
                    } else {
                        done();
                    }
                }
            }).catch(() => {
                this.messages.push({
                    message: this.$t("SystemModule.Notifications.Text.FileCreateCanceled"),
                    status: "warning"
                });
            });
        },
        download(node, data) {
            data.loading["download"] = true;
            const path = data.id;
            this.$http
                .get(this.filesURI, {
                    params: {
                        path
                    }
                })
                .then(r => {
                    if (r.data.status == "success") {
                        const blob = new Blob([r.data.data.data], {type: 'application/octet-stream'});
                        const link = document.createElement('a');
                        link.href = URL.createObjectURL(blob);
                        link.download = data.label;
                        link.click();
                        URL.revokeObjectURL(link.href);
                    } else {
                        throw new Error("response format error");
                    }
                    this.messages.push({
                        message: this.$t("SystemModule.Notifications.Text.DownloadSuccess"),
                        status: "success"
                    });
                    data.loading["download"] = false;
                })
                .catch(e => {
                    console.log(e);
                    this.messages.push({
                        message: this.$t("SystemModule.Notifications.Text.DownloadError"),
                        status: "danger"
                    });
                    data.loading["download"] = false;
                });
        },
        edit(node, data) {
            this.editModuleFile = data.id;
            this.loadFile();
            this.editModuleFileLayout = true;
        },
        move(node, data) {
            const path = data.id;
            this.$prompt('', '', {
                title: this.$t("SystemModule.FilesTab.Title.MoveMsgTitle"),
                message: this.getMessageToMove(data.id.split(this.getFullPrefix())[1]),
                closeOnClickModal: false,
                showCancelButton: true,
                confirmButtonText: this.$t("Common.Pseudo.Button.Move"),
                cancelButtonText: this.$t("Common.Pseudo.Button.Abort"),
                inputValue: this.getNodeRelativeDir(data),
                inputPattern: this.getPatternToUpload(),
                inputErrorMessage: this.$t("SystemModule.FilesTab.Text.MoveMsgError"),
                inputPlaceholder: '/',
                beforeClose: (action, instance, done) => {
                    if (action === 'confirm') {
                        instance.confirmButtonLoading = true;
                        instance.confirmButtonText = this.$t("SystemModule.FilesTab.Button.Moving");
                        let newpath = this.getFullPrefix() + instance.inputValue;
                        if (path[path.length - 1] === "/" && newpath[newpath.length - 1] !== "/") {
                            newpath += "/";
                        }
                        if (path[path.length - 1] !== "/" && newpath[newpath.length - 1] === "/") {
                            newpath += data.label;
                        }
                        const closePrompt = () => {
                            instance.confirmButtonText = this.$t("Common.Pseudo.Button.Move");
                            instance.confirmButtonLoading = false;
                            done();
                        };
                        const index = this.files.indexOf(newpath);
                        if (index !== -1 || path === newpath) {
                            this.messages.push({
                                message: this.$t("SystemModule.Notifications.Text.FileExists"),
                                status: "danger"
                            });
                            closePrompt();
                        } else {
                            data.loading["move"] = true;
                            this.$http
                                .post(this.filesURI, {
                                    action: "move",
                                    path,
                                    newpath
                                }, {
                                    headers: {
                                        'Content-Type': 'application/json'
                                    }
                                })
                                .then(r => {
                                    if (r.data.status == "success") {
                                        this.messages.push({
                                            message: this.$t("SystemModule.Notifications.Text.FileMoveSuccess"),
                                            status: "success"
                                        });
                                        let repls = {};
                                        for (const file of this.files) {
                                            if (file.startsWith(path)) {
                                                repls[file] = file.replace(path, newpath);
                                            }
                                        }
                                        for (const id in repls) {
                                            const index = this.files.indexOf(id);
                                            if (index !== -1) {
                                                this.files.splice(index, 1);
                                            }
                                            this.files.push(repls[id]);
                                        }
                                        this.files.sort();
                                    } else {
                                        throw new Error("response format error");
                                    }
                                    closePrompt();
                                    data.loading["move"] = false;
                                })
                                .catch(e => {
                                    console.log(e);
                                    this.messages.push({
                                        message: this.$t("SystemModule.Notifications.Text.FileMoveError"),
                                        status: "danger"
                                    });
                                    closePrompt();
                                    data.loading["move"] = false;
                                });
                        }
                    } else {
                        done();
                    }
                }
            }).catch(() => {
                this.messages.push({
                    message: this.$t("SystemModule.Notifications.Text.FileMoveCanceled"),
                    status: "warning"
                });
            });
        },
        remove(node, data) {
            const path = data.id;
            this.$confirm('', '', {
                title: this.$t("SystemModule.FilesTab.Title.RemoveMsgTitle"),
                message: this.getMessageToRemove(path.split(this.getFullPrefix())[1]),
                closeOnClickModal: false,
                showCancelButton: true,
                confirmButtonText: this.$t("Common.Pseudo.Button.Remove"),
                cancelButtonText: this.$t("Common.Pseudo.Button.Abort"),
                beforeClose: (action, instance, done) => {
                    if (action === 'confirm') {
                        instance.confirmButtonLoading = true;
                        instance.confirmButtonText = this.$t("SystemModule.FilesTab.Button.Removing");
                        const closeConfirm = () => {
                            instance.confirmButtonText = this.$t("Common.Pseudo.Button.Remove");
                            instance.confirmButtonLoading = false;
                            done();
                        };
                        data.loading["remove"] = true;
                        this.$http
                            .post(this.filesURI, {
                                action: "remove",
                                path
                            }, {
                                headers: {
                                    'Content-Type': 'application/json'
                                }
                            })
                            .then(r => {
                                if (r.data.status == "success") {
                                    this.messages.push({
                                        message: this.$t("SystemModule.Notifications.Text.FileRemoveSuccess"),
                                        status: "success"
                                    });
                                    let ids = [];
                                    for (const file of this.files) {
                                        if (file.startsWith(data.id)) {
                                            ids.push(file);
                                        }
                                    }
                                    for (const id of ids) {
                                        const index = this.files.indexOf(id);
                                        if (index !== -1) {
                                            this.files.splice(index, 1);
                                        }
                                    }
                                } else {
                                    throw new Error("response format error");
                                }
                                closeConfirm();
                                data.loading["remove"] = false;
                            })
                            .catch(e => {
                                console.log(e);
                                this.messages.push({
                                    message: this.$t("SystemModule.Notifications.Text.FileRemoveError"),
                                    status: "danger"
                                });
                                closeConfirm();
                                data.loading["remove"] = false;
                            });
                    } else {
                        done();
                    }
                }
            }).catch(() => {
                this.messages.push({
                    message: this.$t("SystemModule.Notifications.Text.FileRemoveCanceled"),
                    status: "warning"
                });
            });
        },
        filterNode(value, data) {
            if (!value) {
                return true;
            }
            return data.id.indexOf(value) !== -1;
        },
        uploadRequest(param) {
            var reader = new FileReader();
            reader.onload = evt => {
                if (evt.target.readyState != 2) {
                    return;
                }

                const params = Object.assign(param.data, {
                    data: evt.target.result.split("base64,")[1]
                });
                this.$http
                    .post(param.action, params, {
                        headers: {
                            'Content-Type': 'application/json'
                        }
                    })
                    .then(r => {
                        if (r.data.status == "success") {
                            param.onSuccess();
                        } else {
                            throw new Error("response format error");
                        }
                    })
                    .catch(e => {
                        console.log(e);
                        param.onError();
                    });
            };
            reader.onerror = () => {
                console.log('error while reading file');
                param.onError();
                reader.abort();
            };
            reader.readAsDataURL(param.file);
        },
        uploadHook(file) {
            const reset = () => {
                this.$refs.upload.abort(file);
                this.$refs.upload.clearFiles();
                this.uploadCallback();
                this.uploadCallback = () => "";
            };
            return new Promise((resolve, reject) => {
                const uploadCancelToken = this.$http.CancelToken.source();
                this.$prompt('', '', {
                    title: this.$t("SystemModule.FilesTab.Title.UploadMsgTitle"),
                    message: this.getMessageToUpload(file.name),
                    closeOnClickModal: false,
                    showCancelButton: true,
                    confirmButtonText: this.$t("Common.Pseudo.Button.Upload"),
                    cancelButtonText: this.$t("Common.Pseudo.Button.Abort"),
                    inputValue: this.getCurrentRelativeDir(),
                    inputPattern: this.getPatternToUpload(),
                    inputErrorMessage: this.$t("SystemModule.FilesTab.Text.UploadMsgError"),
                    inputPlaceholder: '/',
                    beforeClose: (action, instance, done) => {
                        if (action === 'confirm') {
                            let path = this.getFullPrefix() + instance.inputValue;
                            if (path[path.length - 1] === "/") {
                                path += file.name;
                            }
                            this.uploadData = {
                                cancelToken: uploadCancelToken,
                                action: "save",
                                path
                            }
                            instance.confirmButtonLoading = true;
                            instance.confirmButtonText = this.$t("SystemModule.FilesTab.Button.Uploading");
                            this.uploadCallback = () => {
                                instance.confirmButtonText = this.$t("Common.Pseudo.Button.Upload");
                                instance.confirmButtonLoading = false;
                                done();
                                return path;
                            };
                            const index = this.files.indexOf(path);
                            if (index !== -1) {
                                this.messages.push({
                                    message: this.$t("SystemModule.Notifications.Text.FileExists"),
                                    status: "danger"
                                });
                                reset();
                                reject();
                            } else {
                                resolve();
                            }
                        } else {
                            reject();
                            done();
                        }
                    }
                }).catch(() => {
                    this.messages.push({
                        message: this.$t("SystemModule.Notifications.Text.FileUploadCanceled"),
                        status: "warning"
                    });
                    reset();
                    reject();
                    uploadCancelToken.cancel();
                });
            });
        },
        uploadError() {
            this.messages.push({
                message: this.$t("SystemModule.Notifications.Text.FileUploadError"),
                status: "danger"
            });
            this.$refs.upload.clearFiles();
            this.uploadCallback();
            this.uploadCallback = () => "";
        },
        uploadSuccess() {
            this.messages.push({
                message: this.$t("SystemModule.Notifications.Text.FileUploadSuccess"),
                status: "success"
            });
            this.$refs.upload.clearFiles();
            const path = this.uploadCallback();
            this.uploadCallback = () => "";
            this.files.push(path);
            this.files.sort();
        },
        treeify() {
            let result = [];
            let level = {result};
            let exclude = [];
            let st_level = 3;
            let clibs_paths = [];
            let prefix = this.prefix;
            ["windows", "darwin", "linux"].forEach(
                os => {
                    clibs_paths.push([prefix, "clibs", os, ""].join("/"));
                    ["386", "amd64"].forEach(
                        arch => {
                            clibs_paths.push([prefix, "clibs", os, arch, ""].join("/"));
                            clibs_paths.push([prefix, "clibs", os, arch, "sys", ""].join("/"));
                        }
                    )
                }
            );
            let files = clibs_paths.concat(this.files);
            let static_paths = clibs_paths;

            if (["data", "clibs"].indexOf(this.title) != -1) {
                st_level = 4;
                prefix = [prefix, this.title].join("/");
            } else {
                exclude = [
                    [prefix, "data"].join("/"),
                    [prefix, "clibs"].join("/")
                ]
                if (this.type !== "bmodule") {
                    static_paths = static_paths.concat([
                        [prefix, "args.json"].join("/"),
                        [prefix, "main.lua"].join("/")
                    ]);
                } else {
                    static_paths = static_paths.concat([
                        [prefix, "main.vue"].join("/")
                    ]);
                }
            }

            files.forEach(path => {
                let is_exclude = false;
                exclude.forEach(e => {
                    is_exclude = path.startsWith(e) ? true : is_exclude;
                });
                if (path.startsWith(prefix) && !is_exclude) {
                    path.split('/').slice(st_level).reduce((r, label, i, a) => {
                        let short = path.split('/').slice(0, st_level + i + 1).join("/");
                        short = short != path ? short + "/" : short;
                        if (label != "" && !r[label]) {
                            const arrpath = label.split(".");
                            r[label] = {result: []};
                            r.result.push({
                                label,
                                id: short,
                                children: r[label].result,
                                loading: {
                                    "download": false,
                                    "edit": false,
                                    "move": false,
                                    "remove": false,
                                },
                                can_download: short == path && short[short.length - 1] !== "/",
                                can_edit: arrpath.length > 1 &&
                                this.languages[arrpath[arrpath.length - 1]] ? true : false,
                                can_move: !static_paths.some((p) => p === short),
                                can_remove: !static_paths.some((p) => p === short),
                            })
                        }
                        return r[label];
                    }, level)
                }
            });

            return result;
        }
    }
};
</script>
<style scoped>
.editor {
    height: 450px;
    width: 100%;
}

.hidden {
    position: absolute;
    left: -5000px;
}

.hiddenr {
    position: relative;
    left: -5000px;
}

.tree-block {
    overflow: auto;
    min-height: 160px;
    max-height: 250px;
}

.tree-search {
    width: 160px;
}

.tree-node {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 14px;
    padding-right: 8px;
}
</style>
