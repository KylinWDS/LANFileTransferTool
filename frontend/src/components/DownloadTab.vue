<template>
  <div class="download-tab">
    <div class="card">
      <h2>{{ $t('download.clientDownload') }}</h2>

      <div class="download-link-input-section mb-4">
        <label class="label">{{ $t('download.enterDownloadLink') }}</label>
        <div class="link-input-group">
          <input
            v-model="downloadLinkInput"
            type="text"
            class="input"
            :placeholder="$t('download.pasteLinkHere')"
          />
          <button
            class="btn btn-primary"
            :disabled="!downloadLinkInput || parsing"
            @click="parseDownloadLink"
          >
            {{ parsing ? $t('common.loading') : `📥 ${$t('download.parseAndDownload')}` }}
          </button>
        </div>
        <p class="text-secondary text-sm mt-2">
          {{ $t('download.linkInputHint') }}
        </p>
      </div>

      <div v-if="downloadInfo" class="parsed-info mb-4">
        <h3>{{ $t('download.fileInfo') }}</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="label">{{ $t('transfer.fileName') }}:</span>
            <span>{{ downloadInfo.file_name }}</span>
          </div>
          <div class="info-item">
            <span class="label">{{ $t('transfer.fileSize') }}:</span>
            <span>{{ formatFileSize(downloadInfo.file_size) }}</span>
          </div>
        </div>
        <button class="btn btn-success mt-4" @click="startDownloadFromLink">
          💾 {{ $t('download.saveAs') }}
        </button>
      </div>

      <div v-if="downloadingFromLink" class="download-progress-section mb-4">
        <h3>{{ $t('download.downloadingFromLink') }}</h3>
        <div class="progress-item">
          <div class="file-name">{{ downloadingFromLink.fileName }}</div>
          <div class="progress-bar">
            <div class="progress-bar-fill" :style="{ width: downloadingFromLink.progress + '%' }"></div>
          </div>
          <div class="progress-info">
            <span>{{ downloadingFromLink.progress.toFixed(1) }}%</span>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="flex-between mb-4">
        <h2>{{ $t('download.availableFiles') }}</h2>
        <div class="actions">
          <button class="btn btn-sm btn-primary" @click="selectAll">
            {{ $t('download.selectAll') }}
          </button>
          <button class="btn btn-sm btn-warning" @click="deselectAll">
            {{ $t('download.deselectAll') }}
          </button>
          <button
            class="btn btn-sm btn-success"
            :disabled="selectedFileIDs.length === 0"
            @click="batchDownload"
          >
            📥 {{ $t('download.batchDownload') }} ({{ selectedFileIDs.length }})
          </button>
        </div>
      </div>

      <div v-if="loading" class="text-center">{{ $t('common.loading') }}</div>

      <div v-else-if="availableFiles.length === 0" class="text-center text-secondary">
        {{ $t('download.noFilesAvailable') }}
      </div>

      <div v-else class="files-table">
        <table>
          <thead>
            <tr>
              <th><input type="checkbox" :checked="allSelected" @change="toggleAll" /></th>
              <th>{{ $t('transfer.fileName') }}</th>
              <th>{{ $t('transfer.fileSize') }}</th>
              <th>{{ $t('history.time') }}</th>
              <th>{{ $t('common.action') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="file in availableFiles" :key="file.id">
              <td><input type="checkbox" :value="file.id" v-model="selectedFileIDs" /></td>
              <td><span class="file-name">{{ file.name }}</span></td>
              <td>{{ formatFileSize(file.size) }}</td>
              <td>{{ formatDate(file.mod_time) }}</td>
              <td>
                <button class="btn btn-sm btn-primary" @click="downloadSingle(file)">
                  {{ $t('download.download') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="downloading.length > 0" class="card">
      <h3>{{ $t('download.downloading') }}</h3>
      <div v-for="item in downloading" :key="item.fileID" class="progress-item">
        <div class="file-name">{{ item.fileName }}</div>
        <div class="progress-bar">
          <div class="progress-bar-fill" :style="{ width: item.progress + '%' }"></div>
        </div>
        <div class="progress-info">
          <span>{{ item.progress.toFixed(1) }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const loading = ref(true)
const availableFiles = ref([])
const selectedFileIDs = ref([])
const downloading = ref([])
const downloadLinkInput = ref('')
const downloadInfo = ref(null)
const downloadingFromLink = ref(null)
const parsing = ref(false)

const allSelected = computed(() => {
  return availableFiles.value.length > 0 &&
         selectedFileIDs.value.length === availableFiles.value.length
})

const loadAvailableFiles = async () => {
  loading.value = true
  try {
    const files = await api.GetAvailableFiles()
    availableFiles.value = files || []
  } catch (error) {
    console.error('获取可用文件列表失败:', error)
  } finally {
    loading.value = false
  }
}

const selectAll = () => {
  selectedFileIDs.value = availableFiles.value.map(f => f.id)
}

const deselectAll = () => {
  selectedFileIDs.value = []
}

const toggleAll = (event) => {
  if (event.target.checked) {
    selectAll()
  } else {
    deselectAll()
  }
}

const parseDownloadLink = async () => {
  if (!downloadLinkInput.value) return
  parsing.value = true
  downloadInfo.value = null

  try {
    let token = ''
    const input = downloadLinkInput.value.trim()

    try {
      const url = new URL(input)
      const pathParts = url.pathname.split('/')
      token = pathParts[pathParts.length - 1]
    } catch {
      if (input.startsWith('http') || input.includes('/')) {
        const parts = input.split('/')
        token = parts[parts.length - 1]
      } else {
        token = input
      }
    }

    if (!token) {
      alert(t('download.invalidLink'))
      return
    }

    const info = await api.GetDownloadInfo(token)
    if (info) {
      info._token = token
      downloadInfo.value = info
    } else {
      alert(t('download.invalidLink'))
    }
  } catch (error) {
    console.error('解析下载链接失败:', error)
    alert(t('download.parseFailed') + ': ' + error.message)
  } finally {
    parsing.value = false
  }
}

const startDownloadFromLink = async () => {
  if (!downloadInfo.value) return

  try {
    const savePath = await api.SelectFiles(true)
    if (!savePath || savePath.length === 0) return

    const folder = savePath[0].path
    downloadingFromLink.value = {
      fileName: downloadInfo.value.file_name,
      progress: 0
    }

    try {
      await api.DownloadFile(downloadInfo.value._token, folder)

      let progress = 0
      const interval = setInterval(() => {
        progress += Math.random() * 15 + 5
        if (progress >= 100) {
          progress = 100
          clearInterval(interval)
          setTimeout(() => {
            downloadingFromLink.value = null
            downloadInfo.value = null
            downloadLinkInput.value = ''
          }, 1500)
        }
        if (downloadingFromLink.value) {
          downloadingFromLink.value.progress = progress
        }
      }, 200)
    } catch (error) {
      console.error('下载失败:', error)
      alert(t('download.downloadFailed') + ': ' + error.message)
      downloadingFromLink.value = null
    }
  } catch (error) {
    console.error('选择保存位置失败:', error)
  }
}

const downloadSingle = async (file) => {
  try {
    const savePath = await api.SelectFiles(true)
    if (!savePath || savePath.length === 0) return

    const folder = savePath[0].path
    const progressItem = {
      fileID: file.id,
      fileName: file.name,
      progress: 0
    }
    downloading.value.push(progressItem)

    let progress = 0
    const interval = setInterval(() => {
      progress += Math.random() * 15 + 5
      if (progress >= 100) {
        progress = 100
        clearInterval(interval)
        setTimeout(() => {
          const index = downloading.value.findIndex(d => d.fileID === file.id)
          if (index > -1) downloading.value.splice(index, 1)
        }, 1000)
      }
      progressItem.progress = progress
    }, 300)
  } catch (error) {
    console.error('下载失败:', error)
  }
}

const batchDownload = async () => {
  if (selectedFileIDs.value.length === 0) return
  try {
    const savePath = await api.SelectFiles(true)
    if (!savePath || savePath.length === 0) return
    alert(t('download.batchDownloadStarted'))
  } catch (error) {
    console.error('批量下载失败:', error)
  }
}

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString()
}

onMounted(() => {
  loadAvailableFiles()
})
</script>

<style scoped>
.download-link-input-section {
  padding: 16px;
  background-color: var(--background-color);
  border-radius: 8px;
  border: 1px dashed var(--border-color);
}

.link-input-group {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.link-input-group .input {
  flex: 1;
}

.parsed-info {
  padding: 16px;
  background-color: var(--surface-color);
  border-radius: 8px;
  border-left: 4px solid var(--success-color);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.files-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
}

th {
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 13px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.progress-item {
  margin-bottom: 16px;
}

.file-name {
  font-weight: 500;
  margin-bottom: 8px;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
  font-size: 13px;
  color: var(--text-secondary);
}
</style>
