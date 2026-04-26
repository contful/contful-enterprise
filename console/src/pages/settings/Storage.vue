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

// 各驱动配置表单
const localForm = ref({ root: 'uploads', base_url: '/uploads' })
const s3Form = ref({
  bucket: '',
  endpoint: '',
  region: '',
  path_prefix: '',
  force_path_style: 'false',
  base_url: '',
})
const ossForm = ref({ bucket: '', endpoint: '', region: '', base_url: '' })
const cosForm = ref({ bucket: '', region: '', base_url: '' })
const obsForm = ref({ bucket: '', endpoint: '', region: '', base_url: '' })

const driverOptions = [
  { value: 'local', label: t('storage.driverLocal'), icon: '💾' },
  { value: 's3', label: 'S3 / MinIO', icon: '☁️' },
  { value: 'oss', label: t('storage.driverOss'), icon: '☁️' },
  { value: 'cos', label: t('storage.driverCos'), icon: '☁️' },
  { value: 'obs', label: t('storage.driverObs'), icon: '☁️' },
]

// 当前驱动说明
const driverDesc = computed(() => {
  const map: Record<string, string> = {
    local: t('storage.descLocal'),
    s3: t('storage.descS3'),
    oss: t('storage.descOss'),
    cos: t('storage.descCos'),
    obs: t('storage.descObs'),
  }
  return map[driver.value] || ''
})

// ============ 从配置表读取 ============
const loadStorage = async () => {
  if (!siteStore.currentSiteId) return
  loading.value = true
  try {
    const res = await getConfigs(siteStore.currentSiteId)
    const configs: SiteConfig[] = res.items || []

    const get = (key: string) => configs.find(c => c.config_key === key)?.config_value ?? ''

    // 驱动
    driver.value = get('storage.driver') || 'local'

    // local
    localForm.value.root = get('storage.local.root') || 'uploads'
    localForm.value.base_url = get('storage.local.base_url') || '/uploads'

    // s3
    s3Form.value.bucket = get('storage.s3.bucket')
    s3Form.value.endpoint = get('storage.s3.endpoint')
    s3Form.value.region = get('storage.s3.region')
    s3Form.value.path_prefix = get('storage.s3.path_prefix')
    s3Form.value.force_path_style = get('storage.s3.force_path_style') || 'false'
    s3Form.value.base_url = get('storage.s3.base_url')

    // oss
    ossForm.value.bucket = get('storage.oss.bucket')
    ossForm.value.endpoint = get('storage.oss.endpoint')
    ossForm.value.region = get('storage.oss.region')
    ossForm.value.base_url = get('storage.oss.base_url')

    // cos
    cosForm.value.bucket = get('storage.cos.bucket')
    cosForm.value.region = get('storage.cos.region')
    cosForm.value.base_url = get('storage.cos.base_url')

    // obs
    obsForm.value.bucket = get('storage.obs.bucket')
    obsForm.value.endpoint = get('storage.obs.endpoint')
    obsForm.value.region = get('storage.obs.region')
    obsForm.value.base_url = get('storage.obs.base_url')
  } catch (e: any) {
    showError(e)
  } finally {
    loading.value = false
  }
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
        ['storage.local.root', localForm.value.root],
        ['storage.local.base_url', localForm.value.base_url],
      )
    } else if (driver.value === 's3') {
      updates.push(
        ['storage.s3.bucket', s3Form.value.bucket],
        ['storage.s3.endpoint', s3Form.value.endpoint],
        ['storage.s3.region', s3Form.value.region],
        ['storage.s3.path_prefix', s3Form.value.path_prefix],
        ['storage.s3.force_path_style', s3Form.value.force_path_style],
        ['storage.s3.base_url', s3Form.value.base_url],
      )
    } else if (driver.value === 'oss') {
      updates.push(
        ['storage.oss.bucket', ossForm.value.bucket],
        ['storage.oss.endpoint', ossForm.value.endpoint],
        ['storage.oss.region', ossForm.value.region],
        ['storage.oss.base_url', ossForm.value.base_url],
      )
    } else if (driver.value === 'cos') {
      updates.push(
        ['storage.cos.bucket', cosForm.value.bucket],
        ['storage.cos.region', cosForm.value.region],
        ['storage.cos.base_url', cosForm.value.base_url],
      )
    } else if (driver.value === 'obs') {
      updates.push(
        ['storage.obs.bucket', obsForm.value.bucket],
        ['storage.obs.endpoint', obsForm.value.endpoint],
        ['storage.obs.region', obsForm.value.region],
        ['storage.obs.base_url', obsForm.value.base_url],
      )
    }

    // 批量提交
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

// 切换站点后重新加载
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
            @click="driver = opt.value"
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

      <!-- S3 配置 -->
      <section v-if="driver === 's3'" class="settings-section">
        <h3 class="section-title">{{ t('storage.s3Config') }}</h3>
        <!-- 凭证提示 -->
        <div class="credential-notice">
          <span class="notice-icon">🔑</span>
          <span>{{ t('storage.credentialNotice') }}</span>
        </div>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">{{ t('storage.bucket') }} <span class="required">*</span></label>
            <input v-model="s3Form.bucket" type="text" class="input" placeholder="my-bucket" />
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('storage.region') }}</label>
            <input v-model="s3Form.region" type="text" class="input" placeholder="us-east-1" />
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">{{ t('storage.endpoint') }}</label>
            <input v-model="s3Form.endpoint" type="text" class="input" placeholder="https://s3.amazonaws.com" />
            <p class="form-hint">{{ t('storage.endpointHint') }}</p>
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('storage.pathPrefix') }}</label>
            <input v-model="s3Form.path_prefix" type="text" class="input" placeholder="media/" />
            <p class="form-hint">{{ t('storage.pathPrefixHint') }}</p>
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('storage.baseUrl') }}</label>
            <input v-model="s3Form.base_url" type="text" class="input" placeholder="https://cdn.example.com" />
            <p class="form-hint">{{ t('storage.baseUrlHint') }}</p>
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('storage.forcePathStyle') }}</label>
            <select v-model="s3Form.force_path_style" class="input select">
              <option value="false">{{ t('storage.pathStyleVirtual') }}</option>
              <option value="true">{{ t('storage.pathStyleForce') }}</option>
            </select>
            <p class="form-hint">{{ t('storage.forcePathStyleHint') }}</p>
          </div>
        </div>
      </section>

      <!-- 阿里云 OSS 配置 -->
      <section v-if="driver === 'oss'" class="settings-section">
        <h3 class="section-title">{{ t('storage.ossConfig') }}</h3>
        <div class="credential-notice">
          <span class="notice-icon">🔑</span>
          <span>{{ t('storage.credentialNotice') }}</span>
        </div>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">{{ t('storage.bucket') }} <span class="required">*</span></label>
            <input v-model="ossForm.bucket" type="text" class="input" placeholder="my-oss-bucket" />
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('storage.region') }}</label>
            <input v-model="ossForm.region" type="text" class="input" placeholder="oss-cn-hangzhou" />
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">{{ t('storage.endpoint') }}</label>
            <input v-model="ossForm.endpoint" type="text" class="input" placeholder="https://oss-cn-hangzhou.aliyuncs.com" />
            <p class="form-hint">{{ t('storage.endpointHint') }}</p>
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">{{ t('storage.baseUrl') }}</label>
            <input v-model="ossForm.base_url" type="text" class="input" placeholder="https://your-bucket.oss-cn-hangzhou.aliyuncs.com" />
            <p class="form-hint">{{ t('storage.baseUrlHint') }}</p>
          </div>
        </div>
      </section>

      <!-- 腾讯云 COS 配置 -->
      <section v-if="driver === 'cos'" class="settings-section">
        <h3 class="section-title">{{ t('storage.cosConfig') }}</h3>
        <div class="credential-notice">
          <span class="notice-icon">🔑</span>
          <span>{{ t('storage.credentialNotice') }}</span>
        </div>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">{{ t('storage.bucket') }} <span class="required">*</span></label>
            <input v-model="cosForm.bucket" type="text" class="input" placeholder="my-bucket-1250000000" />
            <p class="form-hint">{{ t('storage.cosBucketHint') }}</p>
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('storage.region') }} <span class="required">*</span></label>
            <input v-model="cosForm.region" type="text" class="input" placeholder="ap-guangzhou" />
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">{{ t('storage.baseUrl') }}</label>
            <input v-model="cosForm.base_url" type="text" class="input" placeholder="https://my-bucket-1250000000.cos.ap-guangzhou.myqcloud.com" />
            <p class="form-hint">{{ t('storage.baseUrlHint') }}</p>
          </div>
        </div>
      </section>

      <!-- 华为云 OBS 配置 -->
      <section v-if="driver === 'obs'" class="settings-section">
        <h3 class="section-title">{{ t('storage.obsConfig') }}</h3>
        <div class="credential-notice">
          <span class="notice-icon">🔑</span>
          <span>{{ t('storage.credentialNotice') }}</span>
        </div>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">{{ t('storage.bucket') }} <span class="required">*</span></label>
            <input v-model="obsForm.bucket" type="text" class="input" placeholder="my-obs-bucket" />
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('storage.region') }}</label>
            <input v-model="obsForm.region" type="text" class="input" placeholder="cn-north-4" />
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">{{ t('storage.endpoint') }}</label>
            <input v-model="obsForm.endpoint" type="text" class="input" placeholder="https://obs.cn-north-4.myhuaweicloud.com" />
            <p class="form-hint">{{ t('storage.endpointHint') }}</p>
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">{{ t('storage.baseUrl') }}</label>
            <input v-model="obsForm.base_url" type="text" class="input" placeholder="https://my-obs-bucket.obs.cn-north-4.myhuaweicloud.com" />
            <p class="form-hint">{{ t('storage.baseUrlHint') }}</p>
          </div>
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
  position: relative;
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

.driver-icon {
  font-size: 16px;
}

.driver-label {
  flex: 1;
}

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
