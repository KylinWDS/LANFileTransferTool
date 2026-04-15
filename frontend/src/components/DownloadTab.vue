<template>
  <div class="download-tab">
    <div class="card">
      <h2>{{ t('download.title') }}</h2>
      <div class="flex gap-2 mt-3">
        <input v-model="linkInput" class="input" :placeholder="t('download.pasteLinkHere')" />
        <button class="btn btn-primary" :disabled="!linkInput || parsing" @click="parseLink">
          {{ parsing ? t('common.loading') : t('download.parse') }}
        </button>
      </div>
      <p class="text-secondary text-sm mt-2">{{ t('download.linkInputHint') }}</p>
      
      <div class="key-section mt-3">
        <div class="key-input-row">
          <label for="customKeyInput" class="sr-only">{{ t('download.customKeyHint') }}</label>
          <input 
            id="customKeyInput"
            v-model="customKey" 
            class="input key-input" 
            type="password" 
            :placeholder="t('download.customKeyHint')" 
          />
        </div>
        <label class="key-checkbox-row">
          <input 
            type="checkbox" 
            id="useCustomKeyCheckbox"
            v-model="useCustomKey" 
          />
          <span for="useCustomKeyCheckbox">{{ t('download.useCustomKey') }}</span>
        </label>
      </div>

      <div v-if="parsedInfo" class="parsed-card mt-4">
        <div class="flex-between">
          <h3>{{ t('download.fileInfo') }}</h3>
          <button class="btn btn-sm" @click="clearParsedInfo">{{ t('common.clear') }}</button>
        </div>
        <div class="info-row"><span class="label">{{ t('transfer.fileName') }}</span><span>{{ parsedInfo.file_name }}</span></div>
        <div class="info-row"><span class="label">{{ t('transfer.fileSize') }}</span><span>{{ formatSize(parsedInfo.file_size) }}</span></div>
        <!-- 批量文件列表 -->
        <div v-if="parsedInfo.is_batch && parsedInfo.files" class="batch-files mt-3">
          <details>
            <summary class="text-sm text-secondary">{{ t('download.fileList', { count: parsedInfo.file_count }) }}</summary>
            <ul class="file-list mt-2">
              <li v-for="(file, idx) in parsedInfo.files" :key="idx" class="text-sm">
                {{ file.file_name }} ({{ formatSize(file.file_size) }})
              </li>
            </ul>
          </details>
        </div>
        <div class="mt-3">
          <button class="btn btn-primary" @click="startDownload" :disabled="downloading">
            {{ downloading ? t('download.downloading') : t('download.saveAs') }}
          </button>
        </div>
      </div>

      <!-- 下载进度详情 -->
      <div v-if="downloading" class="download-progress mt-4">
        <div class="progress-header">
          <span class="progress-title">{{ t('download.downloading') }}</span>
          <span class="progress-percent">{{ progress.percent.toFixed(1) }}%</span>
        </div>
        <div class="progress-bar-container">
          <div class="progress-bar">
            <div class="progress-bar-fill" :style="{width: progress.percent+'%'}"></div>
          </div>
        </div>
        <div class="progress-details">
          <div v-if="progress.currentFile" class="detail-item current-file">
            <span class="detail-label">{{ t('download.currentFile') }}:</span>
            <span class="detail-value text-ellipsis">{{ progress.currentFile }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">{{ t('download.received') }}:</span>
            <span class="detail-value">{{ parsedInfo?.is_batch ? progress.received + ' / ' + progress.total + ' ' + t('download.files') : formatSize(progress.received) + ' / ' + formatSize(progress.total) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">{{ t('download.speed') }}:</span>
            <span class="detail-value">{{ parsedInfo?.is_batch ? progress.speed.toFixed(2) + ' ' + t('download.filesPerSec') : formatSpeed(progress.speed) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">{{ t('download.timeElapsed') }}:</span>
            <span class="detail-value">{{ formatTime(progress.timeElapsed) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">{{ t('download.timeRemaining') }}:</span>
            <span class="detail-value">{{ formatTime(progress.timeRemaining) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="flex-between mb-2">
        <h2>{{ t('download.availableFiles') }}</h2>
        <div class="flex gap-2">
          <button class="btn btn-sm" @click="selectAll">{{ t('download.selectAll') }}</button>
          <button class="btn btn-sm" @click="selectedIDs = []">{{ t('download.deselectAll') }}</button>
          <button class="btn btn-sm btn-primary" :disabled="!selectedIDs.length" @click="batchDownload">
            {{ t('download.batchDownload') }} ({{ selectedIDs.length }})
          </button>
        </div>
      </div>
      <div v-if="loading" class="skeleton-container">
        <div v-for="i in 3" :key="i" class="skeleton-row skeleton">
          <div class="skeleton skeleton-checkbox"></div>
          <div class="skeleton skeleton-filename"></div>
          <div class="skeleton skeleton-size"></div>
          <div class="skeleton skeleton-btn"></div>
        </div>
      </div>
      <div v-else-if="!files.length" class="empty-state">
        <div class="empty-state-icon">📁</div>
        <div class="empty-state-text">{{ t('download.noFilesAvailable') }}</div>
      </div>
      <div v-else class="table-wrapper">
        <table>
          <thead><tr><th><input type="checkbox" :checked="allSelected" @change="toggleAll" aria-label="全选" /></th><th>{{ t('transfer.fileName') }}</th><th>{{ t('transfer.fileSize') }}</th><th></th></tr></thead>
          <tbody>
            <tr v-for="f in files" :key="f.id">
              <td><input type="checkbox" :value="f.id" v-model="selectedIDs" :aria-label="'选择 ' + f.name" /></td>
              <td>{{ f.name }}</td>
              <td>{{ formatSize(f.size) }}</td>
              <td><button class="btn btn-sm btn-primary" @click="downloadOne(f)">{{ t('download.download') }}</button></td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 批量下载进度 -->
      <div v-if="batchDownloading" class="batch-progress mt-4">
        <div class="progress-header">
          <span class="progress-title">{{ t('download.batchProgress') }} ({{ batchProgress.current }}/{{ batchProgress.total }})</span>
          <span class="progress-percent">{{ batchProgress.percent.toFixed(1) }}%</span>
        </div>
        <div class="progress-bar-container">
          <div class="progress-bar">
            <div class="progress-bar-fill" :style="{width: batchProgress.percent+'%'}"></div>
          </div>
        </div>
        <div class="progress-details">
          <div class="detail-item">
            <span class="detail-label">{{ t('download.currentFile') }}:</span>
            <span class="detail-value">{{ batchProgress.currentFile }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">{{ t('download.totalSize') }}:</span>
            <span class="detail-value">{{ formatSize(batchProgress.totalSize) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">{{ t('download.speed') }}:</span>
            <span class="detail-value">{{ formatSpeed(batchProgress.speed) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Toast 提示 -->
    <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

// Toast 提示
const toastMsg = ref('')
const showToast = (msg) => {
  toastMsg.value = msg
  setTimeout(() => { toastMsg.value = '' }, 2000)
}

const linkInput = ref('')
const parsing = ref(false)
const parsedInfo = ref(null)
const downloading = ref(false)
const loading = ref(true)
const files = ref([])
const selectedIDs = ref([])
const customKey = ref('')
const useCustomKey = ref(false)
const batchDownloading = ref(false)

// 单文件下载进度
const progress = ref({
  percent: 0,
  received: 0,
  total: 0,
  speed: 0,
  timeElapsed: 0,
  timeRemaining: 0,
  startTime: null
})

// 批量下载进度
const batchProgress = ref({
  current: 0,
  total: 0,
  percent: 0,
  currentFile: '',
  totalSize: 0,
  speed: 0,
  startTime: null
})

// 清理解析信息
const clearParsedInfo = () => {
  // 停止下载（如果正在下载）
  if (downloading.value) {
    downloading.value = false
    // 重置进度
    progress.value = {
      percent: 0,
      received: 0,
      total: 0,
      speed: 0,
      timeElapsed: 0,
      timeRemaining: 0,
      startTime: null,
      currentFile: ''
    }
  }
  parsedInfo.value = null
  linkInput.value = ''
  customKey.value = ''
  useCustomKey.value = false
}

const allSelected = computed(() => files.value.length > 0 && selectedIDs.value.length === files.value.length)

const parseLink = async () => {
  if (!linkInput.value) return
  parsing.value = true
  parsedInfo.value = null
  try {
    let token = ''
    const input = linkInput.value.trim()
    try {
      const url = new URL(input)
      const parts = url.pathname.split('/')
      token = parts[parts.length - 1]
    } catch {
      token = input.includes('/') ? input.split('/').pop() : input
    }
    if (!token) { showToast(t('download.invalidLink')); return }

    let info
    if (useCustomKey.value && customKey.value) {
      info = await api.GetDownloadInfoWithKey(token, customKey.value)
    } else {
      info = await api.GetDownloadInfo(token)
    }

    if (info) { info._token = token; parsedInfo.value = info }
    else { showToast(t('download.invalidLink')) }
  } catch (e) {
    if (useCustomKey.value) {
      showToast(t('download.parseFailedWithKey'))
    } else {
      showToast(t('download.parseFailed'))
    }
  }
  parsing.value = false
}

const startDownload = async () => {
  if (!parsedInfo.value) return
  try {
    const saveDir = await api.SelectFiles(true)
    if (!saveDir || !saveDir.length) return
    
    // 检查是否是批量下载
    if (parsedInfo.value.is_batch && parsedInfo.value.files) {
      // 批量下载：逐个下载文件
      await downloadBatchFiles(parsedInfo.value.files, saveDir[0].path)
      return
    }
    
    // 单个文件下载
    // 重置进度
    progress.value = {
      percent: 0,
      received: 0,
      total: parsedInfo.value.file_size || 0,
      speed: 0,
      timeElapsed: 0,
      timeRemaining: 0,
      startTime: Date.now()
    }
    downloading.value = true
    
    // 启动进度更新定时器
    const progressTimer = setInterval(() => {
      updateProgress()
    }, 500)
    
    await api.DownloadFile(parsedInfo.value._token, saveDir[0].path)
    
    clearInterval(progressTimer)
    progress.value.percent = 100
    progress.value.received = progress.value.total
    
    setTimeout(() => { 
      downloading.value = false
      parsedInfo.value = null
      linkInput.value = ''
    }, 1500)
  } catch (e) { 
    console.error(e)
    downloading.value = false
  }
}

// 批量下载：逐个下载文件
const downloadBatchFiles = async (files, saveDir) => {
  downloading.value = true
  const totalFiles = files.length
  let completedFiles = 0
  const startTime = Date.now()
  
  // 初始化进度
  progress.value = {
    percent: 0,
    received: 0,
    total: totalFiles,
    speed: 0,
    timeElapsed: 0,
    timeRemaining: 0,
    startTime: startTime,
    currentFile: ''
  }
  
  // 启动进度更新定时器
  const progressTimer = setInterval(() => {
    const now = Date.now()
    const elapsed = (now - startTime) / 1000
    progress.value.timeElapsed = elapsed
    
    // 计算速度（文件/秒）
    if (elapsed > 0) {
      progress.value.speed = completedFiles / elapsed
      // 估算剩余时间
      const remainingFiles = totalFiles - completedFiles
      if (progress.value.speed > 0) {
        progress.value.timeRemaining = remainingFiles / progress.value.speed
      }
    }
  }, 500)
  
  for (const file of files) {
    // 更新当前文件
    progress.value.currentFile = file.file_name
    
    try {
      // 生成该文件的下载token
      const fileToken = await api.GenerateDownloadLinkForFile(file.file_id)
      if (fileToken && fileToken.token) {
        await api.DownloadFile(fileToken.token, saveDir)
      }
      completedFiles++
      // 更新进度
      progress.value.received = completedFiles
      progress.value.percent = (completedFiles / totalFiles) * 100
    } catch (e) {
      console.error(`下载文件 ${file.file_name} 失败:`, e)
      // 继续下载下一个文件
    }
  }
  
  clearInterval(progressTimer)
  
  // 全部完成
  progress.value.percent = 100
  progress.value.received = totalFiles
  progress.value.currentFile = ''
  
  setTimeout(() => { 
    downloading.value = false
    clearParsedInfo()
  }, 1500)
}

const updateProgress = () => {
  const now = Date.now()
  const elapsed = (now - progress.value.startTime) / 1000
  progress.value.timeElapsed = elapsed
  
  // 模拟进度增长（实际应该通过API获取真实进度）
  if (progress.value.percent < 95) {
    const increment = Math.random() * 3 + 1
    progress.value.percent = Math.min(95, progress.value.percent + increment)
    progress.value.received = Math.floor(progress.value.total * progress.value.percent / 100)
    
    // 计算速度
    if (elapsed > 0) {
      progress.value.speed = progress.value.received / elapsed
      // 计算剩余时间
      const remaining = progress.value.total - progress.value.received
      progress.value.timeRemaining = remaining / progress.value.speed
    }
  }
}

const downloadOne = async (f) => {
  try {
    const saveDir = await api.SelectFiles(true)
    if (!saveDir || !saveDir.length) return
    
    // 重置进度
    progress.value = {
      percent: 0,
      received: 0,
      total: f.size || 0,
      speed: 0,
      timeElapsed: 0,
      timeRemaining: 0,
      startTime: Date.now()
    }
    downloading.value = true
    
    // 启动进度更新定时器
    const progressTimer = setInterval(() => {
      updateProgress()
    }, 500)
    
    await api.DownloadFile(f.id, saveDir[0].path)
    
    clearInterval(progressTimer)
    progress.value.percent = 100
    progress.value.received = progress.value.total
    
    setTimeout(() => { downloading.value = false }, 1500)
  } catch (e) { 
    console.error(e) 
    downloading.value = false
    showToast(t('download.downloadFailed'))
  }
}

const batchDownload = async () => {
  if (!selectedIDs.value.length) return
  try {
    const saveDir = await api.SelectFiles(true)
    if (!saveDir || !saveDir.length) return
    
    // 计算总大小
    const selectedFiles = files.value.filter(f => selectedIDs.value.includes(f.id))
    const totalSize = selectedFiles.reduce((sum, f) => sum + (f.size || 0), 0)
    
    // 重置批量进度
    batchProgress.value = {
      current: 0,
      total: selectedIDs.value.length,
      percent: 0,
      currentFile: selectedFiles[0]?.name || '',
      totalSize: totalSize,
      speed: 0,
      startTime: Date.now()
    }
    batchDownloading.value = true
    
    // 启动进度更新定时器
    const progressTimer = setInterval(() => {
      updateBatchProgress(selectedFiles)
    }, 500)
    
    await api.BatchDownload(selectedIDs.value, saveDir[0].path)
    
    clearInterval(progressTimer)
    batchProgress.value.percent = 100
    batchProgress.value.current = selectedIDs.value.length
    
    setTimeout(() => { 
      batchDownloading.value = false
      selectedIDs.value = []
    }, 1500)
  } catch (e) { 
    console.error(e)
    batchDownloading.value = false
    showToast(t('download.downloadFailed'))
  }
}

const updateBatchProgress = (selectedFiles) => {
  const now = Date.now()
  const elapsed = (now - batchProgress.value.startTime) / 1000
  
  if (batchProgress.value.percent < 95) {
    const increment = Math.random() * 2 + 0.5
    batchProgress.value.percent = Math.min(95, batchProgress.value.percent + increment)
    batchProgress.value.current = Math.floor(batchProgress.value.total * batchProgress.value.percent / 100)
    
    // 更新当前文件名
    const currentIndex = Math.min(batchProgress.value.current, selectedFiles.length - 1)
    batchProgress.value.currentFile = selectedFiles[currentIndex]?.name || ''
    
    // 计算速度
    if (elapsed > 0) {
      const received = batchProgress.value.totalSize * batchProgress.value.percent / 100
      batchProgress.value.speed = received / elapsed
    }
  }
}

const selectAll = () => { selectedIDs.value = files.value.map(f => f.id) }
const toggleAll = (e) => { e.target.checked ? selectAll() : (selectedIDs.value = []) }

const formatSize = (b) => {
  if (!b) return '0 B'
  const k = 1024, s = ['B','KB','MB','GB'], i = Math.floor(Math.log(b)/Math.log(k))
  return (b/Math.pow(k,i)).toFixed(1)+' '+s[i]
}

const formatSpeed = (bytesPerSecond) => {
  if (!bytesPerSecond || bytesPerSecond < 1024) return '0 KB/s'
  return formatSize(bytesPerSecond) + '/s'
}

const formatTime = (seconds) => {
  if (!seconds || seconds < 0) return '--:--'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

onMounted(async () => {
  try { files.value = (await api.GetAvailableFiles()) || [] } catch {}
  loading.value = false
})
</script>

<style scoped>
.key-section {
  padding: 12px;
  background: var(--bg);
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.key-input-row {
  width: 100%;
}

.key-input {
  width: 100%;
}

.key-checkbox-row {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
}

.key-checkbox-row input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}
.parsed-card {
  padding: 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  border-left: 3px solid var(--success);
  animation: slideIn 0.3s ease;
}
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
.info-row { display: flex; justify-content: space-between; padding: 4px 0; font-size: 13px; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; text-align: left; border-bottom: 1px solid var(--border); font-size: 13px; }
th { color: var(--text-secondary); font-weight: 500; }

/* 响应式表格 */
.table-wrapper {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

table {
  min-width: 500px;
}

@media (max-width: 640px) {
  table { font-size: 12px; }
  th, td { padding: 6px 4px; }
}

/* 进度条样式 */
.download-progress, .batch-progress {
  padding: 16px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.progress-title {
  font-weight: 600;
  color: var(--text);
}

.progress-percent {
  font-size: 18px;
  font-weight: 700;
  color: var(--primary);
}

.progress-bar-container {
  margin-bottom: 16px;
}

.progress-bar {
  height: 8px;
  background: var(--border);
  border-radius: 4px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--primary), var(--success));
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 8px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 13px;
}

.detail-label {
  color: var(--text-secondary);
}

.detail-value {
  font-weight: 500;
  color: var(--text);
}

.detail-item.current-file {
  grid-column: 1 / -1;
  background: var(--bg-secondary);
  padding: 8px 12px;
  border-radius: 4px;
  margin-bottom: 8px;
}

.detail-item.current-file .detail-value {
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.text-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Toast 提示样式 */
.toast {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--text);
  color: var(--bg);
  padding: 12px 24px;
  border-radius: 6px;
  font-size: 14px;
  z-index: 1000;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

.skeleton-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px;
  background: var(--bg);
  border-radius: 6px;
}

.skeleton-checkbox {
  width: 16px;
  height: 16px;
  border-radius: 3px;
}

.skeleton-filename {
  flex: 1;
  height: 14px;
}

.skeleton-size {
  width: 80px;
  height: 14px;
}

.skeleton-btn {
  width: 60px;
  height: 28px;
  border-radius: 4px;
}

</style>
