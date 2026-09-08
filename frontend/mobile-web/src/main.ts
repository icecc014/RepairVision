import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Vant from 'vant'
import VueKonva from 'vue-konva'
import 'vant/lib/index.css'
import './styles/base.css'
import App from './App.vue'
import router from './router'

createApp(App).use(createPinia()).use(Vant).use(VueKonva).use(router).mount('#app')