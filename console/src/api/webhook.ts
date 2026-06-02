// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
import { get, post, put, del } from '@/utils/request'

export interface Webhook {
  id: string
  site_id: string
  name: string
  url: string
  events: string[]
  secret: string
  is_active: boolean
  created_time: string
  updated_time: string
}

export interface WebhookDelivery {
  id: string
  webhook_id: string
  event: string
  payload: string
  response_status: number
  response_body: string
  status: string
  attempt: number
  error_message: string
  created_time: string
}

export const EVENT_OPTIONS = [
  { value: 'content.created', label: '条目创建' },
  { value: 'content.updated', label: '条目更新' },
  { value: 'content.deleted', label: '条目删除' },
  { value: 'content.published', label: '条目发布' },
  { value: 'content.unpublished', label: '条目下架' },
]

export function listWebhooks() {
  return get<Webhook[]>('/webhooks')
}
export function getWebhook(id: string) {
  return get<Webhook>(`/webhooks/${id}`)
}
export function createWebhook(data: Partial<Webhook>) {
  return post<Webhook>('/webhooks', data)
}
export function updateWebhook(id: string, data: Partial<Webhook>) {
  return put<Webhook>(`/webhooks/${id}`, data)
}
export function deleteWebhook(id: string) {
  return del(`/webhooks/${id}`)
}
export function listDeliveries(webhookId: string) {
  return get<WebhookDelivery[]>(`/webhooks/${webhookId}/deliveries`)
}
export function testWebhook(id: string) {
  return post(`/webhooks/${id}/test`, {})
}
