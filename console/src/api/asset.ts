import request from './request'
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
  created_at: string
  updated_at: string
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
  created_at: string
  updated_at: string
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

    const response = await request.post<AssetResponse>('/admin/v1/assets', formData, {
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
      `/admin/v1/assets?${queryParams.toString()}`
    )

    return response.data
  },

  /**
   * 获取资源详情
   */
  get: async (id: string): Promise<AssetResponse> => {
    const response = await request.get<AssetResponse>(`/admin/v1/assets/${id}`)
    return response.data
  },

  /**
   * 更新资源
   */
  update: async (id: string, data: AssetUpdate): Promise<AssetResponse> => {
    const response = await request.put<AssetResponse>(`/admin/v1/assets/${id}`, data)
    return response.data
  },

  /**
   * 删除资源
   */
  delete: async (id: string): Promise<void> => {
    await request.delete(`/admin/v1/assets/${id}`)
  },

  /**
   * 批量删除资源
   */
  batchDelete: async (ids: string[]): Promise<void> => {
    await request.delete('/admin/v1/assets', { data: { ids } })
  },
}

// ============ Folder API ============

export const folderApi = {
  /**
   * 创建文件夹
   */
  create: async (data: FolderCreate): Promise<FolderResponse> => {
    const response = await request.post<FolderResponse>('/admin/v1/assets/folders', data)
    return response.data
  },

  /**
   * 获取文件夹树
   */
  getTree: async (): Promise<FolderResponse[]> => {
    const response = await request.get<FolderResponse[]>('/admin/v1/assets/folders/tree')
    return response.data
  },

  /**
   * 获取文件夹列表
   */
  list: async (parentId?: string): Promise<FolderResponse[]> => {
    const params = parentId ? { parent_id: parentId } : {}
    const response = await request.get<FolderResponse[]>('/admin/v1/assets/folders', { params })
    return response.data
  },

  /**
   * 获取文件夹详情
   */
  get: async (id: string): Promise<FolderResponse> => {
    const response = await request.get<FolderResponse>(`/admin/v1/assets/folders/${id}`)
    return response.data
  },

  /**
   * 更新文件夹
   */
  update: async (id: string, data: FolderUpdate): Promise<FolderResponse> => {
    const response = await request.put<FolderResponse>(`/admin/v1/assets/folders/${id}`, data)
    return response.data
  },

  /**
   * 删除文件夹
   */
  delete: async (id: string): Promise<void> => {
    await request.delete(`/admin/v1/assets/folders/${id}`)
  },
}

// 便捷导出
export default assetApi
