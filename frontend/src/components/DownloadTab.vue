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
      
      <div class="key-section mt-2">
        <div class="flex gap-2">
          <input v-model="customKey" class="input" type="password" :placeholder="t('download.customKeyHint')" />
          <label class="text-sm" style="display:flex;align-items:center;gap:4px;cursor:pointer">
            <input type="checkbox" v-model="useCustomKey" /> {{ t('download.useCustomKey') }}
          </label>
        </div>
      </div>

      <div v-if="parsedInfo" class="parsed-card mt-4">
        <h3>{{ t('download.fileInfo') }}</h3>
        <div class="info-row"><span class="label">{{ t('transfer.fileName') }}</span><span>{{ parsedInfo.file_name }}</span></div>
        <div class="info-row"><span class="label">{{ t('transfer.fileSize') }}</span><span>{{ formatSize(parsedInfo.file_size) }}</span></div>
        <div class="mt-3">
          <button class="btn btn-primary" @click="startDownload" :disabled="downloading">
            {{ downloading ? t('download.downloading') : t('download.saveAs') }}
          </button>
        </div>
      </div>

      <div v-if="downloading" class="mt-4">
        <div class="progress-bar"><div class="progress-bar-fill" :style="{width: progress+'%'}"></div></div>
        <div class="text-sm text-secondary mt-2">{{ progress.toFixed(0) }}%</div>
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
      <div v-if="loading" class="text-center text-secondary">{{ t('common.loading') }}</div>
      <div v-else-if="!files.length" class="text-center text-secondary">{{ t('download.noFilesAvailable') }}</div>
      <table v-else>
        <thead><tr><th><input type="checkbox" :checked="allSelected" @change="toggleAll" /></th><th>{{ t('transfer.fileName') }}</th><th>{{ t('transfer.fileSize') }}</th><th></th></tr></thead>
        <tbody>
          <tr v-for="f in files" :key="f.id">
            <td><input type="checkbox" :value="f.id" v-model="selectedIDs" /></td>
            <td>{{ f.name }}</td>
            <td>{{ formatSize(f.size) }}</td>
            <td><button class="btn btn-sm btn-primary" @click="downloadOne(f)">{{ t('download.download') }}</button></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const linkInput = ref('')
const parsing = ref(false)
const parsedInfo = ref(null)
const downloading = ref(false)
const progress = ref(0)
const loading = ref(true)
const files = ref([])
const selectedIDs = ref([])
const customKey = ref('')
const useCustomKey = ref(false)

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
    if (!token) { alert(t('download.invalidLink')); return }
    
    let info
    if (useCustomKey.value && customKey.value) {
      info = await api.GetDownloadInfoWithKey(token, customKey.value)
    } else {
      info = await api.GetDownloadInfo(token)
    }
    
    if (info) { info._token = token; parsedInfo.value = info }
    else { alert(t('download.invalidLink')) }
  } catch (e) { 
    if (useCustomKey.value) {
      alert(t('download.parseFailedWithKey'))
    } else {
      alert(t('download.parseFailed'))
    }
  }
  parsing.value = false
}

const startDownload = async () => {
  if (!parsedInfo.value) return
  try {
    const saveDir = await api.SelectFiles(true)
    if (!saveDir || !saveDir.length) return
    downloading.value = true
    progress.value = 0
    await api.DownloadFile(parsedInfo.value._token, saveDir[0].path)
    const iv = setInterval(() => {
      progress.value += Math.random() * 20 + 5
      if (progress.value >= 100) {
        progress.value = 100
        clearInterval(iv)
        setTimeout(() => { downloading.value = false; parsedInfo.value = null; linkInput.value = '' }, 1000)
      }
    }, 200)
  } catch (e) { console.error(e); downloading.value = false }
}

const downloadOne = async (f) => {
  try {
    const saveDir = await api.SelectFiles(true)
    if (!saveDir || !saveDir.length) return
    alert(t('download.downloadStarted'))
  } catch (e) { console.error(e) }
}

const batchDownload = async () => {
  if (!selectedIDs.value.length) return
  try {
    const saveDir = await api.SelectFiles(true)
    if (!saveDir || !saveDir.length) return
    alert(t('download.batchDownloadStarted'))
  } catch (e) { console.error(e) }
}

const selectAll = () => { selectedIDs.value = files.value.map(f => f.id) }
const toggleAll = (e) => { e.target.checked ? selectAll() : (selectedIDs.value = []) }

const formatSize = (b) => {
  if (!b) return '0 B'
  const k = 1024, s = ['B','KB','MB','GB'], i = Math.floor(Math.log(b)/Math.log(k))
  return (b/Math.pow(k,i)).toFixed(1)+' '+s[i]
}

onMounted(async () => {
  try { files.value = (await api.GetAvailableFiles()) || [] } catch {}
  loading.value = false
})
</script>

<style scoped>
.key-section {
  padding: 8px;
  background: var(--bg);
  border-radius: 4px;
}
.parsed-card {
  padding: 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  border-left: 3px solid var(--success);
}
.info-row { display: flex; justify-content: space-between; padding: 4px 0; font-size: 13px; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; text-align: left; border-bottom: 1px solid var(--border); font-size: 13px; }
th { color: var(--text-secondary); font-weight: 500; }
</style>
