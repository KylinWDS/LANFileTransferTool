<template>
  <div class="settings-tab">
    <!-- 外观设置 -->
    <div class="card">
      <h2>{{ t('settings.appearance') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('settings.theme') }}</label>
        <select v-model="settings.theme" class="select" @change="onSettingChange">
          <option value="light">{{ t('settings.lightTheme') }}</option>
          <option value="dark">{{ t('settings.darkTheme') }}</option>
        </select>
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.language') }}</label>
        <select v-model="settings.language" class="select" @change="onLanguageChange">
          <option value="zh-CN">中文</option>
          <option value="en">English</option>
          <option value="ru">Русский</option>
        </select>
      </div>
    </div>

    <!-- 服务器设置 -->
    <div class="card">
      <h2>{{ t('settings.server') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('settings.serverPort') }}</label>
        <input v-model.number="settings.server.port" type="number" class="input" style="max-width:120px" min="1024" max="65535" />
        <span class="text-sm text-secondary">{{ t('settings.range') }}: 1024-65535</span>
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.autoStart') }}</label>
        <input v-model="settings.server.autoStart" type="checkbox" />
      </div>
    </div>

    <!-- 传输设置 -->
    <div class="card">
      <h2>{{ t('settings.transfer') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('settings.chunkSize') }} (MB)</label>
        <input v-model.number="settings.transfer.chunkSize" type="number" class="input" style="max-width:120px" min="1" max="100" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.maxConnections') }}</label>
        <input v-model.number="settings.transfer.maxConnections" type="number" class="input" style="max-width:120px" min="1" max="100" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.enableResume') }}</label>
        <input v-model="settings.transfer.enableResume" type="checkbox" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.defaultProtocol') }}</label>
        <select v-model="settings.transfer.defaultProtocol" class="select">
          <option value="http">HTTP</option>
          <option value="websocket">WebSocket</option>
          <option value="udp">UDP</option>
          <option value="p2p">P2P</option>
        </select>
      </div>
    </div>

    <!-- 协议设置 -->
    <div class="card">
      <h2>{{ t('settings.protocols') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('settings.protocolPreference') }}</label>
        <select v-model="settings.transfer.defaultProtocol" class="select" @change="onProtocolPreferenceChange">
          <option value="auto">{{ t('settings.autoSelect') }}</option>
          <option value="http">HTTP</option>
          <option value="websocket">WebSocket</option>
          <option value="udp">UDP</option>
          <option value="p2p">P2P</option>
        </select>
        <span class="text-sm text-secondary">{{ t('settings.autoSelectHint') }}</span>
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.enableWebSocket') }}</label>
        <input v-model="settings.protocols.websocket" type="checkbox" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.enableUDP') }}</label>
        <input v-model="settings.protocols.udp" type="checkbox" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.enableP2P') }}</label>
        <input v-model="settings.protocols.p2p" type="checkbox" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.enableDiscovery') }}</label>
        <input v-model="settings.protocols.discovery" type="checkbox" />
      </div>
    </div>

    <!-- 安全设置 -->
    <div class="card">
      <h2>{{ t('settings.security') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('settings.tokenExpiry') }} ({{ t('settings.hours') }})</label>
        <input v-model.number="settings.security.tokenExpiry" type="number" class="input" style="max-width:120px" min="1" max="168" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.enableEncryption') }}</label>
        <input v-model="settings.security.enableEncryption" type="checkbox" />
      </div>
    </div>

    <!-- 历史记录设置 -->
    <div class="card">
      <h2>{{ t('settings.history') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('settings.maxHistory') }}</label>
        <input v-model.number="settings.history.maxRecords" type="number" class="input" style="max-width:120px" min="10" max="1000" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.autoClear') }}</label>
        <input v-model="settings.history.autoClear" type="checkbox" />
      </div>
    </div>

    <!-- 线程池设置 -->
    <div class="card">
      <h2>{{ t('settings.threadPool') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('performance.poolSize') }}</label>
        <input v-model.number="settings.pool.size" type="number" class="input" style="max-width:120px" min="1" max="100" />
        <span class="text-sm text-secondary">{{ t('settings.range') }}: 1-100</span>
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.autoStartPool') }}</label>
        <input v-model="settings.pool.autoStart" type="checkbox" />
      </div>
      <div class="setting-row">
        <span class="label">{{ t('settings.poolStatus') }}:</span>
        <span :class="['status-badge', poolRunning ? 'status-ok' : 'status-warning']">
          {{ poolRunning ? t('settings.poolRunning') : t('settings.poolStopped') }}
        </span>
        <button class="btn btn-sm" :disabled="poolRunning" @click="startPool">{{ t('settings.startPool') }}</button>
        <button class="btn btn-sm" :disabled="!poolRunning" @click="stopPool">{{ t('settings.stopPool') }}</button>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="card">
      <div class="flex gap-2">
        <button class="btn btn-primary" @click="saveSettings">{{ t('settings.save') }}</button>
        <button class="btn" @click="resetSettings">{{ t('settings.reset') }}</button>
        <button class="btn btn-secondary" @click="exportSettings">{{ t('settings.export') }}</button>
        <button class="btn btn-secondary" @click="importSettings">{{ t('settings.import') }}</button>
      </div>
      <div v-if="savedMsg" class="text-sm mt-2" :class="saveSuccess ? 'text-success' : 'text-error'">{{ savedMsg }}</div>
    </div>

    <!-- 关于信息 -->
    <div class="card">
      <h2>{{ t('settings.about') }}</h2>
      <div class="info-row">
        <span class="label">{{ t('settings.appName') }}:</span>
        <span>{{ appInfo.name }}</span>
      </div>
      <div class="info-row">
        <span class="label">{{ t('settings.version') }}:</span>
        <span>{{ appInfo.version }}</span>
      </div>
      <div class="info-row">
        <span class="label">{{ t('settings.configPath') }}:</span>
        <span class="text-sm text-secondary">{{ configPath }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const props = defineProps({
  theme: {
    type: String,
    default: 'light'
  },
  language: {
    type: String,
    default: 'zh-CN'
  }
})

const { t, locale } = useI18n()
const emit = defineEmits(['settings-change', 'theme-change', 'language-change'])

// 应用信息
const appInfo = ref({ name: 'LANftt', version: 'v0.2.0' })
const configPath = ref('')

// 设置状态
const savedMsg = ref('')
const saveSuccess = ref(true)

// 默认设置
const defaultSettings = {
  theme: 'light',
  language: 'zh-CN',
  server: {
    port: 8080,
    autoStart: true
  },
  transfer: {
    chunkSize: 1,
    maxConnections: 10,
    enableResume: true,
    defaultProtocol: 'auto' // 默认自动选择
  },
  protocols: {
    websocket: true,
    udp: true,
    p2p: true,
    discovery: true
  },
  security: {
    tokenExpiry: 24,
    enableEncryption: true
  },
  history: {
    maxRecords: 100,
    autoClear: false
  },
  pool: {
    size: 10,
    autoStart: false
  }
}

// 线程池状态
const poolRunning = ref(false)

// 当前设置
const settings = reactive({ ...defaultSettings })

// 加载设置
const loadSettings = async () => {
  try {
    // 获取应用信息
    const info = await api.GetAppInfo()
    if (info) {
      appInfo.value = info
    }

    // 获取用户配置
    const config = await api.GetUserConfig()
    if (config) {
      // 合并配置
      Object.assign(settings, {
        theme: config.theme || defaultSettings.theme,
        language: config.language || defaultSettings.language,
        server: { ...defaultSettings.server, ...config.settings?.server },
        transfer: { ...defaultSettings.transfer, ...config.settings?.transfer },
        protocols: { ...defaultSettings.protocols, ...config.settings?.protocols },
        security: { ...defaultSettings.security, ...config.settings?.security },
        history: { ...defaultSettings.history, ...config.settings?.history },
        pool: { ...defaultSettings.pool, ...config.settings?.pool }
      })

      // 同步到i18n
      locale.value = settings.language
      
      // 通知父组件
      emit('theme-change', settings.theme)
      emit('language-change', settings.language)

      // 检查线程池状态
      await checkPoolStatus()

      // 如果设置了自动启动线程池，则启动
      if (settings.pool.autoStart) {
        await startPool()
      }
    }
  } catch (e) {
    console.error('加载设置失败:', e)
  }
}

// 检查线程池状态
const checkPoolStatus = async () => {
  try {
    // 通过性能监控接口获取线程池状态
    const stats = await api.GetPerformanceStats()
    if (stats) {
      poolRunning.value = stats.pool_running || false
    }
  } catch (e) {
    console.error('检查线程池状态失败:', e)
  }
}

// 启动线程池
const startPool = async () => {
  try {
    await api.InitThreadPool(settings.pool.size)
    poolRunning.value = true
    savedMsg.value = t('settings.poolStarted')
    saveSuccess.value = true
    setTimeout(() => { savedMsg.value = '' }, 2000)
  } catch (e) {
    savedMsg.value = t('settings.poolStartFailed')
    saveSuccess.value = false
    console.error('启动线程池失败:', e)
  }
}

// 停止线程池
const stopPool = async () => {
  try {
    await api.StopThreadPool()
    poolRunning.value = false
    savedMsg.value = t('settings.poolStopped')
    saveSuccess.value = true
    setTimeout(() => { savedMsg.value = '' }, 2000)
  } catch (e) {
    savedMsg.value = t('settings.poolStopFailed')
    saveSuccess.value = false
    console.error('停止线程池失败:', e)
  }
}

// 语言改变
const onLanguageChange = () => {
  locale.value = settings.language
  emit('language-change', settings.language)
  saveSettings()
}

// 协议偏好改变
const onProtocolPreferenceChange = async () => {
  try {
    await api.SetProtocolPreference(settings.transfer.defaultProtocol)
    savedMsg.value = t('settings.protocolPreferenceSaved')
    saveSuccess.value = true
    setTimeout(() => { savedMsg.value = '' }, 2000)
  } catch (e) {
    console.error('设置协议偏好失败:', e)
  }
  saveSettings()
}

// 设置改变
const onSettingChange = () => {
  emit('settings-change', settings)
  saveSettings()
}

// 保存设置
const saveSettings = async () => {
  try {
    await api.SaveUserConfig({
      theme: settings.theme,
      language: settings.language,
      settings: {
        server: settings.server,
        transfer: settings.transfer,
        protocols: settings.protocols,
        security: settings.security,
        history: settings.history,
        pool: settings.pool
      }
    })
    
    // 应用主题
    document.documentElement.setAttribute('data-theme', settings.theme)
    
    savedMsg.value = t('settings.saved')
    saveSuccess.value = true
    setTimeout(() => { savedMsg.value = '' }, 2000)
  } catch (e) {
    savedMsg.value = t('settings.saveFailed')
    saveSuccess.value = false
    console.error('保存设置失败:', e)
  }
}

// 重置设置
const resetSettings = async () => {
  if (!confirm(t('settings.resetConfirm'))) return
  try {
    await api.ResetUserConfig()
    Object.assign(settings, defaultSettings)
    locale.value = settings.language
    document.documentElement.setAttribute('data-theme', settings.theme)
    emit('theme-change', settings.theme)
    emit('language-change', settings.language)
    savedMsg.value = t('settings.resetSuccess')
    saveSuccess.value = true
    setTimeout(() => { savedMsg.value = '' }, 2000)
  } catch (e) {
    console.error('重置设置失败:', e)
  }
}

// 导出设置
const exportSettings = () => {
  const dataStr = JSON.stringify(settings, null, 2)
  const blob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `lanftt-settings-${new Date().toISOString().split('T')[0]}.json`
  a.click()
  URL.revokeObjectURL(url)
}

// 导入设置
const importSettings = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    try {
      const text = await file.text()
      const imported = JSON.parse(text)
      Object.assign(settings, imported)
      await saveSettings()
      savedMsg.value = t('settings.importSuccess')
      saveSuccess.value = true
    } catch (e) {
      savedMsg.value = t('settings.importFailed')
      saveSuccess.value = false
    }
  }
  input.click()
}

// 监听设置变化，同步到全局
watch(() => settings.theme, (newTheme) => {
  document.documentElement.setAttribute('data-theme', newTheme)
  emit('theme-change', newTheme)
})

// 监听父组件传递的props变化，同步到本地设置
watch(() => props.theme, (newTheme) => {
  if (newTheme && newTheme !== settings.theme) {
    settings.theme = newTheme
    document.documentElement.setAttribute('data-theme', newTheme)
  }
})

watch(() => props.language, (newLanguage) => {
  if (newLanguage && newLanguage !== settings.language) {
    settings.language = newLanguage
    locale.value = newLanguage
  }
})

onMounted(loadSettings)
</script>

<style scoped>
.settings-tab {
  padding-bottom: 40px;
}

.setting-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.setting-row .label {
  min-width: 180px;
  font-weight: 500;
}

.info-row {
  display: flex;
  gap: 12px;
  margin-bottom: 8px;
}

.text-success {
  color: var(--success);
}

.text-error {
  color: var(--danger);
}
</style>