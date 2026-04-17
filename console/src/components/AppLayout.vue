<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const sidebarCollapsed = ref(false)

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
