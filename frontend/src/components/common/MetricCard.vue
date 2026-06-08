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
    <p v-if="item.trend" class="metric-card__trend">{{ item.trend }}</p>
  </article>
</template>

<style scoped>
.metric-card {
  position: relative;
  overflow: hidden;
  padding: 18px 18px 16px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    linear-gradient(160deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.82)),
    var(--surface);
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.06);
  transition: transform 0.28s ease, box-shadow 0.28s ease, border-color 0.28s ease;
}

.metric-card::before {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 4px;
  background: linear-gradient(90deg, #0f766e, #2563eb, #fb923c);
}

.metric-card__title,
.metric-card__value,
.metric-card__detail,
.metric-card__trend {
  position: relative;
  z-index: 1;
}

.metric-card__title {
  margin: 0 0 10px;
  font-size: 13px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--muted);
}

.metric-card__value {
  margin: 0 0 8px;
  font-size: 30px;
  line-height: 1;
  font-weight: 800;
  color: var(--text);
}

.metric-card__detail {
  margin: 0;
  font-size: 13px;
  color: var(--text);
  line-height: 1.45;
}

.metric-card__trend {
  margin: 10px 0 0;
  font-size: 12px;
  color: var(--muted);
}

.metric-card--clickable {
  cursor: pointer;
}
.metric-card--clickable:hover {
  transform: translateY(-3px);
  box-shadow: 0 20px 38px rgba(37, 99, 235, 0.1);
  border-color: rgba(37, 99, 235, 0.28);
}

[data-theme="dark"] .metric-card {
  border-color: rgba(91, 192, 190, 0.18);
  background:
    linear-gradient(160deg, rgba(24, 38, 57, 0.98), rgba(16, 27, 42, 0.94)),
    var(--surface);
  box-shadow: 0 18px 36px rgba(0, 0, 0, 0.28);
}

[data-theme="dark"] .metric-card__title {
  color: #8fb4c9;
}

[data-theme="dark"] .metric-card__value {
  color: #f5fbff;
  text-shadow: 0 1px 0 rgba(0, 0, 0, 0.2);
}

[data-theme="dark"] .metric-card__detail {
  color: #d8e8f2;
}

[data-theme="dark"] .metric-card__trend {
  color: #9fc3d7;
}

[data-theme="dark"] .metric-card--clickable:hover {
  box-shadow: 0 22px 42px rgba(0, 0, 0, 0.34);
  border-color: rgba(96, 165, 250, 0.34);
}
</style>
