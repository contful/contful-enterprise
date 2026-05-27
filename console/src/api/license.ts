import { get } from '@/utils/request'

export interface LicenseInfo {
  status: string
  customer: string
  product_name: string
  product_version: string
  product_code: string
  is_trial: boolean
  is_expired: boolean
  issued_date: string
  expiry_date: string
  message?: string
}

export const getLicenseInfo = async (): Promise<LicenseInfo> => {
  const res = await get<LicenseInfo>('/system/license')
  return res.data
}
