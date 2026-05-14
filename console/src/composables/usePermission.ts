// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { useUserStore } from '@/stores/user'

/**
 * 权限控制 composable
 *
 * 用法:
 * const { hasPermission, hasAnyPermission } = usePermission()
 * hasPermission('users:read') // 检查单个权限
 * hasAnyPermission(['users:read', 'users:write']) // 检查多个权限中的任意一个
 */
export function usePermission() {
  const userStore = useUserStore()

  /**
   * 检查是否拥有指定权限
   * @param permission 权限 key，如 'users:read'
   */
  const hasPermission = (permission: string): boolean => {
    return userStore.hasPermission(permission)
  }

  /**
   * 检查是否拥有任意一个指定权限
   * @param permissions 权限 key 数组
   */
  const hasAnyPermission = (permissions: string[]): boolean => {
    return userStore.hasAnyPermission(permissions)
  }

  return {
    hasPermission,
    hasAnyPermission,
  }
}
