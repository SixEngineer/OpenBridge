<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <div class="page-header">
    <div class="page-header__top">
      <div>
        <h2>{{ title }}</h2>
        <p class="page-header__description">{{ description }}</p>
      </div>
      <!-- Actions slot -->
      <div v-if="$slots.actions" class="page-header__actions">
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-header {
  margin-bottom: 24px;
}

.page-header__top {
  display: flex;
  align-items: center;  
  justify-content: space-between;
  width: 100%; 
}

.page-header__top h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.page-header__actions {
  display: flex;
  gap: 12px;
}

.page-header__description {
  margin: 4px 0 0 0;
  color: var(--muted);
  font-size: 14px;
  word-break: keep-all;
}

@media (max-width: 860px) {
  .page-header {
    margin-bottom: 16px;
  }

  .page-header__top {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .page-header__top h2 {
    font-size: 20px;
  }
}
</style>