<!--
  Copyright © 2026-present reepu.com
  SPDX-License-Identifier: Apache-2.0
-->

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { createAdmin } from '@/api/setup'
import type { AdminConfig } from '@/api/setup'

const { t } = useI18n()

defineProps<{ dbConfig: unknown }>()
const emit = defineEmits<{ (e: 'complete'): void }>()

const form = reactive<AdminConfig>({
  email: '',
  password: '',
  password_confirm: '',
  site_name: '',
  site_slug: '',
})
const loading = ref(false)

const handleSubmit = async () => {
  if (form.password !== form.password_confirm) {
    MessagePlugin.warning(t('setup.admin.passwordMismatch'))
    return
  }
  if (form.password.length < 8) {
    MessagePlugin.warning(t('setup.admin.passwordMinLength'))
    return
  }
  loading.value = true
  try {
    await createAdmin({ ...form })
    MessagePlugin.success(t('setup.admin.installSuccess'))
    setTimeout(() => emit('complete'), 2000)
  } catch (err: any) {
    const msg = err?.response?.data?.msg ||
      err?.response?.data?.message ||
      err?.response?.data?.error ||
      t('setup.admin.installFailed')
    MessagePlugin.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <t-form :data="form" label-width="100px">
    <t-form-item :label="$t('setup.admin.email')" name="email">
      <t-input v-model="form.email" type="email" placeholder="admin@example.com" />
    </t-form-item>
    <t-form-item :label="$t('setup.admin.password')" name="password">
      <t-input v-model="form.password" type="password" :placeholder="$t('setup.admin.passwordHint')" />
    </t-form-item>
    <t-form-item :label="$t('setup.admin.passwordConfirm')" name="password_confirm">
      <t-input v-model="form.password_confirm" type="password" :placeholder="$t('setup.admin.passwordConfirmHint')" />
    </t-form-item>
    <t-form-item :label="$t('setup.admin.siteName')" name="site_name">
      <t-input v-model="form.site_name" :placeholder="$t('setup.admin.siteNamePlaceholder')" />
    </t-form-item>
    <t-form-item :label="$t('setup.admin.siteSlug')" name="site_slug">
      <t-input v-model="form.site_slug" :placeholder="$t('setup.admin.siteSlugPlaceholder')" />
    </t-form-item>
  </t-form>
  <div class="setup-actions">
    <t-button theme="default" @click="emit('complete')">
      {{ $t('common.back') }}
    </t-button>
    <t-button theme="primary" :loading="loading" @click="handleSubmit">
      {{ $t('setup.admin.completeInstall') }}
    </t-button>
  </div>
</template>
