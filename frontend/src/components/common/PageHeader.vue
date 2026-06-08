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
  margin-bottom: 14px;
}

.page-header__top {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  width: 100%; 
  gap: 12px 18px;
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
  margin: 6px 0 0 0;
  color: var(--muted);
  font-size: 14px;
  word-break: keep-all;
  max-width: 640px;
}

@media (max-width: 860px) {
  .page-header {
    margin-bottom: 10px;
  }

  .page-header__top {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: 4px 10px;
  }

  .page-header__top h2 {
    font-size: 18px;
    flex: 1;
    min-width: 0;
  }

  .page-header__actions {
    flex-shrink: 0;
  }

  .page-header__description {
    width: 100%;
    margin-top: 0;
  }
}
</style>
