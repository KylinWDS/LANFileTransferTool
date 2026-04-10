<template>
  <div class="history-tab">
    <div class="card">
      <div class="flex-between mb-2">
        <h2>{{ t('history.title') }}</h2>
        <button class="btn btn-sm" :disabled="!records.length || clearing" @click="clearHistory">{{ clearing ? t('common.loading') : t('history.clearHistory') }}</button>
      </div>
      <div v-if="loading" class="text-center text-secondary">{{ t('common.loading') }}</div>
      <div v-else-if="!records.length" class="text-center text-secondary">{{ t('history.noHistory') }}</div>
      <table v-else>
        <thead><tr><th>{{ t('history.fileName') }}</th><th>{{ t('history.fileSize') }}</th><th>{{ t('history.action') }}</th><th>{{ t('history.status') }}</th><th>{{ t('history.savePath') }}</th><th>{{ t('history.time') }}</th></tr></thead>
        <tbody>
          <tr v-for="r in records" :key="r.id">
            <td>{{ r.file_name }}</td>
            <td>{{ formatSize(r.file_size) }}</td>
            <td><span :class="['status-badge', r.action==='upload'?'status-ok':'status-warning']">{{ actionText(r.action) }}</span></td>
            <td><span :class="['status-badge', statusClass(r.status)]">{{ statusText(r.status) }}</span></td>
            <td class="text-sm text-secondary">{{ r.save_path || '-' }}</td>
            <td class="text-sm">{{ formatDate(r.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 1. 新增原生 dialog 元素 -->
    <dialog v-show="dialogShow" ref="confirmDialog" class="native-dialog">
      <div class="dialog-content">
        <p class="dialog-message">{{ t('common.confirm') }}</p>
        <div class="dialog-actions">
          <button class="btn btn-sm" @click="confirmClear">{{ t('common.submit') }}</button>
          <button class="btn btn-sm btn-secondary" @click="cancelClear">{{ t('common.cancel') }}</button>
        </div>
      </div>
    </dialog>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'
const { t } = useI18n()
const records = ref([])
const loading = ref(true)
const clearing = ref(false)
// 2. 使用 ref 获取 dialog 元素的引用
const confirmDialog = ref(null)
const dialogShow = ref(false)

const load = async () => { loading.value = true; try { records.value = (await api.GetHistory(50)) || [] } catch {} loading.value = false }
// const clearHistory = async () => { if (!confirm(t('common.confirm'))) return; clearing.value = true; try { await api.ClearHistory(); records.value = [] } catch {} clearing.value = false }

// 3. 修改 clearHistory 函数，仅用于显示弹窗
const clearHistory = () => {
  if (!records.value.length || clearing.value) return;
  // 使用原生 API 显示模态对话框
  confirmDialog.value?.showModal();
  dialogShow.value = true;
}

// 4. 新增确认按钮的处理函数
const confirmClear = async () => {
  clearing.value = true;
  // 先关闭对话框
  confirmDialog.value?.close();
  dialogShow.value = false;
  
  try {
    await api.ClearHistory();
    records.value = [];
  } catch (error) {
    console.error('清空历史失败', error);
  } finally {
    clearing.value = false;
  }
}

// 5. 新增取消按钮的处理函数
const cancelClear = () => {
  confirmDialog.value?.close();
  dialogShow.value = false; 
}

const actionText = (a) => a === 'upload' ? t('history.upload') : t('history.download')
const statusText = (s) => ({ pending: t('history.pending'), transferring: t('history.transferring'), completed: t('history.completed'), failed: t('history.failed') })[s] || s
const statusClass = (s) => ({ pending: 'status-warning', transferring: 'status-warning', completed: 'status-ok', failed: 'status-error' })[s] || ''
const formatSize = (b) => { if (!b) return '0 B'; const k=1024,s=['B','KB','MB','GB'],i=Math.floor(Math.log(b)/Math.log(k)); return (b/Math.pow(k,i)).toFixed(1)+' '+s[i] }
const formatDate = (d) => d ? new Date(d).toLocaleString() : ''
onMounted(() => {
  load();
  if (confirmDialog.value) {
    confirmDialog.value.close()
    dialogShow.value = false; 
  }
})
</script>

<style scoped>
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; text-align: left; border-bottom: 1px solid var(--border); font-size: 13px; }
th { color: var(--text-secondary); font-weight: 500; }


/* 6. 新增 dialog 相关样式 */
.native-dialog {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  color: var(--text-primary);
  background-color: var(--background);
  
  /* 居中核心代码 */
  display: flex;
  justify-content: center;
  align-items: center;
  
  /* 响应式限制 */
  max-width: 90vw;
  max-height: 90vh;
  margin: auto;
}

.native-dialog::backdrop {
  background-color: rgba(0, 0, 0, 0.5);
  /* 可选：添加淡入淡出效果 */
  transition: opacity 0.2s ease;
}

.dialog-content {
  padding: 24px; /* 稍微增加一点内边距看起来更舒服 */
  min-width: 280px;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 16px; /* 按钮和文字的间距 */
}

.dialog-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}

</style>
