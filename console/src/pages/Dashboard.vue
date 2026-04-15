<template>
  <div>
    <t-page-header title="控制台">
      <template #content>欢迎使用 Contful 内容管理系统</template>
    </t-page-header>
    <div class="dashboard">
      <t-row :gutter="[16, 16]">
        <t-col :span="3">
          <t-card>
            <template #title>内容类型</template>
            <div class="stat-value">{{ stats.contentTypes }}</div>
          </t-card>
        </t-col>
        <t-col :span="3">
          <t-card>
            <template #title>内容条目</template>
            <div class="stat-value">{{ stats.contentEntries }}</div>
          </t-card>
        </t-col>
        <t-col :span="3">
          <t-card>
            <template #title>媒体文件</template>
            <div class="stat-value">{{ stats.mediaFiles }}</div>
          </t-card>
        </t-col>
        <t-col :span="3">
          <t-card>
            <template #title>API Token</template>
            <div class="stat-value">{{ stats.apiTokens }}</div>
          </t-card>
        </t-col>
      </t-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const stats = reactive({
  contentTypes: 0,
  contentEntries: 0,
  mediaFiles: 0,
  apiTokens: 0,
})
</script>

<style scoped>
.dashboard {
  padding: 24px;
}
.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #0052cc;
}
</style>
