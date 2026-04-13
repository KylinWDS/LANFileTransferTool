<template>
  <div class="transfer-tab">
    <div class="card">
      <div class="flex-between mb-2">
        <h2>{{ t('transfer.title') }}</h2>
        <div class="ip-selector">
          <label class="text-sm text-secondary">{{ t('transfer.selectIP') }}:</label>
          <select v-model="selectedIP" @change="onIPChange" class="select">
            <option v-for="ip in availableIPs" :key="ip.ip" :value="ip.ip">
              {{ ip.ip }} ({{ getIPScopeLabel(ip.scope) }})
            </option>
          </select>
        </div>
      </div>
      
      <div class="manual-ip-section mt-2">
        <div class="flex gap-2">
          <input v-model="manualIP" class="input" :placeholder="t('transfer.manualIPHint')" />
          <button class="btn btn-sm" @click="setManualIP">{{ t('transfer.manualIP') }}</button>
        </div>
      </div>
      
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

    <div class="card">
      <div class="flex-between mb-2">
        <h3>{{ t('transfer.discoveredPeers') }}</h3>
        <button class="btn btn-sm" @click="discoverPeers">{{ t('transfer.refreshPeers') }}</button>
      </div>
      <div v-if="discovering" class="text-center text-secondary">{{ t('common.loading') }}</div>
      <div v-else-if="!discoveredPeers.length" class="text-center text-secondary">{{ t('transfer.noPeers') }}</div>
      <div v-else class="peer-list">
        <div v-for="peer in discoveredPeers" :key="peer.ip + ':' + peer.port" class="peer-item" @click="connectToPeer(peer)">
          <div class="peer-info">
            <span class="peer-name">{{ peer.name || 'Unknown' }}</span>
            <span class="text-secondary text-sm">{{ peer.ip }}:{{ peer.port }}</span>
          </div>
          <span class="peer-status status-ok">{{ t('transfer.connect') }}</span>
        </div>
      </div>
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
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import api from '../api'

const { t } = useI18n()

const selectedFiles = ref([])
const generatedLinks = ref([])
const generating = ref(false)
const toast = ref('')
const availableIPs = ref([])
const selectedIP = ref('')
const manualIP = ref('')
const discoveredPeers = ref([])
const discovering = ref(false)
let peerTimer = null

const scopeLabels = {
  'private-a': 'A类私有',
  'private-b': 'B类私有',
  'private-c': 'C类私有',
  'link-local': '链路本地',
  'multicast': '组播',
  'public': '公网',
  'loopback': '本地回环'
}

const getIPScopeLabel = (scope) => scopeLabels[scope] || scope

const loadIPs = async () => {
  try {
    const ips = await api.GetAllIPs()
    availableIPs.value = ips || []
    const savedIP = await api.GetSelectedIP()
    if (savedIP) {
      selectedIP.value = savedIP
    } else if (ips && ips.length > 0) {
      selectedIP.value = ips[0].ip
    }
  } catch (e) { console.error(e) }
}

const onIPChange = async () => {
  try {
    await api.SetSelectedIP(selectedIP.value)
    showToast(t('transfer.ipChanged'))
  } catch (e) { console.error(e) }
}

const setManualIP = async () => {
  if (!manualIP.value) return
  try {
    await api.SetManualIP(manualIP.value)
    selectedIP.value = manualIP.value
    showToast(t('transfer.ipChanged'))
    manualIP.value = ''
  } catch (e) {
    showToast(t('transfer.invalidIP'))
  }
}

const discoverPeers = async () => {
  discovering.value = true
  try {
    const peers = await api.DiscoverPeers()
    discoveredPeers.value = peers || []
  } catch (e) { console.error(e) }
  discovering.value = false
}

const connectToPeer = (peer) => {
  const url = `http://${peer.ip}:${peer.port}`
  window.open(url, '_blank')
}

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

onMounted(() => {
  loadIPs()
  discoverPeers()
  peerTimer = setInterval(discoverPeers, 10000)
})

onUnmounted(() => {
  if (peerTimer) clearInterval(peerTimer)
})
</script>

<style scoped>
.ip-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}
.select {
  padding: 4px 8px;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--text);
  font-size: 13px;
}
.manual-ip-section {
  padding: 8px;
  background: var(--bg);
  border-radius: 4px;
}
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
.peer-list { display: flex; flex-direction: column; gap: 8px; }
.peer-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}
.peer-item:hover { border-color: var(--primary); }
.peer-info { display: flex; flex-direction: column; }
.peer-name { font-weight: 500; }
.peer-status { font-size: 12px; padding: 2px 8px; border-radius: 4px; }
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
