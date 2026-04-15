// Asset 类型别名
export type Asset = {
  id: string
  site_id: string
  folder_id?: string
  name: string
  mime_type: string
  size: number
  url: string
  thumbnail_url?: string
  created_at: string
  updated_at: string
}

// Folder 类型别名
export type Folder = {
  id: string
  site_id: string
  parent_id?: string
  name: string
  slug: string
  path: string
  sort_order: number
  children?: Folder[]
  created_at: string
  updated_at: string
}
