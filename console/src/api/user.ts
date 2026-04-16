import request from '@/utils/request'

export interface User {
  id: string
  name: string
  email: string
  role: string
  status: string
  created_at: string
  updated_at: string
}

export interface UserListParams {
  page?: number
  page_size?: number
  keyword?: string
  role?: string
  status?: string
}

export interface CreateUserData {
  name: string
  email: string
  password: string
  role?: string
}

export interface UpdateUserData {
  name?: string
  email?: string
  password?: string
  role?: string
  status?: string
}

// 获取用户列表
export function getUsers(params: UserListParams) {
  return request<{ items: User[]; total: number; page: number; page_size: number }>({
    url: '/admin/api/v1/users',
    method: 'get',
    params,
  })
}

// 获取当前用户
export function getCurrentUser() {
  return request<User>({
    url: '/admin/api/v1/users/me',
    method: 'get',
  })
}

// 获取用户详情
export function getUser(id: string) {
  return request<User>({
    url: `/admin/api/v1/users/${id}`,
    method: 'get',
  })
}

// 创建用户
export function createUser(data: CreateUserData) {
  return request<User>({
    url: '/admin/api/v1/users',
    method: 'post',
    data,
  })
}

// 更新用户
export function updateUser(id: string, data: UpdateUserData) {
  return request<User>({
    url: `/admin/api/v1/users/${id}`,
    method: 'put',
    data,
  })
}

// 删除用户
export function deleteUser(id: string) {
  return request<void>({
    url: `/admin/api/v1/users/${id}`,
    method: 'delete',
  })
}

// 更新密码
export function updatePassword(id: string, oldPassword: string, newPassword: string) {
  return request<void>({
    url: `/admin/api/v1/users/${id}/password`,
    method: 'put',
    data: { old_password: oldPassword, new_password: newPassword },
  })
}
