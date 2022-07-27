<template>
  <div id="app">
    <file-pond
      name="file"
      ref="pond"
      label-idle="拖放文件或点击"
      labelInvalidField="包含无效字段"
      labelFileWaitingForSize="计算大小中"
      labelFileSizeNotAvailable="文件大小不符合"
      labelFileLoading="加载中"
      labelFileLoadError="文件加载错误"
      labelFileProcessing="上传中"
      labelFileProcessingComplete="上传完成"
      labelFileProcessingAborted="上传失败"
      labelFileProcessingError="上传错误"
      labelFileProcessingRevertError="撤销错误"
      labelFileRemoveError="删除错误"
      labelTapToCancel="点击取消"
      labelTapToRetry="点击重试"
      labelTapToUndo="点击撤销"
      labelButtonRemoveItem="删除"
      labelButtonAbortItemLoad="中止加载"
      labelButtonRetryItemLoad="重试加载"
      labelButtonAbortItemProcessing="取消"
      labelButtonUndoItemProcessing="撤销"
      labelButtonRetryItemProcessing="重试"
      labelButtonProcessItem="上传"
      :allow-remove="true"
      :allow-revert="true"
      :allow-replace="true"
      :allow-multiple="true"
      :chunk-uploads="true"
      :chunk-size="chunkSize"
      :server="serverOptions"
      v-bind:files="uploadFiles"
      :onprocessfile="onProcessFile"
      :onremovefile="onRemoveFile"
    ></file-pond>
  </div>
</template>

<script>
const Domain = "http://127.0.0.1:8888";

function getTempFiles() {
  return fetch(`${Domain}/dummy/tempFiles`)
    .then(async (res) => {
      let data = await res.json();
      return data;
    })
    .catch(console.error);
}

function getSaveFiles() {
  return fetch(`${Domain}/dummy/saveFiles`)
    .then(async (res) => {
      let data = await res.json();
      return data;
    })
    .catch(console.error);
}

export default {
  mounted() {
    // list server temp files (need restore api)
    getTempFiles().then((res) => {
      res.forEach((serverId) => {
        this.uploadFiles.push({
          source: serverId,
          options: {
            type: "limbo",
          },
        });
      });
    });
    // list save server files (need load api)
    getSaveFiles().then((res) => {
      res.forEach((serverId) => {
        this.uploadFiles.push({
          source: serverId,
          options: {
            type: "local",
          },
        });
      });
    });
    // list fake file
    this.uploadFiles.push({
      source: "fakeId",
      options: {
        type: "local",
        file: {
          name: "fake.jpg",
          size: 1024,
          type: "image/png",
        },
      },
    });
  },
  data() {
    return {
      uploadFiles: [],
      chunkSize: 1024 * 1024 * 2,
    };
  },
  computed: {
    serverOptions() {
      return {
        url: `${Domain}/filepond`,
        remove: this.remove,
      };
    },
  },
  methods: {
    remove(source, load, error) {
      fetch(`${Domain}/filepond/${source}`, { method: "DELETE" })
        .then(async (res) => {
          if (res.status == 200) {
            load();
          } else {
            let text = await res.text();
            error(text);
          }
        })
        .catch(error);
    },
    onProcessFile(error, file) {
      if (error) {
        console.error(error);
        return;
      }
      console.debug("upload file", file);
    },
    onRemoveFile(error, file) {
      if (error) {
        console.error(error);
        return;
      }
      console.debug("remove file", file);
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
