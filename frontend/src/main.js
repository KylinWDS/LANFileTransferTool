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

const originalConsole = {
  log: console.log,
  info: console.info,
  warn: console.warn,
  error: console.error,
  debug: console.debug
}

function getLogApi() {
  if (window.go && window.go.app && window.go.app.App) {
    return window.go.app.App
  }
  return null
}

function formatLogArgs(args) {
  return args.map(arg => {
    if (typeof arg === 'object') {
      try {
        return JSON.stringify(arg)
      } catch (e) {
        return String(arg)
      }
    }
    return String(arg)
  }).join(' ')
}

console.log = function(...args) {
  originalConsole.log.apply(console, args)
  try {
    const api = getLogApi()
    if (api && api.LogInfo) {
      api.LogInfo(formatLogArgs(args))
    }
  } catch (e) {}
}

console.info = function(...args) {
  originalConsole.info.apply(console, args)
  try {
    const api = getLogApi()
    if (api && api.LogInfo) {
      api.LogInfo(formatLogArgs(args))
    }
  } catch (e) {}
}

console.warn = function(...args) {
  originalConsole.warn.apply(console, args)
  try {
    const api = getLogApi()
    if (api && api.LogWarn) {
      api.LogWarn(formatLogArgs(args))
    }
  } catch (e) {}
}

console.error = function(...args) {
  originalConsole.error.apply(console, args)
  try {
    const api = getLogApi()
    if (api && api.LogError) {
      api.LogError(formatLogArgs(args))
    }
  } catch (e) {}
}

console.debug = function(...args) {
  originalConsole.debug.apply(console, args)
  try {
    const api = getLogApi()
    if (api && api.LogDebug) {
      api.LogDebug(formatLogArgs(args))
    }
  } catch (e) {}
}

window.addEventListener('error', function(event) {
  try {
    const api = getLogApi()
    if (api && api.LogError) {
      api.LogError(`JavaScript Error: ${event.message} at ${event.filename}:${event.lineno}:${event.colno}`)
    }
  } catch (e) {}
})

window.addEventListener('unhandledrejection', function(event) {
  try {
    const api = getLogApi()
    if (api && api.LogError) {
      api.LogError(`Unhandled Promise Rejection: ${event.reason}`)
    }
  } catch (e) {}
})

function waitForWails() {
  return new Promise((resolve) => {
    if (window.go && window.go.app && window.go.app.App) {
      resolve()
      return
    }
    
    const checkInterval = setInterval(() => {
      if (window.go && window.go.app && window.go.app.App) {
        clearInterval(checkInterval)
        resolve()
      }
    }, 100)
    
    setTimeout(() => {
      clearInterval(checkInterval)
      console.warn('Wails runtime timeout, proceeding anyway...')
      resolve()
    }, 5000)
  })
}

waitForWails().then(() => {
  console.log('应用程序启动')
  const app = createApp(App)
  app.use(i18n)
  app.mount('#app')
}).catch(err => {
  console.error('Failed to initialize app:', err)
  const app = createApp(App)
  app.use(i18n)
  app.mount('#app')
})
