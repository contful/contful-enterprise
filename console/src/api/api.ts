// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import request from '@/utils/request'
import { getContentSchemas } from './schema'
import { getEntries } from './entry'
import { getAssets } from './asset'
import { getUsers } from './user'

export { getContentSchemas, getEntries as getContentEntries, getAssets, getUsers }

// 仪表盘统计（单个接口返回全部数据）
export function getDashboardStats() {
  return request.get('/dashboard/stats')
}
