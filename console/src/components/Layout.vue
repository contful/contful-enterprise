<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { createSite, type CreateSiteParams } from '@/api/site'
import { showError, showSuccess } from '@/utils/request'
import LangSwitcher from './LangSwitcher.vue'

const { t } = useI18n()

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const siteStore = useSiteStore()

const sidebarCollapsed = ref(false)

// 创建站点弹窗
const showCreateSite = ref(false)
const newSiteName = ref('')
const newSiteSlug = ref('')
const newSiteDesc = ref('')
const creating = ref(false)

const menuItems = computed(() => [
  { path: '/', icon: 'dashboard', label: t('menu.dashboard'), name: 'Dashboard', tIcon: 'dashboard' },
  { path: '/content/types', icon: 'schema', label: t('menu.contentTypes'), name: 'ContentTypes', tIcon: 'server' },
  { path: '/content/entries', icon: 'article', label: t('menu.contentEntries'), name: 'Content', tIcon: 'article' },
  { path: '/assets', icon: 'image', label: t('menu.media'), name: 'Media', tIcon: 'image' },
  { path: '/users', icon: 'people', label: t('menu.users'), name: 'Users', tIcon: 'user' },
  { path: '/tokens', icon: 'key', label: t('menu.tokens'), name: 'ApiTokens', tIcon: 'key' },
  { path: '/configs', icon: 'tools', label: t('menu.configs'), name: 'Configs', tIcon: 'tools' },
  { path: '/settings', icon: 'setting', label: t('menu.settings'), name: 'Settings', tIcon: 'setting' },
])

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

const handleLogout = async () => {
  await userStore.logout()
  router.push('/login')
}

const user = computed(() => userStore.user)

const goToProfile = () => {
  router.push('/profile')
}

// 生成 slug
const generateSlug = (name: string) => {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fff]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 100) || 'my-site'
}

// 复制 Site ID 到剪贴板
const copySiteId = async () => {
  if (!siteStore.currentSiteId) return
  try {
    await navigator.clipboard.writeText(siteStore.currentSiteId)
    showSuccess(t('site.siteIdCopied'))
  } catch (e) {
    showError({ response: { data: { msg: t('site.copyFailed') } } } as any)
  }
}

const handleNameInput = (val: string) => {
  const name = String(val || '')
  // 仅自动生成 slug（当 slug 为空或等于自动生成值时才更新）
  if (!newSiteSlug.value || newSiteSlug.value === generateSlug(newSiteSlug.value)) {
    newSiteSlug.value = generateSlug(name)
  }
}

const handleCreateSite = async () => {
  if (!newSiteName.value.trim()) {
    showError({ response: { data: { msg: t('site.enterSiteName') } } } as any)
    return
  }
  const slug = newSiteSlug.value.trim() || generateSlug(newSiteName.value)
  if (!/^[a-z][a-z0-9\-]{0,98}[a-z0-9]$/.test(slug)) {
    showError({ response: { data: { msg: t('site.formatError') } } } as any)
    return
  }
  creating.value = true
  const result = await siteStore.createAndSwitch({
    name: newSiteName.value.trim(),
    slug,
    description: newSiteDesc.value.trim() || undefined,
  })
  creating.value = false
  if (result.success) {
    closeCreateSiteDialog()
  }
}

// 关闭创建站点弹窗
const closeCreateSiteDialog = async () => {
  showCreateSite.value = false
  await nextTick()
  newSiteName.value = ''
  newSiteSlug.value = ''
  newSiteDesc.value = ''
}

// 初始化时加载用户信息和站点（如果已登录）
onMounted(async () => {
  if (userStore.isLoggedIn) {
    // 先加载用户信息
    if (!userStore.user) {
      await userStore.fetchUser()
    }
    // 再加载站点列表
    if (siteStore.sites.length === 0) {
      await siteStore.fetchSites()
    }
  }
})
</script>

<template>
  <div class="app-layout" :class="{ collapsed: sidebarCollapsed }">
    <!-- 顶部 Header -->
    <header class="app-header">
      <div class="header-left">
        <div class="logo">
          <img src="/assets/logo.png" alt="Contful" class="logo-img" />
        </div>
        <t-button
          shape="square" variant="text"
          style="flex-shrink: 0"
          @click="sidebarCollapsed = !sidebarCollapsed"
        >
          <template #icon>
            <Icon :name="sidebarCollapsed ? 'indent-right' : 'indent-left'" />
          </template>
        </t-button>
        <!-- 站点选择器：紧跟折叠按钮，左侧分隔线对齐 -->
        <div class="site-selector">
          <t-select
            v-model="siteStore.currentSiteId"
            :options="siteStore.sites.map(s => ({ label: s.name, value: s.id }))"
            :placeholder="t('site.selectSite')"
            :clearable="false"
            style="width: 220px"
            @change="(val: string) => siteStore.setCurrentSite(val)"
          />
          <!-- 复制 Site ID 按钮 -->
          <t-tooltip
            v-if="siteStore.currentSiteId"
            :content="t('site.copySiteId') + ': ' + siteStore.currentSiteId"
            placement="bottom"
          >
            <t-button
              shape="square" variant="text"
              size="small"
              @click="copySiteId"
            >
              <template #icon>
                <Icon name="file-copy" />
              </template>
            </t-button>
          </t-tooltip>
          <t-button
            v-if="siteStore.currentSiteId"
            shape="square" variant="text"
            size="small"
            :title="t('site.createNewSite')"
            @click="showCreateSite = true"
          >
            <template #icon>
              <Icon name="add" />
            </template>
          </t-button>
          <t-button
            v-else-if="siteStore.sites.length === 0 && userStore.isLoggedIn"
            variant="outline" size="small"
            @click="showCreateSite = true"
          >
            {{ t('site.createSite') }}
          </t-button>
      </div>
      </div>
      <!-- 占位，撑开左右两端的间距 -->
      <div class="header-spacer"></div>
      <div class="header-right">
        <a
          href="https://contful.com"
          target="_blank"
          rel="noopener noreferrer"
          class="header-link"
          :title="t('header.officialSite')"
        >
          {{ t('header.officialSite') }}
        </a>
        <LangSwitcher />
        <t-dropdown trigger="click">
          <div class="user-trigger">
            <div class="avatar" v-if="user">{{ user.nickname?.charAt(0).toUpperCase() || user.email?.charAt(0).toUpperCase() || 'U' }}</div>
            <span class="user-name" v-if="user">{{ user.nickname || user.email }}</span>
            <t-icon name="chevron-down" size="14px" style="color: var(--color-text-secondary)" />
          </div>
          <template #dropdown>
            <t-dropdown-menu>
              <t-dropdown-item class="dropdown-email-item" disabled>
                {{ user?.email }}
              </t-dropdown-item>
              <t-dropdown-item @click="goToProfile">
                <template #prefix-icon><t-icon name="user" /></template>
                {{ t('settings.personalProfile') }}
              </t-dropdown-item>
              <t-divider />
              <t-dropdown-item theme="error" @click="handleLogout">
                <template #prefix-icon><t-icon name="poweroff" /></template>
                {{ t('common.logout') }}
              </t-dropdown-item>
            </t-dropdown-menu>
          </template>
        </t-dropdown>
      </div>
    </header>

    <div class="app-body">
      <!-- 侧边栏 -->
      <aside class="sidebar">
        <nav class="sidebar-nav">
          <router-link
            v-for="item in menuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isActive(item.path) }"
          >
            <span class="nav-icon">
              <Icon :name="item.tIcon" />
            </span>
            <span class="nav-label">{{ item.label }}</span>
          </router-link>
        </nav>
      </aside>

      <!-- 主内容区 -->
      <main class="main-content">
        <div class="content-wrapper">
          <slot />
        </div>
      </main>
    </div>

    <!-- 创建站点弹窗 -->
    <t-dialog
      v-model:visible="showCreateSite"
      :header="t('site.createSite')"
      :close-on-overlay-click="true"
      :close-on-esc-keydown="true"
      :destroy-on-close="false"
      @close="closeCreateSiteDialog"
    >
      <t-form layout="vertical">
        <t-form-item :label="t('site.siteName')" required>
          <t-input
            v-model="newSiteName"
            :placeholder="t('site.siteNamePlaceholder')"
            :maxlength="200"
            @change="handleNameInput"
          />
        </t-form-item>
        <t-form-item :label="t('site.siteSlug')" required>
          <t-input
            v-model="newSiteSlug"
            :placeholder="t('site.siteSlugPlaceholder')"
            :maxlength="100"
          />
        </t-form-item>
        <t-form-item :label="t('site.siteDescription')">
          <t-textarea
            v-model="newSiteDesc"
            :placeholder="t('site.siteDescPlaceholder')"
            :maxlength="2000"
            :autosize="{ minRows: 2, maxRows: 4 }"
          />
        </t-form-item>
      </t-form>
      <template #footer>
        <div class="dialog-footer">
          <t-button variant="outline" @click="closeCreateSiteDialog">{{ t('common.cancel') }}</t-button>
          <t-button theme="primary" :loading="creating" @click="handleCreateSite">{{ t('common.create') }}</t-button>
        </div>
      </template>
    </t-dialog>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--color-bg);
}

/* 顶部 Header */
.app-header {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.header-spacer {
  flex: 1;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo-img {
  height: 32px;
  width: auto;
}

.site-selector {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-left: 12px;
  border-left: 1px solid var(--color-border);
}

.site-selector :deep(.t-select) {
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.header-link {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  font-size: 14px;
  color: var(--color-text-secondary);
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.2s;
}

.header-link:hover {
  color: var(--color-primary);
  background: var(--color-hover);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.user-menu {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 13px;
  flex-shrink: 0;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 120px;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-trigger:hover {
  background: var(--color-hover);
}

:deep(.dropdown-email-item) {
  font-size: 12px;
  color: var(--color-text-secondary);
  cursor: default;
  pointer-events: none;
}

/* 主体区域 */
.app-body {
  display: flex;
  flex: 1;
}

/* 侧边栏 */
.sidebar {
  width: 220px;
  background: #1e293b;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  flex-shrink: 0;
}

.app-layout.collapsed .sidebar {
  width: 64px;
}

.sidebar-nav {
  flex: 1;
  padding: 16px 12px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  margin-bottom: 4px;
  border-radius: 8px;
  color: #94a3b8;
  text-decoration: none;
  transition: all 0.2s;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.nav-item.active {
  background: var(--color-primary);
  color: #fff;
}

.nav-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 20px;
  height: 20px;
}

.nav-label {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
}

.app-layout.collapsed .nav-label {
  display: none;
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--color-bg);
}

.content-wrapper {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}
</style>
