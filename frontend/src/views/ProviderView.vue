<script setup lang="ts">
import { onMounted, ref } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ProviderFormDialog from '@/components/provider/ProviderFormDialog.vue'
import { useConsoleStore } from '@/stores/console'
import type { ProviderRecord } from '@/types/provider'
import { registerProvider, updateProvider } from '@/api/provider'

const store = useConsoleStore()

// 对话框状态
const dialogVisible = ref(false)
const editingProvider = ref<ProviderRecord | null>(null)

// 格式化字节为可读大小
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 打开新增对话框
function openAddDialog() {
  editingProvider.value = null
  dialogVisible.value = true
}

// 打开编辑对话框
function openEditDialog(provider: ProviderRecord) {
  editingProvider.value = provider
  dialogVisible.value = true
}

// 提交表单（区分新增和编辑）
async function handleSubmit(data: Partial<ProviderRecord>) {
  try {
    let res
    if (editingProvider.value?.id) {
      res = await updateProvider({ ...data, id: editingProvider.value.id })
    } else {
      res = await registerProvider(data)
    }
    
    if (res.code === 1000 || res.code === 0) {
      alert(editingProvider.value ? '更新成功！' : '注册成功！')
      dialogVisible.value = false
      editingProvider.value = null
      await store.fetchProviders()
    } else {
      alert('操作失败：' + (res.msg))
    }
  } catch (error) {
    console.error('操作失败', error)
    alert('操作失败，请稍后重试')
  }
}

// 删除处理
async function handleDelete(provider: ProviderRecord) {
  if (!confirm(`确定要删除 "${provider.name}" 吗？`)) {
    return
  }
  
  const success = await store.removeProvider(provider.id)
  if (success) {
    alert('删除成功！')
  } else {
    alert('删除失败，请稍后重试')
  }
}

onMounted(() => {
  store.fetchProviders()
})
</script>

<template>
  <section class="page">
    <PageHeader title="Providers" description="网盘服务商管理">
      <template #actions>
        <button class="btn btn--primary" @click="openAddDialog">
          + 注册 Provider
        </button>
      </template>
    </PageHeader>

    <div v-if="store.providers.length === 0" class="empty-state">
      <p>暂无 Provider，点击上方按钮注册</p>
    </div>

    <div v-else class="provider-grid">
      <article v-for="provider in store.providers" :key="provider.id" class="provider-card">
        <div class="provider-card__header">
          <div>
            <p class="provider-card__name">{{ provider.name }}</p>
            <p class="provider-card__id">{{ provider.provider_type }} · {{ provider.net_disk }}</p>
          </div>
          <div class="provider-card__header-right">
            <StatusBadge :state="provider.status" />
            <button 
              class="provider-card__edit" 
              @click="openEditDialog(provider)"
              title="编辑"
            >
              ✏️
            </button>
            <button 
              class="provider-card__delete" 
              @click="handleDelete(provider)"
              title="删除"
            >
              🗑️
            </button>
          </div>
        </div>
        
        <p class="provider-card__section-title">{{ provider.net_disk === 'local' ? '本地路径' : '账户ID' }}</p>
        <p class="provider-card__text">{{ provider.account_id || '未设置' }}</p>
        
        <p class="provider-card__section-title">配额使用</p>
        <p class="provider-card__text">
          总计: {{ formatBytes(provider.total_quota) }}<br>
          已用: {{ formatBytes(provider.used_quota) }}<br>
          可用: {{ formatBytes(provider.available_quota) }}
        </p>
        
        <p class="provider-card__section-title" v-if="provider.last_error">最近错误</p>
        <p class="provider-card__text provider-card__text--error" v-if="provider.last_error">
          {{ provider.last_error }}
        </p>
      </article>
    </div>

    <ProviderFormDialog
      v-model:visible="dialogVisible"
      :provider="editingProvider"
      @submit="handleSubmit"
    />
  </section>
</template>

<style scoped>
.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.provider-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 1px solid #e5e7eb;
  transition: all 0.2s;
}

.provider-card:hover {
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border-color: #d1d5db;
}

.provider-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.provider-card__header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.provider-card__edit {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  opacity: 0.5;
  transition: all 0.2s;
}

.provider-card__edit:hover {
  opacity: 1;
  background: #e0e7ff;
}

.provider-card__delete {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  opacity: 0.5;
  transition: all 0.2s;
}

.provider-card__delete:hover {
  opacity: 1;
  background: #fee2e2;
}

.provider-card__name {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: #111827;
}

.provider-card__id {
  font-size: 13px;
  color: #6b7280;
  margin: 0;
}

.provider-card__section-title {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #6b7280;
  margin: 16px 0 4px 0;
}

.provider-card__text {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #374151;
}

.provider-card__text--error {
  color: #ef4444;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
  font-size: 16px;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px dashed #d1d5db;
}

.btn {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn--primary {
  background: #3b82f6;
  color: white;
}

.btn--primary:hover {
  background: #2563eb;
}
</style>