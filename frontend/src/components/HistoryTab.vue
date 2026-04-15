<template>
  <div class="history-tab">
    <div class="card">
      <div class="flex-between mb-2">
        <h2>{{ t('history.title') }}</h2>
        <div class="flex gap-2">
          <button class="btn btn-sm" :disabled="loading" @click="load">{{ loading ? t('common.loading') : t('history.refresh') }}</button>
          <button class="btn btn-sm" :disabled="!records.length || clearing" @click="showClearConfirm">{{ clearing ? t('common.loading') : t('history.clearHistory') }}</button>
        </div>
      </div>
      <div class="auto-refresh-row mb-3">
        <label class="flex items-center gap-2 text-sm">
          <input v-model="autoRefresh" type="checkbox" @change="toggleAutoRefresh" />
          <span>{{ t('history.autoRefresh') }}</span>
        </label>
      </div>
      <div v-if="loading" class="skeleton-container">
        <div v-for="i in 3" :key="i" class="skeleton-card skeleton">
          <div class="skeleton skeleton-title"></div>
          <div class="skeleton skeleton-text"></div>
          <div class="skeleton skeleton-text"></div>
        </div>
      </div>
      <div v-else-if="!records.length" class="empty-state">
        <div class="empty-state-icon">📋</div>
        <div class="empty-state-text">{{ t('history.noHistory') }}</div>
      </div>
      <div v-else class="history-list">
        <div 
          v-for="r in records" 
          :key="r.id" 
          v-memo="[r.id, r.status, r.download_link]"
          class="history-item slide-up"
        >
          <div class="history-header">
            <div class="history-main">
              <span class="file-name">{{ r.file_name }}</span>
              <span class="file-size">{{ formatSize(r.file_size) }}</span>
              <span :class="['status-badge', r.status]">{{ t('history.status.' + r.status) }}</span>
            </div>
            <div class="history-actions">
              <button v-if="r.download_link" class="btn btn-sm" @click="copyLink(r.download_link)">{{ t('history.copyLink') }}</button>
              <button v-if="r.file_path" class="btn btn-sm btn-primary" @click="regenerateLink(r)">{{ t('history.regenerate') }}</button>
              <button class="btn btn-sm btn-danger" @click="showDeleteConfirm(r.id)">{{ t('history.delete') }}</button>
            </div>
          </div>
          <div class="history-details">
            <div class="detail-row">
              <span class="detail-label">{{ t('history.filePath') }}:</span>
              <span class="detail-value" :title="r.file_path">{{ r.file_path || '-' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">{{ t('history.protocol') }}:</span>
              <span class="detail-value protocol-badge" :class="'protocol-' + (r.protocol || 'http')">{{ r.protocol || 'HTTP' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">{{ t('history.duration') }}:</span>
              <span class="detail-value">{{ formatDuration(r.duration) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">{{ t('history.createdAt') }}:</span>
              <span class="detail-value">{{ formatDate(r.created_at) }}</span>
            </div>
            <div v-if="r.download_link" class="detail-row">
              <span class="detail-label">{{ t('history.downloadLink') }}:</span>
              <span class="detail-value link-value">{{ r.download_link }}</span>
            </div>
          </div>
        </div>
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

    <div v-if="toast" class="toast">{{ toast }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()
const records = ref([])
const loading = ref(true)
const clearing = ref(false)
const autoRefresh = ref(false)
const toast = ref('')
let autoRefreshTimer = null

// 确认对话框状态
const confirmDialog = ref({
  show: false,
  title: '',
  message: '',
  action: null // 确认后要执行的函数
})

// 显示清除确认对话框
const showClearConfirm = () => {
  confirmDialog.value = {
    show: true,
    title: t('history.confirmClearTitle'),
    message: t('history.confirmClear'),
    action: clearHistory
  }
}

// 显示删除确认对话框
const showDeleteConfirm = (id) => {
  confirmDialog.value = {
    show: true,
    title: t('history.confirmDeleteTitle'),
    message: t('history.confirmDelete'),
    action: () => deleteRecord(id)
  }
}

// 取消确认
const cancelConfirm = () => {
  confirmDialog.value.show = false
  confirmDialog.value.action = null
}

// 确认执行
const confirmAction = () => {
  if (confirmDialog.value.action) {
    confirmDialog.value.action()
  }
  confirmDialog.value.show = false
  confirmDialog.value.action = null
}

const load = async () => { 
  loading.value = true
  try { 
    const data = await api.GetHistory(50)
    records.value = data || [] 
  } catch (e) {
    console.error('加载历史记录失败:', e)
  }
  loading.value = false 
}

// 自动刷新功能
const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    autoRefreshTimer = setInterval(() => {
      load()
    }, 5000)
  } else {
    stopAutoRefresh()
  }
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

const clearHistory = async () => {
  clearing.value = true
  try {
    await api.ClearHistory()
    records.value = []
    showToast(t('history.cleared'))
  } catch (e) {
    console.error(e)
  }
  clearing.value = false
}

const deleteRecord = async (id) => {
  try {
    await api.DeleteHistory(id)
    records.value = records.value.filter(r => r.id !== id)
    showToast(t('history.deleted'))
  } catch (e) {
    console.error(e)
  }
}

const copyLink = async (link) => {
  try {
    await navigator.clipboard.writeText(link)
    showToast(t('history.linkCopied'))
  } catch {
    showToast(t('history.copyFailed'))
  }
}

const regenerateLink = async (record) => {
  try {
    // 调用API重新生成链接
    const result = await api.RegenerateLink(record.id)
    if (result && result.link) {
      // 更新本地记录
      const index = records.value.findIndex(r => r.id === record.id)
      if (index !== -1) {
        records.value[index].download_link = result.link
      }
      showToast(t('history.linkRegenerated'))
    }
  } catch (e) {
    console.error('重新生成链接失败:', e)
    showToast(t('history.regenerateFailed'))
  }
}

const showToast = (msg) => {
  toast.value = msg
  setTimeout(() => { toast.value = '' }, 2000)
}

const formatDate = (d) => d ? new Date(d).toLocaleString() : ''

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const s = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + s[i]
}

const formatDuration = (seconds) => {
  if (!seconds || seconds <= 0) return '-'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  if (mins > 0) {
    return `${mins}分${secs}秒`
  }
  return `${secs}秒`
}

onMounted(() => {
  load()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.auto-refresh-row {
  padding: 8px 12px;
  background: var(--bg);
  border-radius: 4px;
  display: flex;
  align-items: center;
}

.auto-refresh-row label {
  cursor: pointer;
  color: var(--text-secondary);
}

.auto-refresh-row input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-item {
  padding: 16px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  transition: all 0.2s;
}

.history-item:hover {
  border-color: var(--primary);
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.history-main {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.file-name {
  font-weight: 600;
  font-size: 14px;
}

.file-size {
  color: var(--text-secondary);
  font-size: 13px;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.completed {
  background: var(--success);
  color: #fff;
}

.status-badge.failed {
  background: var(--danger);
  color: #fff;
}

.status-badge.pending {
  background: var(--warning);
  color: #000;
}

.history-actions {
  display: flex;
  gap: 8px;
}

.history-details {
  padding-top: 12px;
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-row {
  display: flex;
  gap: 8px;
  font-size: 13px;
}

.detail-label {
  color: var(--text-secondary);
  min-width: 80px;
}

.detail-value {
  color: var(--text);
  word-break: break-all;
}

.detail-value.link-value {
  font-family: monospace;
  font-size: 12px;
  color: var(--primary);
}

.protocol-badge {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.protocol-http {
  background: var(--info);
  color: #fff;
}

.protocol-websocket {
  background: var(--success);
  color: #fff;
}

.protocol-udp {
  background: var(--warning);
  color: #000;
}

.protocol-p2p {
  background: var(--primary);
  color: #fff;
}

.toast {
  position: fixed;
  bottom: 32px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--primary);
  color: #fff;
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 13px;
  z-index: 9999;
  animation: toast-in 2s ease;
}

@keyframes toast-in {
  0% { opacity: 0; transform: translateX(-50%) translateY(8px); }
  10% { opacity: 1; transform: translateX(-50%) translateY(0); }
  80% { opacity: 1; }
  100% { opacity: 0; }
}

/* 确认对话框样式 */
.confirm-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.confirm-dialog {
  background: #ffffff;
  border-radius: 8px;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

[data-theme="dark"] .confirm-dialog {
  background: #2d2d2d;
}

.confirm-dialog-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
}

.confirm-dialog-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.confirm-dialog-body {
  padding: 20px;
}

.confirm-dialog-body p {
  margin: 0;
  font-size: 14px;
  color: var(--text);
  line-height: 1.5;
}

.confirm-dialog-footer {
  padding: 12px 20px 16px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  border-top: 1px solid var(--border);
}

.skeleton-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-card {
  height: 120px;
  padding: 16px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.skeleton-card .skeleton-title {
  width: 40%;
  height: 18px;
  margin-bottom: 12px;
}

.skeleton-card .skeleton-text {
  width: 80%;
  height: 14px;
}

</style>
