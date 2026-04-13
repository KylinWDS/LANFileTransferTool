<template>
  <div class="performance-tab">
    <div class="card">
      <div class="flex-between mb-2">
        <h2>{{ t('performance.title') }}</h2>
        <div class="flex gap-2">
          <button class="btn btn-sm" @click="refresh">{{ t('performance.refresh') }}</button>
          <label class="text-sm" style="display:flex;align-items:center;gap:4px;cursor:pointer">
            <input type="checkbox" v-model="autoRefresh" @change="toggleAuto" /> {{ t('performance.autoRefresh') }}
          </label>
        </div>
      </div>
      <div v-if="loading" class="text-center text-secondary">{{ t('common.loading') }}</div>
      <div v-else>
        <!-- 线程池状态 -->
        <h3 class="section-title">{{ t('performance.poolStats') }}</h3>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.poolStatus') }}</div>
            <div class="stat-val" :class="stats.pool_running ? 'text-success' : 'text-secondary'">
              {{ stats.pool_running ? t('performance.poolRunning') : t('performance.poolStopped') }}
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.poolSize') }}</div>
            <div class="stat-val">{{ stats.pool_size }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.poolTasks') }}</div>
            <div class="stat-val">{{ stats.pool_task_count }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.poolQueue') }}</div>
            <div class="stat-val">{{ stats.pool_queue_size }}</div>
          </div>
        </div>

        <h3 class="section-title mt-4">{{ t('performance.networkStats') }}</h3>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.sendSpeed') }}</div>
            <div class="stat-val">{{ formatSpeed(stats.network_send_speed) }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.recvSpeed') }}</div>
            <div class="stat-val">{{ formatSpeed(stats.network_recv_speed) }}</div>
          </div>
        </div>
        
        <h3 class="section-title mt-4">{{ t('performance.diskStats') }}</h3>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.diskRead') }}</div>
            <div class="stat-val">{{ formatSpeed(stats.disk_read_speed) }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.diskWrite') }}</div>
            <div class="stat-val">{{ formatSpeed(stats.disk_write_speed) }}</div>
          </div>
        </div>
        
        <h3 class="section-title mt-4">{{ t('performance.systemStats') }}</h3>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.cpuUsage') }}</div>
            <div class="stat-val">{{ stats.cpu_usage.toFixed(1) }}%</div>
            <div class="progress-bar mt-2"><div class="progress-bar-fill" :style="{width:stats.cpu_usage+'%'}"></div></div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.memoryUsage') }}</div>
            <div class="stat-val">{{ stats.memory_usage.toFixed(1) }}%</div>
            <div class="progress-bar mt-2"><div class="progress-bar-fill" :style="{width:stats.memory_usage+'%'}"></div></div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ t('performance.goroutines') }}</div>
            <div class="stat-val">{{ stats.active_goroutines }}</div>
          </div>
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
  network_send_speed: 0, 
  network_recv_speed: 0,
  disk_read_speed: 0,
  disk_write_speed: 0,
  active_goroutines: 0,
  pool_running: false,
  pool_size: 0,
  pool_task_count: 0,
  pool_queue_size: 0
})
const autoRefresh = ref(false)
let timer = null

const formatSpeed = (mbps) => {
  if (!mbps || mbps < 0.01) return '0 KB/s'
  if (mbps < 1) return (mbps * 1024).toFixed(1) + ' KB/s'
  return mbps.toFixed(2) + ' MB/s'
}

const refresh = async () => {
  try {
    const r = await api.GetPerformanceStats()
    if (r) stats.value = r
  } catch {}
  loading.value = false
}

const toggleAuto = () => { if (autoRefresh.value) { timer = setInterval(refresh, 2000) } else { clearInterval(timer); timer = null } }

onMounted(refresh)
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<style scoped>
.section-title { font-size: 14px; color: var(--text-secondary); margin-bottom: 8px; font-weight: 500; }
.stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; }
.stat-item { padding: 16px; background: var(--bg); border: 1px solid var(--border); border-radius: 6px; text-align: center; }
.stat-label { font-size: 12px; color: var(--text-secondary); margin-bottom: 4px; }
.stat-val { font-size: 24px; font-weight: 700; color: var(--primary); }
.text-success { color: var(--success); }
.text-secondary { color: var(--text-secondary); }
</style>
