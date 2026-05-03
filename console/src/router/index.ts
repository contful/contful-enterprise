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
    path: '/content/types',
    name: 'ContentTypes',
    component: () => import('@/pages/content-types/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/content/types/:id/fields',
    name: 'ContentTypeFields',
    component: () => import('@/pages/content-types/Fields.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/content/entries',
    name: 'Content',
    component: () => import('@/pages/content/List.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/content/entries/:slug',
    name: 'ContentBySlug',
    component: () => import('@/pages/content/List.vue'),
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
    name: 'Settings',
    component: () => import('@/pages/settings/SiteSettings.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/tokens',
    name: 'ApiTokens',
    component: () => import('@/pages/settings/ApiTokens.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/configs',
    name: 'Configs',
    component: () => import('@/pages/settings/Configs.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/pages/users/List.vue'),
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
