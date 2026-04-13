import { createApp } from 'vue'
import { createI18n } from 'vue-i18n'
import App from './App.vue'
import './style.css'

import zhCN from './i18n/zh-CN.json'
import en from './i18n/en.json'
import ru from './i18n/ru.json'

const i18n = createI18n({
  legacy: false,
  locale: 'zh-CN',
  fallbackLocale: 'en',
  messages: {
    'zh-CN': zhCN,
    'en': en,
    'ru': ru
  }
})

const app = createApp(App)
app.use(i18n)
app.mount('#app')
