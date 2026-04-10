<template>
  <div class="history-tab">
    <div class="card">
      <div class="flex-between mb-4">
        <h2>{{ $t('history.title') }}</h2>
        <button
          class="btn btn-danger"
          :disabled="history.length === 0 || clearing"
          @click="clearHistory"
        >
          {{ clearing ? $t('common.loading') : $t('history.clearHistory') }}
        </button>
      </div>

      <div v-if="loading" class="text-center">{{ $t('common.loading') }}</div>

      <div v-else-if="history.length === 0" class="text-center text-secondary">
        {{ $t('history.noHistory') }}
      </div>

      <div v-else class="history-table">
        <table>
          <thead>
            <tr>
              <th>{{ $t('history.fileName') }}</th>
              <th>{{ $t('history.fileSize') }}</th>
              <th>{{ $t('history.action') }}</th>
              <th>{{ $t('history.status') }}</th>
              <th>{{ $t('history.savePath') }}</th>
              <th>{{ $t('history.time') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="record in history" :key="record.id">
              <td>{{ record.file_name }}</td>
              <td>{{ formatFileSize(record.file_size) }}</td>
              <td>
                <span :class="['status-badge', record.action === 'upload' ? 'status-ok' : 'status-warning']">
                  {{ getActionText(record.action) }}
                </span>
              </td>
              <td>
                <span :class="['status-badge', getStatusClass(record.status)]">
                  {{ getStatusText(record.status) }}
                </span>
              </td>
              <td class="text-sm">{{ record.save_path || '-' }}</td>
              <td>{{ formatDate(record.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const history = ref([])
const loading = ref(true)
const clearing = ref(false)

const loadHistory = async () => {
  loading.value = true
  try {
    const records = await api.GetHistory(50)
    history.value = records || []
  } catch (error) {
    console.error('加载历史记录失败:', error)
  } finally {
    loading.value = false
  }
}

const clearHistory = async () => {
  if (!confirm(t('common.confirm'))) return
  clearing.value = true
  try {
    await api.ClearHistory()
    history.value = []
  } catch (error) {
    console.error('清除历史记录失败:', error)
  } finally {
    clearing.value = false
  }
}

const getActionText = (action) => {
  const actions = {
    upload: t('history.upload'),
    download: t('history.download')
  }
  return actions[action] || action
}

const getStatusText = (status) => {
  const statuses = {
    pending: t('history.pending'),
    transferring: t('history.transferring'),
    completed: t('history.completed'),
    failed: t('history.failed')
  }
  return statuses[status] || status
}

const getStatusClass = (status) => {
  const classes = {
    pending: 'status-warning',
    transferring: 'status-warning',
    completed: 'status-ok',
    failed: 'status-error'
  }
  return classes[status] || ''
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
  loadHistory()
})
</script>

<style scoped>
.history-table {
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
</style>
