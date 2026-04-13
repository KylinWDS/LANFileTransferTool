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

// 等待Wails运行时准备就绪
function waitForWails() {
  return new Promise((resolve) => {
    // 检查Wails是否已就绪
    if (window.go && window.go.app && window.go.app.App) {
      resolve()
      return
    }
    
    // 等待Wails就绪事件
    const checkInterval = setInterval(() => {
      if (window.go && window.go.app && window.go.app.App) {
        clearInterval(checkInterval)
        resolve()
      }
    }, 100)
    
    // 超时处理（5秒）
    setTimeout(() => {
      clearInterval(checkInterval)
      console.warn('Wails runtime timeout, proceeding anyway...')
      resolve()
    }, 5000)
  })
}

// 等待Wails就绪后再挂载Vue应用
waitForWails().then(() => {
  const app = createApp(App)
  app.use(i18n)
  app.mount('#app')
}).catch(err => {
  console.error('Failed to initialize app:', err)
  // 即使出错也尝试挂载应用
  const app = createApp(App)
  app.use(i18n)
  app.mount('#app')
})
