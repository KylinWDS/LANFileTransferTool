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
            <span class="text-secondary text-sm">
              {{ formatSize(file.size) }} · 
              {{ file.is_dir ? t('transfer.folder') : t('transfer.file') }}
              <span v-if="file.download_count !== undefined" class="download-count">
                · {{ t('transfer.downloadCount', { count: file.download_count }) }}
              </span>
            </span>
          </div>
          <button 
            class="btn btn-sm btn-icon" 
            @click="selectedFiles.splice(i, 1)"
            :aria-label="t('accessibility.removeFile', { name: file.name })"
            :title="t('accessibility.remove')"
          >✕</button>
        </div>
        <div class="flex gap-2 mt-4">
          <button v-if="selectedFiles.length === 1" class="btn btn-primary" :disabled="generating" @click="generateLink">
            {{ generating ? t('transfer.generating') : t('transfer.generateLink') }}
          </button>
          <button v-if="selectedFiles.length > 1" class="btn btn-primary" :disabled="generating" @click="generateBatchLink">
            {{ generating ? t('transfer.generating') : t('transfer.batchGenerateLink') }}
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
      <div v-if="discovering" class="skeleton-container">
        <div v-for="i in 2" :key="i" class="skeleton-peer skeleton">
          <div class="skeleton skeleton-avatar"></div>
          <div class="skeleton-content">
            <div class="skeleton skeleton-text"></div>
            <div class="skeleton skeleton-text-sm"></div>
          </div>
        </div>
      </div>
      <div v-else-if="!discoveredPeers.length" class="empty-state">
        <div class="empty-state-icon">🔍</div>
        <div class="empty-state-text">{{ t('transfer.noPeers') }}</div>
      </div>
      <div v-else class="peer-list">
        <div 
          v-for="peer in discoveredPeers" 
          :key="peer.ip + ':' + peer.port" 
          v-memo="[peer.ip, peer.port, peer.name]"
          class="peer-item hover-lift" 
          @click="connectToPeer(peer)"
        >
          <div class="peer-info">
            <span class="peer-name">{{ peer.name || 'Unknown' }}</span>
            <span class="text-secondary text-sm">{{ peer.ip }}:{{ peer.port }}</span>
          </div>
          <span class="peer-status status-ok">{{ t('transfer.connect') }}</span>
        </div>
      </div>
    </div>

    <div v-for="(link, i) in generatedLinks" :key="i" class="card link-card">
      <div class="link-header">
        <h3>{{ t('transfer.linkGenerated') }}</h3>
        <button 
          class="btn-close" 
          @click="generatedLinks.splice(i, 1)" 
          :aria-label="t('accessibility.closeLinkCard')"
          :title="t('common.close')"
        >
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
          </svg>
        </button>
      </div>
      <div class="flex gap-2">
        <input :value="link.url" class="input" readonly />
        <button class="btn btn-primary btn-sm" @click="copyLink(link.url)">{{ t('common.copy') }}</button>
      </div>
      <div class="link-meta mt-2 text-sm text-secondary">
        {{ link.fileName }} · {{ formatSize(link.fileSize) }}
        <span v-if="link.isBatch" class="batch-badge">{{ t('transfer.batchLabel') }}</span>
      </div>
      <!-- 批量文件列表 -->
      <div v-if="link.isBatch && link.files" class="batch-files-list mt-2">
        <details>
          <summary class="text-sm text-secondary">{{ t('transfer.fileList', { count: link.fileCount }) }}</summary>
          <ul class="file-list mt-1">
            <li v-for="(file, idx) in link.files" :key="idx" class="text-xs text-secondary">
              {{ file.file_name }} ({{ formatSize(file.file_size) }})
            </li>
          </ul>
        </details>
      </div>
      <!-- 协议信息 -->
      <div class="protocol-info mt-2">
        <div class="flex-between">
          <span class="text-sm text-secondary">{{ t('transfer.protocol') }}:</span>
          <span class="protocol-badge" :class="'protocol-' + link.protocol">
            {{ getProtocolName(link.protocol) }}
          </span>
        </div>
        <div v-if="link.recommendedProtocol && link.recommendedProtocol !== link.protocol" class="text-sm text-secondary mt-1">
          {{ t('transfer.recommended') }}: {{ getProtocolName(link.recommendedProtocol) }}
        </div>
      </div>
      <!-- 密钥信息 -->
      <div class="secret-key-info mt-3" :class="{ 'is-default': link.isDefaultKey }">
        <div class="flex-between">
          <span class="text-sm">{{ t('transfer.secretKey') }}:</span>
          <span v-if="link.isDefaultKey" class="badge badge-warning">{{ t('transfer.defaultKey') }}</span>
          <span v-else class="badge badge-success">{{ t('transfer.customKey') }}</span>
        </div>
        <div class="secret-key-value mt-1">
          <code class="key-code">{{ link.secretKey }}</code>
          <button class="btn btn-xs" @click="copySecretKey(link.secretKey)">{{ t('common.copy') }}</button>
        </div>
        <p class="text-xs text-secondary mt-1">{{ t('transfer.secretKeyHint') }}</p>
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
const protocolRecommendation = ref(null)
let peerTimer = null

const protocolNames = {
  'http': 'HTTP',
  'websocket': 'WebSocket',
  'udp': 'UDP',
  'p2p': 'P2P'
}

const getProtocolName = (type) => protocolNames[type] || type

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

// 生成单个文件链接
const generateLink = async () => {
  if (selectedFiles.value.length !== 1) return
  generating.value = true
  try {
    const file = selectedFiles.value[0]
    
    // 检查是否已存在该文件的链接卡片
    const existingIndex = generatedLinks.value.findIndex(link => link.fileName === file.name)
    if (existingIndex !== -1) {
      generatedLinks.value.splice(existingIndex, 1)
    }
    
    // 获取协议推荐
    const protoRec = await api.GetProtocolRecommendation(file.size)
    protocolRecommendation.value = protoRec
    
    // 选择协议
    const selectedProtocol = await api.SelectProtocol(file.size, discoveredPeers.value.length > 0, 'auto')
    
    // 生成下载链接
    const result = await api.GenerateDownloadLink(file.path)
    
    // 获取密钥信息
    const securityInfo = await api.GetSecurityInfo()
    
    if (result) {
      const qrDataUrl = await QRCode.toDataURL(result.link, { width: 200, margin: 1 })
      
      generatedLinks.value.unshift({
        url: result.link,
        qrDataUrl,
        fileName: result.file_name,
        fileSize: result.file_size,
        isBatch: false,
        protocol: selectedProtocol,
        recommendedProtocol: protoRec?.recommended,
        secretKey: securityInfo?.secret_key || '',
        isDefaultKey: securityInfo?.is_default || false
      })
    }
  } catch (e) { console.error(e) }
  generating.value = false
}

// 批量生成链接（多个文件一个链接）
const generateBatchLink = async () => {
  if (selectedFiles.value.length <= 1) return
  generating.value = true
  try {
    // 构建唯一标识（用于检查是否已存在）
    const fileNames = selectedFiles.value.map(f => f.name).sort().join(',')
    const existingIndex = generatedLinks.value.findIndex(link => link.fileNames === fileNames)
    if (existingIndex !== -1) {
      generatedLinks.value.splice(existingIndex, 1)
    }
    
    // 计算总大小
    const totalSize = selectedFiles.value.reduce((sum, f) => sum + f.size, 0)
    
    // 获取协议推荐
    const protoRec = await api.GetProtocolRecommendation(totalSize)
    protocolRecommendation.value = protoRec
    
    // 选择协议
    const selectedProtocol = await api.SelectProtocol(totalSize, discoveredPeers.value.length > 0, 'auto')
    
    // 生成批量下载链接
    const filePaths = selectedFiles.value.map(f => f.path)
    const result = await api.GenerateBatchDownloadLink(filePaths)
    
    // 获取密钥信息
    const securityInfo = await api.GetSecurityInfo()
    
    if (result) {
      const qrDataUrl = await QRCode.toDataURL(result.link, { width: 200, margin: 1 })
      
      generatedLinks.value.unshift({
        url: result.link,
        qrDataUrl,
        fileName: t('transfer.batchFiles', { count: result.file_count }),
        fileNames: fileNames,
        fileSize: result.total_size,
        isBatch: true,
        fileCount: result.file_count,
        files: result.files,
        protocol: selectedProtocol,
        recommendedProtocol: protoRec?.recommended,
        secretKey: securityInfo?.secret_key || '',
        isDefaultKey: securityInfo?.is_default || false
      })
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

const copySecretKey = async (key) => {
  try {
    await navigator.clipboard.writeText(key)
    showToast(t('transfer.secretKeyCopied'))
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
  transition: all var(--transition-fast);
}
.file-item:hover {
  border-color: var(--primary);
  box-shadow: var(--shadow);
}
.file-name { font-weight: 500; margin-right: 8px; }
.download-count {
  color: var(--primary);
  font-weight: 500;
}
.link-card { 
  border-left: 3px solid var(--primary); 
  position: relative; 
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
.link-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.btn-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}
.btn-close:hover {
  background: var(--bg);
  color: var(--danger);
}
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
.protocol-info { padding: 8px; background: var(--bg); border-radius: 4px; }
.protocol-badge { font-size: 12px; padding: 2px 8px; border-radius: 4px; font-weight: 500; }
.protocol-http { background: var(--info); color: #fff; }
.protocol-websocket { background: var(--success); color: #fff; }
.protocol-udp { background: var(--warning); color: #000; }
.protocol-p2p { background: var(--primary); color: #fff; }

/* 密钥信息样式 */
.secret-key-info {
  padding: 12px;
  background: var(--bg);
  border-radius: 4px;
  border-left: 3px solid var(--primary);
}
.secret-key-info.is-default {
  border-left-color: var(--warning);
}
.secret-key-value {
  display: flex;
  align-items: center;
  gap: 8px;
}
.key-code {
  flex: 1;
  padding: 6px 10px;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
  font-size: 12px;
  font-family: monospace;
  word-break: break-all;
}
[data-theme="dark"] .key-code {
  background: rgba(255, 255, 255, 0.1);
}
.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}
.badge-warning {
  background: var(--warning);
  color: #000;
}
.badge-success {
  background: var(--success);
  color: #fff;
}
.btn-xs {
  padding: 4px 10px;
  font-size: 12px;
}
.text-xs {
  font-size: 12px;
}

/* 批量文件列表样式 */
.batch-badge {
  display: inline-block;
  font-size: 10px;
  padding: 2px 6px;
  background: var(--primary);
  color: #fff;
  border-radius: 4px;
  margin-left: 8px;
}
.batch-files-list {
  padding: 8px 12px;
  background: var(--bg);
  border-radius: 4px;
}
.batch-files-list summary {
  cursor: pointer;
  user-select: none;
}
.batch-files-list summary:hover {
  color: var(--text);
}
.file-list {
  list-style: none;
  padding: 0;
  margin: 0;
  max-height: 120px;
  overflow-y: auto;
}
.file-list li {
  padding: 4px 0;
  border-bottom: 1px solid var(--border);
}
.file-list li:last-child {
  border-bottom: none;
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

.skeleton-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-peer {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
}

.skeleton-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.skeleton-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
}

.skeleton-text-sm {
  width: 60%;
  height: 12px;
}

</style>
