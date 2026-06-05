<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { userLogin } from '@/api/user'
import { useConsoleStore } from '@/stores/console'

const router = useRouter()
const store = useConsoleStore()
const { t } = useI18n()

const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  if (!username.value.trim() || !password.value.trim()) {
    errorMsg.value = t('login.error_empty')
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const res = await userLogin({
      username: username.value.trim(),
      password: password.value,
    })
    if (res.code === 1000) {
      store.login(username.value.trim())
      router.push('/dashboard')
    } else {
      errorMsg.value = (res.msg as string) || t('login.error_failed')
    }
  } catch (e: any) {
    errorMsg.value = e?.message || t('common.request_error')
  } finally {
    loading.value = false
  }
}
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
          />
        </label>
        <label>
          {{ t('login.password') }}
          <input
            v-model="password"
            type="password"
            placeholder="&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;"
            autocomplete="current-password"
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
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
}

.login-card {
  width: 380px;
  padding: 40px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.15);
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
  color: #111827;
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
  color: #374151;
}

.login-form input {
  padding: 10px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.login-form input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
}

.login-form__error {
  margin: 0;
  padding: 10px 14px;
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
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
</style>
