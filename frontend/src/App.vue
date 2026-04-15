<template>
  <div :data-theme="currentTheme" class="app-container">
    <header class="app-header">
      <div class="header-left">
        <span class="app-title">{{ $t('header.title') }}</span>
        <span class="server-badge" :class="{ on: serverRunning }">
          {{ serverRunning ? $t('header.running') : $t('header.stopped') }}
        </span>
      </div>
      <div class="header-right">
        <select 
          v-model="currentLanguage" 
          class="select" 
          @change="onLanguageChange"
          :aria-label="$t('settings.language')"
        >
          <option value="zh-CN">中文</option>
          <option value="en">English</option>
          <option value="ru">Русский</option>
        </select>
        <button
          class="btn btn-icon theme-toggle"
          @click="toggleTheme"
          :aria-label="currentTheme === 'light' ? $t('accessibility.switchToDark') : $t('accessibility.switchToLight')"
          :title="currentTheme === 'light' ? $t('accessibility.switchToDark') : $t('accessibility.switchToLight')"
        >
          <svg v-if="currentTheme === 'light'" class="theme-icon" viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M12 3a9 9 0 1 0 9 9c0-.46-.04-.92-.1-1.36a5.389 5.389 0 0 1-4.4 2.26 5.403 5.403 0 0 1-3.14-9.8c-.44-.06-.9-.1-1.36-.1z"/>
          </svg>
          <svg v-else class="theme-icon" viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M12 7c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5zM2 13h2c.55 0 1-.45 1-1s-.45-1-1-1H2c-.55 0-1 .45-1 1s.45 1 1 1zm18 0h2c.55 0 1-.45 1-1s-.45-1-1-1h-2c-.55 0-1 .45-1 1s.45 1 1 1zM11 2v2c0 .55.45 1 1 1s1-.45 1-1V2c0-.55-.45-1-1-1s-1 .45-1 1zm0 18v2c0 .55.45 1 1 1s1-.45 1-1v-2c0-.55-.45-1-1-1s-1 .45-1 1zM5.99 4.58a.996.996 0 0 0-1.41 0 .996.996 0 0 0 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0s.39-1.03 0-1.41L5.99 4.58zm12.37 12.37a.996.996 0 0 0-1.41 0 .996.996 0 0 0 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0a.996.996 0 0 0 0-1.41l-1.06-1.06zm1.06-10.96a.996.996 0 0 0 0-1.41.996.996 0 0 0-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0l1.06-1.06zM7.05 18.36a.996.996 0 0 0 0-1.41.996.996 0 0 0-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0l1.06-1.06z"/>
          </svg>
        </button>
      </div>
    </header>

    <main class="container">
      <nav class="tabs" role="tablist" aria-label="主导航">
        <button
          v-for="(label, key) in tabLabels"
          :key="key"
          :class="['tab', { active: activeTab === key }]"
          :role="'tab'"
          :aria-selected="activeTab === key"
          :tabindex="activeTab === key ? 0 : -1"
          @click="activeTab = key"
          @keydown="handleTabKeydown($event, key)"
        >{{ label }}</button>
      </nav>
      <keep-alive>
        <component
          :is="tabComponents[activeTab]"
          :theme="currentTheme"
          :language="currentLanguage"
          @theme-change="onThemeChangeFromSettings"
          @language-change="onLanguageChangeFromSettings"
        />
      </keep-alive>
    </main>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import api from './api'
import TransferTab from './components/TransferTab.vue'
import DownloadTab from './components/DownloadTab.vue'
import HistoryTab from './components/HistoryTab.vue'
import EncryptionTab from './components/EncryptionTab.vue'
import PerformanceTab from './components/PerformanceTab.vue'
import EnvironmentTab from './components/EnvironmentTab.vue'
import SettingsTab from './components/SettingsTab.vue'
import LogTab from './components/LogTab.vue'

const { t, locale } = useI18n()

const currentTheme = ref('light')
const currentLanguage = ref('zh-CN')
const activeTab = ref('transfer')
const serverRunning = ref(false)

const tabLabels = reactive({})
const tabComponents = {
  transfer: TransferTab,
  download: DownloadTab,
  history: HistoryTab,
  encryption: EncryptionTab,
  performance: PerformanceTab,
  environment: EnvironmentTab,
  settings: SettingsTab,
  logs: LogTab
}

const updateTabLabels = () => {
  const keys = ['transfer', 'download', 'history', 'encryption', 'performance', 'environment', 'settings', 'logs']
  keys.forEach(k => { tabLabels[k] = t(`tabs.${k}`) })
}

const tabKeys = ['transfer', 'download', 'history', 'encryption', 'performance', 'environment', 'settings', 'logs']

const handleTabKeydown = (event, currentKey) => {
  const currentIndex = tabKeys.indexOf(currentKey)
  let newIndex = currentIndex

  if (event.key === 'ArrowRight' || event.key === 'ArrowDown') {
    newIndex = (currentIndex + 1) % tabKeys.length
    event.preventDefault()
  } else if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
    newIndex = (currentIndex - 1 + tabKeys.length) % tabKeys.length
    event.preventDefault()
  } else if (event.key === 'Home') {
    newIndex = 0
    event.preventDefault()
  } else if (event.key === 'End') {
    newIndex = tabKeys.length - 1
    event.preventDefault()
  }

  if (newIndex !== currentIndex) {
    activeTab.value = tabKeys[newIndex]
    // 聚焦到新的 tab
    setTimeout(() => {
      const tabs = document.querySelectorAll('.tab')
      if (tabs[newIndex]) tabs[newIndex].focus()
    }, 0)
  }
}

const toggleTheme = async () => {
  currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
  document.documentElement.setAttribute('data-theme', currentTheme.value)
  // 同步保存到用户配置
  await saveThemeToConfig()
}

const onLanguageChange = async () => {
  locale.value = currentLanguage.value
  updateTabLabels()
  // 同步保存到用户配置
  await saveLanguageToConfig()
}

// 从设置页面接收主题变更
const onThemeChangeFromSettings = async (theme) => {
  if (currentTheme.value !== theme) {
    currentTheme.value = theme
    document.documentElement.setAttribute('data-theme', theme)
    await saveThemeToConfig()
  }
}

// 从设置页面接收语言变更
const onLanguageChangeFromSettings = async (language) => {
  if (currentLanguage.value !== language) {
    currentLanguage.value = language
    locale.value = language
    updateTabLabels()
    await saveLanguageToConfig()
  }
}

// 保存主题到配置
const saveThemeToConfig = async () => {
  try {
    const config = await api.GetUserConfig()
    if (config) {
      await api.SaveUserConfig({
        theme: currentTheme.value,
        language: currentLanguage.value,
        settings: config.settings || {}
      })
    }
  } catch (e) {
    console.error('保存主题配置失败:', e)
  }
}

// 保存语言到配置
const saveLanguageToConfig = async () => {
  try {
    const config = await api.GetUserConfig()
    if (config) {
      await api.SaveUserConfig({
        theme: currentTheme.value,
        language: currentLanguage.value,
        settings: config.settings || {}
      })
    }
  } catch (e) {
    console.error('保存语言配置失败:', e)
  }
}

watch(currentTheme, (v) => {
  document.documentElement.setAttribute('data-theme', v)
})

onMounted(async () => {
  updateTabLabels()
  try {
    const info = await api.GetServerInfo()
    if (info) serverRunning.value = info.running
    const config = await api.GetUserConfig()
    if (config) {
      currentTheme.value = config.theme || 'light'
      currentLanguage.value = config.language || 'zh-CN'
      locale.value = currentLanguage.value
      document.documentElement.setAttribute('data-theme', currentTheme.value)
      updateTabLabels()
    }
  } catch {}
})
</script>

<style scoped>
.app-container { min-height: 100vh; }

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  border-bottom: 1px solid var(--border);
  background: var(--bg-secondary);
}

.header-left { display: flex; align-items: center; gap: 12px; }
.app-title { font-size: 15px; font-weight: 600; }

.server-badge {
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
  background: rgba(239,68,68,0.1);
  color: var(--danger);
}
.server-badge.on { background: rgba(16,185,129,0.1); color: var(--success); }

.header-right { display: flex; align-items: center; gap: 8px; }

.tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--border);
  padding-bottom: 0;
  overflow-x: auto;
}

.tab {
  padding: 10px 16px;
  border: none;
  background: none;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  border-bottom: 2px solid transparent;
  transition: all 0.15s;
  white-space: nowrap;
}

.tab:hover { color: var(--primary); }
.tab.active { color: var(--primary); border-bottom-color: var(--primary); }

.theme-toggle {
  transition: transform 0.2s ease, background-color 0.2s ease;
}
.theme-toggle:hover {
  transform: rotate(15deg);
  background: var(--bg-tertiary);
}
.theme-toggle:active {
  transform: rotate(0deg) scale(0.95);
}
.theme-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}
</style>
