// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getAccessToken, initializeSession } from '@/utils/request'

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
    path: '/configs',
    name: 'Configs',
    component: () => import('@/pages/sites/Config.vue'),
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
    path: '/sites/:siteId/config',
    name: 'SiteConfig',
    component: () => import('@/pages/sites/Config.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach(async (to, _from, next) => {
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
    next('/login')
  } else if (to.path === '/login' && token) {
    // 已登录用户访问登录页，跳转到首页
    next('/')
  } else {
    next()
  }
})

export default router
