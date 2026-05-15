// 系统配置相关类型定义

export interface PasswordPolicy {
  min_length: number
  require_uppercase: boolean
  require_lowercase: boolean
  require_number: boolean
  require_special: boolean
  expire_days: number
}

export interface SystemConfig {
  id: string
  config_key: string
  config_value: string
  value_type: string
  description: string
  is_public: boolean
  is_system: boolean
  created_time: string
  updated_time: string
}
