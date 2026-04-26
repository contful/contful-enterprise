<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { getConfigs, setConfig, type SiteConfig } from '@/api/config'
import { showSuccess, showError } from '@/utils/request'

const { t } = useI18n()
const siteStore = useSiteStore()

// ============ 状态 ============
const loading = ref(false)
const saving = ref(false)

// 当前选中驱动
const driver = ref('local')

// 本地存储表单
const localForm = ref({ root: 'uploads', base_url: '/uploads' })

// 云存储统一表单（s3/oss/cos/obs 共用）
const cloudForm = ref({
  bucket: '',
  endpoint: '',
  region: '',
  base_url: '',
  // s3 独有
  path_prefix: '',
  force_path_style: 'false',
})

const driverOptions = [
  { value: 'local', label: t('storage.driverLocal'), icon: '💾' },
  { value: 's3',   label: 'S3 / MinIO',              icon: '☁️' },
  { value: 'oss',  label: t('storage.driverOss'),    icon: '☁️' },
  { value: 'cos',  label: t('storage.driverCos'),    icon: '☁️' },
  { value: 'obs',  label: t('storage.driverObs'),    icon: '☁️' },
]

// 各驱动说明文案
const driverDesc = computed(() => ({
  local: t('storage.descLocal'),
  s3:    t('storage.descS3'),
  oss:   t('storage.descOss'),
  cos:   t('storage.descCos'),
  obs:   t('storage.descObs'),
}[driver.value] ?? ''))

// 云存储区段标题
const cloudConfigTitle = computed(() => ({
  s3:  t('storage.s3Config'),
  oss: t('storage.ossConfig'),
  cos: t('storage.cosConfig'),
  obs: t('storage.obsConfig'),
}[driver.value] ?? ''))

// cos 无 endpoint；s3 多路径前缀和路径样式
const showEndpoint   = computed(() => driver.value !== 'cos')
const showS3Extra    = computed(() => driver.value === 's3')
const isCloud        = computed(() => driver.value !== 'local')

// endpoint placeholder
const endpointPlaceholder = computed(() => ({
  s3:  'https://s3.amazonaws.com',
  oss: 'https://oss-cn-hangzhou.aliyuncs.com',
  obs: 'https://obs.cn-north-4.myhuaweicloud.com',
}[driver.value] ?? ''))

// bucket placeholder
const bucketPlaceholder = computed(() => ({
  s3:  'my-bucket',
  oss: 'my-oss-bucket',
  cos: 'my-bucket-1250000000',
  obs: 'my-obs-bucket',
}[driver.value] ?? ''))

// region placeholder
const regionPlaceholder = computed(() => ({
  s3:  'us-east-1',
  oss: 'oss-cn-hangzhou',
  cos: 'ap-guangzhou',
  obs: 'cn-north-4',
}[driver.value] ?? ''))

// cos bucket 额外提示
const showCosBucketHint = computed(() => driver.value === 'cos')

// ============ 从配置表读取 ============
const loadStorage = async () => {
  if (!siteStore.currentSiteId) return
  loading.value = true
  try {
    const res = await getConfigs(siteStore.currentSiteId)
    const configs: SiteConfig[] = res.items || []
    const get = (key: string) => configs.find(c => c.config_key === key)?.config_value ?? ''

    driver.value = get('storage.driver') || 'local'

    // local
    localForm.value.root     = get('storage.local.root')     || 'uploads'
    localForm.value.base_url = get('storage.local.base_url') || '/uploads'

    // 云存储：读取当前 driver 对应的配置
    const d = driver.value
    if (d !== 'local') {
      cloudForm.value.bucket           = get(`storage.${d}.bucket`)
      cloudForm.value.endpoint         = get(`storage.${d}.endpoint`)
      cloudForm.value.region           = get(`storage.${d}.region`)
      cloudForm.value.base_url         = get(`storage.${d}.base_url`)
      cloudForm.value.path_prefix      = get('storage.s3.path_prefix')
      cloudForm.value.force_path_style = get('storage.s3.force_path_style') || 'false'
    }
  } catch (e: any) {
    showError(e)
  } finally {
    loading.value = false
  }
}

// 切换驱动时重新读取该驱动对应的已存配置（如果有）
const handleDriverChange = async (val: string) => {
  driver.value = val
  if (val === 'local' || !siteStore.currentSiteId) return
  try {
    const res = await getConfigs(siteStore.currentSiteId)
    const configs: SiteConfig[] = res.items || []
    const get = (key: string) => configs.find(c => c.config_key === key)?.config_value ?? ''
    cloudForm.value.bucket           = get(`storage.${val}.bucket`)
    cloudForm.value.endpoint         = get(`storage.${val}.endpoint`)
    cloudForm.value.region           = get(`storage.${val}.region`)
    cloudForm.value.base_url         = get(`storage.${val}.base_url`)
    cloudForm.value.path_prefix      = get('storage.s3.path_prefix')
    cloudForm.value.force_path_style = get('storage.s3.force_path_style') || 'false'
  } catch { /* ignore */ }
}

// ============ 保存配置 ============
const handleSave = async () => {
  if (!siteStore.currentSiteId || saving.value) return
  saving.value = true
  try {
    const siteId = siteStore.currentSiteId
    const updates: [string, string][] = [['storage.driver', driver.value]]

    if (driver.value === 'local') {
      updates.push(
        ['storage.local.root',     localForm.value.root],
        ['storage.local.base_url', localForm.value.base_url],
      )
    } else {
      const d = driver.value
      updates.push(
        [`storage.${d}.bucket`,   cloudForm.value.bucket],
        [`storage.${d}.region`,   cloudForm.value.region],
        [`storage.${d}.base_url`, cloudForm.value.base_url],
      )
      if (showEndpoint.value) {
        updates.push([`storage.${d}.endpoint`, cloudForm.value.endpoint])
      }
      if (showS3Extra.value) {
        updates.push(
          ['storage.s3.path_prefix',      cloudForm.value.path_prefix],
          ['storage.s3.force_path_style', cloudForm.value.force_path_style],
        )
      }
    }

    for (const [key, value] of updates) {
      await setConfig(siteId, key, { config_value: value })
    }
    showSuccess(t('storage.saved'))
  } catch (e: any) {
    showError(e)
  } finally {
    saving.value = false
  }
}

onMounted(loadStorage)
watch(() => siteStore.currentSiteId, () => { loadStorage() })
</script>

<template>
  <div class="storage-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('storage.title') }}</h1>
        <p class="page-subtitle">{{ t('storage.subtitle') }}</p>
      </div>
    </div>

    <!-- 无站点 -->
    <div v-if="!siteStore.currentSiteId" class="empty-state">
      <p>{{ t('settings.noSite') }}</p>
    </div>

    <!-- 加载中 -->
    <div v-else-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('common.loading') }}</p>
    </div>

    <!-- 配置表单 -->
    <div v-else class="settings-form">

      <!-- 驱动选择 -->
      <section class="settings-section">
        <h3 class="section-title">{{ t('storage.driver') }}</h3>
        <p class="section-desc">{{ t('storage.driverDesc') }}</p>
        <div class="driver-cards">
          <button
            v-for="opt in driverOptions"
            :key="opt.value"
            class="driver-card"
            :class="{ active: driver === opt.value }"
            @click="handleDriverChange(opt.value)"
          >
            <span class="driver-icon">{{ opt.icon }}</span>
            <span class="driver-label">{{ opt.label }}</span>
            <span v-if="driver === opt.value" class="driver-check">✓</span>
          </button>
        </div>
        <p v-if="driverDesc" class="driver-hint">{{ driverDesc }}</p>
      </section>

      <!-- 本地存储配置 -->
      <section v-if="driver === 'local'" class="settings-section">
        <h3 class="section-title">{{ t('storage.localConfig') }}</h3>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">{{ t('storage.localRoot') }}</label>
            <input v-model="localForm.root" type="text" class="input" placeholder="uploads" />
            <p class="form-hint">{{ t('storage.localRootHint') }}</p>
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('storage.baseUrl') }}</label>
            <input v-model="localForm.base_url" type="text" class="input" placeholder="/uploads" />
            <p class="form-hint">{{ t('storage.baseUrlHint') }}</p>
          </div>
        </div>
      </section>

      <!-- 云存储统一配置（s3 / oss / cos / obs） -->
      <section v-if="isCloud" class="settings-section">
        <h3 class="section-title">{{ cloudConfigTitle }}</h3>

        <!-- 凭证提示 -->
        <div class="credential-notice">
          <span class="notice-icon">🔑</span>
          <span>{{ t('storage.credentialNotice') }}</span>
        </div>

        <div class="form-grid">
          <!-- Bucket -->
          <div class="form-item">
            <label class="form-label">{{ t('storage.bucket') }} <span class="required">*</span></label>
            <input v-model="cloudForm.bucket" type="text" class="input" :placeholder="bucketPlaceholder" />
            <p v-if="showCosBucketHint" class="form-hint">{{ t('storage.cosBucketHint') }}</p>
          </div>

          <!-- Region -->
          <div class="form-item">
            <label class="form-label">{{ t('storage.region') }}</label>
            <input v-model="cloudForm.region" type="text" class="input" :placeholder="regionPlaceholder" />
          </div>

          <!-- Endpoint（cos 不显示） -->
          <div v-if="showEndpoint" class="form-item form-item-full">
            <label class="form-label">{{ t('storage.endpoint') }}</label>
            <input v-model="cloudForm.endpoint" type="text" class="input" :placeholder="endpointPlaceholder" />
            <p class="form-hint">{{ t('storage.endpointHint') }}</p>
          </div>

          <!-- 公开访问 URL -->
          <div class="form-item form-item-full">
            <label class="form-label">{{ t('storage.baseUrl') }}</label>
            <input v-model="cloudForm.base_url" type="text" class="input" placeholder="https://cdn.example.com" />
            <p class="form-hint">{{ t('storage.baseUrlHint') }}</p>
          </div>

          <!-- S3 独有：路径前缀 + 路径样式 -->
          <template v-if="showS3Extra">
            <div class="form-item">
              <label class="form-label">{{ t('storage.pathPrefix') }}</label>
              <input v-model="cloudForm.path_prefix" type="text" class="input" placeholder="media/" />
              <p class="form-hint">{{ t('storage.pathPrefixHint') }}</p>
            </div>
            <div class="form-item">
              <label class="form-label">{{ t('storage.forcePathStyle') }}</label>
              <select v-model="cloudForm.force_path_style" class="input select">
                <option value="false">{{ t('storage.pathStyleVirtual') }}</option>
                <option value="true">{{ t('storage.pathStyleForce') }}</option>
              </select>
              <p class="form-hint">{{ t('storage.forcePathStyleHint') }}</p>
            </div>
          </template>
        </div>
      </section>

      <!-- 提交按钮 -->
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="saving" @click="handleSave">
          {{ saving ? t('settings.savingBtn') : t('settings.saveBtn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.storage-page {
  min-height: 200px;
}

.empty-state,
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  color: var(--color-text-secondary);
}

.spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 12px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.settings-section {
  padding: 24px 0;
  border-bottom: 1px solid var(--color-border);
}

.settings-section:last-of-type {
  border-bottom: none;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  margin: 0 0 8px;
}

.section-desc {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin: 0 0 16px;
}

/* 驱动卡片 */
.driver-cards {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.driver-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: 1.5px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-bg);
  cursor: pointer;
  font-size: 14px;
  color: var(--color-text);
  transition: all 0.15s;
  min-width: 120px;
}

.driver-card:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
  background: var(--color-hover, rgba(0, 82, 217, 0.04));
}

.driver-card.active {
  border-color: var(--color-primary);
  background: var(--color-hover, rgba(0, 82, 217, 0.06));
  color: var(--color-primary);
  font-weight: 500;
}

.driver-icon { font-size: 16px; }
.driver-label { flex: 1; }
.driver-check {
  font-size: 13px;
  color: var(--color-primary);
  font-weight: 700;
}

.driver-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin: 0;
  padding: 6px 10px;
  background: var(--color-bg-secondary, #f8f9fa);
  border-radius: 6px;
  border-left: 3px solid var(--color-border);
}

/* 凭证提示框 */
.credential-notice {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 16px;
  background: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: 8px;
  font-size: 13px;
  color: #92400e;
  margin-bottom: 16px;
}

.notice-icon {
  font-size: 16px;
  flex-shrink: 0;
  margin-top: 1px;
}

/* 表单 */
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-item-full {
  grid-column: 1 / -1;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
}

.required {
  color: var(--color-error, #e53935);
}

.form-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin: 0;
}

.input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  color: var(--color-text);
  background: var(--color-bg);
  transition: border-color 0.15s;
  box-sizing: border-box;
}

.input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23666' d='M2 4l4 4 4-4'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
  cursor: pointer;
}

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

.form-actions {
  padding: 24px 0 0;
  display: flex;
  justify-content: flex-end;
}
</style>
