<template>
  <div class="environment-tab">
    <div class="card">
      <div class="flex-between mb-2">
        <h2>{{ t('environment.title') }}</h2>
        <button class="btn btn-primary" :disabled="checking" @click="runCheck">{{ checking ? t('common.loading') : t('environment.startCheck') }}</button>
      </div>

      <div v-if="checking" class="skeleton-container">
        <div v-for="i in 3" :key="i" class="skeleton-check skeleton">
          <div class="skeleton skeleton-title-sm"></div>
          <div class="skeleton skeleton-text"></div>
        </div>
      </div>
      <div v-else-if="results" class="check-list">
        <div 
          v-for="item in checkItems" 
          :key="item.key" 
          v-memo="[item.key, item.status]"
          class="check-row slide-up"
        >
          <div class="flex-between">
            <span class="check-label">{{ item.icon }} {{ item.label }}</span>
            <span :class="['status-badge', item.cls]">{{ item.status }}</span>
          </div>
          <div v-if="item.message" class="text-sm text-secondary mt-2">{{ item.message }}</div>
          <div v-if="item.details" class="text-sm mt-2">{{ item.details }}</div>
        </div>
        <div v-if="results.solutions && results.solutions.length" class="mt-4">
          <h3>{{ t('environment.solutions') }}</h3>
          <ul class="solution-list">
            <li v-for="(s, i) in results.solutions" :key="i">{{ s }}</li>
          </ul>
        </div>
      </div>
      <div v-else class="empty-state">
        <div class="empty-state-icon">🔍</div>
        <div class="empty-state-text">{{ t('environment.startCheck') }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'
const { t } = useI18n()
const checking = ref(false)
const results = ref(null)

const checkItems = computed(() => {
  if (!results.value) return []
  const r = results.value
  return [
    { key: 'firewall', icon: '🔥', label: t('environment.firewall'), cls: statusCls(r.firewall), status: statusText(r.firewall), message: r.firewall?.message, details: r.firewall?.details },
    { key: 'network', icon: '🌐', label: t('environment.network'), cls: statusCls(r.network), status: statusText(r.network), message: r.network?.message, details: r.network?.details },
    { key: 'port', icon: '🔌', label: t('environment.port'), cls: statusCls(r.port), status: statusText(r.port), message: r.port?.message, details: r.port?.details },
  ]
})

const statusCls = (item) => {
  if (!item) return 'status-ok'
  const s = item.status
  if (s === 'ok' || s === 'normal') return 'status-ok'
  if (s === 'warning') return 'status-warning'
  if (s === 'error' || s === 'blocked') return 'status-error'
  return 'status-ok'
}

const statusText = (item) => {
  if (!item) return t('environment.statusOk')
  const s = item.status
  if (s === 'ok' || s === 'normal') return t('environment.statusOk')
  if (s === 'warning') return t('environment.statusWarning')
  if (s === 'error' || s === 'blocked') return t('environment.statusError')
  return s
}

const runCheck = async () => {
  checking.value = true
  try { results.value = await api.CheckEnvironment() } catch {}
  checking.value = false
}
</script>

<style scoped>
.check-list { display: flex; flex-direction: column; gap: 12px; }
.check-row { padding: 12px; background: var(--bg); border: 1px solid var(--border); border-radius: 6px; transition: all var(--transition-fast); }
.check-row:hover { border-color: var(--primary); box-shadow: var(--shadow); }
.check-label { font-weight: 500; }
.solution-list { list-style: none; padding: 0; margin-top: 8px; }
.solution-list li { padding: 6px 10px; margin-bottom: 4px; background: var(--bg); border: 1px solid var(--border); border-radius: 4px; font-size: 13px; }

.skeleton-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-check {
  padding: 12px;
  background: var(--bg);
  border-radius: 6px;
}

.skeleton-title-sm {
  width: 150px;
  height: 16px;
  margin-bottom: 8px;
}

.skeleton-check .skeleton-text {
  width: 80%;
  height: 12px;
}
</style>
