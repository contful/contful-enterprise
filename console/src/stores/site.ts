// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getMySites, createSite, type Site, type CreateSiteParams } from '@/api/site'
import { showSuccess, showError } from '@/utils/request'

export const useSiteStore = defineStore('site', () => {
  const { t } = useI18n()

  // 当前选中的站点 ID
  const currentSiteId = ref<string | null>(localStorage.getItem('currentSiteId'))

  // 用户所属站点列表
  const sites = ref<Site[]>([])
  const loading = ref(false)

  // 当前站点对象
  const currentSite = computed(() =>
    sites.value.find(s => s.id === currentSiteId.value) || null
  )

  // 设置当前站点
  function setCurrentSite(siteId: string | null) {
    currentSiteId.value = siteId
    if (siteId) {
      localStorage.setItem('currentSiteId', siteId)
    } else {
      localStorage.removeItem('currentSiteId')
    }
  }

  // 加载用户站点列表
  async function fetchSites() {
    loading.value = true
    try {
      const res = await getMySites({ page: 1, page_size: 100 })
      if (res.code === 200) {
        sites.value = res.data.items || []

        // 如果当前没有选中站点，自动选中第一个
        if (!currentSiteId.value && sites.value.length > 0) {
          setCurrentSite(sites.value[0].id)
        }

        // 如果当前选中的站点已不在列表中，清除选择
        if (currentSiteId.value && !sites.value.find(s => s.id === currentSiteId.value)) {
          setCurrentSite(sites.value.length > 0 ? sites.value[0].id : null)
        }

        return sites.value
      }
      return []
    } catch (error) {
      showError(error)
      return []
    } finally {
      loading.value = false
    }
  }

  // 创建站点并自动设为当前站点
  async function createAndSwitch(data: CreateSiteParams) {
    try {
      const res = await createSite(data)
      if (res.code === 200) {
        showSuccess(t('site.created'))
        // 返回新创建的站点（响应中直接包含）
        const newSite = res.data
        if (newSite?.id) {
          setCurrentSite(newSite.id)
        }
        // 异步刷新站点列表（不阻塞返回）
        fetchSites().catch(() => {})
        return { success: true }
      }
      return { success: false, message: res.message }
    } catch (error: unknown) {
      const err = error as { response?: { data?: { msg?: string } } }
      const msg = err.response?.data?.msg || t('site.createFailed') || '创建站点失败'
      return { success: false, message: msg }
    }
  }

  // 清除站点状态（登出时调用）
  function clearSites() {
    sites.value = []
    currentSiteId.value = null
    localStorage.removeItem('currentSiteId')
  }

  return {
    currentSiteId,
    sites,
    loading,
    currentSite,
    setCurrentSite,
    fetchSites,
    createAndSwitch,
    clearSites,
  }
})
