<template>
  <div class="transfer-tab">
    <div class="card">
      <h2>{{ $t('transfer.selectFile') }} / {{ $t('transfer.selectFolder') }}</h2>

      <div class="file-selection">
        <button class="btn btn-primary" @click="selectFile(false)">
          📁 {{ $t('transfer.selectFile') }}
        </button>
        <button class="btn btn-success" @click="selectFile(true)">
          📂 {{ $t('transfer.selectFolder') }}
        </button>
      </div>

      <div v-if="selectedFiles.length > 0" class="selected-files-list mt-4">
        <div class="flex-between">
          <h3>{{ $t('transfer.selectedFiles') }} ({{ selectedFiles.length }})</h3>
          <button class="btn btn-sm btn-danger" @click="clearAllFiles">{{ $t('transfer.clearAll') }}</button>
        </div>

        <div v-for="(file, index) in selectedFiles" :key="index" class="file-item">
          <div class="file-info">
            <strong>{{ file.name }}</strong>
            <span class="text-secondary">{{ formatFileSize(file.size) }} · {{ file.is_dir ? $t('transfer.folder') : $t('transfer.file') }}</span>
          </div>
          <button class="btn btn-danger btn-sm" @click="removeFile(index)">✕</button>
        </div>

        <div class="generate-actions mt-4">
          <button
            class="btn btn-primary"
            :disabled="generating"
            @click="generateLink"
          >
            {{ generating ? $t('transfer.generating') : `🔗 ${$t('transfer.generateLink')}` }}
          </button>
          <button
            v-if="selectedFiles.length > 1"
            class="btn btn-success"
            :disabled="generating"
            @click="generateBatchLink"
          >
            {{ generating ? $t('transfer.generating') : `📦 ${$t('transfer.batchGenerateLink')}` }}
          </button>
        </div>
      </div>

      <p v-else class="text-secondary text-center mt-4">
        {{ $t('transfer.noFileSelected') }}
      </p>
    </div>

    <div v-for="(link, index) in generatedLinks" :key="index" class="card download-link-card">
      <div class="flex-between">
        <h2>✅ {{ $t('transfer.linkGenerated') }}</h2>
        <button class="btn btn-sm btn-danger" @click="removeLink(index)">✕</button>
      </div>

      <div class="link-display">
        <input
          type="text"
          :value="link.url"
          class="input link-input"
          readonly
        />
        <button class="btn btn-primary" @click="copyToClipboard(link.url)">
          {{ $t('common.copy') }}
        </button>
      </div>

      <div v-if="link.qrCode" class="qr-code-section mt-4">
        <h3>{{ $t('transfer.qrCode') }}</h3>
        <div class="qr-code-container" v-html="link.qrCode"></div>
      </div>

      <div class="file-info-section mt-4">
        <h3>{{ $t('transfer.fileInfo') }}</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="label">{{ $t('transfer.fileName') }}:</span>
            <span>{{ link.fileName }}</span>
          </div>
          <div class="info-item">
            <span class="label">{{ $t('transfer.fileSize') }}:</span>
            <span>{{ formatFileSize(link.fileSize) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="copyToast" class="toast">{{ copyToast }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const selectedFiles = ref([])
const generatedLinks = ref([])
const generating = ref(false)
const copyToast = ref('')

const selectFile = async (isDirectory) => {
  try {
    const files = await api.SelectFiles(isDirectory)
    if (files && files.length > 0) {
      selectedFiles.value.push(...files)
    }
  } catch (error) {
    console.error('选择文件失败:', error)
  }
}

const removeFile = (index) => {
  selectedFiles.value.splice(index, 1)
}

const clearAllFiles = () => {
  selectedFiles.value = []
}

const generateLink = async () => {
  if (selectedFiles.value.length === 0) return
  generating.value = true
  try {
    const file = selectedFiles.value[0]
    const result = await api.GenerateDownloadLink(file.path)
    if (result) {
      generatedLinks.value.unshift({
        url: result.link,
        qrCode: result.qr_code || '',
        fileName: result.file_name,
        fileSize: result.file_size
      })
    }
  } catch (error) {
    console.error('生成链接失败:', error)
  } finally {
    generating.value = false
  }
}

const generateBatchLink = async () => {
  if (selectedFiles.value.length === 0) return
  generating.value = true
  try {
    for (const file of selectedFiles.value) {
      const result = await api.GenerateDownloadLink(file.path)
      if (result) {
        generatedLinks.value.unshift({
          url: result.link,
          qrCode: result.qr_code || '',
          fileName: result.file_name,
          fileSize: result.file_size
        })
      }
    }
  } catch (error) {
    console.error('批量生成链接失败:', error)
  } finally {
    generating.value = false
  }
}

const removeLink = (index) => {
  generatedLinks.value.splice(index, 1)
}

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    showToast(t('common.copied'))
  } catch {
    try {
      const textarea = document.createElement('textarea')
      textarea.value = text
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
      showToast(t('common.copied'))
    } catch {
      showToast(t('common.copyFailed'))
    }
  }
}

const showToast = (msg) => {
  copyToast.value = msg
  setTimeout(() => { copyToast.value = '' }, 2000)
}

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<style scoped>
.file-selection {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

.selected-files-list {
  border-top: 1px solid var(--border-color);
  padding-top: 16px;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background-color: var(--background-color);
  border-radius: 6px;
  margin-bottom: 8px;
}

.file-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.generate-actions {
  display: flex;
  gap: 12px;
}

.download-link-card {
  border-left: 4px solid var(--success-color);
}

.link-display {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

.link-input {
  flex: 1;
  font-family: monospace;
  font-size: 13px;
}

.qr-code-container {
  display: flex;
  justify-content: center;
  padding: 20px;
  background-color: white;
  border-radius: 8px;
}

.qr-code-container img {
  max-width: 200px;
  height: auto;
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

.toast {
  position: fixed;
  bottom: 40px;
  left: 50%;
  transform: translateX(-50%);
  background-color: var(--primary-color);
  color: white;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  z-index: 9999;
  animation: fadeInOut 2s ease;
}

@keyframes fadeInOut {
  0% { opacity: 0; transform: translateX(-50%) translateY(10px); }
  15% { opacity: 1; transform: translateX(-50%) translateY(0); }
  85% { opacity: 1; transform: translateX(-50%) translateY(0); }
  100% { opacity: 0; transform: translateX(-50%) translateY(-10px); }
}
</style>
