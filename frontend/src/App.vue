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
        <select v-model="currentLanguage" class="select" @change="onLanguageChange">
          <option value="zh-CN">中文</option>
          <option value="en">English</option>
          <option value="ru">Русский</option>
        </select>
        <button class="btn" @click="toggleTheme">
          {{ currentTheme === 'light' ? '🌙' : '☀️' }}
        </button>
      </div>
    </header>

    <main class="container">
      <nav class="tabs">
        <button
          v-for="(label, key) in tabLabels"
          :key="key"
          :class="['tab', { active: activeTab === key }]"
          @click="activeTab = key"
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
  settings: SettingsTab
}

const updateTabLabels = () => {
  const keys = ['transfer', 'download', 'history', 'encryption', 'performance', 'environment', 'settings']
  keys.forEach(k => { tabLabels[k] = t(`tabs.${k}`) })
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
</style>
