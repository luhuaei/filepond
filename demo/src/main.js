import Vue from "vue";
import App from "./App.vue";

Vue.config.productionTip = false;

// begin filepond
import vueFilePond from "vue-filepond";
import "filepond/dist/filepond.min.css";
const FilePond = vueFilePond();
Vue.component("file-pond", FilePond);
// end filepond

new Vue({
  render: (h) => h(App),
}).$mount("#app");
