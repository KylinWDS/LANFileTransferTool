<template>
  <div class="log-tab">
    <div class="card">
      <div class="log-header">
        <h2>{{ t('logs.title') }}</h2>
        <div class="log-actions">
          <select v-model="filterLevel" class="select select-sm" :aria-label="t('logs.filterLevel')">
            <option value="all">{{ t('logs.allLevels') }}</option>
            <option value="DEBUG">DEBUG</option>
            <option value="INFO">INFO</option>
            <option value="WARN">WARN</option>
            <option value="ERROR">ERROR</option>
          </select>
          <button class="btn btn-sm" @click="refreshLogs" :disabled="loading">
            {{ loading ? t('common.loading') : t('logs.refresh') }}
          </button>
          <button class="btn btn-sm" @click="showClearConfirm" :disabled="loading">
            {{ t('logs.clear') }}
          </button>
          <button class="btn btn-sm" @click="exportLogs" :disabled="filteredLogs.length === 0">
            {{ t('logs.export') }}
          </button>
          <label class="checkbox-label">
            <input type="checkbox" v-model="autoRefresh" />
            {{ t('logs.autoRefresh') }}
          </label>
        </div>
      </div>

      <div class="log-stats">
        <span class="stat-item">
          <span class="stat-label">{{ t('logs.total') }}:</span>
          <span class="stat-value">{{ logs.length }}</span>
        </span>
        <span class="stat-item stat-error">
          <span class="stat-label">ERROR:</span>
          <span class="stat-value">{{ errorCount }}</span>
        </span>
        <span class="stat-item stat-warn">
          <span class="stat-label">WARN:</span>
          <span class="stat-value">{{ warnCount }}</span>
        </span>
        <span class="stat-item stat-info">
          <span class="stat-label">INFO:</span>
          <span class="stat-value">{{ infoCount }}</span>
        </span>
      </div>

      <div class="log-container" ref="logContainer">
        <div v-if="loading && logs.length === 0" class="skeleton-container">
          <div v-for="i in 5" :key="i" class="skeleton-log skeleton">
            <div class="skeleton skeleton-time"></div>
            <div class="skeleton skeleton-text"></div>
          </div>
        </div>
        <div v-else-if="filteredLogs.length === 0" class="empty-state">
          <div class="empty-state-icon">📝</div>
          <div class="empty-state-text">{{ t('logs.noLogs') }}</div>
        </div>
        <div v-else>
          <div
            v-for="(log, index) in filteredLogs"
            :key="index"
            v-memo="[log.level, log.message]"
            class="log-entry fade-in"
            :class="'log-' + log.level.toLowerCase()"
          >
            <span class="log-time">{{ formatTime(log.timestamp) }}</span>
            <span class="log-level">[{{ log.level }}]</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="confirmDialog.show" class="confirm-dialog-overlay" @click="cancelConfirm">
      <div class="confirm-dialog" @click.stop>
        <div class="confirm-dialog-header">
          <h3>{{ t('logs.clearConfirmTitle') }}</h3>
        </div>
        <div class="confirm-dialog-body">
          <p>{{ t('logs.clearConfirm') }}</p>
        </div>
        <div class="confirm-dialog-footer">
          <button class="btn" @click="cancelConfirm">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" @click="confirmClear">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </div>

    <div v-if="toast" class="toast" :class="toastType">{{ toast }}</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const logs = ref([])
const filterLevel = ref('all')
const loading = ref(false)
const autoRefresh = ref(true)
const logContainer = ref(null)
const toast = ref('')
const toastType = ref('')

const confirmDialog = ref({
  show: false
})

let refreshInterval = null

const showToast = (message, type = 'info') => {
  toast.value = message
  toastType.value = type
  setTimeout(() => { toast.value = '' }, 3000)
}

const filteredLogs = computed(() => {
  if (filterLevel.value === 'all') {
    return logs.value
  }
  return logs.value.filter(log => log.level === filterLevel.value)
})

const errorCount = computed(() => logs.value.filter(l => l.level === 'ERROR').length)
const warnCount = computed(() => logs.value.filter(l => l.level === 'WARN').length)
const infoCount = computed(() => logs.value.filter(l => l.level === 'INFO').length)

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

const refreshLogs = async () => {
  loading.value = true
  try {
    const result = await api.GetLogs()
    if (result && Array.isArray(result)) {
      logs.value = result
      if (autoRefresh.value) {
        await nextTick()
        scrollToBottom()
      }
    }
  } catch (e) {
    console.error('获取日志失败:', e)
    showToast(t('logs.refreshFailed'), 'error')
  } finally {
    loading.value = false
  }
}

const scrollToBottom = () => {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

const showClearConfirm = () => {
  confirmDialog.value.show = true
}

const cancelConfirm = () => {
  confirmDialog.value.show = false
}

const confirmClear = async () => {
  confirmDialog.value.show = false
  try {
    await api.ClearLogs()
    logs.value = []
    showToast(t('logs.cleared'), 'success')
  } catch (e) {
    console.error('清除日志失败:', e)
    showToast(t('logs.clearFailed'), 'error')
  }
}

const exportLogs = async () => {
  try {
    const defaultFilename = `lanftt-logs-${new Date().toISOString().split('T')[0]}.json`
    const filePath = await api.SelectSaveFile(defaultFilename)
    if (!filePath) {
      return
    }
    const dataStr = JSON.stringify(filteredLogs.value, null, 2)
    await api.SaveTextFile(filePath, dataStr)
    showToast(t('logs.exported'), 'success')
  } catch (e) {
    console.error('导出日志失败:', e)
    showToast(t('logs.exportFailed'), 'error')
  }
}

watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshInterval = setInterval(refreshLogs, 2000)
  } else {
    if (refreshInterval) {
      clearInterval(refreshInterval)
      refreshInterval = null
    }
  }
})

onMounted(() => {
  refreshLogs()
  if (autoRefresh.value) {
    refreshInterval = setInterval(refreshLogs, 2000)
  }
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.log-tab {
  padding-bottom: 40px;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}

.log-header h2 {
  margin: 0;
}

.log-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.select-sm {
  padding: 4px 8px;
  font-size: 12px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  cursor: pointer;
}

.log-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
  font-size: 12px;
}

.stat-item {
  display: flex;
  gap: 4px;
}

.stat-label {
  color: var(--text-secondary);
}

.stat-error .stat-value {
  color: var(--danger);
  font-weight: 600;
}

.stat-warn .stat-value {
  color: var(--warning);
  font-weight: 600;
}

.stat-info .stat-value {
  color: var(--primary);
  font-weight: 600;
}

.log-container {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 12px;
  max-height: 500px;
  overflow-y: auto;
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
  font-size: 12px;
}

.log-empty {
  text-align: center;
  color: var(--text-secondary);
  padding: 40px;
}

.log-entry {
  display: flex;
  gap: 8px;
  padding: 4px 0;
  border-bottom: 1px solid var(--border);
}

.log-entry:last-child {
  border-bottom: none;
}

.log-time {
  color: var(--text-secondary);
  white-space: nowrap;
}

.log-level {
  font-weight: 600;
  white-space: nowrap;
}

.log-debug .log-level {
  color: #9ca3af;
}

.log-info .log-level {
  color: var(--primary);
}

.log-warn .log-level {
  color: var(--warning);
}

.log-error .log-level {
  color: var(--danger);
}

.log-error {
  background: rgba(239, 68, 68, 0.1);
  margin: 0 -12px;
  padding: 4px 12px;
}

.log-message {
  flex: 1;
  word-break: break-word;
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

@keyframes toast-in {
  0% { opacity: 0; transform: translateX(-50%) translateY(8px); }
  10% { opacity: 1; transform: translateX(-50%) translateY(0); }
  80% { opacity: 1; }
  100% { opacity: 0; }
}

.skeleton-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-log {
  display: flex;
  gap: 8px;
  padding: 4px 0;
  border-bottom: 1px solid var(--border);
}

.skeleton-time {
  width: 70px;
  height: 14px;
}

.skeleton-log .skeleton-text {
  flex: 1;
  height: 14px;
}

</style>
