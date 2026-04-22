<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { createSite, type CreateSiteParams } from '@/api/site'
import { showError, showSuccess } from '@/utils/request'

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

const menuItems = [
  { path: '/', icon: 'dashboard', label: '仪表盘', name: 'Dashboard', tIcon: 'dashboard' },
  { path: '/content/types', icon: 'schema', label: '类型', name: 'ContentTypes', tIcon: 'server' },
  { path: '/content/entries', icon: 'article', label: '内容', name: 'Content', tIcon: 'article' },
  { path: '/assets', icon: 'image', label: '媒体', name: 'Media', tIcon: 'image' },
  { path: '/users', icon: 'people', label: '用户', name: 'Users', tIcon: 'user' },
  { path: '/settings', icon: 'settings', label: '设置', name: 'Settings', tIcon: 'setting' },
]

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

const handleLogout = async () => {
  await userStore.logout()
  router.push('/login')
}

const user = computed(() => userStore.user)

// 生成 slug
const generateSlug = (name: string) => {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fff]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 100) || 'my-site'
}

const handleNameInput = (val: string) => {
  newSiteName.value = val
  if (!newSiteSlug.value || newSiteSlug.value === generateSlug(newSiteSlug.value)) {
    newSiteSlug.value = generateSlug(val)
  }
}

const handleCreateSite = async () => {
  if (!newSiteName.value.trim()) {
    showError({ response: { data: { msg: '请输入站点名称' } } } as any)
    return
  }
  const slug = newSiteSlug.value.trim() || generateSlug(newSiteName.value)
  if (!/^[a-z][a-z0-9\-]{0,98}[a-z0-9]$/.test(slug)) {
    showError({ response: { data: { msg: 'Slug 格式不正确：字母开头，仅支持小写字母、数字和连字符' } } } as any)
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
    showCreateSite.value = false
    newSiteName.value = ''
    newSiteSlug.value = ''
    newSiteDesc.value = ''
  }
}

// 初始化时加载站点（如果已登录）
onMounted(async () => {
  if (userStore.isLoggedIn && siteStore.sites.length === 0) {
    await siteStore.fetchSites()
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
          @click="sidebarCollapsed = !sidebarCollapsed"
        >
          <template #icon>
            <t-icon :name="sidebarCollapsed ? 'indent-right' : 'indent-left'" />
          </template>
        </t-button>
      </div>
      <div class="header-center">
        <!-- 站点选择器 -->
        <t-select
          v-model="siteStore.currentSiteId"
          :options="siteStore.sites.map(s => ({ label: s.name, value: s.id }))"
          placeholder="选择站点"
          :clearable="false"
          style="width: 200px"
          @change="(val: string) => siteStore.setCurrentSite(val)"
        >
          <template #suffixIcon>
            <t-icon name="chevron-down" />
          </template>
        </t-select>
        <t-button
          v-if="siteStore.currentSiteId"
          shape="square" variant="text"
          size="small"
          title="创建新站点"
          @click="showCreateSite = true"
        >
          <template #icon>
            <t-icon name="add" />
          </template>
        </t-button>
        <t-button
          v-else-if="siteStore.sites.length === 0 && userStore.isLoggedIn"
          variant="outline" size="small"
          @click="showCreateSite = true"
        >
          创建站点
        </t-button>
      </div>
      <div class="header-right">
        <div class="user-menu">
          <div class="avatar" v-if="user">{{ user.nickname?.charAt(0).toUpperCase() || user.email?.charAt(0).toUpperCase() || 'U' }}</div>
          <span class="user-name" v-if="user">{{ user.nickname || user.email }}</span>
          <t-button
            shape="square" variant="text"
            @click="handleLogout"
          >
            <template #icon>
              <t-icon name="logout" />
            </template>
          </t-button>
        </div>
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
              <t-icon :name="item.tIcon" />
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
      header="创建站点"
      :confirm-btn="{ loading: creating, theme: 'primary' }"
      @confirm="handleCreateSite"
    >
      <t-form layout="vertical">
        <t-form-item label="站点名称" required>
          <t-input
            v-model="newSiteName"
            placeholder="例如：我的博客"
            :maxlength="200"
            @input="handleNameInput"
          />
        </t-form-item>
        <t-form-item label="站点标识 (Slug)" required>
          <t-input
            v-model="newSiteSlug"
            placeholder="字母开头，小写字母+数字+连字符"
            :maxlength="100"
          />
        </t-form-item>
        <t-form-item label="描述">
          <t-textarea
            v-model="newSiteDesc"
            placeholder="站点描述（可选）"
            :maxlength="2000"
            :autosize="{ minRows: 2, maxRows: 4 }"
          />
        </t-form-item>
      </t-form>
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
  justify-content: space-between;
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
  gap: 16px;
}

.collapse-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.collapse-btn:hover {
  background: var(--color-hover);
  color: var(--color-text);
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

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
}

.breadcrumb-item {
  font-size: 15px;
  font-weight: 500;
  color: var(--color-text);
}

.header-right {
  display: flex;
  align-items: center;
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
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
}

.logout-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: var(--color-error-light);
  color: var(--color-error);
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
