<template>
  <div class="settings-tab">
    <div class="card">
      <h2>{{ t('settings.general') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('settings.theme') }}</label>
        <select v-model="theme" class="select" @change="save">
          <option value="light">{{ t('settings.lightTheme') }}</option>
          <option value="dark">{{ t('settings.darkTheme') }}</option>
        </select>
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.language') }}</label>
        <select v-model="language" class="select" @change="onLangChange">
          <option value="zh-CN">中文</option>
          <option value="en">English</option>
          <option value="ru">Русский</option>
        </select>
      </div>
    </div>

    <div class="card">
      <h2>{{ t('settings.advanced') }}</h2>
      <div class="setting-row">
        <label class="label">{{ t('settings.defaultPort') }}</label>
        <input v-model.number="port" type="number" class="input" style="max-width:120px" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.chunkSize') }} (MB)</label>
        <input v-model.number="chunkSize" type="number" class="input" style="max-width:120px" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.maxConnections') }}</label>
        <input v-model.number="maxConn" type="number" class="input" style="max-width:120px" />
      </div>
      <div class="setting-row">
        <label class="label">{{ t('settings.enableResume') }}</label>
        <input v-model="resume" type="checkbox" />
      </div>
      <div class="flex gap-2 mt-4">
        <button class="btn btn-primary" @click="save">{{ t('settings.save') }}</button>
        <button class="btn" @click="reset">{{ t('settings.reset') }}</button>
      </div>
      <div v-if="savedMsg" class="text-sm mt-2" style="color:var(--success)">{{ savedMsg }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'
const { t, locale } = useI18n()
const theme = ref('light')
const language = ref('zh-CN')
const port = ref(8080)
const chunkSize = ref(1)
const maxConn = ref(10)
const resume = ref(true)
const savedMsg = ref('')

const emit = defineEmits(['theme-change', 'language-change'])

const load = async () => {
  try {
    const c = await api.GetUserConfig()
    if (c) {
      theme.value = c.theme || 'light'
      language.value = c.language || 'zh-CN'
      if (c.settings) {
        port.value = c.settings.default_port || 8080
        chunkSize.value = (c.settings.chunk_size || 1048576) / 1048576
        maxConn.value = c.settings.max_connections || 10
        resume.value = c.settings.enable_resume !== false
      }
    }
  } catch {}
}

const onLangChange = () => { locale.value = language.value; save() }

const save = async () => {
  try {
    await api.SaveUserConfig({
      theme: theme.value,
      language: language.value,
      settings: {
        default_port: port.value,
        chunk_size: chunkSize.value * 1048576,
        max_connections: maxConn.value,
        enable_resume: resume.value
      }
    })
    document.documentElement.setAttribute('data-theme', theme.value)
    savedMsg.value = t('settings.saved')
    setTimeout(() => { savedMsg.value = '' }, 2000)
  } catch {}
}

const reset = async () => {
  if (!confirm(t('settings.resetConfirm'))) return
  try { await api.ResetUserConfig(); await load() } catch {}
}

onMounted(load)
</script>

<style scoped>
.setting-row { margin-bottom: 16px; }
</style>
