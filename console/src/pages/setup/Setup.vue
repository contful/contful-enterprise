<!--
  Copyright © 2026-present reepu.com
  SPDX-License-Identifier: Apache-2.0
-->

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import StepDatabase from './StepDatabase.vue'
import StepAdmin from './StepAdmin.vue'
import { getSetupStatus } from '@/api/setup'
import type { DatabaseConfig } from '@/api/setup'

const router = useRouter()
const currentStep = ref(1)
const dbConfig = ref<DatabaseConfig>({
  host: 'localhost',
  port: 5432,
  user: 'postgres',
  password: '',
  db_name: 'contful',
  ssl_mode: 'disable',
})

onMounted(async () => {
  try {
    const res = await getSetupStatus()
    if (!res?.data?.setup_required) {
      router.replace('/login')
    }
  } catch {
    // API 不可达时继续显示安装向导
  }
})

const handleDbNext = (config: DatabaseConfig) => {
  dbConfig.value = config
  currentStep.value = 2
}

const handleComplete = () => {
  router.replace('/login')
}
</script>

<template>
  <div class="setup-container">
    <div class="setup-card">
      <div class="setup-logo">Contful</div>
      <div class="setup-subtitle">{{ $t('setup.subtitle') }}</div>
      <t-steps class="setup-steps" :current="currentStep - 1">
        <t-step-item :title="$t('setup.stepDatabase')" />
        <t-step-item :title="$t('setup.stepAdmin')" />
      </t-steps>
      <div class="setup-form">
        <StepDatabase
          v-if="currentStep === 1"
          :initial-config="dbConfig"
          @next="handleDbNext"
        />
        <StepAdmin
          v-if="currentStep === 2"
          :db-config="dbConfig"
          @complete="handleComplete"
        />
      </div>
    </div>
  </div>
</template>

<style scoped src="@/styles/setup.css"></style>
