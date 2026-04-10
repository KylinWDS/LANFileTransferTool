<template>
  <div class="encryption-tab">
    <div class="card">
      <h2>{{ t('encryption.title') }}</h2>
      <div class="mb-4">
        <label class="label">{{ t('encryption.key') }}</label>
        <div class="flex gap-2">
          <input v-model="key" class="input" readonly />
          <button class="btn btn-primary" @click="genKey">{{ t('encryption.generateKey') }}</button>
        </div>
      </div>
      <div class="grid-2 mb-4">
        <div>
          <label class="label">{{ t('encryption.plainText') }}</label>
          <textarea v-model="plainText" class="input" rows="4" style="resize:vertical"></textarea>
          <button class="btn btn-primary mt-2" @click="encrypt">{{ t('encryption.encrypt') }}</button>
        </div>
        <div>
          <label class="label">{{ t('encryption.cipherText') }}</label>
          <textarea v-model="cipherText" class="input" rows="4" style="resize:vertical"></textarea>
          <button class="btn mt-2" @click="decrypt">{{ t('encryption.decrypt') }}</button>
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
const key = ref('')
const plainText = ref('')
const cipherText = ref('')

const genKey = async () => { try { key.value = await api.GenerateEncryptionKey() } catch {} }
const encrypt = async () => { try { cipherText.value = await api.EncryptData(plainText.value, key.value) } catch { alert(t('encryption.encryptFailed')) } }
const decrypt = async () => { try { plainText.value = await api.DecryptData(cipherText.value, key.value) } catch { alert(t('encryption.decryptFailed')) } }
</script>
