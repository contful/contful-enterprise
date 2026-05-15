import request from '@/utils/request'
import type { PasswordPolicy, SystemConfig } from '@/types/system-config'

// 获取密码策略（公开 API，无需认证）
export const getPasswordPolicy = async (): Promise<PasswordPolicy> => {
  const res = await request.get('/system/config/password/policy')
  return (res as any).data
}

// 获取公开配置
export const getPublicConfig = async (): Promise<Record<string, string>> => {
  const res = await request.get('/system/config/public')
  return (res as any).data
}

// 获取所有配置（需要 settings:read 权限）
export const getSystemConfigs = async (): Promise<SystemConfig[]> => {
  const res = await request.get('/system/config')
  return (res as any).data
}

// 获取单个配置
export const getSystemConfig = async (key: string): Promise<SystemConfig> => {
  const res = await request.get(`/system/config/${key}`)
  return (res as any).data
}

// 更新配置（需要 settings:write 权限）
export const updateSystemConfig = (key: string, data: Partial<SystemConfig>): Promise<void> => {
  return request.put(`/system/config/${key}`, data)
}
