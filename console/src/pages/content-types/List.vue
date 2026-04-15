<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  Space,
  Button,
  Card,
  Table,
  TableColumn,
  Dialog,
  Form,
  FormItem,
  Input,
  Select,
  Option,
  Switch,
  MessagePlugin,
  Pagination,
  Icon,
  Tooltip,
  Popconfirm,
  Badge,
} from 'tdesign-vue-next'
import {
  AddIcon,
  DeleteIcon,
  EditIcon,
  RefreshIcon,
  UndoIcon,
  CheckCircleIcon,
  ErrorCircleIcon,
  SettingIcon,
} from 'tdesign-icons-vue-next'
import {
  getContentTypes,
  createContentType,
  updateContentType,
  deleteContentType,
  type ContentType,
  type ContentTypeCreate,
  type ContentTypeUpdate,
} from '@/api/content-type'

const router = useRouter()

// 状态
const loading = ref(false)
const dataList = ref<ContentType[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 创建/编辑对话框
const dialogVisible = ref(false)
const dialogTitle = ref('创建内容类型')
const isEditing = ref(false)
const editingId = ref('')

const formData = ref<ContentTypeCreate>({
  name: '',
  slug: '',
  description: '',
  kind: 'collection',
  versioning_enabled: false,
})

// 表单规则
const formRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  slug: [
    { required: true, message: '请输入标识符', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*$/, message: '只能包含小写字母、数字和连字符，必须以字母开头', trigger: 'blur' },
  ],
}

// 自动生成 slug
const generateSlug = () => {
  if (formData.value.name && !isEditing.value) {
    formData.value.slug = formData.value.name
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')
  }
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const res = await getContentTypes({ page: currentPage.value, page_size: pageSize.value })
    if (res.data.code === 0) {
      dataList.value = res.data.data.items
      total.value = res.data.data.total
    }
  } catch (e) {
    MessagePlugin.error('加载失败')
  } finally {
    loading.value = false
  }
}

// 分页变化
const onPageChange = (page: number) => {
  currentPage.value = page
  loadData()
}

const onPageSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadData()
}

// 打开创建对话框
const openCreateDialog = () => {
  isEditing.value = false
  dialogTitle.value = '创建内容类型'
  formData.value = {
    name: '',
    slug: '',
    description: '',
    kind: 'collection',
    versioning_enabled: false,
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const openEditDialog = (row: ContentType) => {
  isEditing.value = true
  dialogTitle.value = '编辑内容类型'
  editingId.value = row.id
  formData.value = {
    name: row.name,
    slug: row.slug,
    description: row.description,
    kind: row.kind,
    versioning_enabled: row.versioning_enabled,
  }
  dialogVisible.value = true
}

// 提交表单
const submitForm = async () => {
  try {
    if (isEditing.value) {
      await updateContentType(editingId.value, formData.value as ContentTypeUpdate)
      MessagePlugin.success('更新成功')
    } else {
      await createContentType(formData.value)
      MessagePlugin.success('创建成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (e: any) {
    MessagePlugin.error(e?.response?.data?.msg || '操作失败')
  }
}

// 删除内容类型
const handleDelete = async (row: ContentType) => {
  try {
    await deleteContentType(row.id)
    MessagePlugin.success('删除成功')
    loadData()
  } catch (e: any) {
    MessagePlugin.error(e?.response?.data?.msg || '删除失败')
  }
}

// 切换启用状态
const toggleActive = async (row: ContentType) => {
  try {
    await updateContentType(row.id, { is_active: !row.is_active })
    MessagePlugin.success(row.is_active ? '已禁用' : '已启用')
    loadData()
  } catch (e: any) {
    MessagePlugin.error(e?.response?.data?.msg || '操作失败')
    // 回滚 UI
    row.is_active = !row.is_active
  }
}

// 跳转到字段管理
const goToFields = (row: ContentType) => {
  router.push(`/content-types/${row.id}/fields`)
}

// 格式化时间
const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN')
}

// 格式化 kind
const formatKind = (kind: string) => {
  return kind === 'collection' ? '集合' : '单条'
}

// 获取 kind 标签类型
const getKindType = (kind: string) => {
  return kind === 'collection' ? 'primary' : 'warning'
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="content-types-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="title-section">
        <h1>内容类型</h1>
        <p class="subtitle">定义和管理内容的数据结构</p>
      </div>
      <Space>
        <Button theme="default" @click="loadData">
          <template #icon><RefreshIcon /></template>
          刷新
        </Button>
        <Button theme="primary" @click="openCreateDialog">
          <template #icon><AddIcon /></template>
          创建内容类型
        </Button>
      </Space>
    </div>

    <!-- 数据表格 -->
    <Card>
      <Table
        :data="dataList"
        :loading="loading"
        :pagination="{
          count: total,
          page: currentPage,
          pageSize: pageSize,
          pageSizeOptions: [20, 50, 100],
        }"
        row-key="id"
        hover
      >
        <TableColumn label="名称" min-width="200">
          <template #default="{ row }">
            <div class="name-cell">
              <span class="name">{{ row.name }}</span>
              <span class="slug">{{ row.slug }}</span>
            </div>
          </template>
        </TableColumn>

        <TableColumn label="类型" width="100">
          <template #default="{ row }">
            <Badge
              :theme="row.kind === 'collection' ? 'primary' : 'warning'"
              :content="formatKind(row.kind)"
            />
          </template>
        </TableColumn>

        <TableColumn label="状态" width="100">
          <template #default="{ row }">
            <Switch
              :value="row.is_active"
              size="small"
              @change="() => toggleActive(row)"
            />
          </template>
        </TableColumn>

        <TableColumn label="版本控制" width="100">
          <template #default="{ row }">
            <Tooltip :content="row.versioning_enabled ? '已启用' : '未启用'">
              <CheckCircleIcon v-if="row.versioning_enabled" style="color: var(--td-success-color)" />
              <ErrorCircleIcon v-else style="color: var(--td-text-color-disabled)" />
            </Tooltip>
          </template>
        </TableColumn>

        <TableColumn label="描述" min-width="200">
          <template #default="{ row }">
            <span class="description">{{ row.description || '-' }}</span>
          </template>
        </TableColumn>

        <TableColumn label="更新时间" width="180">
          <template #default="{ row }">
            <span class="time">{{ formatDate(row.updated_at) }}</span>
          </template>
        </TableColumn>

        <TableColumn label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <Space>
              <Tooltip content="管理字段">
                <Button size="small" variant="text" @click="goToFields(row)">
                  <template #icon><SettingIcon /></template>
                </Button>
              </Tooltip>
              <Tooltip content="编辑">
                <Button size="small" variant="text" @click="openEditDialog(row)">
                  <template #icon><EditIcon /></template>
                </Button>
              </Tooltip>
              <Popconfirm @confirm="handleDelete(row)">
                <template #content>
                  <p>确定删除「{{ row.name }}」吗？</p>
                  <p style="color: var(--td-warning-color); font-size: 12px">此操作不可恢复</p>
                </template>
                <Tooltip content="删除">
                  <Button size="small" variant="text" theme="danger">
                    <template #icon><DeleteIcon /></template>
                  </Button>
                </Tooltip>
              </Popconfirm>
            </Space>
          </template>
        </TableColumn>
      </Table>

      <!-- 分页 -->
      <div class="pagination">
        <Pagination
          v-model:page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-size-options="[20, 50, 100]"
          @change="onPageChange"
          @page-size-change="onPageSizeChange"
        />
      </div>
    </Card>

    <!-- 创建/编辑对话框 -->
    <Dialog
      v-model:visible="dialogVisible"
      :header="dialogTitle"
      width="600px"
      :close-on-overlay-click="false"
    >
      <Form :data="formData" :rules="formRules" label-width="100px" colon>
        <FormItem label="名称" name="name">
          <Input
            v-model="formData.name"
            placeholder="请输入内容类型名称"
            @blur="generateSlug"
          />
        </FormItem>

        <FormItem label="标识符" name="slug">
          <Input
            v-model="formData.slug"
            placeholder="如：article、product"
            :disabled="isEditing"
          />
        </FormItem>

        <FormItem label="类型" name="kind">
          <Select v-model="formData.kind" :disabled="isEditing">
            <Option value="collection" label="集合类型" />
            <Option value="single" label="单条类型" />
          </Select>
        </FormItem>

        <FormItem label="描述">
          <Input
            v-model="formData.description"
            type="textarea"
            placeholder="可选，简要描述这个内容类型的用途"
            :rows="3"
          />
        </FormItem>

        <FormItem label="版本控制">
          <Switch v-model="formData.versioning_enabled" />
          <span style="margin-left: 8px; color: var(--td-text-color-secondary)">
            启用后可保留内容的历史版本
          </span>
        </FormItem>
      </Form>

      <template #footer>
        <Space>
          <Button @click="dialogVisible = false">取消</Button>
          <Button theme="primary" @click="submitForm">
            {{ isEditing ? '保存' : '创建' }}
          </Button>
        </Space>
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.content-types-page {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.title-section h1 {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 4px 0;
}

.subtitle {
  color: var(--td-text-color-secondary);
  margin: 0;
}

.name-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.name-cell .name {
  font-weight: 500;
}

.name-cell .slug {
  font-size: 12px;
  color: var(--td-text-color-secondary);
  font-family: monospace;
}

.description {
  color: var(--td-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time {
  color: var(--td-text-color-secondary);
  font-size: 13px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
