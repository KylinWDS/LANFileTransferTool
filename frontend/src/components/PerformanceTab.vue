<template>
  <div class="performance-tab">
    <div class="card">
      <div class="flex-between mb-4">
        <h2>{{ $t('performance.title') }}</h2>
        <div class="actions">
          <button class="btn btn-sm btn-primary" @click="refreshStats">
            {{ $t('performance.refresh') }}
          </button>
          <label class="auto-refresh-label">
            <input type="checkbox" v-model="autoRefresh" @change="toggleAutoRefresh" />
            {{ $t('performance.autoRefresh') }}
          </label>
        </div>
      </div>

      <div v-if="loading" class="text-center">{{ $t('common.loading') }}</div>

      <div v-else class="stats-grid">
        <div class="stat-card">
          <div class="stat-label">{{ $t('performance.cpuUsage') }}</div>
          <div class="stat-value">{{ stats.cpu_usage.toFixed(1) }}%</div>
          <div class="progress-bar mt-2">
            <div class="progress-bar-fill" :style="{ width: stats.cpu_usage + '%' }"></div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-label">{{ $t('performance.memoryUsage') }}</div>
          <div class="stat-value">{{ stats.memory_usage.toFixed(1) }}%</div>
          <div class="progress-bar mt-2">
            <div class="progress-bar-fill" :style="{ width: stats.memory_usage + '%' }"></div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-label">{{ $t('performance.networkSpeed') }}</div>
          <div class="stat-value">{{ stats.network_speed.toFixed(2) }} MB/s</div>
        </div>

        <div class="stat-card">
          <div class="stat-label">{{ $t('performance.goroutines') }}</div>
          <div class="stat-value">{{ stats.active_goroutines }}</div>
        </div>
      </div>
    </div>

    <div class="card">
      <h2>{{ $t('performance.poolSize') }}</h2>
      <div class="pool-controls">
        <div class="pool-input-group">
          <label class="label">{{ $t('performance.poolSize') }}</label>
          <input
            type="number"
            v-model.number="poolSize"
            class="input"
            min="1"
            max="100"
            style="max-width: 120px"
          />
        </div>
        <div class="pool-actions">
          <button
            class="btn btn-success"
            :disabled="poolRunning"
            @click="initPool"
          >
            {{ $t('performance.initPool') }}
          </button>
          <button
            class="btn btn-danger"
            :disabled="!poolRunning"
            @click="stopPool"
          >
            {{ $t('performance.stopPool') }}
          </button>
        </div>
        <div class="pool-status">
          <span :class="['status-badge', poolRunning ? 'status-ok' : 'status-warning']">
            {{ poolRunning ? $t('performance.poolRunning') : $t('performance.poolStopped') }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const loading = ref(true)
const stats = ref({
  cpu_usage: 0,
  memory_usage: 0,
  network_speed: 0,
  active_goroutines: 0
})
const poolSize = ref(10)
const poolRunning = ref(false)
const autoRefresh = ref(false)
let refreshInterval = null

const refreshStats = async () => {
  try {
    const result = await api.GetPerformanceStats()
    if (result) {
      stats.value = result
    }
  } catch (error) {
    console.error('获取性能统计失败:', error)
  } finally {
    loading.value = false
  }
}

const initPool = async () => {
  try {
    await api.InitThreadPool(poolSize.value)
    poolRunning.value = true
  } catch (error) {
    console.error('启动线程池失败:', error)
  }
}

const stopPool = async () => {
  try {
    await api.StopThreadPool()
    poolRunning.value = false
  } catch (error) {
    console.error('停止线程池失败:', error)
  }
}

const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    refreshInterval = setInterval(refreshStats, 2000)
  } else {
    if (refreshInterval) {
      clearInterval(refreshInterval)
      refreshInterval = null
    }
  }
}

onMounted(() => {
  refreshStats()
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.stat-card {
  padding: 20px;
  background-color: var(--background-color);
  border-radius: 8px;
  text-align: center;
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--primary-color);
}

.pool-controls {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-top: 16px;
  flex-wrap: wrap;
}

.pool-input-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pool-actions {
  display: flex;
  gap: 12px;
}

.auto-refresh-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  cursor: pointer;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}
</style>
