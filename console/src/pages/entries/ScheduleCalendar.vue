<script setup lang="ts">
// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { showError } from '@/utils/request'
import {
  getScheduledEntries,
  type Entry,
} from '@/api/entry'
import type { ContentSchema } from '@/api/schema'

const props = defineProps<{
  selectedType: ContentSchema | null
  contentSchemas: ContentSchema[]
}>()

const emit = defineEmits<{
  (e: 'editEntry', entry: Entry): void
}>()

const { t } = useI18n()
const siteStore = useSiteStore()

// 当前月
const currentYear = ref(new Date().getFullYear())
const currentMonth = ref(new Date().getMonth() + 1) // 0-indexed → 1-indexed

// 排期数据
const scheduledEntries = ref<Entry[]>([])
const loading = ref(false)

// 计算月视图的日期网格
const calendarDays = computed(() => {
  const year = currentYear.value
  const month = currentMonth.value

  // 本月第一天
  const firstDay = new Date(year, month - 1, 1)
  const startDayOfWeek = firstDay.getDay() // 0=Sun

  // 本月总天数
  const daysInMonth = new Date(year, month, 0).getDate()

  const days: (number | null)[] = []

  // 上月补齐
  for (let i = 0; i < startDayOfWeek; i++) {
    days.push(null)
  }

  // 本月日期
  for (let d = 1; d <= daysInMonth; d++) {
    days.push(d)
  }

  return days
})

// 每天的排期条目
const entriesByDay = computed(() => {
  const map: Record<number, Entry[]> = {}
  for (const entry of scheduledEntries.value) {
    const date = entry.scheduled_publish_time || entry.scheduled_unpublish_time
    if (!date) continue
    const d = new Date(date)
    const year = d.getFullYear()
    const month = d.getMonth() + 1
    if (year !== currentYear.value || month !== currentMonth.value) continue
    const day = d.getDate()
    if (!map[day]) map[day] = []
    // 最多显示 3 条
    if (map[day].length < 3) {
      map[day].push(entry)
    }
  }
  return map
})

// 月名
const monthLabel = computed(() => {
  const d = new Date(currentYear.value, currentMonth.value - 1, 1)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long' })
})

// 周标题
const weekDays = ['日', '一', '二', '三', '四', '五', '六']

// 加载排期
const loadScheduled = async () => {
  if (!siteStore.currentSiteId) return
  loading.value = true
  try {
    const year = currentYear.value
    const month = currentMonth.value
    const from = new Date(year, month - 1, 1).toISOString()
    const to = new Date(year, month, 0, 23, 59, 59).toISOString()
    const res = await getScheduledEntries({
      status: 'all',
      from,
      to,
      page: 1,
      page_size: 200,
    })
    scheduledEntries.value = res.data?.items || []
  } catch (error) {
    showError(error as any)
  } finally {
    loading.value = false
  }
}

// 导航
const prevMonth = () => {
  if (currentMonth.value === 1) {
    currentYear.value--
    currentMonth.value = 12
  } else {
    currentMonth.value--
  }
}

const nextMonth = () => {
  if (currentMonth.value === 12) {
    currentYear.value++
    currentMonth.value = 1
  } else {
    currentMonth.value++
  }
}

// 判断今天
const isToday = (day: number) => {
  const now = new Date()
  return (
    currentYear.value === now.getFullYear() &&
    currentMonth.value === now.getMonth() + 1 &&
    day === now.getDate()
  )
}

// 点击条目
const handleEntryClick = (entry: Entry) => {
  emit('editEntry', entry)
}

// 监听年月变化重新加载
watch([currentYear, currentMonth], () => {
  loadScheduled()
})

// 获取排期标签文本
const getScheduleTag = (entry: Entry): string => {
  if (entry.scheduled_publish_time && entry.status === 'draft') return t('schedule.pendingPublish')
  if (entry.scheduled_unpublish_time && entry.status === 'published') return t('schedule.pendingUnpublish')
  return ''
}

const getScheduleTagTheme = (entry: Entry): string => {
  if (entry.scheduled_publish_time && entry.status === 'draft') return 'primary'
  if (entry.scheduled_unpublish_time && entry.status === 'published') return 'warning'
  return 'default'
}

const getEntryTitle = (entry: Entry): string => {
  return entry.values?.title?.value || entry.values?.title || entry.id.slice(0, 8)
}

onMounted(() => {
  loadScheduled()
})
</script>

<template>
  <div class="calendar-container">
    <!-- 月导航 -->
    <div class="calendar-header">
      <t-button variant="text" @click="prevMonth">
        <t-icon name="chevron-left" />
      </t-button>
      <h3 class="month-label">{{ monthLabel }}</h3>
      <t-button variant="text" @click="nextMonth">
        <t-icon name="chevron-right" />
      </t-button>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="calendar-loading">
      <t-loading size="small" />
      <span>{{ t('common.loading') }}</span>
    </div>

    <!-- 日历网格 -->
    <div v-else class="calendar-grid">
      <!-- 周标题 -->
      <div
        v-for="wd in weekDays"
        :key="wd"
        class="weekday-header"
      >{{ wd }}</div>

      <!-- 日期格子 -->
      <div
        v-for="(day, idx) in calendarDays"
        :key="idx"
        class="day-cell"
        :class="{ today: day && isToday(day), empty: !day }"
      >
        <template v-if="day">
          <span class="day-number" :class="{ today: isToday(day) }">{{ day }}</span>
          <div class="day-entries">
            <div
              v-for="entry in entriesByDay[day]"
              :key="entry.id"
              class="day-entry-item"
              @click="handleEntryClick(entry)"
            >
              <span class="entry-title">{{ getEntryTitle(entry) }}</span>
              <t-tag
                :theme="(getScheduleTagTheme(entry) as any)"
                variant="light"
                size="small"
              >{{ getScheduleTag(entry) }}</t-tag>
            </div>
            <div
              v-if="(entriesByDay[day]?.length || 0) === 0 && day"
              class="day-empty"
            />
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.calendar-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: auto;
}

.calendar-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 12px 0;
}

.month-label {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  min-width: 140px;
  text-align: center;
}

.calendar-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: var(--color-text-secondary);
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 1px;
  background: var(--color-border);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}

.weekday-header {
  background: var(--color-card);
  padding: 8px 4px;
  text-align: center;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.day-cell {
  background: var(--color-card);
  min-height: 100px;
  padding: 6px;
  display: flex;
  flex-direction: column;
}

.day-cell.empty {
  background: var(--color-hover);
  opacity: 0.5;
}

.day-cell.today {
  background: var(--color-primary-light);
}

.day-number {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
  margin-bottom: 4px;
}

.day-number.today {
  color: var(--color-primary);
  font-weight: 700;
}

.day-entries {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 3px;
  overflow: hidden;
}

.day-entry-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 4px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: background 0.15s;
}

.day-entry-item:hover {
  background: var(--color-hover);
}

.entry-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text);
}

.day-empty {
  flex: 1;
}
</style>
