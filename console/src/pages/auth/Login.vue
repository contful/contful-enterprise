<template>
  <div class="auth-container">
    <t-card class="auth-card">
      <template #header>
        <div class="auth-header">
          <img
            :src="logoUrl"
            alt="Contful"
            class="auth-logo"
            role="heading"
            aria-level="1"
          />
          <p class="auth-subtitle">{{ t('auth.openSource') }} Headless CMS</p>
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
        <div class="auth-footer">
          <span class="auth-copyright">© 2026 Contful. Powered by <a href="https://reepu.com" target="_blank" rel="noopener">reepu.com</a></span>
        </div>
      </template>
    </t-card>
  </div>
</template>

<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

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
      // 使用 sessionStorage 传递敏感 token（不出现在 URL 中）
      sessionStorage.setItem('mfa_token', (result as any).mfa_token)
      sessionStorage.setItem('mfa_email', loginForm.email)
      router.push({
        path: '/mfa',
        // 不再通过 query 传递敏感信息
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
/* 仅页面特有样式，通用样式已抽取到 src/styles/auth.css */
</style>
