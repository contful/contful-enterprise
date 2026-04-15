<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const sidebarCollapsed = ref(false)

const menuItems = [
  { path: '/', icon: 'dashboard', label: '仪表盘', name: 'Dashboard' },
  { path: '/content-types', icon: 'schema', label: '内容类型', name: 'ContentTypes' },
  { path: '/content', icon: 'article', label: '内容', name: 'Content' },
  { path: '/media', icon: 'image', label: '媒体库', name: 'Media' },
  { path: '/users', icon: 'people', label: '用户管理', name: 'Users' },
  { path: '/settings', icon: 'settings', label: '设置', name: 'Settings' },
]

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

const user = computed(() => userStore.user)
</script>

<template>
  <div class="app-layout" :class="{ collapsed: sidebarCollapsed }">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="logo">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
            <rect width="32" height="32" rx="8" fill="var(--color-primary)"/>
            <path d="M8 10h16M8 16h12M8 22h8" stroke="white" stroke-width="2" stroke-linecap="round"/>
          </svg>
          <span v-if="!sidebarCollapsed" class="logo-text">Contful</span>
        </div>
        <button class="collapse-btn" @click="sidebarCollapsed = !sidebarCollapsed">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
            <path v-if="sidebarCollapsed" d="M7 4l6 6-6 6" stroke="currentColor" stroke-width="2" fill="none"/>
            <path v-else d="M13 4l-6 6 6 6" stroke="currentColor" stroke-width="2" fill="none"/>
          </svg>
        </button>
      </div>

      <nav class="sidebar-nav">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
        >
          <span class="nav-icon">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <!-- Dashboard -->
              <template v-if="item.icon === 'dashboard'">
                <path d="M3 4a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V4zM3 14a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H4a1 1 0 01-1-1v-4zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
              </template>
              <!-- Schema -->
              <template v-else-if="item.icon === 'schema'">
                <path d="M4 5a2 2 0 012-2h8a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm0 8a2 2 0 012-2h8a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2z"/>
              </template>
              <!-- Article -->
              <template v-else-if="item.icon === 'article'">
                <path d="M4 4a2 2 0 012-2h8a2 2 0 012 2v2a2 2 0 01-2 2h-2l-2 2-2-2H6a2 2 0 01-2-2V4zm2 0v10h8V4H6zm2 2h4v2H8V6z"/>
              </template>
              <!-- Image -->
              <template v-else-if="item.icon === 'image'">
                <path d="M4 4a2 2 0 012-2h8a2 2 0 012 2v8a2 2 0 01-2 2H8l-4-4-2 2zm8 0l2-2 4 4v2a2 2 0 01-2 2h-8a2 2 0 01-2-2V4h6zm-4 5a2 2 0 110-4 2 2 0 010 4z"/>
              </template>
              <!-- People -->
              <template v-else-if="item.icon === 'people'">
                <path d="M12 4a4 4 0 100 8 4 4 0 000-8zM6 8a6 6 0 1112 0A6 6 0 016 8zm2 10a2 2 0 100-4 2 2 0 000 4zm8 0a2 2 0 100-4 2 2 0 000 4z"/>
              </template>
              <!-- Settings -->
              <template v-else-if="item.icon === 'settings'">
                <path d="M10.325 4.317a1 1 0 011.772 0l1.128 2.29a1 1 0 001.218.582l2.514-.837a1 1 0 011.347 1.415l-.725 2.822a1 1 0 01-.99.682l-2.828-.235a1 1 0 00-.933.603l-.942 1.884a1 1 0 01-1.786-.326L9.4 10.82a1 1 0 00-.933-.603l-2.828.235a1 1 0 01-.99-.682l.725-2.822a1 1 0 011.347-1.415l2.514.837a1 1 0 001.218-.582l1.128-2.29z"/>
              </template>
            </svg>
          </span>
          <span v-if="!sidebarCollapsed" class="nav-label">{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <div class="user-info" v-if="user">
          <div class="avatar">{{ user.name?.charAt(0).toUpperCase() || 'U' }}</div>
          <div v-if="!sidebarCollapsed" class="user-details">
            <div class="user-name">{{ user.name || '用户' }}</div>
            <div class="user-email">{{ user.email }}</div>
          </div>
        </div>
        <button class="logout-btn" @click="handleLogout" :title="sidebarCollapsed ? '退出登录' : ''">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 102 0V4a1 1 0 00-1-1zm10.293 4.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L14.586 5H7a1 1 0 100 2h7.586l-1.293 1.293z"/>
          </svg>
          <span v-if="!sidebarCollapsed">退出</span>
        </button>
      </div>
    </aside>

    <!-- 主内容区 -->
    <main class="main-content">
      <header class="main-header">
        <div class="breadcrumb">
          <span class="breadcrumb-item">{{ route.meta?.title || route.name }}</span>
        </div>
        <div class="header-actions">
          <button class="icon-btn" title="通知">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <path d="M10 2a6 6 0 00-6 6v3.586l-.707.707A1 1 0 004 14h12a1 1 0 00.707-1.707L16 11.586V8a6 6 0 00-6-6zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z"/>
            </svg>
          </button>
        </div>
      </header>
      <div class="content-wrapper">
        <slot />
      </div>
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg);
}

.sidebar {
  width: 240px;
  background: var(--color-sidebar);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
}

.app-layout.collapsed .sidebar {
  width: 64px;
}

.sidebar-header {
  padding: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--color-border);
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
}

.collapse-btn {
  width: 28px;
  height: 28px;
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

.sidebar-nav {
  flex: 1;
  padding: 12px 8px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  margin-bottom: 4px;
  border-radius: 8px;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: all 0.2s;
}

.nav-item:hover {
  background: var(--color-hover);
  color: var(--color-text);
}

.nav-item.active {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.nav-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.nav-label {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--color-border);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  flex-shrink: 0;
}

.user-details {
  overflow: hidden;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-email {
  font-size: 12px;
  color: var(--color-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.logout-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: var(--color-error-light);
  border-color: var(--color-error);
  color: var(--color-error);
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.main-header {
  height: 60px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
}

.breadcrumb-item {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: var(--color-hover);
  color: var(--color-text);
}

.content-wrapper {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}
</style>
