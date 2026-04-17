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
          <p class="login-subtitle">开源 Headless CMS</p>
        </div>
      </template>

      <t-form
        :data="loginForm"
        :required-mark="false"
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
import { MessagePlugin } from 'tdesign-vue-next'
import { MailIcon, LockOnIcon } from 'tdesign-icons-vue-next'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const loginFormRef = ref()

const loginForm = reactive({
  email: '',
  password: '',
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

// Logo URL - 本地资源
const logoUrl = '/assets/logo.png'

const onLogin = async () => {
  const result = await userStore.login(loginForm.email, loginForm.password)
  if (result.success) {
    MessagePlugin.success('登录成功')
    router.push('/')
  } else {
    MessagePlugin.error(result.message || '登录失败')
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

.t-form :deep(.t-form__item) {
  display: flex;
  justify-content: center;
}
</style>
