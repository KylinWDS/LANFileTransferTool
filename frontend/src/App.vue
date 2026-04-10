<template>
  <div :data-theme="currentTheme" class="app-container">
    <HeaderComponent
      :current-theme="currentTheme"
      :current-language="currentLanguage"
      @toggle-theme="toggleTheme"
      @change-language="changeLanguage"
    />

    <main class="container">
      <div class="tabs">
        <button
          v-for="(tab, key) in tabs"
          :key="key"
          :class="['tab-btn', { active: activeTab === key }]"
          @click="activeTab = key"
        >
          {{ $t(`tabs.${key}`) }}
        </button>
      </div>

      <component :is="currentTabComponent" />
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import api from './api'
import HeaderComponent from './components/HeaderComponent.vue'
import TransferTab from './components/TransferTab.vue'
import DownloadTab from './components/DownloadTab.vue'
import HistoryTab from './components/HistoryTab.vue'
import EncryptionTab from './components/EncryptionTab.vue'
import PerformanceTab from './components/PerformanceTab.vue'
import EnvironmentTab from './components/EnvironmentTab.vue'
import SettingsTab from './components/SettingsTab.vue'

const { locale } = useI18n()

const currentTheme = ref('light')
const currentLanguage = ref('zh-CN')
const activeTab = ref('transfer')

const tabs = {
  transfer: 'transfer',
  download: 'download',
  history: 'history',
  encryption: 'encryption',
  performance: 'performance',
  environment: 'environment',
  settings: 'settings'
}

const currentTabComponent = computed(() => {
  const components = {
    transfer: TransferTab,
    download: DownloadTab,
    history: HistoryTab,
    encryption: EncryptionTab,
    performance: PerformanceTab,
    environment: EnvironmentTab,
    settings: SettingsTab
  }
  return components[activeTab.value]
})

const toggleTheme = () => {
  currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
  // 同时更新到 body 标签，确保全局生效
  document.documentElement.setAttribute('data-theme', currentTheme.value)
}

const changeLanguage = (lang) => {
  currentLanguage.value = lang
  locale.value = lang
}

// 监听主题变化，同步到 HTML 标签
watch(currentTheme, (newTheme) => {
  document.documentElement.setAttribute('data-theme', newTheme)
  // 同时保存配置
  saveThemePreference(newTheme)
})

const saveThemePreference = async (theme) => {
  try {
    const config = await api.GetUserConfig()
    if (config) {
      await api.SaveUserConfig({
        ...config,
        theme
      })
    }
  } catch (error) {
    console.error('保存主题偏好失败:', error)
  }
}

onMounted(async () => {
  try {
    const config = await api.GetUserConfig()
    if (config) {
      currentTheme.value = config.theme || 'light'
      currentLanguage.value = config.language || 'zh-CN'
      locale.value = currentLanguage.value
      // 设置到 HTML 标签
      document.documentElement.setAttribute('data-theme', currentTheme.value)
    }
  } catch (error) {
    console.error('加载用户配置失败:', error)
  }
})
</script>

<style scoped>
.app-container {
  min-height: 100vh;
  width: 100%;
}

.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  flex-wrap: wrap;
  border-bottom: 2px solid var(--border-color);
  padding-bottom: 12px;
}

.tab-btn {
  padding: 10px 20px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  border-radius: 6px 6px 0 0;
  transition: all 0.3s ease;
  position: relative;
}

.tab-btn:hover {
  color: var(--primary-color);
  background-color: var(--surface-color);
}

.tab-btn.active {
  color: var(--primary-color);
  font-weight: 600;
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: -14px;
  left: 0;
  right: 0;
  height: 3px;
  background-color: var(--primary-color);
  border-radius: 3px 3px 0 0;
}
</style>
