<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSiteStore } from '@/stores/site'
import { getSite, updateSite, type Site, type SiteConfig, type SiteSEO } from '@/api/site'
import { showError } from '@/utils/request'

// ============ 状态 ============
const siteStore = useSiteStore()
const loading = ref(false)
const saving = ref(false)

// 站点数据
const site = ref<Site | null>(null)

// 表单数据
const form = ref({
  name: '',
  slug: '',
  description: '',
  logo_url: '',
  favicon_url: '',
  config: {
    timezone: 'Asia/Shanghai',
    locale: 'zh-CN',
  } as SiteConfig,
  seo: {
    meta_title: '',
    meta_description: '',
    keywords: '',
  } as SiteSEO,
  custom_domains: [] as string[],
})

// 原始 slug（不可修改）
const originalSlug = ref('')

// 自定义域名输入
const newDomain = ref('')

// 时区选项
const timezoneOptions = [
  { value: 'Asia/Shanghai', label: 'Asia/Shanghai (UTC+8)' },
  { value: 'Asia/Tokyo', label: 'Asia/Tokyo (UTC+9)' },
  { value: 'Asia/Singapore', label: 'Asia/Singapore (UTC+8)' },
  { value: 'Asia/Hong_Kong', label: 'Asia/Hong_Kong (UTC+8)' },
  { value: 'America/New_York', label: 'America/New_York (UTC-5)' },
  { value: 'America/Los_Angeles', label: 'America/Los_Angeles (UTC-8)' },
  { value: 'Europe/London', label: 'Europe/London (UTC+0)' },
  { value: 'Europe/Paris', label: 'Europe/Paris (UTC+1)' },
  { value: 'UTC', label: 'UTC' },
]

// 语言选项
const localeOptions = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'zh-TW', label: '繁體中文' },
  { value: 'en-US', label: 'English' },
  { value: 'ja-JP', label: '日本語' },
  { value: 'ko-KR', label: '한국어' },
]

// ============ 加载数据 ============
const loadSite = async () => {
  if (!siteStore.currentSiteId) return

  loading.value = true
  try {
    const res = await getSite(siteStore.currentSiteId)
    if (res.data.code === 200) {
      site.value = res.data.data

      // 填充表单
      const data = res.data.data
      form.value = {
        name: data.name || '',
        slug: data.slug || '',
        description: data.description || '',
        logo_url: data.logo_url || '',
        favicon_url: data.favicon_url || '',
        config: (data.config as SiteConfig) || { timezone: 'Asia/Shanghai', locale: 'zh-CN' },
        seo: (data.seo as SiteSEO) || { meta_title: '', meta_description: '', keywords: '' },
        custom_domains: data.custom_domains || [],
      }
      originalSlug.value = data.slug || ''
    }
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
}

// ============ 保存 ============
const handleSave = async () => {
  if (!siteStore.currentSiteId) return

  saving.value = true
  try {
    const res = await updateSite(siteStore.currentSiteId, {
      name: form.value.name,
      slug: form.value.slug,
      description: form.value.description,
      logo_url: form.value.logo_url || undefined,
      favicon_url: form.value.favicon_url || undefined,
      config: form.value.config,
      seo: form.value.seo,
      custom_domains: form.value.custom_domains,
    })

    if (res.data.code === 200) {
      site.value = res.data.data
      originalSlug.value = res.data.data.slug
      window.__showSuccess?.('站点设置已保存') || (window as any).__messagePlugin?.success?.('站点设置已保存')
    }
  } catch (error) {
    showError(error)
  } finally {
    saving.value = false
  }
}

// ============ 自定义域名 ============
const addDomain = () => {
  const domain = newDomain.value.trim().toLowerCase()
  if (!domain) return
  if (form.value.custom_domains.includes(domain)) {
    newDomain.value = ''
    return
  }
  form.value.custom_domains.push(domain)
  newDomain.value = ''
}

const removeDomain = (index: number) => {
  form.value.custom_domains.splice(index, 1)
}

// ============ SEO 预览 ============
const seoPreview = () => {
  const title = form.value.seo.meta_title || form.value.name || '站点名称'
  const desc = form.value.seo.meta_description || form.value.description || ''
  return { title, desc }
}

// ============ 挂载 ============
onMounted(() => {
  loadSite()
})
</script>

<template>
  <div class="site-settings">
    <div v-if="loading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="!siteStore.currentSiteId" class="empty-state">
      <p class="text-secondary">请先选择一个站点</p>
    </div>

    <div v-else class="settings-form">
      <!-- 基础信息 -->
      <section class="settings-section">
        <h3 class="section-title">基础信息</h3>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">站点名称 <span class="required">*</span></label>
            <input
              v-model="form.name"
              type="text"
              class="input"
              placeholder="输入站点名称"
              maxlength="200"
            />
          </div>
          <div class="form-item">
            <label class="form-label">
              站点标识 (Slug) <span class="required">*</span>
            </label>
            <input
              v-model="form.slug"
              type="text"
              class="input"
              placeholder="site-identifier"
              maxlength="100"
            />
            <p class="form-hint">用于 API 路径，只允许小写字母、数字和连字符</p>
            <p v-if="form.slug !== originalSlug" class="form-warning">
              修改 Slug 会影响 API 路由，请谨慎操作
            </p>
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">站点描述</label>
            <textarea
              v-model="form.description"
              class="input textarea"
              placeholder="简短描述站点的用途"
              rows="3"
              maxlength="2000"
            ></textarea>
          </div>
        </div>
      </section>

      <!-- 品牌设置 -->
      <section class="settings-section">
        <h3 class="section-title">品牌设置</h3>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">Logo URL</label>
            <input
              v-model="form.logo_url"
              type="url"
              class="input"
              placeholder="https://example.com/logo.png"
            />
            <p class="form-hint">站点 Logo 图片链接，建议尺寸 200x60</p>
          </div>
          <div class="form-item">
            <label class="form-label">Favicon URL</label>
            <input
              v-model="form.favicon_url"
              type="url"
              class="input"
              placeholder="https://example.com/favicon.ico"
            />
            <p class="form-hint">浏览器标签页图标，建议 32x32 或 64x64</p>
          </div>
        </div>
        <!-- Logo 预览 -->
        <div v-if="form.logo_url" class="brand-preview">
          <img :src="form.logo_url" alt="Logo 预览" class="logo-preview" @error="e => (e.target as HTMLImageElement).style.display='none'" />
        </div>
      </section>

      <!-- 区域设置 -->
      <section class="settings-section">
        <h3 class="section-title">区域设置</h3>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">时区</label>
            <select v-model="form.config.timezone" class="input select">
              <option v-for="opt in timezoneOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
          <div class="form-item">
            <label class="form-label">默认语言</label>
            <select v-model="form.config.locale" class="input select">
              <option v-for="opt in localeOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
        </div>
      </section>

      <!-- SEO 设置 -->
      <section class="settings-section">
        <h3 class="section-title">SEO 设置</h3>
        <p class="section-desc">配置站点在搜索引擎中的展示信息</p>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">Meta 标题</label>
            <input
              v-model="form.seo.meta_title"
              type="text"
              class="input"
              placeholder="站点名称 - 副标题"
              maxlength="60"
            />
            <p class="form-hint">{{ form.seo.meta_title.length }}/60 字符</p>
          </div>
          <div class="form-item">
            <label class="form-label">Meta 关键词</label>
            <input
              v-model="form.seo.keywords"
              type="text"
              class="input"
              placeholder="关键词1, 关键词2, 关键词3"
              maxlength="200"
            />
            <p class="form-hint">多个关键词用逗号分隔</p>
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">Meta 描述</label>
            <textarea
              v-model="form.seo.meta_description"
              class="input textarea"
              placeholder="简短描述站点内容，建议 120-160 字符"
              rows="3"
              maxlength="200"
            ></textarea>
            <p class="form-hint">{{ form.seo.meta_description.length }}/200 字符</p>
          </div>
        </div>

        <!-- SEO 预览 -->
        <div class="seo-preview">
          <p class="preview-label">搜索结果预览</p>
          <div class="google-preview">
            <p class="preview-title">{{ seoPreview().title }}</p>
            <p class="preview-url">{{ site?.slug || 'your-site' }}.contful.com</p>
            <p class="preview-desc">{{ seoPreview().desc || '站点描述将显示在这里...' }}</p>
          </div>
        </div>
      </section>

      <!-- 自定义域名 -->
      <section class="settings-section">
        <h3 class="section-title">自定义域名</h3>
        <p class="section-desc">为站点绑定独立域名，实现品牌统一</p>
        <div class="domain-list">
          <div v-for="(domain, index) in form.custom_domains" :key="domain" class="domain-item">
            <span class="domain-text">{{ domain }}</span>
            <button class="btn-remove" @click="removeDomain(index)">移除</button>
          </div>
          <div v-if="form.custom_domains.length === 0" class="domain-empty">
            暂无绑定域名
          </div>
        </div>
        <div class="domain-add">
          <input
            v-model="newDomain"
            type="text"
            class="input"
            placeholder="输入域名，如 example.com"
            @keydown.enter.prevent="addDomain"
          />
          <button class="btn btn-default" @click="addDomain">添加域名</button>
        </div>
      </section>

      <!-- 提交按钮 -->
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="saving || !form.name || !form.slug" @click="handleSave">
          {{ saving ? '保存中...' : '保存设置' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.site-settings {
  min-height: 200px;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  color: var(--color-text-secondary);
}

.loading-spinner {
  width: 32px;
  height: 32px;
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
  margin: 0 0 16px;
}

.section-desc {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin: -8px 0 16px;
}

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
  color: var(--color-error);
}

.form-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin: 0;
}

.form-warning {
  font-size: 12px;
  color: var(--color-warning, #e6a23c);
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
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.input:disabled {
  background: var(--color-bg-secondary, #f5f5f5);
  color: var(--color-text-secondary);
  cursor: not-allowed;
}

.textarea {
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
  line-height: 1.5;
}

.select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23666' d='M2 4l4 4 4-4'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
  cursor: pointer;
}

.brand-preview {
  margin-top: 12px;
  padding: 12px;
  background: var(--color-bg-secondary, #f5f5f5);
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo-preview {
  max-height: 48px;
  max-width: 200px;
  object-fit: contain;
}

/* SEO 预览 */
.seo-preview {
  margin-top: 20px;
  padding: 16px;
  background: var(--color-bg-secondary, #f5f5f5);
  border-radius: 8px;
}

.preview-label {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin: 0 0 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.google-preview {
  font-family: Arial, sans-serif;
}

.preview-title {
  font-size: 18px;
  color: #1a0dab;
  margin: 0 0 2px;
  text-decoration: underline;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-url {
  font-size: 12px;
  color: #006621;
  margin: 0 0 4px;
}

.preview-desc {
  font-size: 13px;
  color: #545454;
  margin: 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 自定义域名 */
.domain-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.domain-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--color-bg-secondary, #f5f5f5);
  border-radius: 6px;
  border: 1px solid var(--color-border);
}

.domain-text {
  font-size: 14px;
  color: var(--color-text);
  font-family: monospace;
}

.btn-remove {
  background: none;
  border: none;
  color: var(--color-error);
  font-size: 13px;
  cursor: pointer;
  padding: 2px 8px;
  border-radius: 4px;
}

.btn-remove:hover {
  background: rgba(230, 62, 62, 0.1);
}

.domain-empty {
  font-size: 13px;
  color: var(--color-text-secondary);
  padding: 8px 0;
}

.domain-add {
  display: flex;
  gap: 8px;
}

.domain-add .input {
  flex: 1;
}

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
  white-space: nowrap;
  flex-shrink: 0;
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

.btn-default {
  background: var(--color-bg);
  color: var(--color-text);
  border-color: var(--color-border);
}

.btn-default:hover:not(:disabled) {
  background: var(--color-hover);
}

/* 表单提交区 */
.form-actions {
  padding: 24px 0 0;
  display: flex;
  justify-content: flex-end;
}

.text-secondary {
  color: var(--color-text-secondary);
  font-size: 14px;
}
</style>
