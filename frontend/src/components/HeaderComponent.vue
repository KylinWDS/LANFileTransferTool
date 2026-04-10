<template>
  <header class="header">
    <div class="header-left">
      <h1>{{ $t('header.title') }}</h1>
      <span class="server-status" :class="{ running: serverRunning }">
        {{ serverStatusText }}
      </span>
    </div>

    <div class="header-actions">
      <select
        v-model="selectedLanguage"
        class="select language-select"
        @change="$emit('changeLanguage', selectedLanguage)"
      >
        <option value="zh-CN">中文</option>
        <option value="en">English</option>
        <option value="ru">Русский</option>
      </select>

      <button class="btn btn-primary" @click="$emit('toggleTheme')">
        {{ currentTheme === 'light' ? '🌙' : '☀️' }}
        {{ $t('header.theme') }}
      </button>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../api'

const props = defineProps({
  currentTheme: String,
  currentLanguage: String
})

defineEmits(['toggleTheme', 'changeLanguage'])

const selectedLanguage = ref(props.currentLanguage)
const serverRunning = ref(false)
const serverInfo = ref(null)

const serverStatusText = computed(() => {
  return serverRunning.value ? `🟢 ${serverInfo.value?.url || 'Running'}` : '🔴 Stopped'
})

onMounted(async () => {
  try {
    const info = await api.GetServerInfo()
    if (info) {
      serverInfo.value = info
      serverRunning.value = info.running
    }
  } catch (error) {
    console.error('获取服务器信息失败:', error)
  }
})
</script>

<style scoped>
.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.server-status {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  background-color: rgba(244, 67, 54, 0.1);
  color: var(--danger-color);
}

.server-status.running {
  background-color: rgba(76, 175, 80, 0.1);
  color: var(--success-color);
}

.language-select {
  min-width: 120px;
}
</style>
