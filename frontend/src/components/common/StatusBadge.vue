<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { HealthState } from '@/types/common'

const { t } = useI18n()

const props = defineProps<{
  state: HealthState
}>()

const label = computed(() => {
  switch (props.state) {
    case 'active':
      return t('common.status_active')
    case 'disabled':
      return t('common.status_disabled')
    case 'expired':
      return t('common.status_expired')
    case 'error':
      return t('common.status_error')
    default:
      return props.state
  }
})
</script>

<template>
  <span class="status-badge" :class="`status-badge--${state}`">{{ label }}</span>
</template>