<template>
  <div class="auth-container" :style="loginBackgroundUrl ? { backgroundImage: `url(${loginBackgroundUrl})` } : undefined">
    <t-card class="auth-card">
      <template #header>
        <div class="auth-header">
          <img
            :src="logoUrl"
            :alt="siteName"
            class="auth-logo"
            role="heading"
            aria-level="1"
          />
          <p class="auth-subtitle">{{ siteDescription }}</p>
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

    <!-- 密码过期强制修改弹窗 -->
    <t-dialog
      v-model:visible="passwordExpiredDialogVisible"
      :header="t('auth.passwordExpiredTitle')"
      :width="480"
      :close-on-overlay-click="false"
      :close-on-esc-keydown="false"
      :show-overlay="true"
      :footer="false"
    >
      <div class="password-expired-content">
        <t-icon name="error-circle-filled" size="48px" color="var(--td-error-color)" />
        <h3>{{ t('auth.passwordExpired') }}</h3>
        <p>{{ t('auth.passwordExpiredHint', { days: passwordExpireDays }) }}</p>

        <t-form :data="passwordForm" label-align="top" :rules="passwordRules" ref="passwordFormRef">
          <t-form-item :label="t('auth.newPassword')" name="newPassword">
            <t-input v-model="passwordForm.newPassword" type="password" :placeholder="t('auth.enterPassword')" clearable />
            <!-- 密码强度条 -->
            <div class="password-strength">
              <div class="strength-bar">
                <div class="strength-fill" :class="passwordStrength.level" :style="{ width: passwordStrength.width }"></div>
              </div>
              <span class="strength-text" :class="passwordStrength.level">{{ passwordStrength.label }}</span>
            </div>
          </t-form-item>

          <t-form-item :label="t('auth.confirmPassword')" name="confirmPassword">
            <t-input v-model="passwordForm.confirmPassword" type="password" :placeholder="t('auth.confirmPasswordHint')" clearable />
          </t-form-item>

          <t-form-item>
            <t-button theme="primary" block :loading="changingPassword" @click="handleChangePassword">
              {{ t('auth.changePassword') }}
            </t-button>
          </t-form-item>
        </t-form>
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import JSEncrypt from 'jsencrypt'
import { useUserStore } from '@/stores/user'
import { updatePassword } from '@/api/user'
import { getSiteConfig } from '@/api/system-config'
import { get } from '@/utils/request'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

// @ts-expect-error template ref
const loginFormRef = ref()
// @ts-expect-error template ref
const passwordFormRef = ref()

const loginForm = reactive({
  email: '',
  password: '',
})

const logoUrl = ref('/assets/logo.png')
const siteName = ref('Contful')
const siteDescription = ref(t('auth.openSource') + ' Headless CMS')
const loginBackgroundUrl = ref('')
const mfaEnforced = ref(false)
const loginMaxAttempts = ref(5)
const loginLockDuration = ref(30)
const passwordPolicy = ref({
  min_length: 8,
  require_uppercase: true,
  require_lowercase: true,
  require_number: true,
  require_special: false,
  expire_days: 90,
})

// 组件挂载时获取站点配置（品牌 + 策略 + 密码规则）
onMounted(async () => {
  try {
    const config = await getSiteConfig()
    if (config.site_name) siteName.value = config.site_name
    if (config.site_description) siteDescription.value = config.site_description
    if (config.logo_url) logoUrl.value = config.logo_url
    if (config.login_background_url) loginBackgroundUrl.value = config.login_background_url
    mfaEnforced.value = config.mfa_enforced
    loginMaxAttempts.value = config.login_max_attempts
    loginLockDuration.value = config.login_lock_duration
    passwordPolicy.value = {
      min_length: config.password_min_length,
      require_uppercase: config.password_require_uppercase,
      require_lowercase: config.password_require_lowercase,
      require_number: config.password_require_number,
      require_special: config.password_require_special,
      expire_days: config.password_expire_days,
    }
  } catch (error) {
    console.warn('Failed to load site config, using defaults', error)
  }
})

const loginRules = computed(() => ({
  email: [
    { required: true, message: t('auth.enterEmail') },
    { email: true, message: t('auth.invalidEmail') },
  ],
  password: [
    { required: true, message: t('auth.enterPassword') },
    { min: passwordPolicy.value.min_length, message: t('auth.passwordMinLength', { min: passwordPolicy.value.min_length }) },
  ],
}))

// 密码过期相关状态
const passwordExpiredDialogVisible = ref(false)
const passwordExpireDays = ref(0)
const changingPassword = ref(false)
const passwordForm = reactive({
  newPassword: '',
  confirmPassword: '',
})

const passwordRules = computed(() => ({
  newPassword: [
    { required: true, message: t('auth.enterPassword') },
    { min: passwordPolicy.value.min_length, message: t('auth.passwordMinLength', { min: passwordPolicy.value.min_length }) },
  ],
  confirmPassword: [
    { required: true, message: t('auth.confirmPasswordHint') },
  ],
}))

// 密码强度计算（使用动态策略）
const passwordStrength = computed(() => {
  const pwd = passwordForm.newPassword
  if (!pwd) return { level: '', width: '0%', label: '' }

  let score = 0
  if (pwd.length >= passwordPolicy.value.min_length) score++
  if (pwd.length >= passwordPolicy.value.min_length + 4) score++
  if (passwordPolicy.value.require_lowercase && /[a-z]/.test(pwd)) score++
  if (passwordPolicy.value.require_uppercase && /[A-Z]/.test(pwd)) score++
  if (passwordPolicy.value.require_number && /[0-9]/.test(pwd)) score++
  if (passwordPolicy.value.require_special && /[^a-zA-Z0-9]/.test(pwd)) score++

  if (score < 3) return { level: 'weak', width: '33%', label: t('users.passwordWeak') }
  if (score < 5) return { level: 'medium', width: '66%', label: t('users.passwordMedium') }
  return { level: 'strong', width: '100%', label: t('users.passwordStrong') }
})

// 密码强度检查（使用动态策略）
const checkPasswordStrength = (pwd: string): boolean => {
  if (pwd.length < passwordPolicy.value.min_length) return false
  if (passwordPolicy.value.require_lowercase && !/[a-z]/.test(pwd)) return false
  if (passwordPolicy.value.require_uppercase && !/[A-Z]/.test(pwd)) return false
  if (passwordPolicy.value.require_number && !/[0-9]/.test(pwd)) return false
  return true
}

const onLogin = async () => {
  // 1. 获取 RSA 公钥 + Anti-Replay Token
  let encryptedPassword = ''
  let tokenId = ''
  let rsaToken = ''
  try {
    const keyRes: any = await get('/auth/public/key')
    if (keyRes.code === 200 && keyRes.data?.public_key) {
      const { public_key, token_id, token } = keyRes.data
      tokenId = token_id
      rsaToken = token
      // 2. RSA 加密「密码@@token」
      const encrypt = new JSEncrypt()
      encrypt.setPublicKey(public_key)
      const plaintext = loginForm.password + '@@' + token
      encryptedPassword = encrypt.encrypt(plaintext) || ''
    }
  } catch {
    // 降级：加密失败时使用明文（保持兼容）
  }

  const result: any = await userStore.login(
    loginForm.email,
    loginForm.password,
    encryptedPassword,
    tokenId,
    rsaToken,
  )
  if (result.success) {
    if (result.mfa_required) {
      // 使用 sessionStorage 传递敏感 token（不出现在 URL 中）
      sessionStorage.setItem('mfa_token', result.mfa_token)
      sessionStorage.setItem('mfa_email', loginForm.email)
      router.push({
        path: '/mfa',
        // 不再通过 query 传递敏感信息
      })
    } else if (result.mfa_setup_required) {
      // MFA 强制开启但用户未设置，跳转到 MFA 设置页
      sessionStorage.removeItem('mfa_token')
      sessionStorage.removeItem('mfa_email')
      router.push('/mfa')
    } else if (result.password_expired) {
      // 密码已过期，强制修改密码
      MessagePlugin.warning(t('auth.passwordExpired'))
      passwordExpireDays.value = result.password_expire_days || 0
      passwordExpiredDialogVisible.value = true
    } else {
      MessagePlugin.success(t('auth.loginSuccess'))
      router.push('/')
    }
  } else {
    MessagePlugin.error(result.message || t('auth.loginFailed'))
  }
}

const handleChangePassword = async () => {
  if (!passwordForm.newPassword) {
    MessagePlugin.error(t('auth.enterPassword'))
    return
  }
  if (!checkPasswordStrength(passwordForm.newPassword)) {
    MessagePlugin.error(t('auth.passwordWeak'))
    return
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    MessagePlugin.error(t('auth.passwordMismatch'))
    return
  }

  changingPassword.value = true
  try {
    // 获取当前用户 ID
    const userId = userStore.user?.id
    if (!userId) {
      MessagePlugin.error(t('auth.sessionExpired'))
      passwordExpiredDialogVisible.value = false
      return
    }

    // 调用修改密码 API（需要旧密码，但这里密码已过期，应该使用 reset API）
    // 注意：这里需要管理员权限，或者创建一个特殊的 API 端点
    // 临时方案：使用当前登录用户的 token 调用 updatePassword API
    await updatePassword(userId, loginForm.password, passwordForm.newPassword)

    MessagePlugin.success(t('auth.passwordChanged'))
    passwordExpiredDialogVisible.value = false

    // 重新登录（使用新密码）
    loginForm.password = passwordForm.newPassword
    await userStore.login(loginForm.email, loginForm.password)
  } catch (error: any) {
    MessagePlugin.error(error?.response?.data?.msg || t('auth.passwordChangeFailed'))
  } finally {
    changingPassword.value = false
  }
}
</script>

<style scoped>
/* 仅页面特有样式，通用样式已抽取到 src/styles/auth.css */

.password-expired-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 20px 0;
}

.password-expired-content h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.password-expired-content p {
  margin: 0;
  font-size: 14px;
  color: var(--td-text-color-secondary);
  text-align: center;
}

.password-strength {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.strength-bar {
  flex: 1;
  height: 4px;
  background: var(--td-component-stroke);
  border-radius: 2px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  transition: width 0.3s, background-color 0.3s;
}

.strength-fill.weak { background: var(--td-error-color); }
.strength-fill.medium { background: var(--td-warning-color); }
.strength-fill.strong { background: var(--td-success-color); }

.strength-text {
  font-size: 12px;
  min-width: 36px;
}

.strength-text.weak { color: var(--td-error-color); }
.strength-text.medium { color: var(--td-warning-color); }
.strength-text.strong { color: var(--td-success-color); }
</style>
