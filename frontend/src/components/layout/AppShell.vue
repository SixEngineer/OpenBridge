<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import AppSidebar from './AppSidebar.vue'
import AppTopbar from './AppTopbar.vue'

const route = useRoute()
const contentRef = ref<HTMLElement | null>(null)
const resetTimers: number[] = []

function resetScrollPosition() {
  contentRef.value?.scrollTo({ top: 0, behavior: 'auto' })
  document.documentElement.scrollTop = 0
  document.body.scrollTop = 0
  window.scrollTo({ top: 0, behavior: 'auto' })
}

function clearResetTimers() {
  while (resetTimers.length > 0) {
    const timer = resetTimers.pop()
    if (timer !== undefined) {
      window.clearTimeout(timer)
    }
  }
}

function scheduleScrollReset() {
  clearResetTimers()
  resetScrollPosition()
  void nextTick(() => {
    resetScrollPosition()
    requestAnimationFrame(() => resetScrollPosition())
    ;[60, 180, 360, 720].forEach((delay) => {
      const timer = window.setTimeout(() => resetScrollPosition(), delay)
      resetTimers.push(timer)
    })
  })
}

onMounted(async () => {
  scheduleScrollReset()
})

watch(() => route.fullPath, async () => {
  scheduleScrollReset()
})

onBeforeUnmount(() => {
  clearResetTimers()
})
</script>

<template>
  <div class="shell">
    <AppSidebar />
    <div class="shell__body">
      <AppTopbar />
      <main ref="contentRef" class="shell__content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
