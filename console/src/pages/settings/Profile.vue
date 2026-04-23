<template>
  <div class="profile-page">
    <!-- 头像区域 -->
    <div class="card avatar-section">
      <div class="avatar-container">
        <t-avatar size="80px" :image="userStore.user?.avatar_url || undefined">
          <template #icon>
            <t-icon name="user" />
          </template>
        </t-avatar>
        <div class="avatar-info">
          <h3 class="user-name">{{ userStore.user?.nickname || userStore.user?.email || 'User' }}</h3>
          <p class="user-email">{{ userStore.user?.email }}</p>
          <div class="avatar-actions">
            <t-upload
              :action="uploadUrl"
              :headers="uploadHeaders"
              :format-response="formatUploadResponse"
              @success="onAvatarUploadSuccess"
              @fail="onAvatarUploadFail"
            >
              <template #file-list-display>
                <t-button size="small" variant="outline">
                  <template #icon><t-icon name="upload" /></template>
                  {{ t('settings.profileChangeAvatar') }}
                </t-button>
              </template>
            </t-upload>
          </div>
        </div>
      </div>
    </div>

    <!-- 基本信息 -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">{{ t('settings.profileBasicInfo') }}</h3>
      </div>
      <t-form
        :data="profileForm"
        :rules="profileRules"
        @submit="onUpdateProfile"
        label-align="right"
        :label-width="120"
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

    <!-- 账号信息 -->
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/stores/user'
import request from '@/utils/request'

const { t } = useI18n()
const userStore = useUserStore()

// 头像上传
const uploadUrl = import.meta.env.VITE_API_URL + '/admin/api/v1/users/me/avatar'
const uploadHeaders = {
  Authorization: `Bearer ${localStorage.getItem('access_token')}`,
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

const onResetPasswordForm = () => {
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
}

// 头像上传成功
const formatUploadResponse = (res: any) => {
  return {
    ...res,
    response: res,
  }
}

const onAvatarUploadSuccess = (context: any) => {
  const res = context.response?.response
  if (res?.code === 200) {
    MessagePlugin.success(t('settings.profileAvatarSuccess'))
    // 更新 store 中的头像
    if (userStore.user && res.data?.avatar_url) {
      userStore.user.avatar_url = res.data.avatar_url
    }
  } else {
    MessagePlugin.error(res?.message || t('settings.profileAvatarFailed'))
  }
}

const onAvatarUploadFail = (context: any) => {
  MessagePlugin.error(context?.response?.message || t('settings.profileAvatarFailed'))
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

onMounted(() => {
  loadProfile()
})
</script>

<style scoped>
.profile-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
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

/* 头像区域 */
.avatar-section {
  padding: 32px;
}

.avatar-container {
  display: flex;
  align-items: center;
  gap: 24px;
}

.avatar-info {
  flex: 1;
}

.user-name {
  font-size: 20px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin: 0 0 4px;
}

.user-email {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin: 0 0 12px;
}

.avatar-actions {
  display: flex;
  gap: 8px;
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
</style>
