import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import './styles/index.css'
import { useConsoleStore } from './stores/console'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)

const store = useConsoleStore(pinia)
store.startSessionMonitor(router)

app.mount('#app')
