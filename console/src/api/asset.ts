import request from '@/utils/request'
import type { AxiosProgressEvent } from 'axios'

// ============ Types ============

export type AssetType = 'image' | 'video' | 'audio' | 'document' | 'file'

export type AssetVisibility = 'public' | 'private'

export interface AssetResponse {
  id: string
  site_id: string
  folder_id?: string
  uuid: string
  name: string
  original_name: string
  slug: string
  type: AssetType
  mime_type: string
  extension: string
  size: number
  width?: number
  height?: number
  duration?: number
  path: string
  url: string
  thumbnail_url?: string
  alt?: string
  title?: string
  caption?: string
  alt_text?: string
  description?: string
  tags?: string[]
  visibility: AssetVisibility
  file_hash: string
  disk: string
  download_count: number
  used_count: number
  created_by?: string
  created_time: string
  updated_time: string
}

export interface AssetListFilter {
  folder_id?: string
  type?: AssetType
  extension?: string
  tag?: string
  keyword?: string
}

export interface AssetListResponse {
  items: AssetResponse[]
  total: number
  page: number
  page_size: number
}

export interface AssetUpdate {
  folder_id?: string
  name?: string
  alt?: string
  title?: string
  caption?: string
  alt_text?: string
  description?: string
  tags?: string[]
  visibility?: AssetVisibility
}

export interface FolderResponse {
  id: string
  site_id: string
  parent_id?: string
  name: string
  slug: string
  path: string
  sort_order: number
  children?: FolderResponse[]
  assets?: AssetResponse[]
  created_by?: string
  created_time: string
  updated_time: string
}

export interface FolderCreate {
  parent_id?: string
  name: string
  sort_order?: number
}

export interface FolderUpdate {
  parent_id?: string
  name?: string
  sort_order?: number
}

// ============ Asset API ============

export const assetApi = {
  /**
   * 上传资源
   */
  upload: async (
    file: File,
    options?: {
      folder_id?: string
      alt?: string
      title?: string
      onProgress?: (event: AxiosProgressEvent) => void
    }
  ): Promise<AssetResponse> => {
    const formData = new FormData()
    formData.append('file', file)

    if (options?.folder_id) {
      formData.append('folder_id', options.folder_id)
    }
    if (options?.alt) {
      formData.append('alt', options.alt)
    }
    if (options?.title) {
      formData.append('title', options.title)
    }

    const response = await request.post<AssetResponse>('/assets', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: options?.onProgress,
    })

    return response.data
  },

  /**
   * 批量上传资源
   */
  uploadMultiple: async (
    files: File[],
    options?: {
      folder_id?: string
      alt?: string
      onProgress?: (event: AxiosProgressEvent) => void
    }
  ): Promise<AssetResponse[]> => {
    const results: AssetResponse[] = []

    for (const file of files) {
      const result = await assetApi.upload(file, options)
      results.push(result)
    }

    return results
  },

  /**
   * 获取资源列表
   */
  list: async (params?: {
    page?: number
    page_size?: number
    filter?: AssetListFilter
  }): Promise<AssetListResponse> => {
    const { page = 1, page_size = 20, filter } = params || {}

    const queryParams = new URLSearchParams()
    queryParams.append('page', String(page))
    queryParams.append('page_size', String(page_size))

    if (filter) {
      if (filter.folder_id) queryParams.append('folder_id', filter.folder_id)
      if (filter.type) queryParams.append('type', filter.type)
      if (filter.extension) queryParams.append('extension', filter.extension)
      if (filter.tag) queryParams.append('tag', filter.tag)
      if (filter.keyword) queryParams.append('keyword', filter.keyword)
    }

    const response = await request.get<AssetListResponse>(
      `/assets?${queryParams.toString()}`
    )

    return response.data
  },

  /**
   * 获取资源详情
   */
  get: async (id: string): Promise<AssetResponse> => {
    const response = await request.get<AssetResponse>(`/assets/${id}`)
    return response.data
  },

  /**
   * 更新资源
   */
  update: async (id: string, data: AssetUpdate): Promise<AssetResponse> => {
    const response = await request.put<AssetResponse>(`/assets/${id}`, data)
    return response.data
  },

  /**
   * 删除资源
   */
  delete: async (id: string): Promise<void> => {
    await request.delete(`/assets/${id}`)
  },

  /**
   * 批量删除资源
   */
  batchDelete: async (ids: string[]): Promise<void> => {
    await request.delete('/assets/batch-delete', { data: { ids } })
  },
}

// ============ Folder API ============

export const folderApi = {
  /**
   * 创建文件夹
   */
  create: async (data: FolderCreate): Promise<FolderResponse> => {
    const response = await request.post<FolderResponse>('/assets/folders', data)
    return response.data
  },

  /**
   * 获取文件夹树
   */
  getTree: async (): Promise<FolderResponse[]> => {
    const response = await request.get<FolderResponse[]>('/assets/folders/tree')
    // API 返回 { code, message, data: [...] }，需要取 response.data.data
    return response.data.data || []
  },

  /**
   * 获取文件夹列表
   */
  list: async (parentId?: string): Promise<FolderResponse[]> => {
    const params = parentId ? { parent_id: parentId } : {}
    const response = await request.get<FolderResponse[]>('/assets/folders', { params })
    // API 返回 { code, message, data: [...] }，需要取 response.data.data
    return response.data.data || []
  },

  /**
   * 获取文件夹详情
   */
  get: async (id: string): Promise<FolderResponse> => {
    const response = await request.get<FolderResponse>(`/assets/folders/${id}`)
    return response.data
  },

  /**
   * 更新文件夹
   */
  update: async (id: string, data: FolderUpdate): Promise<FolderResponse> => {
    const response = await request.put<FolderResponse>(`/assets/folders/${id}`, data)
    return response.data
  },

  /**
   * 删除文件夹
   */
  delete: async (id: string): Promise<void> => {
    await request.delete(`/assets/folders/${id}`)
  },
}

// 便捷导出
export default assetApi

// ============ 便捷函数导出 ============
// 与 media/List.vue 兼容的导出
export const getAssets = (params?: { page?: number; page_size?: number; folder_id?: string; type?: string; keyword?: string }) => {
  return assetApi.list({ page: params?.page, page_size: params?.page_size, filter: params })
}

export const getAsset = (id: string) => {
  return assetApi.get(id)
}

export const createAsset = (data: { file: File; folder_id?: string; alt?: string; title?: string }) => {
  return assetApi.upload(data.file, { folder_id: data.folder_id, alt: data.alt, title: data.title })
}

export const updateAsset = (id: string, data: AssetUpdate) => {
  return assetApi.update(id, data)
}

export const deleteAsset = (id: string) => {
  return assetApi.delete(id)
}

export const batchDeleteAssets = (ids: string[]) => {
  return assetApi.batchDelete(ids)
}

// 文件夹相关
export const getAssetFolders = () => {
  return folderApi.getTree()
}

export const createFolder = (data: FolderCreate) => {
  return folderApi.create(data)
}

export const updateFolder = (id: string, data: FolderUpdate) => {
  return folderApi.update(id, data)
}

export const deleteFolder = (id: string) => {
  return folderApi.delete(id)
}

// 导出类型
export type { Asset, Folder } from './asset-types'
