// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get } from '@/utils/request'
import { getContentSchemas } from './schema'
import { getAssets } from './asset'
import { getUsers } from './user'

export { getContentSchemas, getAssets, getUsers }

// 仪表盘统计（单个接口返回全部数据）
export function getDashboardStats() {
  return get('/dashboard/stats')
}
