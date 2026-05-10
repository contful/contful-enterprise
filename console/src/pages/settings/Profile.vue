<template>
  <div class="page page--padded">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">{{ t('settings.personalProfile') }}</h1>
      <p class="page-subtitle">{{ t('settings.profileSubtitle') }}</p>
    </div>

    <!-- 第一行：两列布局 -->
    <div class="profile-grid">
      <!-- 左列：账号信息 -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">{{ t('settings.profileAccountInfo') }}</h3>
        </div>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">{{ t('settings.profileUserId') }}</span>
            <span class="info-value">{{ userStore.user?.id || '-' }}</span>
            <t-button size="small" variant="text" @click="copyUserId">
              <t-icon name="file-copy" />
            </t-button>
          </div>
          <div class="info-item">
            <span class="info-label">{{ t('settings.profileStatus') }}</span>
            <t-tag v-if="userStore.user?.status === 'active'" theme="success" variant="light">
              {{ t('settings.profileStatusActive') }}
            </t-tag>
            <t-tag v-else theme="warning" variant="light">
              {{ userStore.user?.status }}
            </t-tag>
          </div>
          <div class="info-item">
            <span class="info-label">{{ t('settings.profileRole') }}</span>
            <t-tag v-if="userStore.user?.is_super_admin" theme="danger" variant="light">
              {{ t('settings.profileSuperAdmin') }}
            </t-tag>
            <t-tag v-else theme="primary" variant="light">
              {{ t('settings.profileUser') }}
            </t-tag>
          </div>
          <div class="info-item">
            <span class="info-label">{{ t('settings.profileMFA') }}</span>
            <t-tag v-if="userStore.user?.mfa_enabled" theme="success" variant="light">
              {{ t('settings.mfaEnabled') }}
            </t-tag>
            <t-tag v-else theme="default" variant="light">
              {{ t('settings.mfaDisabled') }}
            </t-tag>
          </div>
          <div class="info-item">
            <span class="info-label">{{ t('settings.profileCreatedAt') }}</span>
            <span class="info-value">{{ formatTime(userStore.user?.created_time) }}</span>
          </div>
        </div>
      </div>

      <!-- 右列：基本信息 + 头像 -->
      <div class="right-column">
        <!-- 头像 + 基本信息 -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">{{ t('settings.profileBasicInfo') }}</h3>
          </div>
          <div class="basic-info-layout">
            <!-- 头像 -->
            <div class="avatar-col">
              <div class="avatar-wrapper" @click="triggerAvatarUpload">
                <t-avatar size="80px" :image="userStore.user?.avatar_url || undefined">
                  <template #icon>
                    <t-icon name="user" />
                  </template>
                </t-avatar>
                <div class="avatar-overlay">
                  <t-icon name="upload" />
                </div>
              </div>
              <input
                ref="avatarInputRef"
                type="file"
                accept="image/*"
                style="display:none"
                @change="onAvatarFileChange"
              />
              <span class="avatar-hint">{{ t('settings.profileClickToChange') }}</span>
            </div>
            <!-- 基本信息表单 -->
            <div class="basic-form-col">
              <t-form
                :data="profileForm"
                :rules="profileRules"
                @submit="onUpdateProfile"
                label-align="right"
                :label-width="100"
              >
                <t-form-item :label="t('settings.profileNickname')" name="nickname">
                  <t-input
                    v-model="profileForm.nickname"
                    :placeholder="t('settings.profileNicknamePlaceholder')"
                    clearable
                  />
                </t-form-item>
                <t-form-item :label="t('settings.profileEmail')" name="email">
                  <t-input
                    v-model="profileForm.email"
                    :placeholder="t('auth.emailPlaceholder')"
                    clearable
                    :disabled="true"
                  />
                  <template #tips>
                    <span class="form-tip">{{ t('settings.profileEmailImmutable') }}</span>
                  </template>
                </t-form-item>
                <t-form-item>
                  <t-button theme="primary" type="submit" :loading="profileLoading">
                    {{ t('common.save') }}
                  </t-button>
                </t-form-item>
              </t-form>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 修改密码 -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">{{ t('settings.profileChangePassword') }}</h3>
        <p class="card-desc">{{ t('settings.profilePasswordHint') }}</p>
      </div>
      <t-form
        :data="passwordForm"
        :rules="passwordRules"
        @submit="onChangePassword"
        label-align="right"
        :label-width="120"
        @reset="onResetPasswordForm"
      >
        <t-form-item :label="t('settings.profileOldPassword')" name="old_password">
          <t-input
            v-model="passwordForm.old_password"
            type="password"
            :placeholder="t('settings.profileOldPasswordPlaceholder')"
            clearable
          />
        </t-form-item>
        <t-form-item :label="t('settings.profileNewPassword')" name="new_password">
          <t-input
            v-model="passwordForm.new_password"
            type="password"
            :placeholder="t('auth.passwordPlaceholder')"
            clearable
          />
        </t-form-item>
        <t-form-item :label="t('settings.profileConfirmPassword')" name="confirm_password">
          <t-input
            v-model="passwordForm.confirm_password"
            type="password"
            :placeholder="t('settings.profileConfirmPasswordPlaceholder')"
            clearable
          />
        </t-form-item>
        <t-form-item>
          <t-space>
            <t-button theme="primary" type="submit" :loading="passwordLoading">
              {{ t('settings.profileUpdatePassword') }}
            </t-button>
            <t-button type="reset" variant="outline">
              {{ t('common.reset') }}
            </t-button>
          </t-space>
        </t-form-item>
      </t-form>
    </div>

    <!-- MFA 双因子认证 -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">{{ t('settings.mfaSection') }}</h3>
        <p class="card-desc">{{ t('settings.mfaDesc') }}</p>
      </div>

      <div class="mfa-status-row">
        <div class="mfa-status-info">
          <t-tag v-if="mfaEnabled" theme="success" variant="light">
            <template #icon><t-icon name="check-circle" /></template>
            {{ t('settings.mfaEnabled') }}
          </t-tag>
          <t-tag v-else theme="default" variant="light">
            {{ t('settings.mfaDisabled') }}
          </t-tag>
        </div>
        <div class="mfa-actions">
          <t-button
            v-if="!mfaEnabled"
            theme="primary"
            @click="openSetupDialog"
            :loading="setupLoading"
          >
            {{ t('settings.mfaEnable') }}
          </t-button>
          <t-button
            v-else
            theme="danger"
            variant="outline"
            @click="openDisableDialog"
          >
            {{ t('settings.mfaDisableBtn') }}
          </t-button>
        </div>
      </div>

      <template v-if="mfaEnabled">
        <t-divider />
        <div class="recovery-section">
          <h4 class="recovery-title">{{ t('settings.mfaRecoveryCodes') }}</h4>
          <p class="recovery-desc">{{ t('settings.mfaRecoveryDesc') }}</p>
          <p class="recovery-remaining">{{ t('settings.mfaRemainingCodes', { count: remainingCodes }) }}</p>
        </div>
      </template>
    </div>

    <!-- MFA Setup 弹窗 -->
    <t-dialog
      v-model:visible="setupVisible"
      :header="mfaSetupStep === 1 ? t('settings.mfaSetupStep1') : t('settings.mfaSetupStep2')"
      :footer="false"
      width="480px"
      @close="onSetupClose"
    >
      <template v-if="mfaSetupStep === 1">
        <div class="setup-step1">
          <p class="step-hint">{{ t('settings.mfaScanHint') }}</p>
          <div class="qr-container">
            <t-image
              v-if="setupData?.qr_code_url"
              :src="setupData.qr_code_url"
              :alt="t('settings.mfaSection')"
              width="200"
              height="200"
              style="display:block; margin:0 auto;"
            />
            <t-loading v-else size="medium" />
          </div>
          <div class="manual-key-section">
            <p class="manual-key-label">{{ t('settings.mfaManualKey') }}</p>
            <div class="manual-key-row">
              <t-input
                :value="setupData?.totp_secret || ''"
                readonly
                class="key-input"
              />
              <t-button variant="outline" @click="copySecret">{{ t('common.copy') }}</t-button>
            </div>
          </div>
          <div class="step-footer">
            <t-button theme="primary" @click="mfaSetupStep = 2">
              {{ t('common.nextPage') }} →
            </t-button>
          </div>
        </div>
      </template>
      <template v-if="mfaSetupStep === 2">
        <div class="setup-step2">
          <p class="step-hint">{{ t('settings.mfaConfirmCode') }}</p>
          <t-input
            v-model="enableCode"
            :placeholder="t('auth.mfaCodePlaceholder')"
            size="large"
            maxlength="6"
            :autofocus="true"
            style="margin-bottom: 16px;"
          />
          <t-button
            theme="primary"
            block
            :loading="enableLoading"
            :disabled="enableCode.length !== 6"
            @click="doEnable"
          >
            {{ enableLoading ? t('settings.mfaActivating') : t('settings.mfaEnableBtn') }}
          </t-button>
          <div v-if="recoveryCodesVisible" class="recovery-codes-container">
            <p class="save-hint">⚠ {{ t('settings.mfaSaveRecovery') }}</p>
            <div class="recovery-codes-grid">
              <div v-for="code in generatedRecoveryCodes" :key="code" class="recovery-code-item">
                {{ code }}
              </div>
            </div>
            <div class="recovery-buttons">
              <t-button variant="outline" @click="copyAllCodes">{{ t('settings.mfaCopyAll') }}</t-button>
              <t-button variant="outline" @click="downloadCodes">{{ t('settings.mfaDownload') }}</t-button>
            </div>
            <t-button theme="primary" block style="margin-top:16px;" @click="finishSetup">
              {{ t('common.confirm') }}
            </t-button>
          </div>
        </div>
      </template>
    </t-dialog>

    <!-- MFA Disable 弹窗 -->
    <t-dialog
      v-model:visible="disableVisible"
      :header="t('settings.mfaDisableBtn')"
      :confirm-btn="{ content: disableLoading ? t('settings.mfaDisabling') : t('common.confirm'), loading: disableLoading }"
      :cancel-btn="t('common.cancel')"
      @confirm="doDisable"
      @close="disableCode = ''"
    >
      <p>{{ t('settings.mfaDisableConfirm') }}</p>
      <t-input
        v-model="disableCode"
        :placeholder="t('auth.mfaCodePlaceholder')"
        size="large"
        maxlength="6"
        style="margin-top: 12px;"
      />
    </t-dialog>
  </div>
</template>

<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/stores/user'
import request from '@/utils/request'

const { t } = useI18n()
const userStore = useUserStore()

// 头像上传
const uploadUrl = '/admin/api/v1/users/me/avatar'

// 头像上传 refs
const avatarInputRef = ref<HTMLInputElement | null>(null)

const triggerAvatarUpload = () => {
  avatarInputRef.value?.click()
}

const onAvatarFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  const xhr = new XMLHttpRequest()
  xhr.open('POST', uploadUrl, true)
  xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('access_token')}`)

  xhr.onload = () => {
    if (xhr.status === 200) {
      const res = JSON.parse(xhr.responseText)
      if (res.code === 200) {
        MessagePlugin.success(t('settings.profileAvatarSuccess'))
        if (userStore.user && res.data?.avatar_url) {
          userStore.user.avatar_url = res.data.avatar_url
        }
      } else {
        MessagePlugin.error(res.message || t('settings.profileAvatarFailed'))
      }
    } else {
      MessagePlugin.error(t('settings.profileAvatarFailed'))
    }
    // 清空 input，允许重复选择同一文件
    target.value = ''
  }

  xhr.onerror = () => {
    MessagePlugin.error(t('settings.profileAvatarFailed'))
    target.value = ''
  }

  xhr.send(formData)
}

// 基本信息
const profileForm = reactive({
  nickname: '',
  email: '',
})
const profileLoading = ref(false)
const profileRules = {
  nickname: [{ required: true, message: t('settings.profileNicknameRequired'), trigger: 'blur' }],
}

// 修改密码
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})
const passwordLoading = ref(false)
const validateConfirmPassword = (val: string) => {
  if (val !== passwordForm.new_password) {
    return { result: false, message: t('settings.profilePasswordMismatch'), type: 'error' }
  }
  return { result: true }
}
const passwordRules = {
  old_password: [{ required: true, message: t('settings.profileOldPasswordRequired'), trigger: 'blur' }],
  new_password: [
    { required: true, message: t('auth.passwordRequired'), trigger: 'blur' },
    { min: 8, message: t('auth.passwordMinLength'), trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: t('settings.profileConfirmPasswordRequired'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

// 加载用户信息
const loadProfile = () => {
  if (userStore.user) {
    profileForm.nickname = userStore.user.nickname || ''
    profileForm.email = userStore.user.email || ''
  }
}

// 更新基本信息
const onUpdateProfile = async () => {
  profileLoading.value = true
  try {
    const res = await request.patch('/users/me', {
      nickname: profileForm.nickname,
    })
    if (res.data.code === 200) {
      MessagePlugin.success(t('settings.profileUpdateSuccess'))
      // 更新 store 中的用户信息
      if (userStore.user) {
        userStore.user.nickname = profileForm.nickname
      }
    } else {
      MessagePlugin.error(res.data.message || t('settings.profileUpdateFailed'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.message || t('settings.profileUpdateFailed'))
  } finally {
    profileLoading.value = false
  }
}

// 修改密码
const onChangePassword = async () => {
  passwordLoading.value = true
  try {
    const res = await request.put('/users/me/password', {
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    if (res.data.code === 200) {
      MessagePlugin.success(t('settings.profilePasswordSuccess'))
      // 重置表单
      passwordForm.old_password = ''
      passwordForm.new_password = ''
      passwordForm.confirm_password = ''
    } else {
      MessagePlugin.error(res.data.message || t('settings.profilePasswordFailed'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.message || t('settings.profilePasswordFailed'))
  } finally {
    passwordLoading.value = false
  }
}

// ============ MFA 双因子认证 ============
const mfaEnabled = ref(false)
const remainingCodes = ref(0)

// Setup 弹窗
const setupVisible = ref(false)
const setupLoading = ref(false)
const mfaSetupStep = ref<1 | 2>(1)
const setupData = ref<{ totp_secret: string; otpauth_uri: string; qr_code_url: string } | null>(null)
const enableCode = ref('')
const enableLoading = ref(false)
const recoveryCodesVisible = ref(false)
const generatedRecoveryCodes = ref<string[]>([])

// Disable 弹窗
const disableVisible = ref(false)
const disableCode = ref('')
const disableLoading = ref(false)

// 获取当前用户 MFA 状态
const fetchMFAStatus = async () => {
  try {
    const res = await request.get('/users/me')
    if (res.data.code === 200) {
      const u = res.data.data
      mfaEnabled.value = u.mfa_enabled ?? false
    }
  } catch {
    // ignore
  }
}

// 打开 Setup 弹窗
const openSetupDialog = async () => {
  setupLoading.value = true
  try {
    const res = await request.post('/auth/mfa/setup')
    if (res.data.code === 200) {
      setupData.value = res.data.data
      mfaSetupStep.value = 1
      enableCode.value = ''
      recoveryCodesVisible.value = false
      generatedRecoveryCodes.value = []
      setupVisible.value = true
    } else {
      MessagePlugin.error(res.data.message || t('settings.mfaSetupFailed'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.message || t('settings.mfaSetupFailed'))
  } finally {
    setupLoading.value = false
  }
}

const onSetupClose = () => {
  mfaSetupStep.value = 1
  enableCode.value = ''
  recoveryCodesVisible.value = false
}

// 复制 Secret
const copySecret = () => {
  if (setupData.value?.totp_secret) {
    navigator.clipboard.writeText(setupData.value.totp_secret).then(() => {
      MessagePlugin.success(t('common.copied'))
    })
  }
}

// 激活 MFA
const doEnable = async () => {
  if (enableCode.value.length !== 6) return
  enableLoading.value = true
  try {
    const res = await request.post('/auth/mfa/enable', { totp_code: enableCode.value })
    if (res.data.code === 200) {
      generatedRecoveryCodes.value = res.data.data.recovery_codes || []
      recoveryCodesVisible.value = true
    } else {
      MessagePlugin.error(res.data.message || t('settings.mfaEnableFailed'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.message || t('settings.mfaEnableFailed'))
  } finally {
    enableLoading.value = false
  }
}

const finishSetup = () => {
  mfaEnabled.value = true
  setupVisible.value = false
  remainingCodes.value = generatedRecoveryCodes.value.length
  MessagePlugin.success(t('settings.mfaEnableSuccess'))
  // 更新 store
  if (userStore.user) {
    userStore.user.mfa_enabled = true
  }
}

// 复制全部恢复码
const copyAllCodes = () => {
  const text = generatedRecoveryCodes.value.join('\n')
  navigator.clipboard.writeText(text).then(() => {
    MessagePlugin.success(t('common.copied'))
  })
}

// 下载恢复码
const downloadCodes = () => {
  const text = generatedRecoveryCodes.value.join('\n')
  const blob = new Blob([text], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'contful-recovery-codes.txt'
  a.click()
  URL.revokeObjectURL(url)
}

// 打开关闭 MFA 弹窗
const openDisableDialog = () => {
  disableCode.value = ''
  disableVisible.value = true
}

// 关闭 MFA
const doDisable = async () => {
  if (disableCode.value.length !== 6) {
    MessagePlugin.warning(t('auth.mfaCodePlaceholder'))
    return
  }
  disableLoading.value = true
  try {
    const res = await request.post('/auth/mfa/disable', { totp_code: disableCode.value })
    if (res.data.code === 200) {
      mfaEnabled.value = false
      remainingCodes.value = 0
      disableVisible.value = false
      MessagePlugin.success(t('settings.mfaDisableSuccess'))
      if (userStore.user) {
        userStore.user.mfa_enabled = false
      }
    } else {
      MessagePlugin.error(res.data.message || t('settings.mfaDisableFailed'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.message || t('settings.mfaDisableFailed'))
  } finally {
    disableLoading.value = false
  }
}

const onResetPasswordForm = () => {
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
}

// 复制 User ID
const copyUserId = () => {
  if (userStore.user?.id) {
    navigator.clipboard.writeText(userStore.user.id)
    MessagePlugin.success(t('common.copied'))
  }
}

// 格式化时间
const formatTime = (time?: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(async () => {
  if (!userStore.user) {
    await userStore.fetchUser()
  }
  loadProfile()
  fetchMFAStatus()
})
</script>

<style scoped>
/* 页面特有样式：个人资料 */

/* 第一行两列布局 */
.profile-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  align-items: stretch;
}

.right-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.right-column > .card {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.right-column > .card > .card-header {
  flex-shrink: 0;
}

.right-column > .card > .basic-info-layout {
  flex: 1;
}

/* 头像 + 基本信息 横向布局 */
.basic-info-layout {
  display: flex;
  gap: 32px;
  align-items: flex-start;
}

.avatar-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
}

.avatar-wrapper {
  position: relative;
  cursor: pointer;
  border-radius: 50%;
}

.avatar-wrapper :deep(.t-avatar) {
  transition: opacity 0.2s;
}

.avatar-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
  color: white;
  font-size: 20px;
}

.avatar-wrapper:hover .avatar-overlay {
  opacity: 1;
}

.avatar-hint {
  margin-top: 12px;
  font-size: 13px;
  color: var(--td-text-color-secondary);
}

.basic-form-col {
  flex: 1;
  min-width: 0;
}

.page-header {
  margin-bottom: 4px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin: 0 0 4px;
}

.page-subtitle {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  margin: 0;
}

.card {
  background: var(--td-bg-color-container);
  border-radius: 8px;
  border: 1px solid var(--td-border-level-1-color);
  padding: 24px;
}

.card-header {
  margin-bottom: 20px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin: 0 0 4px;
}

.card-desc {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  margin: 0;
}

/* 表单提示 */
.form-tip {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
}

/* 账号信息 */
.info-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.info-label {
  width: 100px;
  font-size: 14px;
  color: var(--td-text-color-secondary);
  flex-shrink: 0;
}

.info-value {
  font-size: 14px;
  color: var(--td-text-color-primary);
  font-family: monospace;
}

/* MFA */
.mfa-status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.mfa-status-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.recovery-section {
  margin-top: 16px;
}

.recovery-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin: 0 0 6px;
}

.recovery-desc {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  margin: 0 0 6px;
}

.recovery-remaining {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  margin: 0;
}

/* MFA Setup 弹窗 */
.setup-step1,
.setup-step2 {
  padding: 8px 0;
}

.step-hint {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin-bottom: 16px;
}

.qr-container {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
  min-height: 200px;
  align-items: center;
}

.manual-key-section {
  margin-bottom: 20px;
}

.manual-key-label {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  margin-bottom: 8px;
}

.manual-key-row {
  display: flex;
  gap: 8px;
}

.key-input {
  flex: 1;
  font-family: monospace;
}

.step-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.recovery-codes-container {
  margin-top: 20px;
  border: 1px solid var(--td-border-level-1-color);
  border-radius: 8px;
  padding: 16px;
}

.save-hint {
  font-size: 13px;
  color: var(--td-warning-color);
  margin-bottom: 12px;
}

.recovery-codes-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 12px;
}

.recovery-code-item {
  font-family: monospace;
  font-size: 14px;
  background: var(--td-bg-color-page);
  border: 1px solid var(--td-border-level-1-color);
  border-radius: 4px;
  padding: 6px 10px;
  text-align: center;
  color: var(--td-text-color-primary);
}

.recovery-buttons {
  display: flex;
  gap: 8px;
}
</style>
