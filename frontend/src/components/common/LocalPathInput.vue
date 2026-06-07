<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { pickLocalPath } from '@/api/system'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  modelValue: string
  mode?: 'file' | 'directory'
  placeholder?: string
  title?: string
  filter?: string
  disabled?: boolean
}>(), {
  mode: 'directory',
  placeholder: '',
  title: '',
  filter: '',
  disabled: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const picking = ref(false)
const error = ref('')

function handleInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

async function handleBrowse() {
  if (props.disabled || picking.value) return

  picking.value = true
  error.value = ''
  try {
    const res = await pickLocalPath({
      kind: props.mode,
      title: props.title || props.placeholder || '',
      current_path: props.modelValue,
      filter: props.filter,
    })
    if (res.data.path) {
      emit('update:modelValue', res.data.path)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.request_error')
  } finally {
    picking.value = false
  }
}
</script>

<template>
  <div class="local-path-input">
    <div class="local-path-input__row">
      <input
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        class="local-path-input__control"
        type="text"
        @input="handleInput"
      />
      <button
        type="button"
        class="local-path-input__browse"
        :disabled="disabled || picking"
        @click="handleBrowse"
      >
        {{ picking ? t('common.browsing') : t('common.browse') }}
      </button>
    </div>
    <p v-if="error" class="local-path-input__error">{{ error }}</p>
  </div>
</template>

<style scoped>
.local-path-input {
  width: 100%;
}

.local-path-input__row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.local-path-input__control {
  flex: 1;
  min-width: 0;
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  box-sizing: border-box;
  background: var(--surface);
  color: var(--text);
  font-size: 14px;
}

.local-path-input__control:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
}

.local-path-input__browse {
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
  white-space: nowrap;
}

.local-path-input__browse:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.local-path-input__error {
  margin: 6px 0 0;
  color: #dc2626;
  font-size: 12px;
}

@media (max-width: 860px) {
  .local-path-input__row {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
