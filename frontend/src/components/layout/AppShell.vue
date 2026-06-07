<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppSidebar from './AppSidebar.vue'
import AppTopbar from './AppTopbar.vue'
</script>

<template>
  <div class="shell">
    <AppSidebar />
    <div class="shell__body">
      <AppTopbar />
      <main class="shell__content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
