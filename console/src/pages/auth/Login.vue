<template>
  <div class="login-container">
    <t-card class="login-card">
      <template #header>
        <div class="login-header">
          <t-image
            :src="logoUrl"
            :style="{ width: '48px', height: '48px' }"
            shape="round"
          />
          <h1 class="login-title">Contful</h1>
          <p class="login-subtitle">开源 Headless CMS</p>
        </div>
      </template>

      <t-tabs v-model="activeTab" :default-value="activeTab">
        <t-tab-panel value="login" label="登录">
          <t-form
            :data="loginForm"
            :rules="loginRules"
            ref="loginFormRef"
            @submit="onLogin"
          >
            <t-form-item name="email">
              <t-input
                v-model="loginForm.email"
                placeholder="邮箱"
                size="large"
              >
                <template #prefix-icon>
                  <MailIcon />
                </template>
              </t-input>
            </t-form-item>

            <t-form-item name="password">
              <t-input
                v-model="loginForm.password"
                type="password"
                placeholder="密码"
                size="large"
                autocomplete="current-password"
              >
                <template #prefix-icon>
                  <LockOnIcon />
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
                登录
              </t-button>
            </t-form-item>
          </t-form>
        </t-tab-panel>

        <t-tab-panel value="register" label="注册">
          <t-form
            :data="registerForm"
            :rules="registerRules"
            ref="registerFormRef"
            @submit="onRegister"
          >
            <t-form-item name="nickname">
              <t-input
                v-model="registerForm.nickname"
                placeholder="昵称（可选）"
                size="large"
              >
                <template #prefix-icon>
                  <UserIcon />
                </template>
              </t-input>
            </t-form-item>

            <t-form-item name="email">
              <t-input
                v-model="registerForm.email"
                placeholder="邮箱"
                size="large"
              >
                <template #prefix-icon>
                  <MailIcon />
                </template>
              </t-input>
            </t-form-item>

            <t-form-item name="password">
              <t-input
                v-model="registerForm.password"
                type="password"
                placeholder="密码（至少8位）"
                size="large"
                autocomplete="new-password"
              >
                <template #prefix-icon>
                  <LockOnIcon />
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
                注册
              </t-button>
            </t-form-item>
          </t-form>
        </t-tab-panel>
      </t-tabs>

      <template #footer>
        <div class="login-footer">
          <span class="copyright">© 2026 Contful. MIT License.</span>
        </div>
      </template>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { MessagePlugin } from 'tdesign-vue-next'
import { MailIcon, LockOnIcon, UserIcon } from 'tdesign-icons-vue-next'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref('login')
const loginFormRef = ref()
const registerFormRef = ref()

const loginForm = reactive({
  email: '',
  password: '',
})

const registerForm = reactive({
  email: '',
  password: '',
  nickname: '',
})

const loginRules = {
  email: [
    { required: true, message: '请输入邮箱' },
    { email: true, message: '请输入有效的邮箱地址' },
  ],
  password: [
    { required: true, message: '请输入密码' },
    { min: 8, message: '密码至少8位' },
  ],
}

const registerRules = {
  email: [
    { required: true, message: '请输入邮箱' },
    { email: true, message: '请输入有效的邮箱地址' },
  ],
  password: [
    { required: true, message: '请输入密码' },
    { min: 8, message: '密码至少8位' },
  ],
}

// Logo URL - 可以替换为实际的 logo
const logoUrl = 'https://www.contful.dev/logo.png'

const onLogin = async () => {
  const result = await userStore.login(loginForm.email, loginForm.password)
  if (result.success) {
    MessagePlugin.success('登录成功')
    router.push('/')
  } else {
    MessagePlugin.error(result.message || '登录失败')
  }
}

const onRegister = async () => {
  const result = await userStore.register(
    registerForm.email,
    registerForm.password,
    registerForm.nickname
  )
  if (result.success) {
    MessagePlugin.success('注册成功，请登录')
    activeTab.value = 'login'
    loginForm.email = registerForm.email
  } else {
    MessagePlugin.error(result.message || '注册失败')
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
  padding: 20px 0;
}

.login-title {
  font-size: 28px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin: 0;
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
</style>
