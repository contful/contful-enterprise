<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getContentTypes, getContentEntries, getAssets, getUsers } from '@/api/api'
import { showError } from '@/utils/request'

const router = useRouter()

const stats = ref({
  contentTypes: 0,
  entries: 0,
  assets: 0,
  users: 0,
})
const recentEntries = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [typesRes, entriesRes, assetsRes, usersRes] = await Promise.all([
      getContentTypes({ page: 1, page_size: 1 }),
      getContentEntries({ page: 1, page_size: 5 }),
      getAssets({ page: 1, page_size: 1 }),
      getUsers({ page: 1, page_size: 1 }),
    ])

    stats.value = {
      contentTypes: typesRes.data.total || 0,
      entries: entriesRes.data.total || 0,
      assets: assetsRes.data.total || 0,
      users: usersRes.data.total || 0,
    }
    recentEntries.value = entriesRes.data.items || []
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
})

const quickActions = [
  { icon: 'add', label: '创建内容', path: '/content/entries', color: '#3b82f6' },
  { icon: 'upload', label: '上传媒体', path: '/assets', color: '#10b981' },
  { icon: 'schema', label: '管理类型', path: '/content/types', color: '#8b5cf6' },
  { icon: 'token', label: 'API Token', path: '/settings', color: '#f59e0b' },
]

const getStatusClass = (status: string) => {
  const map: Record<string, string> = {
    published: 'badge-success',
    draft: 'badge-warning',
    archived: 'badge-default',
  }
  return map[status] || 'badge-default'
}

const getStatusLabel = (status: string) => {
  const map: Record<string, string> = {
    published: '已发布',
    draft: '草稿',
    archived: '已归档',
  }
  return map[status] || status
}
</script>

<template>
  <div class="dashboard">
    <div class="page-header">
      <div>
        <h1 class="page-title">仪表盘</h1>
        <p class="page-subtitle">欢迎回来！以下是您的内容概览。</p>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" @click="router.push('/content/entries')">
        <div class="stat-icon" style="background: #eff6ff; color: #3b82f6;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M4 4a2 2 0 012-2h8a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 0v10h8V4H6z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.entries }}</div>
          <div class="stat-label">内容条目</div>
        </div>
      </div>

      <div class="stat-card" @click="router.push('/content/types')">
        <div class="stat-icon" style="background: #f3e8ff; color: #8b5cf6;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1v-2zm0 6a1 1 0 011-1h6a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1v-2z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.contentTypes }}</div>
          <div class="stat-label">内容类型</div>
        </div>
      </div>

      <div class="stat-card" @click="router.push('/assets')">
        <div class="stat-icon" style="background: #ecfdf5; color: #10b981;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm0 2h12v7l-4-3-2 1.5L6 12V5z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.assets }}</div>
          <div class="stat-label">媒体文件</div>
        </div>
      </div>

      <div class="stat-card" @click="router.push('/users')">
        <div class="stat-icon" style="background: #fef3c7; color: #f59e0b;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.users }}</div>
          <div class="stat-label">用户数</div>
        </div>
      </div>
    </div>

    <div class="dashboard-grid">
      <!-- 快速操作 -->
      <div class="card quick-actions">
        <h3 class="card-title">快速操作</h3>
        <div class="actions-list">
          <button
            v-for="action in quickActions"
            :key="action.label"
            class="action-item"
            @click="router.push(action.path)"
          >
            <span class="action-icon" :style="{ background: action.color }">
              <svg width="16" height="16" viewBox="0 0 20 20" fill="white">
                <path v-if="action.icon === 'add'" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"/>
                <path v-else-if="action.icon === 'upload'" d="M10 3a1 1 0 011 1v5.586l1.707-1.707a1 1 0 011.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 011.414-1.414L9 9.586V4a1 1 0 011-1zm5 10a1 1 0 100-2 1 1 0 000 2z"/>
                <path v-else-if="action.icon === 'schema'" d="M4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h6a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1v-2z"/>
                <path v-else-if="action.icon === 'token'" d="M7 7a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zm-4 4a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zM4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5z"/>
              </svg>
            </span>
            <span class="action-label">{{ action.label }}</span>
            <svg class="action-arrow" width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- 最近内容 -->
      <div class="card recent-entries">
        <h3 class="card-title">最近内容</h3>
        <div v-if="loading" class="loading">加载中...</div>
        <div v-else-if="recentEntries.length === 0" class="empty-state">
          <p>暂无内容</p>
          <button class="btn btn-primary btn-sm" @click="router.push('/content/entries')">
            创建第一个内容
          </button>
        </div>
        <table v-else class="table">
          <thead>
            <tr>
              <th>标题</th>
              <th>类型</th>
              <th>状态</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="entry in recentEntries"
              :key="entry.id"
              @click="router.push(`/content/entries?type=${entry.content_type_id}&id=${entry.id}`)"
              style="cursor: pointer;"
            >
              <td>{{ entry.title || entry.id.slice(0, 8) }}</td>
              <td>{{ entry.content_type_id?.slice(0, 8) || '-' }}</td>
              <td>
                <span :class="['badge', getStatusClass(entry.status)]">
                  {{ getStatusLabel(entry.status) }}
                </span>
              </td>
              <td>{{ new Date(entry.updated_time).toLocaleDateString('zh-CN') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: 1400px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.stat-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.1);
}

.stat-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--color-text);
}

.stat-label {
  font-size: 14px;
  color: var(--color-text-secondary);
}

.dashboard-grid {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 20px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 16px;
}

.quick-actions .actions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
  width: 100%;
}

.action-item:hover {
  background: var(--color-hover);
  border-color: var(--color-primary);
}

.action-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.action-label {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
}

.action-arrow {
  color: var(--color-text-secondary);
}

.loading {
  text-align: center;
  padding: 40px;
  color: var(--color-text-secondary);
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
