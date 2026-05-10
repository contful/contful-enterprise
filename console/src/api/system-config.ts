import request from '@/utils/request'
import type { PasswordPolicy, SystemConfig } from '@/types/system-config'

// 获取密码策略（公开 API，无需认证）
export const getPasswordPolicy = (): Promise<PasswordPolicy> => {
  return request.get('/admin/api/v1/system/config/password-policy')
}

// 获取公开配置
export const getPublicConfig = (): Promise<Record<string, string>> => {
  return request.get('/admin/api/v1/system/config/public')
}

// 获取所有配置（需要 settings:read 权限）
export const getSystemConfigs = (): Promise<SystemConfig[]> => {
  return request.get('/admin/api/v1/system/config')
}

// 获取单个配置
export const getSystemConfig = (key: string): Promise<SystemConfig> => {
  return request.get(`/admin/api/v1/system/config/${key}`)
}

// 更新配置（需要 settings:write 权限）
export const updateSystemConfig = (key: string, data: Partial<SystemConfig>): Promise<void> => {
  return request.put(`/admin/api/v1/system/config/${key}`, data)
}
