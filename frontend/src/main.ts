import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import './styles/index.css'
import { useConsoleStore } from './stores/console'
import { applyMotionPreference } from './utils/uiPreferences'

const app = createApp(App)
const pinia = createPinia()

applyMotionPreference()

app.use(pinia)
app.use(router)
app.use(i18n)

const store = useConsoleStore(pinia)
store.startSessionMonitor(router)

app.mount('#app')
