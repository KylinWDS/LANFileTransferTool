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
        <button class="btn" @click="showResetConfirm">{{ t('settings.reset') }}</button>
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

    <!-- 自定义确认对话框 -->
    <div v-if="confirmDialog.show" class="confirm-dialog-overlay" @click="cancelConfirm">
      <div class="confirm-dialog" @click.stop>
        <div class="confirm-dialog-header">
          <h3>{{ confirmDialog.title }}</h3>
        </div>
        <div class="confirm-dialog-body">
          <p>{{ confirmDialog.message }}</p>
        </div>
        <div class="confirm-dialog-footer">
          <button class="btn" @click="cancelConfirm">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" @click="confirmAction">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </div>

    <!-- Toast 提示 -->
    <div v-if="toast" class="toast" :class="toastType">{{ toast }}</div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
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

const appInfo = ref({ name: 'LANftt', version: 'v0.2.0' })
const configPath = ref('')

const savedMsg = ref('')
const saveSuccess = ref(true)

const toast = ref('')
const toastType = ref('')

const showToast = (message, type = 'info') => {
  toast.value = message
  toastType.value = type
  setTimeout(() => { toast.value = '' }, 3000)
}

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
    defaultProtocol: 'auto'
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

const poolRunning = ref(false)

const settings = reactive({ ...defaultSettings })

const confirmDialog = ref({
  show: false,
  title: '',
  message: '',
  action: null
})

const showResetConfirm = () => {
  confirmDialog.value = {
    show: true,
    title: t('settings.resetConfirmTitle'),
    message: t('settings.resetConfirm'),
    action: 'reset'
  }
}

const cancelConfirm = () => {
  confirmDialog.value.show = false
}

const confirmAction = async () => {
  const action = confirmDialog.value.action
  confirmDialog.value.show = false
  
  if (action === 'reset') {
    await doResetSettings()
  }
}

const loadSettings = async () => {
  try {
    const info = await api.GetAppInfo()
    if (info) {
      appInfo.value = info
    }

    const config = await api.GetUserConfig()
    if (config) {
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

      locale.value = settings.language
      
      emit('theme-change', settings.theme)
      emit('language-change', settings.language)

      await checkPoolStatus()

      if (settings.pool.autoStart) {
        await startPool()
      }
    }
  } catch (e) {
    console.error('加载设置失败:', e)
    showToast(t('settings.loadFailed'), 'error')
  }
}

const checkPoolStatus = async () => {
  try {
    const stats = await api.GetPerformanceStats()
    if (stats) {
      poolRunning.value = stats.pool_running || false
    }
  } catch (e) {
    console.error('检查线程池状态失败:', e)
  }
}

const startPool = async () => {
  try {
    await api.InitThreadPool(settings.pool.size)
    poolRunning.value = true
    showToast(t('settings.poolStarted'), 'success')
  } catch (e) {
    showToast(t('settings.poolStartFailed'), 'error')
    console.error('启动线程池失败:', e)
  }
}

const stopPool = async () => {
  try {
    await api.StopThreadPool()
    poolRunning.value = false
    showToast(t('settings.poolStopped'), 'success')
  } catch (e) {
    showToast(t('settings.poolStopFailed'), 'error')
    console.error('停止线程池失败:', e)
  }
}

const onLanguageChange = () => {
  locale.value = settings.language
  emit('language-change', settings.language)
  saveSettings()
}

const onProtocolPreferenceChange = async () => {
  try {
    await api.SetProtocolPreference(settings.transfer.defaultProtocol)
    showToast(t('settings.protocolPreferenceSaved'), 'success')
  } catch (e) {
    console.error('设置协议偏好失败:', e)
    showToast(t('settings.protocolPreferenceFailed'), 'error')
  }
  saveSettings()
}

const onSettingChange = () => {
  emit('settings-change', settings)
  saveSettings()
}

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

const doResetSettings = async () => {
  try {
    await api.ResetUserConfig()
    Object.assign(settings, defaultSettings)
    locale.value = settings.language
    document.documentElement.setAttribute('data-theme', settings.theme)
    emit('theme-change', settings.theme)
    emit('language-change', settings.language)
    showToast(t('settings.resetSuccess'), 'success')
  } catch (e) {
    console.error('重置设置失败:', e)
    showToast(t('settings.resetFailed'), 'error')
  }
}

const exportSettings = async () => {
  try {
    const defaultFilename = `lanftt-settings-${new Date().toISOString().split('T')[0]}.json`
    const filePath = await api.SelectSaveFile(defaultFilename)
    if (!filePath) {
      return
    }
    const dataStr = JSON.stringify(settings, null, 2)
    await api.SaveTextFile(filePath, dataStr)
    showToast(t('settings.exportSuccess'), 'success')
  } catch (e) {
    console.error('导出设置失败:', e)
    showToast(t('settings.exportFailed'), 'error')
  }
}

const importSettings = async () => {
  try {
    const files = await api.SelectFiles(false)
    if (!files || files.length === 0) {
      return
    }
    const filePath = files[0].path
    const text = await api.ReadTextFile(filePath)
    const imported = JSON.parse(text)
    Object.assign(settings, imported)
    await saveSettings()
    showToast(t('settings.importSuccess'), 'success')
  } catch (e) {
    console.error('导入设置失败:', e)
    showToast(t('settings.importFailed'), 'error')
  }
}

watch(() => settings.theme, (newTheme) => {
  document.documentElement.setAttribute('data-theme', newTheme)
  emit('theme-change', newTheme)
})

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
  padding: 8px 12px;
  border-radius: 6px;
  transition: background var(--transition-fast);
}

.setting-row:hover {
  background: var(--bg);
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

.confirm-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.confirm-dialog {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
  max-width: 400px;
  width: 90%;
}

.confirm-dialog-header h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
}

.confirm-dialog-body p {
  margin: 0 0 20px 0;
  color: var(--text-secondary);
}

.confirm-dialog-footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.toast {
  position: fixed;
  bottom: 32px;
  left: 50%;
  transform: translateX(-50%);
  padding: 10px 24px;
  border-radius: 6px;
  font-size: 13px;
  z-index: 9999;
  animation: toast-in 3s ease;
}

.toast.success {
  background: var(--success);
  color: #fff;
}

.toast.error {
  background: var(--danger);
  color: #fff;
}

.toast.warning {
  background: var(--warning);
  color: #000;
}

.toast.info {
  background: var(--primary);
  color: #fff;
}

@keyframes toast-in {
  0% { opacity: 0; transform: translateX(-50%) translateY(8px); }
  10% { opacity: 1; transform: translateX(-50%) translateY(0); }
  80% { opacity: 1; }
  100% { opacity: 0; }
}
</style>
