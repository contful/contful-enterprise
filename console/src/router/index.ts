// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { getAccessToken, initializeSession } from '@/utils/request'
import { useUserStore } from '@/stores/user'
import { MessagePlugin } from 'tdesign-vue-next'
import { i18n } from '@/locales'

// ─────────────────────────────────────────────────────────────
// 路由 → 权限映射
// ─────────────────────────────────────────────────────────────
const ROUTE_PERMISSION_MAP: Record<string, string> = {
  '/': 'dashboard:read',
  '/dashboard': 'dashboard:read',
  '/users': 'users:read',
  '/sites': 'sites:read',
  '/content/schemas': 'schema:read',
  '/content/entries': 'entry:read',
  '/assets': 'asset:read',
  '/tokens': 'tokens:read',
  '/settings/tokens': 'tokens:read',
  '/system/config': 'settings:read',
  '/system/roles': 'roles:read',
  '/system/permissions': 'roles:read',
  '/audit/logs': 'audit:read',
}

// 获取路由对应的权限
function getRoutePermission(path: string): string | null {
  // 精确匹配
  if (ROUTE_PERMISSION_MAP[path]) {
    return ROUTE_PERMISSION_MAP[path]
  }
  // 模糊匹配（处理 /content/schemas/:id/fields 这种情况）
  for (const [route, permission] of Object.entries(ROUTE_PERMISSION_MAP)) {
    if (path.startsWith(route + '/') || path.startsWith(route + '?')) {
      return permission
    }
  }
  return null
}

// ─────────────────────────────────────────────────────────────
// 路由配置
// ─────────────────────────────────────────────────────────────
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/auth/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/mfa',
    name: 'MFA',
    component: () => import('@/pages/auth/MFA.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/pages/Dashboard.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/pages/settings/Profile.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/system/config',
    name: 'SystemConfig',
    component: () => import('@/pages/system/Config.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/content/schemas',
    name: 'ContentSchemas',
    component: () => import('@/pages/schemas/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/content/schemas/:id/fields',
    name: 'ContentSchemaFields',
    component: () => import('@/pages/schemas/Fields.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/content/entries',
    name: 'Content',
    component: () => import('@/pages/entries/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/content/entries/:slug',
    name: 'ContentBySlug',
    component: () => import('@/pages/entries/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/assets',
    name: 'Media',
    component: () => import('@/pages/media/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/settings',
    redirect: '/sites',
  },
  {
    path: '/tokens',
    name: 'ApiTokens',
    component: () => import('@/pages/tokens/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/pages/users/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/sites',
    name: 'Sites',
    component: () => import('@/pages/sites/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/sites/:siteId/setting',
    name: 'SiteSetting',
    component: () => import('@/pages/sites/Setting.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/system/roles',
    name: 'SystemRoles',
    component: () => import('@/pages/system/Roles.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/system/permissions',
    name: 'SystemPermissions',
    component: () => import('@/pages/system/Permissions.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/audit/logs',
    name: 'AuditLogs',
    component: () => import('@/pages/audit/List.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// ─────────────────────────────────────────────────────────────
// 路由守卫
// ─────────────────────────────────────────────────────────────
router.beforeEach(async (to, _from) => {
  const requiresAuth = to.meta.requiresAuth !== false
  let token = getAccessToken()

  // 如果没有 AccessToken，尝试从 Cookie 刷新恢复会话
  if (requiresAuth && !token) {
    const restored = await initializeSession()
    if (restored) {
      token = getAccessToken()
    }
  }

  if (requiresAuth && !token) {
    return '/login'
  } else if (to.path === '/login' && token) {
    // 已登录用户访问登录页，跳转到首页
    return '/'
  } else if (requiresAuth && token) {
    // 权限检查
    const userStore = useUserStore()
    const t = i18n.global.t

    // 如果还没有获取过用户信息，先获取
    if (!userStore.user) {
      const ok = await userStore.fetchUser()
      if (!ok) {
        return '/login'
      }
    }

    // 如果还没有获取过权限列表，先获取
    if (!userStore.isSuperAdmin && userStore.permissions.length === 0) {
      await userStore.fetchPermissions()
    }

    // 获取目标路由需要的权限
    const requiredPermission = getRoutePermission(to.path)

    // 如果路由没有定义权限要求，直接放行
    if (requiredPermission === null) {
      return
    }

    // 检查权限
    if (userStore.hasPermission(requiredPermission)) {
      return
    } else {
      // 无权访问，显示提示并跳转到首页
      MessagePlugin.warning(t('permission.denied'))
      return '/'
    }
  }
})

export default router
