<template>
  <div class="transfer-tab">
    <div class="card">
      <h2>{{ t('transfer.title') }}</h2>
      <div class="flex gap-2 mt-3">
        <button class="btn btn-primary" @click="selectFile(false)">{{ t('transfer.selectFile') }}</button>
        <button class="btn" @click="selectFile(true)">{{ t('transfer.selectFolder') }}</button>
        <button v-if="selectedFiles.length" class="btn" @click="selectedFiles = []">{{ t('transfer.clearAll') }}</button>
      </div>

      <div v-if="selectedFiles.length" class="mt-4">
        <div v-for="(file, i) in selectedFiles" :key="i" class="file-item">
          <div>
            <span class="file-name">{{ file.name }}</span>
            <span class="text-secondary text-sm">{{ formatSize(file.size) }} · {{ file.is_dir ? t('transfer.folder') : t('transfer.file') }}</span>
          </div>
          <button class="btn btn-sm" @click="selectedFiles.splice(i, 1)">✕</button>
        </div>
        <div class="flex gap-2 mt-4">
          <button class="btn btn-primary" :disabled="generating" @click="generateLink">
            {{ generating ? t('transfer.generating') : t('transfer.generateLink') }}
          </button>
          <button v-if="selectedFiles.length > 1" class="btn" :disabled="generating" @click="generateBatchLink">
            {{ t('transfer.batchGenerateLink') }}
          </button>
        </div>
      </div>
      <p v-else class="text-secondary text-center mt-4">{{ t('transfer.noFileSelected') }}</p>
    </div>

    <div v-for="(link, i) in generatedLinks" :key="i" class="card link-card">
      <div class="flex-between mb-2">
        <h3>{{ t('transfer.linkGenerated') }}</h3>
        <button class="btn btn-sm" @click="generatedLinks.splice(i, 1)">✕</button>
      </div>
      <div class="flex gap-2">
        <input :value="link.url" class="input" readonly />
        <button class="btn btn-primary btn-sm" @click="copyLink(link.url)">{{ t('common.copy') }}</button>
      </div>
      <div class="link-meta mt-2 text-sm text-secondary">
        {{ link.fileName }} · {{ formatSize(link.fileSize) }}
      </div>
      <div v-if="link.qrDataUrl" class="mt-3 text-center">
        <img :src="link.qrDataUrl" alt="QR" style="width:160px;height:160px;border-radius:6px;" />
      </div>
    </div>

    <div v-if="toast" class="toast">{{ toast }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import api from '../api'

const { t } = useI18n()

const selectedFiles = ref([])
const generatedLinks = ref([])
const generating = ref(false)
const toast = ref('')

const selectFile = async (isDir) => {
  try {
    const files = await api.SelectFiles(isDir)
    if (files && files.length) selectedFiles.value.push(...files)
  } catch (e) { console.error(e) }
}

const generateLink = async () => {
  if (!selectedFiles.value.length) return
  generating.value = true
  try {
    const file = selectedFiles.value[0]
    const result = await api.GenerateDownloadLink(file.path)
    if (result) {
      const qrDataUrl = await QRCode.toDataURL(result.link, { width: 200, margin: 1 })
      generatedLinks.value.unshift({
        url: result.link,
        qrDataUrl,
        fileName: result.file_name,
        fileSize: result.file_size
      })
    }
  } catch (e) { console.error(e) }
  generating.value = false
}

const generateBatchLink = async () => {
  generating.value = true
  try {
    for (const file of selectedFiles.value) {
      const result = await api.GenerateDownloadLink(file.path)
      if (result) {
        const qrDataUrl = await QRCode.toDataURL(result.link, { width: 200, margin: 1 })
        generatedLinks.value.unshift({
          url: result.link,
          qrDataUrl,
          fileName: result.file_name,
          fileSize: result.file_size
        })
      }
    }
  } catch (e) { console.error(e) }
  generating.value = false
}

const copyLink = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    showToast(t('common.copied'))
  } catch {
    showToast(t('common.copyFailed'))
  }
}

const showToast = (msg) => {
  toast.value = msg
  setTimeout(() => { toast.value = '' }, 2000)
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const s = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + s[i]
}
</script>

<style scoped>
.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  margin-bottom: 6px;
}
.file-name { font-weight: 500; margin-right: 8px; }
.link-card { border-left: 3px solid var(--primary); }
.link-meta { padding: 4px 0; }
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
</style>
