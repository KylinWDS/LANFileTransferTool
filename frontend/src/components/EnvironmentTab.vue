<template>
  <div class="environment-tab">
    <div class="card">
      <div class="flex-between mb-4">
        <h2>{{ $t('environment.title') }}</h2>
        <button class="btn btn-primary" :disabled="checking" @click="runCheck">
          {{ checking ? $t('common.loading') : $t('environment.startCheck') }}
        </button>
      </div>

      <div v-if="checking" class="text-center">{{ $t('common.loading') }}</div>

      <div v-else-if="results" class="check-results">
        <div class="check-item card">
          <div class="check-header">
            <h3>🔥 {{ $t('environment.firewall') }}</h3>
            <span :class="['status-badge', getStatusClass(results.firewall)]">
              {{ getStatusText(results.firewall) }}
            </span>
          </div>
          <p class="text-secondary text-sm mt-2">{{ results.firewall?.message || '' }}</p>
          <p v-if="results.firewall?.details" class="text-sm mt-2">{{ results.firewall.details }}</p>
        </div>

        <div class="check-item card">
          <div class="check-header">
            <h3>🌐 {{ $t('environment.network') }}</h3>
            <span :class="['status-badge', getStatusClass(results.network)]">
              {{ getStatusText(results.network) }}
            </span>
          </div>
          <p class="text-secondary text-sm mt-2">{{ results.network?.message || '' }}</p>
          <p v-if="results.network?.details" class="text-sm mt-2">{{ results.network.details }}</p>
        </div>

        <div class="check-item card">
          <div class="check-header">
            <h3>🔌 {{ $t('environment.port') }}</h3>
            <span :class="['status-badge', getStatusClass(results.port)]">
              {{ getStatusText(results.port) }}
            </span>
          </div>
          <p class="text-secondary text-sm mt-2">{{ results.port?.message || '' }}</p>
          <p v-if="results.port?.details" class="text-sm mt-2">{{ results.port.details }}</p>
        </div>

        <div v-if="results.solutions && results.solutions.length > 0" class="card solutions-card">
          <h3>💡 {{ $t('environment.solutions') }}</h3>
          <ul class="solutions-list">
            <li v-for="(solution, index) in results.solutions" :key="index">
              {{ solution }}
            </li>
          </ul>
        </div>
      </div>

      <div v-else class="text-center text-secondary">
        {{ $t('environment.startCheck') }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const checking = ref(false)
const results = ref(null)

const runCheck = async () => {
  checking.value = true
  try {
    const checkResults = await api.CheckEnvironment()
    results.value = checkResults
  } catch (error) {
    console.error('环境检测失败:', error)
  } finally {
    checking.value = false
  }
}

const getStatusClass = (item) => {
  if (!item) return ''
  const status = item.status
  if (status === 'ok' || status === 'normal') return 'status-ok'
  if (status === 'warning') return 'status-warning'
  if (status === 'error' || status === 'blocked') return 'status-error'
  return 'status-ok'
}

const getStatusText = (item) => {
  if (!item) return ''
  const status = item.status
  if (status === 'ok' || status === 'normal') return t('environment.statusOk')
  if (status === 'warning') return t('environment.statusWarning')
  if (status === 'error' || status === 'blocked') return t('environment.statusError')
  return status
}
</script>

<style scoped>
.check-results {
  display: grid;
  gap: 16px;
}

.check-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.check-item {
  margin-bottom: 0;
}

.solutions-card {
  border-left: 4px solid var(--warning-color);
}

.solutions-list {
  list-style: none;
  padding: 0;
  margin-top: 12px;
}

.solutions-list li {
  padding: 8px 12px;
  margin-bottom: 8px;
  background-color: var(--background-color);
  border-radius: 6px;
  font-size: 14px;
}
</style>
