<template>
  <div class="history-tab">
    <div class="card">
      <div class="flex-between mb-2">
        <h2>{{ t('history.title') }}</h2>
        <button class="btn btn-sm" :disabled="!records.length || clearing" @click="clearHistory">{{ clearing ? t('common.loading') : t('history.clearHistory') }}</button>
      </div>
      <div v-if="loading" class="text-center text-secondary">{{ t('common.loading') }}</div>
      <div v-else-if="!records.length" class="text-center text-secondary">{{ t('history.noHistory') }}</div>
      <table v-else>
        <thead><tr><th>{{ t('history.fileName') }}</th><th>{{ t('history.fileSize') }}</th><th>{{ t('history.action') }}</th><th>{{ t('history.status') }}</th><th>{{ t('history.savePath') }}</th><th>{{ t('history.time') }}</th></tr></thead>
        <tbody>
          <tr v-for="r in records" :key="r.id">
            <td>{{ r.file_name }}</td>
            <td>{{ formatSize(r.file_size) }}</td>
            <td><span :class="['status-badge', r.action==='upload'?'status-ok':'status-warning']">{{ actionText(r.action) }}</span></td>
            <td><span :class="['status-badge', statusClass(r.status)]">{{ statusText(r.status) }}</span></td>
            <td class="text-sm text-secondary">{{ r.save_path || '-' }}</td>
            <td class="text-sm">{{ formatDate(r.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'
const { t } = useI18n()
const records = ref([])
const loading = ref(true)
const clearing = ref(false)

const load = async () => { loading.value = true; try { records.value = (await api.GetHistory(50)) || [] } catch {} loading.value = false }
const clearHistory = async () => { if (!confirm(t('common.confirm'))) return; clearing.value = true; try { await api.ClearHistory(); records.value = [] } catch {} clearing.value = false }
const actionText = (a) => a === 'upload' ? t('history.upload') : t('history.download')
const statusText = (s) => ({ pending: t('history.pending'), transferring: t('history.transferring'), completed: t('history.completed'), failed: t('history.failed') })[s] || s
const statusClass = (s) => ({ pending: 'status-warning', transferring: 'status-warning', completed: 'status-ok', failed: 'status-error' })[s] || ''
const formatSize = (b) => { if (!b) return '0 B'; const k=1024,s=['B','KB','MB','GB'],i=Math.floor(Math.log(b)/Math.log(k)); return (b/Math.pow(k,i)).toFixed(1)+' '+s[i] }
const formatDate = (d) => d ? new Date(d).toLocaleString() : ''
onMounted(load)
</script>

<style scoped>
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; text-align: left; border-bottom: 1px solid var(--border); font-size: 13px; }
th { color: var(--text-secondary); font-weight: 500; }
</style>
