// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post } from '@/utils/request'

/** 安装状态响应 */
export interface SetupStatus {
  setup_required: boolean
  version: string
}

/** 数据库连接配置 */
export interface DatabaseConfig {
  host: string
  port: number
  user: string
  password: string
  db_name: string
  ssl_mode: string
}

/** 管理员账号配置 */
export interface AdminConfig {
  email: string
  password: string
  password_confirm: string
  site_name: string
  site_slug: string
}

/** 操作结果 */
export interface SetupResult {
  success: boolean
  message: string
}

/**
 * 获取安装状态
 * 用于判断是否需要进入安装向导
 */
export const getSetupStatus = () =>
  get<SetupStatus>('/setup/status')

/**
 * 测试数据库连接
 */
export const testDatabase = (config: DatabaseConfig) =>
  post<SetupResult>('/setup/database', config)

/**
 * 初始化数据库表结构
 */
export const initializeDatabase = (config: DatabaseConfig) =>
  post<SetupResult>('/setup/initialize', config)

/**
 * 创建管理员账号并完成安装
 */
export const createAdmin = (config: AdminConfig) =>
  post<SetupResult>('/setup/admin', config)
