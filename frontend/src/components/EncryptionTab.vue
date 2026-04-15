<template>
  <div class="encryption-tab">
    <div class="card">
      <h2>{{ t('encryption.title') }}</h2>
      <div class="mb-4">
        <label class="label">{{ t('encryption.key') }}</label>
        <div class="flex gap-2">
          <input v-model="key" class="input" :placeholder="t('encryption.keyPlaceholder')" />
          <button 
            class="btn btn-primary" 
            @click="genKey" 
            :disabled="generatingKey"
          >
            {{ generatingKey ? t('common.loading') : t('encryption.generateKey') }}
          </button>
        </div>
      </div>
      <div class="grid-2 mb-4">
        <div>
          <label class="label">{{ t('encryption.plainText') }}</label>
          <textarea 
            v-model="plainText" 
            class="input" 
            rows="4" 
            style="resize:vertical"
            :placeholder="t('encryption.plainTextPlaceholder')"
          ></textarea>
          <button 
            class="btn btn-primary mt-2" 
            @click="encrypt" 
            :disabled="encrypting || !plainText || !key"
          >
            {{ encrypting ? t('common.loading') : t('encryption.encrypt') }}
          </button>
        </div>
        <div>
          <label class="label">{{ t('encryption.cipherText') }}</label>
          <textarea 
            v-model="cipherText" 
            class="input" 
            rows="4" 
            style="resize:vertical"
            :placeholder="t('encryption.cipherTextPlaceholder')"
          ></textarea>
          <button 
            class="btn mt-2" 
            @click="decrypt" 
            :disabled="decrypting || !cipherText || !key"
          >
            {{ decrypting ? t('common.loading') : t('encryption.decrypt') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="toast" class="toast" :class="toastType">{{ toast }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()
const key = ref('')
const plainText = ref('')
const cipherText = ref('')

const generatingKey = ref(false)
const encrypting = ref(false)
const decrypting = ref(false)
const toast = ref('')
const toastType = ref('')

const showToast = (message, type = 'info') => {
  toast.value = message
  toastType.value = type
  setTimeout(() => { toast.value = '' }, 3000)
}

const genKey = async () => {
  generatingKey.value = true
  try {
    key.value = await api.GenerateEncryptionKey()
    plainText.value = ''
    cipherText.value = ''
    showToast(t('encryption.keyGenerated'), 'success')
  } catch (e) {
    console.error('生成密钥失败:', e)
    showToast(t('encryption.keyGenerateFailed'), 'error')
  } finally {
    generatingKey.value = false
  }
}

const encrypt = async () => {
  if (!plainText.value || !key.value) {
    showToast(t('encryption.emptyInput'), 'warning')
    return
  }
  
  encrypting.value = true
  try {
    cipherText.value = ''
    cipherText.value = await api.EncryptData(plainText.value, key.value)
    showToast(t('encryption.encryptSuccess'), 'success')
  } catch (e) {
    console.error('加密失败:', e)
    showToast(t('encryption.encryptFailed'), 'error')
  } finally {
    encrypting.value = false
  }
}

const decrypt = async () => {
  if (!cipherText.value || !key.value) {
    showToast(t('encryption.emptyInput'), 'warning')
    return
  }
  
  decrypting.value = true
  try {
    plainText.value = ''
    plainText.value = await api.DecryptData(cipherText.value, key.value)
    showToast(t('encryption.decryptSuccess'), 'success')
  } catch (e) {
    console.error('解密失败:', e)
    showToast(t('encryption.decryptFailed'), 'error')
  } finally {
    decrypting.value = false
  }
}
</script>

<style scoped>
.toast {
  position: fixed;
  bottom: 32px;
  left: 50%;
  transform: translateX(-50%);
  padding: 10px 24px;
  border-radius: 6px;
  font-size: 13px;
  z-index: 9999;
  animation: toast-in 3s ease;
}

.toast.success {
  background: var(--success);
  color: #fff;
}

.toast.error {
  background: var(--danger);
  color: #fff;
}

.toast.warning {
  background: var(--warning);
  color: #000;
}

.toast.info {
  background: var(--primary);
  color: #fff;
}

@keyframes toast-in {
  0% { opacity: 0; transform: translateX(-50%) translateY(8px); }
  10% { opacity: 1; transform: translateX(-50%) translateY(0); }
  80% { opacity: 1; }
  100% { opacity: 0; }
}
</style>
