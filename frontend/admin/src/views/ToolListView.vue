<script setup lang="ts">
import axios from 'axios'
import { Close, EditPen, Plus, Refresh } from '@element-plus/icons-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import {
  createTool,
  listAdminTools,
  updateTool,
  type AdminTool,
} from '../api/admin'
import CenteredModal from '../components/CenteredModal.vue'

type EditorMode = 'create' | 'edit'

const tools = ref<AdminTool[]>([])
const loading = ref(true)
const refreshing = ref(false)
const saving = ref(false)
const editorMode = ref<EditorMode | null>(null)
const errorMessage = ref('')
const successMessage = ref('')
const formError = ref('')

const form = reactive({
  code: '',
  name: '',
  description: '',
  priceYuan: '',
  currency: 'CNY',
  lifetime: true,
  active: true,
})

const editorTitle = computed(() => editorMode.value === 'create' ? '新增工具' : '修改工具')
const activeCount = computed(() => tools.value.filter((tool) => tool.active).length)

onMounted(() => load(false))

async function load(isRefresh: boolean) {
  if (isRefresh) {
    refreshing.value = true
  } else {
    loading.value = true
  }
  errorMessage.value = ''

  try {
    tools.value = await listAdminTools()
  } catch (error) {
    errorMessage.value = errorText(error, '工具列表加载失败，请稍后重试。')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function openCreate() {
  editorMode.value = 'create'
  form.code = ''
  form.name = ''
  form.description = ''
  form.priceYuan = ''
  form.currency = 'CNY'
  form.lifetime = true
  form.active = true
  formError.value = ''
  successMessage.value = ''
}

function openEdit(tool: AdminTool) {
  editorMode.value = 'edit'
  form.code = tool.code
  form.name = tool.name
  form.description = tool.description
  form.priceYuan = (tool.priceCents / 100).toFixed(2)
  form.currency = tool.currency
  form.lifetime = tool.lifetime
  form.active = tool.active
  formError.value = ''
  successMessage.value = ''
}

function closeEditor() {
  if (saving.value) return
  editorMode.value = null
  formError.value = ''
}

async function submitEditor() {
  const code = form.code.trim().toLowerCase()
  const name = form.name.trim()
  const description = form.description.trim()
  const price = form.priceYuan.trim()

  if (!/^[a-z][a-z0-9_]{0,63}$/.test(code)) {
    formError.value = '工具编码须以小写字母开头，只能包含小写字母、数字和下划线。'
    return
  }
  if (!name || [...name].length > 100) {
    formError.value = '工具名称不能为空，且不能超过 100 个字符。'
    return
  }
  if ([...description].length > 2000) {
    formError.value = '工具描述不能超过 2000 个字符。'
    return
  }
  if (!/^\d+(?:\.\d{1,2})?$/.test(price)) {
    formError.value = '价格须为非负金额，最多保留两位小数。'
    return
  }

  const priceCents = Math.round(Number(price) * 100)
  if (!Number.isSafeInteger(priceCents)) {
    formError.value = '价格数值过大。'
    return
  }

  saving.value = true
  formError.value = ''
  errorMessage.value = ''

  const payload = {
    name,
    description,
    priceCents,
    currency: form.currency,
    lifetime: form.lifetime,
    active: form.active,
  }

  try {
    let saved: AdminTool
    if (editorMode.value === 'create') {
      saved = await createTool({ code, ...payload })
      tools.value = [...tools.value, saved].sort((left, right) => left.id - right.id)
      successMessage.value = `工具 ${saved.name} 已创建。`
    } else {
      saved = await updateTool(code, payload)
      const index = tools.value.findIndex((tool) => tool.code === saved.code)
      if (index >= 0) {
        tools.value.splice(index, 1, saved)
      }
      successMessage.value = `工具 ${saved.name} 已更新。`
    }
    closeEditor()
  } catch (error) {
    formError.value = errorText(error, '保存工具失败，请稍后重试。')
  } finally {
    saving.value = false
  }
}

function formatPrice(tool: AdminTool) {
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: tool.currency,
    minimumFractionDigits: 2,
  }).format(tool.priceCents / 100)
}

function formatTime(timestamp: number) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(timestamp * 1000))
}

function errorText(error: unknown, fallback: string) {
  if (axios.isAxiosError(error)) {
    const message = error.response?.data?.error
    if (typeof message === 'string') return message
  }
  return fallback
}
</script>

<template>
  <section class="tools-page">
    <header class="section-heading">
      <div>
        <h2>工具目录</h2>
        <p>共 {{ tools.length }} 个工具，{{ activeCount }} 个正在上架。</p>
      </div>
      <div class="section-actions">
        <button
          class="secondary-action"
          type="button"
          :disabled="refreshing"
          @click="load(true)"
        >
          <Refresh aria-hidden="true" />
          {{ refreshing ? '刷新中' : '刷新' }}
        </button>
        <button class="primary-action" type="button" @click="openCreate">
          <Plus aria-hidden="true" />
          新增工具
        </button>
      </div>
    </header>

    <p v-if="errorMessage" class="page-message error-message" role="alert">
      {{ errorMessage }}
    </p>
    <p v-if="successMessage" class="page-message success-message" role="status">
      {{ successMessage }}
    </p>

    <div class="tool-management">
      <div class="tool-table-wrap">
        <table class="tool-table">
          <caption class="sr-only">工具目录列表</caption>
          <thead>
            <tr>
              <th scope="col">工具</th>
              <th scope="col">价格</th>
              <th scope="col">授权</th>
              <th scope="col">状态</th>
              <th scope="col">创建时间</th>
              <th scope="col"><span class="sr-only">操作</span></th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr v-for="index in 4" :key="index" class="skeleton-row">
              <td colspan="6"><span /></td>
            </tr>
          </tbody>
          <tbody v-else-if="tools.length === 0">
            <tr>
              <td colspan="6" class="empty-table">
                <strong>还没有工具</strong>
                <span>新增工具后即可配置价格并对外上架。</span>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr v-for="tool in tools" :key="tool.code">
              <td>
                <div class="tool-identity">
                  <span>{{ tool.name.slice(0, 1) }}</span>
                  <div>
                    <strong>{{ tool.name }}</strong>
                    <code>{{ tool.code }}</code>
                  </div>
                </div>
              </td>
              <td class="price-cell">{{ formatPrice(tool) }}</td>
              <td>
                <span class="license-label">{{ tool.lifetime ? '永久使用' : '限时授权' }}</span>
              </td>
              <td>
                <span class="status-badge" :class="{ active: tool.active, disabled: !tool.active }">
                  {{ tool.active ? '已上架' : '已下架' }}
                </span>
              </td>
              <td class="date-cell">{{ formatTime(tool.createdAt) }}</td>
              <td class="operation-cell">
                <button type="button" title="修改工具" @click="openEdit(tool)">
                  <EditPen aria-hidden="true" />
                  <span>修改</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </div>

    <CenteredModal
      :open="Boolean(editorMode)"
      labelledby="tool-editor-title"
      width="600px"
      @close="closeEditor"
    >
      <aside v-if="editorMode" class="tool-editor">
          <header>
            <div>
              <p>{{ editorMode === 'create' ? '创建工具' : form.code }}</p>
              <h3 id="tool-editor-title">{{ editorTitle }}</h3>
            </div>
            <button
              type="button"
              aria-label="关闭编辑弹窗"
              :disabled="saving"
              @click="closeEditor"
            >
              <Close aria-hidden="true" />
            </button>
          </header>

          <form @submit.prevent="submitEditor">
            <label for="tool-code">工具编码</label>
            <input
              id="tool-code"
              v-model="form.code"
              type="text"
              maxlength="64"
              autocomplete="off"
              :disabled="saving || editorMode === 'edit'"
              placeholder="例如 auto_click"
            >
            <span class="field-hint">
              创建后不可修改，用于关联订单和用户授权。
            </span>

            <label for="tool-name">工具名称</label>
            <input
              id="tool-name"
              v-model="form.name"
              type="text"
              maxlength="100"
              autocomplete="off"
              :disabled="saving"
              placeholder="例如 自动点击"
            >

            <label for="tool-description">工具描述</label>
            <textarea
              id="tool-description"
              v-model="form.description"
              rows="4"
              maxlength="2000"
              :disabled="saving"
              placeholder="简要说明工具用途"
            />

            <div class="form-row">
              <div>
                <label for="tool-price">价格（元）</label>
                <input
                  id="tool-price"
                  v-model="form.priceYuan"
                  type="text"
                  inputmode="decimal"
                  :disabled="saving"
                  placeholder="10.00"
                >
              </div>
              <div>
                <label for="tool-currency">货币</label>
                <select id="tool-currency" v-model="form.currency" :disabled="saving">
                  <option value="CNY">人民币 CNY</option>
                </select>
              </div>
            </div>

            <label for="tool-license">授权方式</label>
            <select id="tool-license" v-model="form.lifetime" :disabled="saving">
              <option :value="true">购买后永久使用</option>
              <option :value="false">限时授权</option>
            </select>

            <label for="tool-status">上架状态</label>
            <select id="tool-status" v-model="form.active" :disabled="saving">
              <option :value="true">上架</option>
              <option :value="false">下架</option>
            </select>

            <p v-if="formError" class="form-error" role="alert">{{ formError }}</p>

            <div class="editor-actions">
              <button class="cancel-action" type="button" :disabled="saving" @click="closeEditor">
                取消
              </button>
              <button class="save-action" type="submit" :disabled="saving">
                {{ saving ? '保存中' : '保存工具' }}
              </button>
            </div>
          </form>
      </aside>
    </CenteredModal>
  </section>
</template>

<style scoped>
.tools-page {
  min-width: 0;
  padding: 34px clamp(24px, 4vw, 48px) 48px;
}

.section-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 24px;
}

.section-heading h2,
.section-heading p {
  margin: 0;
}

.section-heading h2 {
  color: oklch(27% 0.018 165);
  font-size: 18px;
}

.section-heading p {
  margin-top: 7px;
  color: oklch(53% 0.018 165);
  font-size: 13px;
}

.section-actions {
  display: flex;
  gap: 10px;
}

.section-actions button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  min-height: 38px;
  padding: 0 13px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 700;
}

.section-actions svg {
  width: 16px;
  height: 16px;
}

.secondary-action {
  color: oklch(42% 0.025 165);
  background: oklch(99% 0.003 150);
  border: 1px solid oklch(82% 0.018 158);
}

.secondary-action:hover:not(:disabled) {
  background: oklch(94% 0.02 158);
}

.primary-action {
  color: oklch(98% 0.004 150);
  background: oklch(38% 0.075 165);
  border: 1px solid oklch(38% 0.075 165);
}

.primary-action:hover {
  background: oklch(33% 0.075 165);
}

.section-actions button:disabled {
  cursor: wait;
  opacity: 0.62;
}

.page-message {
  margin: 0 0 16px;
  padding: 10px 12px;
  border: 1px solid;
  border-radius: 6px;
  font-size: 13px;
}

.error-message,
.form-error {
  color: oklch(42% 0.13 28);
  background: oklch(95% 0.035 28);
  border-color: oklch(82% 0.065 28);
}

.success-message {
  color: oklch(35% 0.08 150);
  background: oklch(94% 0.035 150);
  border-color: oklch(80% 0.055 150);
}

.tool-management {
  min-width: 0;
}

.tool-table-wrap {
  width: 100%;
  min-width: 0;
  overflow-x: auto;
  background: oklch(99% 0.003 150);
  border: 1px solid oklch(87% 0.014 155);
  border-radius: 8px;
}

.tool-table {
  width: 100%;
  min-width: 820px;
  border-collapse: collapse;
  font-size: 13px;
}

.tool-table th {
  height: 44px;
  padding: 0 16px;
  color: oklch(51% 0.018 165);
  background: oklch(96% 0.01 155);
  border-bottom: 1px solid oklch(87% 0.014 155);
  font-size: 12px;
  text-align: left;
}

.tool-table td {
  height: 66px;
  padding: 10px 16px;
  border-bottom: 1px solid oklch(91% 0.01 155);
}

.tool-table tbody tr:last-child td {
  border-bottom: 0;
}

.tool-table tbody tr:hover {
  background: oklch(98% 0.008 155);
}

.tool-identity {
  display: flex;
  align-items: center;
  gap: 11px;
}

.tool-identity > span {
  display: grid;
  flex: none;
  place-items: center;
  width: 34px;
  height: 34px;
  color: oklch(35% 0.09 165);
  background: oklch(91% 0.045 158);
  border-radius: 6px;
  font-weight: 800;
}

.tool-identity div {
  display: grid;
  gap: 4px;
}

.tool-identity strong {
  color: oklch(29% 0.018 165);
}

.tool-identity code {
  color: oklch(55% 0.02 165);
  font-size: 11px;
}

.price-cell {
  color: oklch(30% 0.03 165);
  font-weight: 700;
  white-space: nowrap;
}

.license-label {
  color: oklch(44% 0.028 165);
  white-space: nowrap;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 8px;
  border-radius: 5px;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
}

.status-badge::before {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  content: "";
}

.status-badge.active {
  color: oklch(37% 0.085 150);
  background: oklch(93% 0.038 150);
}

.status-badge.active::before {
  background: oklch(55% 0.12 150);
}

.status-badge.disabled {
  color: oklch(47% 0.065 60);
  background: oklch(94% 0.035 75);
}

.status-badge.disabled::before {
  background: oklch(60% 0.1 65);
}

.date-cell {
  color: oklch(48% 0.016 165);
  white-space: nowrap;
}

.operation-cell {
  width: 104px;
  text-align: right;
}

.operation-cell button {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-height: 30px;
  padding: 0 8px;
  color: oklch(44% 0.035 165);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 5px;
  cursor: pointer;
  font-size: 12px;
  white-space: nowrap;
}

.operation-cell button:hover {
  color: oklch(34% 0.07 165);
  background: oklch(93% 0.03 158);
}

.operation-cell svg {
  width: 14px;
  height: 14px;
}

.empty-table {
  height: 180px !important;
  color: oklch(51% 0.018 165);
  text-align: center;
}

.empty-table strong,
.empty-table span {
  display: block;
}

.empty-table strong {
  color: oklch(34% 0.018 165);
  font-size: 14px;
}

.empty-table span {
  margin-top: 7px;
  font-size: 12px;
}

.skeleton-row span {
  display: block;
  width: 100%;
  height: 18px;
  background: oklch(92% 0.012 160);
  border-radius: 4px;
  animation: skeleton-pulse 1.2s ease-in-out infinite alternate;
}

.tool-editor {
  padding: 22px;
}

.tool-editor > header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid oklch(89% 0.012 155);
}

.tool-editor header p,
.tool-editor header h3 {
  margin: 0;
}

.tool-editor header p {
  margin-bottom: 5px;
  color: oklch(49% 0.07 165);
  font-size: 11px;
  font-weight: 800;
}

.tool-editor header h3 {
  color: oklch(27% 0.018 165);
  font-size: 17px;
}

.tool-editor header button {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  padding: 0;
  color: oklch(52% 0.018 165);
  background: transparent;
  border: 0;
  border-radius: 5px;
  cursor: pointer;
}

.tool-editor header button:hover {
  background: oklch(93% 0.018 158);
}

.tool-editor header button:disabled {
  cursor: wait;
  opacity: 0.5;
}

.tool-editor header svg {
  width: 17px;
  height: 17px;
}

.tool-editor form {
  display: grid;
  padding-top: 18px;
}

.tool-editor label {
  margin: 16px 0 8px;
  color: oklch(35% 0.018 165);
  font-size: 12px;
  font-weight: 700;
}

.tool-editor label:first-child {
  margin-top: 0;
}

.tool-editor input,
.tool-editor select,
.tool-editor textarea {
  width: 100%;
  padding: 0 11px;
  color: oklch(27% 0.018 165);
  background: oklch(99% 0.003 150);
  border: 1px solid oklch(82% 0.018 158);
  border-radius: 6px;
  outline: none;
  font: inherit;
  font-size: 13px;
}

.tool-editor input,
.tool-editor select {
  height: 42px;
}

.tool-editor textarea {
  min-height: 88px;
  padding-top: 10px;
  resize: vertical;
  line-height: 1.5;
}

.tool-editor input:focus,
.tool-editor select:focus,
.tool-editor textarea:focus {
  border-color: oklch(51% 0.085 165);
  box-shadow: 0 0 0 3px oklch(74% 0.08 160 / 0.16);
}

.tool-editor input:disabled,
.tool-editor select:disabled,
.tool-editor textarea:disabled {
  cursor: wait;
  opacity: 0.68;
}

.form-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 130px;
  gap: 10px;
}

.form-row > div {
  display: grid;
}

.field-hint {
  margin-top: 6px;
  color: oklch(57% 0.014 165);
  font-size: 10px;
  line-height: 1.5;
}

.form-error {
  margin: 16px 0 0;
  padding: 9px 10px;
  border: 1px solid oklch(82% 0.065 28);
  border-radius: 5px;
  font-size: 11px;
  line-height: 1.5;
}

.editor-actions {
  display: grid;
  grid-template-columns: 1fr 1.4fr;
  gap: 8px;
  margin-top: 22px;
}

.editor-actions button {
  min-height: 40px;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 700;
}

.cancel-action {
  color: oklch(44% 0.025 165);
  background: transparent;
  border: 1px solid oklch(82% 0.018 158);
}

.save-action {
  color: oklch(98% 0.004 150);
  background: oklch(38% 0.075 165);
  border: 1px solid oklch(38% 0.075 165);
}

.editor-actions button:disabled {
  cursor: wait;
  opacity: 0.65;
}

.section-actions button:focus-visible,
.operation-cell button:focus-visible,
.tool-editor button:focus-visible,
.tool-editor input:focus-visible,
.tool-editor select:focus-visible,
.tool-editor textarea:focus-visible {
  outline: 2px solid oklch(51% 0.085 165);
  outline-offset: 2px;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@keyframes skeleton-pulse {
  to {
    opacity: 0.45;
  }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton-row span {
    animation: none;
  }
}
</style>
