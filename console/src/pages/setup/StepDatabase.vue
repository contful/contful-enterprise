<!--
  Copyright © 2026-present reepu.com
  SPDX-License-Identifier: Apache-2.0
-->

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import { testDatabase, initializeDatabase } from '@/api/setup'
import type { DatabaseConfig } from '@/api/setup'

const props = defineProps<{ initialConfig: DatabaseConfig }>()
const emit = defineEmits<{ (e: 'next', config: DatabaseConfig): void }>()

const form = reactive<DatabaseConfig>({ ...props.initialConfig })
const testing = ref(false)
const initializing = ref(false)

const handleTest = async () => {
  testing.value = true
  try {
    await testDatabase({ ...form })
    MessagePlugin.success('数据库连接成功')
  } catch (err: any) {
    const msg = err?.response?.data?.msg ||
      err?.response?.data?.message ||
      err?.response?.data?.error ||
      '连接失败，请检查配置'
    MessagePlugin.error(msg)
  } finally {
    testing.value = false
  }
}

const handleNext = async () => {
  initializing.value = true
  try {
    await initializeDatabase({ ...form })
    emit('next', { ...form })
  } catch (err: any) {
    const msg = err?.response?.data?.msg ||
      err?.response?.data?.message ||
      err?.response?.data?.error ||
      '初始化失败'
    MessagePlugin.error(msg)
  } finally {
    initializing.value = false
  }
}
</script>

<template>
  <t-form :data="form" label-width="100px">
    <t-form-item :label="$t('setup.database.host')" name="host">
      <t-input v-model="form.host" placeholder="localhost" />
    </t-form-item>
    <t-form-item :label="$t('setup.database.port')" name="port">
      <t-input-number v-model="form.port" :min="1" :max="65535" />
    </t-form-item>
    <t-form-item :label="$t('setup.database.user')" name="user">
      <t-input v-model="form.user" placeholder="postgres" />
    </t-form-item>
    <t-form-item :label="$t('setup.database.password')" name="password">
      <t-input v-model="form.password" type="password" :placeholder="$t('setup.database.passwordPlaceholder')" />
    </t-form-item>
    <t-form-item :label="$t('setup.database.dbName')" name="db_name">
      <t-input v-model="form.db_name" placeholder="contful" />
    </t-form-item>
    <t-form-item :label="$t('setup.database.sslMode')" name="ssl_mode">
      <t-select v-model="form.ssl_mode">
        <t-option value="disable" :label="$t('setup.database.sslDisable')" />
        <t-option value="require" :label="$t('setup.database.sslRequire')" />
        <t-option value="verify-ca" :label="$t('setup.database.sslVerifyCa')" />
        <t-option value="verify-full" :label="$t('setup.database.sslVerifyFull')" />
      </t-select>
    </t-form-item>
  </t-form>
  <div class="setup-actions">
    <t-button theme="default" :loading="testing" @click="handleTest">
      {{ $t('setup.database.testConnection') }}
    </t-button>
    <t-button theme="primary" :loading="initializing" @click="handleNext">
      {{ $t('common.next') }}
    </t-button>
  </div>
</template>
