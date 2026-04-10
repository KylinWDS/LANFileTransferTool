<template>
  <div class="encryption-tab">
    <div class="card">
      <h2>{{ $t('encryption.title') }}</h2>

      <div class="key-section mb-4">
        <label class="label">{{ $t('encryption.key') }}</label>
        <div class="key-input-group">
          <input
            type="text"
            v-model="encryptionKey"
            class="input"
            :placeholder="$t('encryption.generateKey')"
          />
          <button class="btn btn-primary" @click="generateNewKey">
            🔑 {{ $t('encryption.generateKey') }}
          </button>
        </div>
      </div>

      <div class="grid grid-2 mt-4">
        <div class="encrypt-section">
          <h3>{{ $t('encryption.encrypt') }}</h3>

          <label class="label">{{ $t('encryption.plainText') }}</label>
          <textarea
            v-model="plainText"
            class="input textarea"
            rows="5"
            :placeholder="$t('encryption.inputText')"
          ></textarea>

          <button
            class="btn btn-primary mt-4"
            :disabled="!plainText || !encryptionKey"
            @click="encrypt"
          >
            🔒 {{ $t('encryption.encrypt') }}
          </button>

          <div v-if="encryptedText" class="result-section mt-4">
            <label class="label">{{ $t('encryption.cipherText') }}</label>
            <textarea
              :value="encryptedText"
              class="input textarea result-text"
              rows="3"
              readonly
            ></textarea>
            <button class="btn btn-sm btn-success mt-2" @click="copyResult(encryptedText)">
              {{ $t('common.copy') }}
            </button>
          </div>
        </div>

        <div class="decrypt-section">
          <h3>{{ $t('encryption.decrypt') }}</h3>

          <label class="label">{{ $t('encryption.cipherText') }}</label>
          <textarea
            v-model="cipherTextInput"
            class="input textarea"
            rows="5"
            :placeholder="$t('encryption.inputText')"
          ></textarea>

          <button
            class="btn btn-success mt-4"
            :disabled="!cipherTextInput || !encryptionKey"
            @click="decrypt"
          >
            🔓 {{ $t('encryption.decrypt') }}
          </button>

          <div v-if="decryptedText" class="result-section mt-4">
            <label class="label">{{ $t('encryption.plainText') }}</label>
            <textarea
              :value="decryptedText"
              class="input textarea result-text"
              rows="3"
              readonly
            ></textarea>
            <button class="btn btn-sm btn-success mt-2" @click="copyResult(decryptedText)">
              {{ $t('common.copy') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'

const { t } = useI18n()

const encryptionKey = ref('')
const plainText = ref('')
const cipherTextInput = ref('')
const encryptedText = ref('')
const decryptedText = ref('')

const generateNewKey = async () => {
  try {
    const key = await api.GenerateEncryptionKey()
    encryptionKey.value = key
  } catch (error) {
    console.error('生成密钥失败:', error)
  }
}

const encrypt = async () => {
  if (!plainText.value || !encryptionKey.value) return

  try {
    const result = await api.EncryptData(plainText.value, encryptionKey.value)
    encryptedText.value = result
  } catch (error) {
    console.error('加密失败:', error)
    alert(t('encryption.encryptFailed'))
  }
}

const decrypt = async () => {
  if (!cipherTextInput.value || !encryptionKey.value) return

  try {
    const result = await api.DecryptData(cipherTextInput.value, encryptionKey.value)
    decryptedText.value = result
  } catch (error) {
    console.error('解密失败:', error)
    alert(t('encryption.decryptFailed'))
  }
}

const copyResult = (text) => {
  navigator.clipboard.writeText(text).then(() => {
    alert(t('common.copied'))
  })
}
</script>

<style scoped>
.key-input-group {
  display: flex;
  gap: 12px;
}

.key-input-group .input {
  flex: 1;
}

.textarea {
  resize: vertical;
  min-height: 100px;
  font-family: monospace;
  font-size: 13px;
}

.result-text {
  background-color: var(--surface-color);
  border-color: var(--success-color);
}

.result-section {
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}
</style>
