// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getLicenseInfo, type LicenseInfo } from '@/api/license'

export const useLicenseStore = defineStore('license', () => {
  const info = ref<LicenseInfo | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 派生状态
  const isUnlicensed = computed(() => !info.value || info.value.status === 'unlicensed')
  const isExpired = computed(() => !!info.value?.is_expired)
  const isTrial = computed(() => !!info.value?.is_trial)
  const productName = computed(() => info.value?.product_name || null)
  const productVersion = computed(() => info.value?.product_version || null)
  const productCode = computed(() => info.value?.product_code || null)
  const customer = computed(() => info.value?.customer || null)
  const issuedDate = computed(() => info.value?.issued_date || null)
  const expiryDate = computed(() => info.value?.expiry_date || null)
  const status = computed(() => info.value?.status || 'unlicensed')

  // 加载 License 信息
  async function fetchLicense() {
    loading.value = true
    error.value = null
    try {
      const data = await getLicenseInfo()
      info.value = data
    } catch (err: unknown) {
      const e = err as { response?: { data?: { msg?: string } } }
      error.value = e.response?.data?.msg || 'Failed to load license info'
      info.value = null
    } finally {
      loading.value = false
    }
  }

  return {
    info,
    loading,
    error,
    isUnlicensed,
    isExpired,
    isTrial,
    productName,
    productVersion,
    productCode,
    customer,
    issuedDate,
    expiryDate,
    status,
    fetchLicense,
  }
})
