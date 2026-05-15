import request from '@/utils/request'
import type { PasswordPolicy, SystemConfig } from '@/types/system/config'

// 站点公开配置（登录页使用，无需认证）
export interface SiteConfig {
  site_name: string
  site_description: string
  logo_url: string
  login_background_url: string
  mfa_enforced: boolean
  login_max_attempts: number
  login_lock_duration: number
  password_min_length: number
  password_require_uppercase: boolean
  password_require_lowercase: boolean
  password_require_number: boolean
  password_require_special: boolean
  password_expire_days: number
}

// 获取站点公开配置（无需认证，含品牌/策略/密码规则）
export const getSiteConfig = async (): Promise<SiteConfig> => {
  const res = await request.get('/system/config/site')
  return res.data.data
}

// 获取密码策略（已合并到 getSiteConfig，保留兼容）
export const getPasswordPolicy = async (): Promise<PasswordPolicy> => {
  const site = await getSiteConfig()
  return {
    min_length: site.password_min_length,
    require_uppercase: site.password_require_uppercase,
    require_lowercase: site.password_require_lowercase,
    require_number: site.password_require_number,
    require_special: site.password_require_special,
    expire_days: site.password_expire_days,
  }
}

// 获取公开配置
export const getPublicConfig = async (): Promise<Record<string, string>> => {
  const res = await request.get('/system/config/public')
  return res.data.data
}

// 获取所有配置（需要 settings:read 权限）
export const getSystemConfigs = async (): Promise<SystemConfig[]> => {
  const res = await request.get('/system/config')
  return res.data.data
}

// 获取单个配置
export const getSystemConfig = async (key: string): Promise<SystemConfig> => {
  const res = await request.get(`/system/config/${key}`)
  return res.data.data
}

// 更新配置（需要 settings:write 权限）
export const updateSystemConfig = (key: string, data: Partial<SystemConfig>): Promise<void> => {
  return request.put(`/system/config/${key}`, data)
}

// 创建配置（需要 settings:write 权限）
export const createSystemConfig = async (data: Partial<SystemConfig>): Promise<SystemConfig> => {
  const res = await request.post('/system/config', data)
  return res.data.data
}

// 删除配置（需要 settings:write 权限，仅自定义配置可删除）
export const deleteSystemConfig = (key: string): Promise<void> => {
  return request.delete(`/system/config/${key}`)
}
