<template>
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

      <!-- 恢复码区域（MFA 已启用时显示） -->
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
      <!-- 步骤 1：扫码 -->
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

      <!-- 步骤 2：验证激活 -->
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

          <!-- 恢复码展示（Enable 成功后） -->
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
</template>

<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import request from '@/utils/request'

const { t } = useI18n()

// MFA 状态
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
      // 剩余恢复码暂从本地状态维护（API 可扩展）
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
    } else {
      MessagePlugin.error(res.data.message || t('settings.mfaDisableFailed'))
    }
  } catch (e: any) {
    MessagePlugin.error(e.response?.data?.message || t('settings.mfaDisableFailed'))
  } finally {
    disableLoading.value = false
  }
}

onMounted(() => {
  fetchMFAStatus()
})
</script>

<style src="@/styles/mfa.css"></style>
<style scoped>
/* 页面特有样式：安全设置 — card/mfa 样式已提取到 common.css / mfa.css */
</style>
