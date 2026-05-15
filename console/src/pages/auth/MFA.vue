<template>
  <div class="auth-container">
    <t-card class="auth-card">
      <template #header>
        <div class="auth-header">
          <img :src="logoUrl" alt="Contful" class="auth-logo" />
          <h2 class="mfa-title">{{ title }}</h2>
          <p class="mfa-subtitle">{{ subtitle }}</p>
        </div>
      </template>

      <!-- 设置模式：扫码绑定 MFA -->
      <div v-if="mode === 'setup'" class="mfa-form">
        <div class="setup-qr">
          <div v-if="qrCodeUrl" class="qr-wrapper">
            <img :src="qrCodeUrl" alt="TOTP QR Code" class="qr-image" />
          </div>
          <div v-else class="qr-placeholder">
            <t-loading size="large" />
          </div>
        </div>

        <div class="setup-secret">
          <span class="secret-label">{{ t('auth.mfaSecret') }}</span>
          <code class="secret-value">{{ totpSecret }}</code>
        </div>

        <p class="setup-hint">{{ t('auth.mfaSetupHint') }}</p>

        <t-input
          v-model="setupCode"
          :placeholder="t('auth.mfaCodePlaceholder')"
          size="large"
          maxlength="6"
          inputmode="numeric"
          :autofocus="true"
        >
          <template #prefix-icon><t-icon name="lock-on" /></template>
        </t-input>

        <t-button
          theme="primary"
          size="large"
          block
          :loading="loading"
          :disabled="setupCode.length !== 6"
          style="margin-top: 16px"
          @click="onSetupVerify"
        >
          {{ loading ? t('auth.mfaVerifying') : t('auth.mfaEnableAndVerify') }}
        </t-button>
      </div>

      <!-- 验证模式：输入 TOTP 验证码 -->
      <div v-if="mode === 'verify' && !showRecovery" class="mfa-form">
        <t-input
          v-model="totpCode"
          :placeholder="t('auth.mfaCodePlaceholder')"
          size="large"
          maxlength="6"
          :autofocus="true"
        >
          <template #prefix-icon><t-icon name="lock-on" /></template>
        </t-input>

        <t-button
          theme="primary"
          size="large"
          block
          :loading="loading"
          :disabled="totpCode.length !== 6"
          style="margin-top: 16px"
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
      <div v-if="mode === 'verify' && showRecovery" class="mfa-form">
        <t-input
          v-model="recoveryCode"
          :placeholder="t('auth.mfaRecoveryPlaceholder')"
          size="large"
          :autofocus="true"
        >
          <template #prefix-icon><t-icon name="key" /></template>
        </t-input>

        <t-button
          theme="primary"
          size="large"
          block
          :loading="loading"
          :disabled="!recoveryCode.trim()"
          style="margin-top: 16px"
          @click="onRecover"
        >
          {{ loading ? t('auth.mfaVerifying') : t('auth.mfaRecoverySubmit') }}
        </t-button>

        <div class="mfa-links">
          <a href="#" @click.prevent="showRecovery = false">{{ t('auth.mfaBackToLogin') }}</a>
        </div>
      </div>

      <template #footer>
        <div class="auth-footer">
          <span class="auth-copyright">© 2026 Contful. Powered by <a href="https://reepu.com" target="_blank" rel="noopener">reepu.com</a></span>
        </div>
      </template>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import request, { setAccessToken, getAccessToken } from '@/utils/request'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const siteStore = useSiteStore()

const logoUrl = '/assets/logo.png'
const loading = ref(false)
const showRecovery = ref(false)
const totpCode = ref('')
const recoveryCode = ref('')

// 设置模式
const qrCodeUrl = ref('')
const totpSecret = ref('')
const setupCode = ref('')

// 模式判断：有 mfa_token → 验证模式；无 token 但已登录 → 设置模式
const mode = ref<'verify' | 'setup'>('verify')

const title = computed(() => {
  if (mode.value === 'setup') return t('auth.mfaSetupTitle')
  if (showRecovery.value) return t('auth.mfaRecoveryTitle')
  return t('auth.mfaVerifyTitle')
})

const subtitle = computed(() => {
  if (mode.value === 'setup') return t('auth.mfaSetupSubtitle')
  if (showRecovery.value) return t('auth.mfaRecoveryTip')
  return t('auth.mfaVerifySubtitle')
})

const handleLoginSuccess = async (data: { access_token: string; user: any }) => {
  setAccessToken(data.access_token)
  if (data.user) userStore.setUser(data.user)
  await siteStore.fetchSites()
  MessagePlugin.success(t('auth.mfaSuccess'))
  router.replace('/')
}

// ─── 设置模式：获取 TOTP 密钥和二维码 ─────────────────────────
const initSetup = async () => {
  loading.value = true
  try {
    const res = await request.post('/auth/mfa/setup')
    if (res.data.code === 200) {
      qrCodeUrl.value = res.data.data.qr_code_url || ''
      totpSecret.value = res.data.data.totp_secret || ''
    } else {
      MessagePlugin.error(res.data.msg || t('auth.mfaSetupFailed'))
      router.replace('/login')
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.msg || t('auth.mfaSetupFailed'))
    router.replace('/login')
  } finally {
    loading.value = false
  }
}

const onSetupVerify = async () => {
  if (setupCode.value.length !== 6) return
  loading.value = true
  try {
    const res = await request.post('/auth/mfa/enable', {
      totp_code: setupCode.value,
    })
    if (res.data.code === 200) {
      MessagePlugin.success(t('auth.mfaSetupSuccess'))
      await handleLoginSuccess(res.data.data)
    } else {
      MessagePlugin.error(res.data.msg || t('auth.mfaInvalidCode'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.msg || t('auth.mfaInvalidCode'))
  } finally {
    loading.value = false
  }
}

// ─── 验证模式 ──────────────────────────────────────────────
const mfaToken = ref('')

const onVerify = async () => {
  if (totpCode.value.length !== 6) return
  loading.value = true
  try {
    const res = await request.post('/auth/mfa/verify', {
      mfa_token: mfaToken.value,
      totp_code: totpCode.value,
    })
    if (res.data.code === 200) {
      sessionStorage.removeItem('mfa_token')
      sessionStorage.removeItem('mfa_email')
      await handleLoginSuccess(res.data.data)
    } else {
      MessagePlugin.error(res.data.msg || t('auth.mfaInvalidCode'))
    }
  } catch (e: any) {
    const msg = e.response?.data?.msg || t('auth.mfaInvalidCode')
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
    const email = sessionStorage.getItem('mfa_email') || ''
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
      sessionStorage.removeItem('mfa_token')
      sessionStorage.removeItem('mfa_email')
      await handleLoginSuccess(res.data.data)
    } else {
      MessagePlugin.error(res.data.msg || t('auth.mfaInvalidCode'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.msg || t('auth.mfaInvalidCode'))
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.replace('/login')
}

onMounted(async () => {
  mfaToken.value = sessionStorage.getItem('mfa_token') || ''

  if (mfaToken.value) {
    // 有 mfa_token → 验证模式
    mode.value = 'verify'
  } else if (getAccessToken()) {
    // 无 mfa_token 但已登录（有 JWT）→ 设置模式
    mode.value = 'setup'
    await initSetup()
  } else {
    // 都没 → 回登录
    router.replace('/login')
  }
})
</script>

<style scoped>
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
  margin-top: 12px;
}

.mfa-links a {
  font-size: 13px;
  color: var(--td-brand-color);
}

.setup-qr {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}

.qr-wrapper {
  border: 1px solid var(--td-component-stroke);
  border-radius: 8px;
  padding: 8px;
}

.qr-image {
  width: 200px;
  height: 200px;
  display: block;
}

.qr-placeholder {
  width: 200px;
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--td-component-stroke);
  border-radius: 8px;
}

.setup-secret {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.secret-label {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  white-space: nowrap;
}

.secret-value {
  font-family: monospace;
  font-size: 14px;
  background: var(--td-bg-color-component);
  padding: 4px 8px;
  border-radius: 4px;
  user-select: all;
}

.setup-hint {
  font-size: 13px;
  color: var(--td-text-color-placeholder);
  margin-bottom: 16px;
  text-align: center;
}
</style>
