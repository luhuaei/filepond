<template>
  <div id="app">
    <file-pond
      name="file"
      ref="pond"
      label-idle="拖放文件或点击"
      labelInvalidField="字段包含无效文件"
      labelFileWaitingForSize="计算大小中"
      labelFileSizeNotAvailable="文件大小不符合"
      labelFileLoading="加载中"
      labelFileLoadError="上传错误"
      labelFileProcessing="上传中"
      labelFileProcessingComplete="上传完成"
      labelFileProcessingAborted="上传失败"
      labelFileProcessingError="上传错误"
      labelFileProcessingRevertError="撤销错误"
      labelFileRemoveError="删除错误"
      labelTapToCancel="点击取消上传"
      labelTapToRetry="点击重试"
      labelTapToUndo="点击撤销"
      labelButtonRemoveItem="删除"
      labelButtonAbortItemLoad="中止"
      labelButtonRetryItemLoad="重试"
      labelButtonAbortItemProcessing="取消"
      labelButtonUndoItemProcessing="撤销"
      labelButtonRetryItemProcessing="重试"
      labelButtonProcessItem="上传"
      credits="false"
      :allow-remove="true"
      :allow-revert="true"
      :allow-replace="true"
      :allow-multiple="false"
      :server="serverOptions"
      v-bind:files="uploadFiles"
      :onprocessfile="onProcessFile"
      :onremovefile="onRemoveFile"
    ></file-pond>
  </div>
</template>

<script>
function fileToLocalMetadata(file) {
  return {
    source: file.serverId,
    options: {
      type: "local",
      file: file.file,
    },
  };
}

export default {
  model: {
    prop: "originFile",
    event: "updateFile",
  },
  props: {
    originFile: Object,
  },
  mounted() {
    if (this.originFile?.source) {
      console.log("file", this.originFile);
      this.uploadFiles.push(this.originFile);
    }
  },
  data() {
    return {
      uploadFiles: [],
    };
  },
  computed: {
    serverOptions() {
      return {
        url: `http://127.0.0.1:8888/filepond`,
        remove: this.remove,
      };
    },
  },
  methods: {
    remove(source, load) {
      // TODO: delete from backend
      console.log("source", source);
      load();
    },
    onProcessFile(error, file) {
      if (error) {
        console.error(error);
        return;
      }
      // only one file
      this.$emit("updateFile", fileToLocalMetadata(file));
    },
    onRemoveFile(error, file) {
      if (error) {
        console.error(error);
        return;
      }
      console.log("remove file", file);
      this.$emit("updateFile", {});
    },
  },
};
</script>

<style scoped>
#app {
  width: 100%;
  display: flex;
  flex-direction: column;
  min-width: 230px;
  max-width: 800px;
  margin: auto;
}
.app > div {
  margin: 50px 0px;
}
</style>
