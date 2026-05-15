// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post, put, del } from '@/utils/request'

export interface User {
  id: string
  name: string
  email: string
  role: string
  status: string
  created_time: string
  updated_time: string
}

export interface UserListParams {
  page?: number
  page_size?: number
  keyword?: string
  role?: string
  status?: string
}

export interface UserListResponse {
  items: User[]
  total: number
  page: number
  page_size: number
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
export function getUsers(params?: UserListParams) {
  return get<UserListResponse>('/users', { params })
}

// 获取当前用户
export function getCurrentUser() {
  return get<User>('/users/me')
}

// 获取用户详情
export function getUser(id: string) {
  return get<User>(`/users/${id}`)
}

// 创建用户
export function createUser(data: CreateUserData) {
  return post<User>('/users', data)
}

// 更新用户
export function updateUser(id: string, data: UpdateUserData) {
  return put<User>(`/users/${id}`, data)
}

// 删除用户
export function deleteUser(id: string) {
  return del(`/users/${id}`)
}

// 更新密码
export function updatePassword(id: string, oldPassword: string, newPassword: string) {
  return put<void>(`/users/${id}/password`, { old_password: oldPassword, new_password: newPassword })
}

// 管理员重置用户密码（不需要旧密码）
export function resetPassword(id: string, newPassword: string) {
  return post<void>(`/users/${id}/reset-password`, { new_password: newPassword })
}

// 数据签名/验签
export interface VerifyResult {
  valid: boolean
  algorithm: string
  signature: string
  payload: string
  reason?: string
}

export function signUser(id: string) {
  return post(`/users/${id}/sign`)
}

export function verifyUser(id: string) {
  return post<VerifyResult>(`/users/${id}/verify`)
}
