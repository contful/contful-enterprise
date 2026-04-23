<template>
  <div class="login-container">
    <t-card class="login-card">
      <template #header>
        <div class="login-header">
          <img
            :src="logoUrl"
            alt="Contful"
            class="login-logo"
            role="heading"
            aria-level="1"
          />
          <p class="login-subtitle">{{ t('auth.openSource') }} Headless CMS</p>
        </div>
      </template>

      <t-form
        :data="loginForm"
        :required-mark="false"
        :rules="loginRules"
        label-width="0"
        label-align="top"
        ref="loginFormRef"
        @submit="onLogin"
      >
        <t-form-item name="email">
          <t-input
            v-model="loginForm.email"
            :placeholder="t('auth.email')"
            size="large"
          >
            <template #prefix-icon>
              <Icon name="mail" />
            </template>
          </t-input>
        </t-form-item>

        <t-form-item name="password">
          <t-input
            v-model="loginForm.password"
            type="password"
            :placeholder="t('auth.password')"
            size="large"
            autocomplete="current-password"
          >
            <template #prefix-icon>
              <Icon name="lock-on" />
            </template>
          </t-input>
        </t-form-item>

        <t-form-item>
          <t-button
            type="submit"
            theme="primary"
            size="large"
            block
            :loading="userStore.isLoading"
          >
            {{ t('auth.login') }}
          </t-button>
        </t-form-item>
      </t-form>

      <template #footer>
        <div class="login-footer">
          <span class="copyright">© 2026 Contful. Powered by <a href="https://reepu.com" target="_blank" rel="noopener">reepu</a></span>
        </div>
      </template>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

const loginFormRef = ref()

const loginForm = reactive({
  email: '',
  password: '',
})

const loginRules = {
  email: [
    { required: true, message: t('auth.enterEmail') },
    { email: true, message: t('auth.invalidEmail') },
  ],
  password: [
    { required: true, message: t('auth.enterPassword') },
    { min: 8, message: t('auth.passwordMinLength', { min: 8 }) },
  ],
}

const logoUrl = '/assets/logo.png'

const onLogin = async () => {
  const result = await userStore.login(loginForm.email, loginForm.password)
  if (result.success) {
    if ((result as any).mfa_required) {
      // 跳转 MFA 验证页，携带 mfa_token 和 email
      router.push({
        path: '/mfa',
        query: {
          mfa_token: (result as any).mfa_token,
          email: loginForm.email,
        },
      })
    } else {
      MessagePlugin.success(t('auth.loginSuccess'))
      router.push('/')
    }
  } else {
    MessagePlugin.error((result as any).message || t('auth.loginFailed'))
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(200,240,255,0.7) 0%, rgba(180,240,200,0.6) 50%, rgba(220,255,210,0.7) 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 420px;
}

.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 20px 0;
  width: 100%;
}

.login-logo {
  height: 48px;
  width: auto;
  object-fit: contain;
}

.login-subtitle {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin: 0;
}

.login-footer {
  text-align: center;
  padding: 16px 0;
}

.copyright {
  font-size: 12px;
  color: var(--td-text-color-secondary);
}

.copyright a {
  color: var(--td-brand-color);
  text-decoration: none;
}
</style>
