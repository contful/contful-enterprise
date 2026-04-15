<template>
  <div class="login-page">
    <t-card class="login-card">
      <template #title>Contful Console</template>
      <t-form :data="formData" :rules="rules" @submit="handleSubmit">
        <t-form-item label="邮箱" name="email">
          <t-input v-model="formData.email" placeholder="请输入邮箱" />
        </t-form-item>
        <t-form-item label="密码" name="password">
          <t-input v-model="formData.password" type="password" placeholder="请输入密码" />
        </t-form-item>
        <t-form-item>
          <t-button type="submit" theme="primary" :loading="loading">登录</t-button>
        </t-form-item>
      </t-form>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { MessagePlugin, type FormRule } from 'tdesign-vue-next'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)

const formData = reactive({
  email: '',
  password: '',
})

const rules: Record<string, FormRule[]> = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码至少8位', trigger: 'blur' },
  ],
}

const handleSubmit = async ({ validateResult }: { validateResult: boolean }) => {
  if (validateResult !== true) return
  
  loading.value = true
  try {
    const success = await userStore.login(formData.email, formData.password)
    if (success) {
      MessagePlugin.success('登录成功')
      router.push('/')
    } else {
      MessagePlugin.error('登录失败')
    }
  } catch (err) {
    MessagePlugin.error('登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f5f5;
}
.login-card {
  width: 400px;
}
</style>
