import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import AppHeader from './components/AppHeader.vue'
import AppFooter from './components/AppFooter.vue'

const app = createApp(App)
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.component('QuillEditor', QuillEditor)
app.component('AppHeader', AppHeader)
app.component('AppFooter', AppFooter)
app.mount('#app')
