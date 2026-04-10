<template>
  <div class="settings-tab">
    <div class="card">
      <h2>{{ $t('settings.title') }}</h2>

      <div class="settings-sections">
        <section class="section">
          <h3>{{ $t('settings.general') }}</h3>

          <div class="setting-item">
            <label class="label">{{ $t('settings.theme') }}</label>
            <div class="theme-options">
              <button
                :class="['option-btn', { active: localTheme === 'light' }]"
                @click="localTheme = 'light'"
              >
                ☀️ {{ $t('settings.lightTheme') }}
              </button>
              <button
                :class="['option-btn', { active: localTheme === 'dark' }]"
                @click="localTheme = 'dark'"
              >
                🌙 {{ $t('settings.darkTheme') }}
              </button>
            </div>
          </div>

          <div class="setting-item">
            <label class="label">{{ $t('settings.language') }}</label>
            <select v-model="localLanguage" class="select">
              <option value="zh-CN">中文</option>
              <option value="en">English</option>
              <option value="ru">Русский</option>
            </select>
          </div>

          <div class="setting-item">
            <label class="checkbox-label">
              <input
                type="checkbox"
                v-model="settings.auto_start_server"
              />
              <span>{{ $t('settings.autoStartServer') }}</span>
            </label>
          </div>
        </section>

        <section class="section">
          <h3>{{ $t('settings.advanced') }}</h3>

          <div class="setting-item">
            <label class="label">{{ $t('settings.defaultPort') }}</label>
            <input
              type="number"
              v-model.number="settings.default_port"
              class="input"
              min="1024"
              max="65535"
            />
          </div>

          <div class="setting-item">
            <label class="label">{{ $t('settings.chunkSize') }} (bytes)</label>
            <input
              type="number"
              v-model.number="settings.chunk_size"
              class="input"
              min="1024"
              step="1024"
            />
          </div>

          <div class="setting-item">
            <label class="label">{{ $t('settings.maxConnections') }}</label>
            <input
              type="number"
              v-model.number="settings.max_connections"
              class="input"
              min="1"
              max="100"
            />
          </div>

          <div class="setting-item">
            <label class="checkbox-label">
              <input
                type="checkbox"
                v-model="settings.enable_resume"
              />
              <span>{{ $t('settings.enableResume') }}</span>
            </label>
          </div>
        </section>
      </div>

      <div class="settings-actions mt-4">
        <button class="btn btn-primary" @click="saveSettings">
          💾 {{ $t('settings.save') }}
        </button>
        <button class="btn btn-danger" @click="resetSettings">
          🔄 {{ $t('settings.reset') }}
        </button>
      </div>

      <div v-if="savedMessage" class="alert alert-success mt-4">
        {{ savedMessage }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const localTheme = ref('light')
const localLanguage = ref('zh-CN')
const settings = ref({
  auto_start_server: true,
  default_port: 8080,
  chunk_size: 1048576,
  max_connections: 10,
  enable_resume: true
})

const savedMessage = ref('')

const loadSettings = async () => {
  try {
    const config = await api.GetUserConfig()
    if (config) {
      localTheme.value = config.theme || 'light'
      localLanguage.value = config.language || 'zh-CN'
      if (config.settings) {
        Object.assign(settings.value, config.settings)
      }
    }
  } catch (error) {
    console.error('加载设置失败:', error)
  }
}

const saveSettings = async () => {
  try {
    const configData = {
      theme: localTheme.value,
      language: localLanguage.value,
      settings: settings.value
    }

    await api.SaveUserConfig(configData)

    savedMessage.value = t('settings.saved')
    setTimeout(() => {
      savedMessage.value = ''
    }, 3000)
  } catch (error) {
    console.error('保存设置失败:', error)
    alert('保存设置失败: ' + error.message)
  }
}

const resetSettings = async () => {
  if (!confirm(t('settings.resetConfirm'))) return

  try {
    await api.ResetUserConfig()
    await loadSettings()

    savedMessage.value = '配置已重置为默认值'
    setTimeout(() => {
      savedMessage.value = ''
    }, 3000)
  } catch (error) {
    console.error('重置设置失败:', error)
    alert('重置设置失败: ' + error.message)
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.settings-sections {
  margin-top: 24px;
}

.section {
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border-color);
}

.section:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.section h3 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--primary-color);
}

.setting-item {
  margin-bottom: 16px;
}

.theme-options {
  display: flex;
  gap: 12px;
}

.option-btn {
  padding: 10px 20px;
  border: 2px solid var(--border-color);
  background-color: var(--background-color);
  color: var(--text-primary);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 14px;
}

.option-btn:hover {
  border-color: var(--primary-color);
}

.option-btn.active {
  border-color: var(--primary-color);
  background-color: var(--primary-color);
  color: white;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.settings-actions {
  display: flex;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid var(--border-color);
}

.alert {
  padding: 12px 16px;
  border-radius: 6px;
  font-weight: 500;
}

.alert-success {
  background-color: rgba(76, 175, 80, 0.1);
  color: var(--success-color);
  border: 1px solid var(--success-color);
}
</style>
