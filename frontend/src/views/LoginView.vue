<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getSessionStatus, userLogin } from '@/api/user'
import { useConsoleStore } from '@/stores/console'

const router = useRouter()
const store = useConsoleStore()
const { t } = useI18n()

const username = ref('')
const password = ref('')
const passwordInput = ref<HTMLInputElement | null>(null)
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  if (loading.value) {
    return
  }

  const name = username.value.trim()
  const pass = password.value

  if (!name || !pass) {
    if (name && !pass) {
      passwordInput.value?.focus()
      return
    }
    errorMsg.value = t('login.error_empty')
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const res = await userLogin({ username: name, password: pass })
    if (res.code === 1000) {
      const sessionStatus = await getSessionStatus()
      if (!sessionStatus.data.authenticated) {
        throw new Error(t('login.error_session'))
      }
      store.login(sessionStatus.data.username || name, sessionStatus.data)
      await store.fetchCurrentUser()
      router.push('/dashboard')
    } else {
      errorMsg.value = (res.msg as string) || t('login.error_failed')
    }
  } catch (e: any) {
    errorMsg.value = formatLoginError(e?.message || '')
  } finally {
    loading.value = false
  }
}

function formatLoginError(message: string) {
  if (message.startsWith('device_limit_reached:')) {
    const matched = message.match(/(\d+)/)
    return t('login.error_device_limit', { count: matched ? Number(matched[1]) : 5 })
  }
  if (message.includes('Too many unsuccessful sign-in attempts')) {
    return t('login.error_too_many_attempts')
  }
  return message || t('common.request_error')
}

function mapLogoutReason(reason: string) {
  switch (reason) {
    case 'expired':
      return t('login.reason_expired')
    case 'openlist_token_missing':
    case 'openlist_auth_invalid':
      return t('login.reason_openlist')
    case 'openlist_base_url_missing':
    case 'source_changed':
      return t('login.reason_source_changed')
    case 'device_session_missing':
      return t('login.reason_backend_changed')
    case 'session_changed':
      return t('login.reason_session_changed')
    case 'manual':
      return ''
    default:
      return reason ? t('login.reason_backend_changed') : ''
  }
}

onMounted(() => {
  const reason = store.consumeLogoutReason()
  if (reason) {
    errorMsg.value = mapLogoutReason(reason)
  }
})
</script>

<template>
  <section class="login-page">
    <div class="login-card">
      <p class="login-card__eyebrow">{{ t('app.name') }}</p>
      <h1 class="login-card__title">{{ t('login.title') }}</h1>

      <form class="login-form" @submit.prevent="handleLogin">
        <label>
          {{ t('login.username') }}
          <input
            v-model="username"
            type="text"
            :placeholder="t('login.username_placeholder')"
            autocomplete="username"
            autofocus
            @keydown.enter="passwordInput?.focus()"
          />
        </label>
        <label>
          {{ t('login.password') }}
          <input
            v-model="password"
            type="password"
            placeholder="&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;"
            autocomplete="current-password"
            ref="passwordInput"
          />
        </label>

        <p v-if="errorMsg" class="login-form__error">{{ errorMsg }}</p>

        <button type="submit" :disabled="loading">
          {{ loading ? t('login.submitting') : t('login.submit') }}
        </button>
      </form>
    </div>
  </section>
</template>

<style scoped>
.login-page {
  height: 100vh;
  height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
}

.login-card {
  width: 380px;
  padding: 40px;
  background: var(--surface);
  border-radius: 16px;
  box-shadow: var(--shadow);
}

.login-card__eyebrow {
  font-size: 12px;
  font-weight: 600;
  color: #3b82f6;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin: 0 0 8px;
}

.login-card__title {
  font-size: 22px;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 24px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.login-form label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
}

.login-form input {
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  background: var(--surface);
  color: var(--text);
  transition: border-color 0.2s;
}

.login-form input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
}

.login-form__error {
  margin: 0;
  padding: 10px 14px;
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 8px;
  font-size: 13px;
}

.login-form button {
  padding: 12px 20px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.login-form button:hover:not(:disabled) {
  background: #2563eb;
}

.login-form button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 860px) {
  .login-page {
    padding: 16px;
    overflow: hidden;
  }

  .login-card {
    width: 100%;
    max-width: 100%;
    padding: 28px 24px;
  }

  .login-card__title {
    font-size: 20px;
  }
}
</style>
