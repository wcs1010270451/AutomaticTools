<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { ArrowRight, Hide, Lock, User, View } from '@element-plus/icons-vue'
import { useAdminAuthStore } from '../stores/adminAuth'

const route = useRoute()
const router = useRouter()
const auth = useAdminAuthStore()

const username = ref('admin')
const password = ref('')
const passwordVisible = ref(false)
const loading = ref(false)
const errorMessage = ref('')

async function submit() {
  if (!username.value.trim() || !password.value) {
    errorMessage.value = '请输入管理员账号和密码。'
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    await auth.login(username.value, password.value)
    const redirect = typeof route.query.redirect === 'string'
      && route.query.redirect.startsWith('/')
      && !route.query.redirect.startsWith('//')
      ? route.query.redirect
      : '/'
    await router.replace(redirect)
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const message = error.response?.data?.error
      errorMessage.value = typeof message === 'string'
        ? message
        : '无法连接管理服务，请稍后重试。'
    } else {
      errorMessage.value = '登录失败，请稍后重试。'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-brand" aria-label="AutomaticTools 管理后台">
      <div class="login-brand-content">
        <div class="login-logo">
          <span class="login-logo-mark">A</span>
          <div>
            <strong>AutomaticTools</strong>
            <span>管理后台</span>
          </div>
        </div>

        <div class="login-intro">
          <p class="login-eyebrow">OPERATIONS CONSOLE</p>
          <h1>让每一次管理操作清晰可控。</h1>
          <p>统一处理用户、订单与工具授权，关键操作全程留痕。</p>
        </div>

        <p class="login-security">
          <Lock aria-hidden="true" />
          管理员身份验证
        </p>
      </div>
    </section>

    <section class="login-panel">
      <div class="login-form-wrap">
        <header class="login-heading">
          <p>欢迎回来</p>
          <h2>登录管理后台</h2>
          <span>请输入管理员账号继续。</span>
        </header>

        <form class="login-form" novalidate @submit.prevent="submit">
          <label for="admin-username">管理员账号</label>
          <div class="login-field">
            <User aria-hidden="true" />
            <input
              id="admin-username"
              v-model="username"
              name="username"
              type="text"
              autocomplete="username"
              autofocus
              :disabled="loading"
              placeholder="请输入管理员账号"
            >
          </div>

          <label for="admin-password">密码</label>
          <div class="login-field">
            <Lock aria-hidden="true" />
            <input
              id="admin-password"
              v-model="password"
              name="password"
              :type="passwordVisible ? 'text' : 'password'"
              autocomplete="current-password"
              :disabled="loading"
              placeholder="请输入密码"
            >
            <button
              class="password-toggle"
              type="button"
              :aria-label="passwordVisible ? '隐藏密码' : '显示密码'"
              :disabled="loading"
              @click="passwordVisible = !passwordVisible"
            >
              <Hide v-if="passwordVisible" aria-hidden="true" />
              <View v-else aria-hidden="true" />
            </button>
          </div>

          <p v-if="errorMessage" class="login-error" role="alert">
            {{ errorMessage }}
          </p>

          <button class="login-submit" type="submit" :disabled="loading">
            <span class="submit-icon" aria-hidden="true">
              <span v-if="loading" class="loading-indicator" />
            </span>
            <span>{{ loading ? '正在登录' : '登录' }}</span>
            <span class="submit-icon" aria-hidden="true">
              <ArrowRight v-if="!loading" />
            </span>
          </button>
        </form>

        <p class="login-footnote">仅限授权管理员访问</p>
      </div>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  display: grid;
  grid-template-columns: minmax(360px, 0.9fr) minmax(480px, 1.1fr);
  min-height: 100vh;
  background: oklch(98.5% 0.004 150);
}

.login-brand {
  position: relative;
  display: flex;
  min-height: 100vh;
  padding: clamp(28px, 4vw, 56px);
  overflow: hidden;
  color: oklch(96% 0.008 155);
  background:
    radial-gradient(circle at 12% 85%, oklch(46% 0.09 155 / 0.52), transparent 38%),
    oklch(27% 0.038 165);
}

.login-brand::after {
  position: absolute;
  top: 12%;
  right: -96px;
  width: 270px;
  height: 270px;
  border: 1px solid oklch(78% 0.06 155 / 0.18);
  border-radius: 50%;
  box-shadow:
    0 0 0 52px oklch(78% 0.06 155 / 0.045),
    0 0 0 104px oklch(78% 0.06 155 / 0.025);
  content: "";
}

.login-brand-content {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1;
  flex-direction: column;
  max-width: 560px;
}

.login-logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.login-logo-mark {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  color: oklch(27% 0.038 165);
  background: oklch(82% 0.14 86);
  border-radius: 7px;
  font-size: 20px;
  font-weight: 800;
}

.login-logo div {
  display: grid;
  gap: 2px;
}

.login-logo strong {
  font-size: 15px;
}

.login-logo span:last-child {
  color: oklch(76% 0.014 160);
  font-size: 12px;
}

.login-intro {
  margin-block: auto;
}

.login-eyebrow {
  margin: 0 0 20px;
  color: oklch(82% 0.14 86);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.16em;
}

.login-intro h1 {
  max-width: 12ch;
  margin: 0;
  color: oklch(98% 0.004 150);
  font-size: clamp(34px, 4vw, 52px);
  line-height: 1.18;
  letter-spacing: -0.035em;
}

.login-intro > p:last-child {
  max-width: 33ch;
  margin: 24px 0 0;
  color: oklch(78% 0.018 160);
  font-size: 15px;
  line-height: 1.8;
}

.login-security {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: oklch(72% 0.014 160);
  font-size: 12px;
}

.login-security svg {
  width: 15px;
  height: 15px;
}

.login-panel {
  display: grid;
  place-items: center;
  min-height: 100vh;
  padding: 48px clamp(28px, 7vw, 104px);
}

.login-form-wrap {
  width: min(100%, 400px);
}

.login-heading {
  margin-bottom: 36px;
}

.login-heading p,
.login-heading h2,
.login-heading span {
  display: block;
  margin: 0;
}

.login-heading p {
  margin-bottom: 9px;
  color: oklch(43% 0.078 165);
  font-size: 13px;
  font-weight: 700;
}

.login-heading h2 {
  color: oklch(25% 0.018 165);
  font-size: 28px;
  line-height: 1.2;
  letter-spacing: -0.025em;
}

.login-heading span {
  margin-top: 11px;
  color: oklch(53% 0.018 165);
  font-size: 14px;
}

.login-form {
  display: grid;
}

.login-form > label {
  margin: 0 0 9px;
  color: oklch(34% 0.018 165);
  font-size: 13px;
  font-weight: 700;
}

.login-form > label:not(:first-child) {
  margin-top: 22px;
}

.login-field {
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr) auto;
  align-items: center;
  min-height: 48px;
  padding: 0 14px;
  background: oklch(99% 0.003 150);
  border: 1px solid oklch(84% 0.016 158);
  border-radius: 7px;
  transition:
    border-color 180ms ease-out,
    box-shadow 180ms ease-out,
    background 180ms ease-out;
}

.login-field:focus-within {
  background: oklch(99.5% 0.003 150);
  border-color: oklch(51% 0.085 165);
  box-shadow: 0 0 0 3px oklch(74% 0.08 160 / 0.18);
}

.login-field > svg {
  width: 17px;
  height: 17px;
  color: oklch(55% 0.022 165);
}

.login-field input {
  width: 100%;
  min-width: 0;
  height: 46px;
  padding: 0 10px;
  color: oklch(25% 0.018 165);
  background: transparent;
  border: 0;
  outline: none;
  font: inherit;
  font-size: 14px;
}

.login-field input::placeholder {
  color: oklch(64% 0.012 165);
}

.login-field input:disabled {
  cursor: wait;
}

.password-toggle {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  padding: 0;
  color: oklch(52% 0.02 165);
  background: transparent;
  border: 0;
  border-radius: 5px;
  cursor: pointer;
}

.password-toggle:hover {
  color: oklch(36% 0.065 165);
  background: oklch(94% 0.025 158);
}

.password-toggle:focus-visible {
  outline: 2px solid oklch(51% 0.085 165);
  outline-offset: 2px;
}

.password-toggle svg {
  width: 17px;
  height: 17px;
}

.login-error {
  margin: 16px 0 0;
  padding: 10px 12px;
  color: oklch(42% 0.13 28);
  background: oklch(95% 0.035 28);
  border: 1px solid oklch(82% 0.065 28);
  border-radius: 6px;
  font-size: 13px;
  line-height: 1.55;
}

.login-submit {
  display: grid;
  grid-template-columns: 18px auto 18px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 48px;
  margin-top: 28px;
  padding: 0 18px;
  color: oklch(98% 0.004 150);
  background: oklch(38% 0.075 165);
  border: 0;
  border-radius: 7px;
  cursor: pointer;
  font-weight: 700;
  transition:
    background 180ms ease-out,
    transform 180ms ease-out,
    box-shadow 180ms ease-out;
}

.login-submit:hover:not(:disabled) {
  background: oklch(33% 0.075 165);
  box-shadow: 0 8px 20px oklch(30% 0.05 165 / 0.2);
  transform: translateY(-1px);
}

.login-submit:active:not(:disabled) {
  transform: translateY(0);
}

.login-submit:focus-visible {
  outline: 3px solid oklch(69% 0.09 160 / 0.36);
  outline-offset: 3px;
}

.login-submit:disabled {
  cursor: wait;
  opacity: 0.72;
}

.login-submit svg {
  width: 17px;
  height: 17px;
}

.submit-icon {
  display: grid;
  place-items: center;
  width: 18px;
  height: 18px;
}

.loading-indicator {
  width: 16px;
  height: 16px;
  border: 2px solid oklch(98% 0.004 150 / 0.35);
  border-top-color: oklch(98% 0.004 150);
  border-radius: 50%;
  animation: login-spin 700ms linear infinite;
}

.login-footnote {
  margin: 24px 0 0;
  color: oklch(59% 0.014 165);
  font-size: 12px;
  text-align: center;
}

@keyframes login-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 820px) {
  .login-page {
    grid-template-columns: 1fr;
  }

  .login-brand {
    min-height: auto;
    padding: 22px 24px;
  }

  .login-brand::after,
  .login-intro,
  .login-security {
    display: none;
  }

  .login-panel {
    min-height: calc(100vh - 82px);
    padding: 42px 24px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .login-field,
  .login-submit {
    transition: none;
  }
}
</style>
