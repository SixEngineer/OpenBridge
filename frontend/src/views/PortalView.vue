<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { getDrivers } from '@/api/storage'

const router = useRouter()
const { t } = useI18n()
const isConnected = ref(false)
const isLaunching = ref(false)

onMounted(async () => {
  try {
    const res = await getDrivers()
    // 有驱动列表数据才算真正连接（OpenList 启动了但未登录时列表为空）
    isConnected.value = (res.code === 1000 || res.code === 0) && Array.isArray(res.data) && res.data.length > 0
  } catch {
    isConnected.value = false
  }
})

const buttonLabel = computed(() => {
  if (!isConnected.value) return t('portal.openlist_offline')
  if (isLaunching.value) return t('portal.opening_console')
  return t('portal.tap_to_enter')
})

function enterConsole() {
  if (!isConnected.value || isLaunching.value) {
    return
  }

  isLaunching.value = true

  window.setTimeout(() => {
    router.push('/dashboard')
  }, 900)
}
</script>

<template>
  <main class="portal" :class="{ 'portal--launching': isLaunching }">
    <div class="portal__grid"></div>
    <div class="portal__halo portal__halo--left"></div>
    <div class="portal__halo portal__halo--right"></div>

    <section class="portal__status-panel">
      <div class="status-chip" :class="{ 'status-chip--offline': !isConnected }">
        <span class="status-chip__dot"></span>
        {{ isConnected ? t('portal.openlist_connected') : t('portal.openlist_disconnected') }}
      </div>
    </section>

    <section class="hero">
      <p class="hero__eyebrow">{{ t('portal.eyebrow') }}</p>

      <button class="wordmark" type="button" @click="enterConsole">
        <span class="wordmark__open">Open</span>
        <span class="wordmark__bridge">Bridge</span>
      </button>

      <p class="hero__subtitle">{{ t('portal.subtitle') }}</p>

      <div class="hero__actions">
        <button
          class="enter-button"
          type="button"
          :disabled="!isConnected || isLaunching"
          @click="enterConsole"
        >
          {{ buttonLabel }}
        </button>
        <router-link to="/login" class="portal-login-link">{{ t('portal.sign_in') }}</router-link>
      </div>
    </section>
  </main>
</template>

<style scoped>
.portal {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  display: grid;
  grid-template-rows: auto 1fr;
  padding: 28px;
  isolation: isolate;
  color: #eff8ff;
  background:
    radial-gradient(circle at 20% 20%, rgba(103, 185, 255, 0.18), transparent 28%),
    radial-gradient(circle at 80% 30%, rgba(132, 255, 214, 0.15), transparent 26%),
    linear-gradient(145deg, #03131f 0%, #071c2c 52%, #0d2636 100%);
}

.portal__grid {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(rgba(255, 255, 255, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.06) 1px, transparent 1px);
  background-size: 72px 72px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.7), transparent 88%);
  opacity: 0.24;
  z-index: -3;
}

.portal__halo {
  position: absolute;
  border-radius: 999px;
  filter: blur(28px);
  opacity: 0.75;
  z-index: -2;
  animation: drift 12s ease-in-out infinite;
}

.portal__halo--left {
  width: 320px;
  height: 320px;
  left: -40px;
  top: 22%;
  background: rgba(115, 213, 255, 0.22);
}

.portal__halo--right {
  width: 420px;
  height: 420px;
  right: -120px;
  bottom: 5%;
  background: rgba(134, 255, 209, 0.18);
  animation-delay: -4s;
}

.portal__status-panel {
  display: flex;
  justify-content: flex-start;
  gap: 16px;
  align-items: center;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 999px;
  color: #f4fbff;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(18px);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.18);
}

.status-chip__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #86ffd1;
  box-shadow: 0 0 18px rgba(134, 255, 209, 0.8);
}

.status-chip--offline .status-chip__dot {
  background: #ff8a8a;
  box-shadow: 0 0 18px rgba(255, 138, 138, 0.72);
}

.hero {
  display: grid;
  place-items: center;
  text-align: center;
  gap: 18px;
  padding: 48px 0 56px;
}

.hero__eyebrow {
  margin: 0;
  font-size: 0.88rem;
  letter-spacing: 0.26em;
  text-transform: uppercase;
  color: #ffd67f;
}

.wordmark {
  margin: 0;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
  color: inherit;
  display: inline-flex;
  align-items: baseline;
  gap: 0.2em;
  overflow: visible;
  transition: transform 240ms ease, filter 240ms ease, opacity 240ms ease;
}

.wordmark:hover {
  transform: scale(1.02);
  filter: brightness(1.08);
}

.wordmark__open,
.wordmark__bridge {
  font-size: clamp(3.7rem, 12vw, 9rem);
  line-height: 1.06;
  font-weight: 800;
  letter-spacing: -0.035em;
}

.wordmark__open {
  color: rgba(255, 255, 255, 0.92);
  text-shadow: 0 0 42px rgba(255, 255, 255, 0.14);
}

.wordmark__bridge {
  background: linear-gradient(135deg, #73d5ff 0%, #86ffd1 45%, #ffffff 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  text-shadow: 0 0 56px rgba(115, 213, 255, 0.34);
  padding: 0 0.06em 0.14em 0;
}

.hero__subtitle {
  max-width: 820px;
  margin: 0;
  font-size: clamp(1rem, 2.4vw, 1.2rem);
  color: rgba(229, 242, 252, 0.72);
}

.hero__actions {
  display: grid;
  justify-items: center;
  gap: 14px;
}

.enter-button {
  min-width: 260px;
  padding: 14px 20px;
  border-radius: 999px;
  border: 1px solid rgba(115, 213, 255, 0.36);
  background: linear-gradient(135deg, rgba(115, 213, 255, 0.18), rgba(134, 255, 209, 0.16));
  color: #eefbff;
  box-shadow: 0 18px 34px rgba(16, 120, 162, 0.22);
  cursor: pointer;
  transition: transform 180ms ease, opacity 180ms ease, box-shadow 180ms ease;
}

.enter-button:hover:enabled {
  transform: translateY(-2px);
  box-shadow: 0 26px 44px rgba(16, 120, 162, 0.28);
}

.enter-button:disabled {
  cursor: not-allowed;
  opacity: 0.42;
}

.portal--launching .wordmark {
  transform: scale(1.03);
  filter: brightness(1.18);
}

.portal--launching .enter-button {
  animation: pulse 0.9s ease forwards;
}

@keyframes drift {
  0%,
  100% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(0, -24px, 0);
  }
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 rgba(115, 213, 255, 0.18);
  }
  100% {
    box-shadow: 0 0 0 26px rgba(115, 213, 255, 0);
  }
}

.portal-login-link {
  display: inline-block;
  margin-top: 16px;
  color: rgba(229, 242, 252, 0.6);
  font-size: 14px;
  transition: color 0.2s;
}

.portal-login-link:hover {
  color: rgba(229, 242, 252, 0.9);
}

@media (max-width: 920px) {
  .portal {
    padding: 18px;
  }

  .portal__status-panel {
    display: grid;
  }

  .wordmark {
    flex-direction: column;
    align-items: center;
    gap: 0;
  }
}
</style>
