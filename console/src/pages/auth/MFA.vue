<template>
  <div class="login-container">
    <t-card class="login-card">
      <template #header>
        <div class="login-header">
          <img :src="logoUrl" alt="Contful" class="login-logo" />
          <h2 class="mfa-title">
            {{ showRecovery ? t('auth.mfaRecoveryTitle') : t('auth.mfaVerifyTitle') }}
          </h2>
          <p class="mfa-subtitle">
            {{ showRecovery ? t('auth.mfaRecoveryTip') : t('auth.mfaVerifySubtitle') }}
          </p>
        </div>
      </template>

      <!-- TOTP 验证 -->
      <div v-if="!showRecovery" class="mfa-form">
        <t-input
          v-model="totpCode"
          :placeholder="t('auth.mfaCodePlaceholder')"
          size="large"
          maxlength="6"
          :autofocus="true"
          inputmode="numeric"
          pattern="[0-9]*"
          @keyup.enter="onVerify"
        >
          <template #prefix-icon>
            <t-icon name="lock-on" />
          </template>
        </t-input>

        <t-button
          theme="primary"
          size="large"
          block
          :loading="loading"
          :disabled="totpCode.length !== 6"
          style="margin-top: 16px;"
          @click="onVerify"
        >
          {{ loading ? t('auth.mfaVerifying') : t('auth.mfaVerifyBtn') }}
        </t-button>

        <div class="mfa-links">
          <a href="#" @click.prevent="showRecovery = true">{{ t('auth.mfaUseRecovery') }}</a>
          <a href="#" @click.prevent="goBack">{{ t('auth.mfaBackToLogin') }}</a>
        </div>
      </div>

      <!-- Recovery Code 恢复 -->
      <div v-else class="mfa-form">
        <t-input
          v-model="recoveryCode"
          :placeholder="t('auth.mfaRecoveryPlaceholder')"
          size="large"
          :autofocus="true"
          @keyup.enter="onRecover"
        >
          <template #prefix-icon>
            <t-icon name="key" />
          </template>
        </t-input>

        <t-button
          theme="primary"
          size="large"
          block
          :loading="loading"
          :disabled="!recoveryCode.trim()"
          style="margin-top: 16px;"
          @click="onRecover"
        >
          {{ loading ? t('auth.mfaVerifying') : t('auth.mfaRecoverySubmit') }}
        </t-button>

        <div class="mfa-links">
          <a href="#" @click.prevent="showRecovery = false">{{ t('auth.mfaBackToLogin') }}</a>
        </div>
      </div>

      <template #footer>
        <div class="login-footer">
          <span class="copyright">© 2026 Contful. Powered by <a href="https://reepu.com" target="_blank" rel="noopener">reepu</a></span>
        </div>
      </template>
    </t-card>
  </div>
</template>

<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import request, { setAccessToken, setRefreshToken } from '@/utils/request'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const siteStore = useSiteStore()

const logoUrl = '/assets/logo.png'
const loading = ref(false)
const showRecovery = ref(false)
const totpCode = ref('')
const recoveryCode = ref('')

// mfa_token 从路由 query 参数获取
const mfaToken = ref('')

onMounted(() => {
  // 从 sessionStorage 读取敏感信息（避免出现在 URL 中）
  mfaToken.value = sessionStorage.getItem('mfa_token') || ''
  const email = sessionStorage.getItem('mfa_email') || ''
  
  if (!mfaToken.value) {
    // 没有 mfa_token，跳回登录
    router.replace('/login')
  }
  
  // 清理：MFA 验证完成后应清除 sessionStorage
})

// Recovery Code 恢复需要 email
const getEmailFromSession = () => {
  return sessionStorage.getItem('mfa_email') || ''
}

const handleLoginSuccess = async (data: { access_token: string; refresh_token: string; user: any }) => {
  setAccessToken(data.access_token)
  setRefreshToken(data.refresh_token)
  if (data.user) {
    userStore.setUser(data.user)
  }
  await siteStore.fetchSites()
  MessagePlugin.success(t('auth.mfaSuccess'))
  router.replace('/')
}

const onVerify = async () => {
  if (totpCode.value.length !== 6) return
  loading.value = true
  try {
    const res = await request.post('/auth/mfa/verify', {
      mfa_token: mfaToken.value,
      totp_code: totpCode.value,
    })
    if (res.data.code === 200) {
      await handleLoginSuccess(res.data.data)
    } else {
      MessagePlugin.error(res.data.message || t('auth.mfaInvalidCode'))
    }
  } catch (e: any) {
    const msg = e.response?.data?.message || t('auth.mfaInvalidCode')
    if (e.response?.status === 401) {
      MessagePlugin.error(t('auth.mfaTokenExpired'))
      router.replace('/login')
    } else {
      MessagePlugin.error(msg)
    }
  } finally {
    loading.value = false
  }
}

const onRecover = async () => {
  if (!recoveryCode.value.trim()) return
  loading.value = true
  try {
    // Recovery Code 恢复需要 email（从 sessionStorage 读取）
    const email = getEmailFromSession()
    if (!email) {
      MessagePlugin.error(t('auth.mfaEmailMissing'))
      router.replace('/login')
      return
    }
    const res = await request.post('/auth/mfa/recover', {
      email,
      recovery_code: recoveryCode.value.trim(),
    })
    if (res.data.code === 200) {
      // 验证成功，清除 sessionStorage
      sessionStorage.removeItem('mfa_token')
      sessionStorage.removeItem('mfa_email')
      await handleLoginSuccess(res.data.data)
    } else {
      MessagePlugin.error(res.data.message || t('auth.mfaInvalidCode'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.message || t('auth.mfaInvalidCode'))
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.replace('/login')
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
  gap: 12px;
  padding: 20px 0 8px;
}

.login-logo {
  height: 40px;
  width: auto;
  object-fit: contain;
}

.mfa-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin: 0;
}

.mfa-subtitle {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin: 0;
  text-align: center;
}

.mfa-form {
  padding: 8px 0;
}

.mfa-links {
  display: flex;
  justify-content: space-between;
  margin-top: 16px;
}

.mfa-links a {
  font-size: 13px;
  color: var(--td-brand-color);
  text-decoration: none;
  cursor: pointer;
}

.mfa-links a:hover {
  text-decoration: underline;
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
