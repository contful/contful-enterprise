<template>
  <t-card :title="t('audit.anomaly.trend')" :loading="loading" class="anomaly-trend">
    <div v-if="trend.length === 0" class="no-data">{{ t('audit.anomaly.noAnomalies') }}</div>
    <svg v-else viewBox="0 0 680 220" class="trend-svg">
      <line v-for="(_, i) in gridLines" :key="'g'+i" :x1="p.l" :y1="p.t + i*stepY" :x2="680-p.r" :y2="p.t + i*stepY" stroke="#e7e7e7" stroke-dasharray="4 4"/>
      <text v-for="(val,i) in yLabels" :key="'yl'+i" :x="p.l-8" :y="p.t + i*stepY + 4" text-anchor="end" font-size="11" fill="#999">{{ val }}</text>
      <text v-for="(item,i) in trend" :key="'xl'+i" :x="p.l + i*stepX" :y="p.t+h+20" text-anchor="middle" font-size="10" fill="#999">{{ fmt(item.date) }}</text>
      <polyline :points="pts" fill="none" stroke="#0052d9" stroke-width="2" stroke-linejoin="round"/>
      <circle v-for="(item,i) in trend" :key="'d'+i" :cx="p.l + i*stepX" :cy="p.t+h - (item.count/yMax)*h" r="4" fill="#0052d9"/>
    </svg>
  </t-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const props = defineProps<{ trend: Array<{ date: string; count: number }>; loading: boolean }>()
const p = { t: 10, r: 20, b: 30, l: 45 }; const h = 180
const yMax = computed(() => { if (!props.trend.length) return 1; const m = Math.max(...props.trend.map(d=>d.count)); return m===0?1:Math.ceil(m*1.1) })
const stepX = computed(() => props.trend.length<=1?0:(660-p.l-p.r)/(props.trend.length-1))
const stepY = h/5
const gridLines = [0,1,2,3,4,5]
const yLabels = computed(() => { const m = yMax.value; return gridLines.map(i=>Math.round((m/5)*(5-i))) })
const pts = computed(() => { if(!props.trend.length) return ''; return props.trend.map((item,i)=>`${p.l + i*stepX.value},${p.t+h-(item.count/yMax.value)*h}`).join(' ') })
function fmt(date: string) { const parts = date.split('-'); return parts.length>=3?parts[1]+'-'+parts[2]:date }
</script>

<style scoped>
.anomaly-trend { margin-bottom: 16px; }
.trend-svg { width: 100%; height: 220px; }
.no-data { font-size: 13px; color: var(--td-text-color-placeholder); text-align: center; padding: 40px 0; }
</style>
