<script setup lang="ts">
import type { MetricCardData } from '@/types/dashboard'

defineProps<{
  item: MetricCardData
}>()

defineEmits<{
  click: []
}>()
</script>

<template>
  <article class="metric-card" :class="{ 'metric-card--clickable': true }" @click="$emit('click')">
    <p class="metric-card__title">{{ item.title }}</p>
    <p class="metric-card__value">{{ item.value }}</p>
    <p class="metric-card__detail">{{ item.detail }}</p>
    <p class="metric-card__trend">{{ item.trend }}</p>
  </article>
</template>

<style scoped>
.metric-card--clickable {
  cursor: pointer;
}
.metric-card--clickable:hover {
  opacity: 0.85;
}
</style>
