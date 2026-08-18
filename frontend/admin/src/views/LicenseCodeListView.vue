<script setup lang="ts">
import {
  Close,
  CopyDocument,
  Download,
  Key,
  Plus,
  Refresh,
  Search,
} from '@element-plus/icons-vue'
import axios from 'axios'
import { computed, onMounted, reactive, ref } from 'vue'
import {
  generateLicenseCodes,
  listAdminTools,
  listLicenseCodes,
  revokeLicenseCode,
  type AdminTool,
  type GeneratedLicenseCode,
  type LicenseCodeRecord,
  type LicenseCodeStatus,
} from '../api/admin'
import CenteredModal from '../components/CenteredModal.vue'

const codes = ref<LicenseCodeRecord[]>([])
const tools = ref<AdminTool[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const status = ref<'' | LicenseCodeStatus>('')
const toolCode = ref('')
const search = ref('')
const loading = ref(false)
const refreshing = ref(false)
const generating = ref(false)
const revokingId = ref<number | null>(null)
const generatorOpen = ref(false)
const generatedCodes = ref<GeneratedLicenseCode[]>([])
const generatedBatchNo = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const generatorError = ref('')
const copyState = ref('')

const generator = reactive({
  toolCode: '',
  count: 1,
  note: '',
})

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const generatedText = computed(() => generatedCodes.value.map((item) => item.code).join('\n'))

onMounted(async () => {
  await Promise.all([loadTools(), load(false)])
})

async function loadTools() {
  try {
    tools.value = await listAdminTools()
  } catch (error) {
    errorMessage.value = errorText(error, '工具列表加载失败。')
  }
}

async function load(isRefresh: boolean) {
  if (isRefresh) refreshing.value = true
  else loading.value = true
  errorMessage.value = ''
  try {
    const response = await listLicenseCodes({
      page: page.value,
      pageSize: pageSize.value,
      status: status.value,
      toolCode: toolCode.value || undefined,
      search: search.value.trim() || undefined,
    })
    codes.value = response.codes
    total.value = response.total
    if (page.value > pageCount.value) {
      page.value = pageCount.value
      await load(false)
    }
  } catch (error) {
    errorMessage.value = errorText(error, '授权码列表加载失败，请稍后重试。')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

async function applyFilters() {
  page.value = 1
  await load(false)
}

async function resetFilters() {
  status.value = ''
  toolCode.value = ''
  search.value = ''
  page.value = 1
  await load(false)
}

function openGenerator() {
  generatedCodes.value = []
  generatedBatchNo.value = ''
  generator.toolCode = tools.value.find((item) => item.active)?.code || ''
  generator.count = 1
  generator.note = ''
  generatorError.value = ''
  copyState.value = ''
  generatorOpen.value = true
}

function closeGenerator() {
  if (generating.value) return
  if (generatedCodes.value.length > 0 && !window.confirm('关闭后无法再次查看本批授权码明文，确认关闭？')) {
    return
  }
  generatorOpen.value = false
}

async function submitGenerator() {
  if (!generator.toolCode) {
    generatorError.value = '请选择授权工具。'
    return
  }
  if (!Number.isInteger(generator.count) || generator.count < 1 || generator.count > 100) {
    generatorError.value = '生成数量必须是 1 到 100 的整数。'
    return
  }
  if ([...generator.note.trim()].length > 200) {
    generatorError.value = '备注不能超过 200 个字符。'
    return
  }

  generating.value = true
  generatorError.value = ''
  try {
    const response = await generateLicenseCodes({
      toolCode: generator.toolCode,
      count: generator.count,
      note: generator.note.trim(),
    })
    generatedBatchNo.value = response.batchNo
    generatedCodes.value = response.codes
    successMessage.value = `已生成 ${response.codes.length} 个授权码。`
    page.value = 1
    await load(false)
  } catch (error) {
    generatorError.value = errorText(error, '授权码生成失败，请稍后重试。')
  } finally {
    generating.value = false
  }
}

async function copyGenerated() {
  try {
    await navigator.clipboard.writeText(generatedText.value)
    copyState.value = '已复制全部授权码'
  } catch {
    copyState.value = '复制失败，请选中文本后手动复制'
  }
}

function downloadGenerated() {
  const content = [
    `批次：${generatedBatchNo.value}`,
    `工具：${generator.toolCode}`,
    `备注：${generator.note.trim() || '-'}`,
    '',
    generatedText.value,
    '',
  ].join('\n')
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `${generatedBatchNo.value}.txt`
  anchor.click()
  URL.revokeObjectURL(url)
}

async function revoke(record: LicenseCodeRecord) {
  if (record.status !== 'active') return
  if (!window.confirm(`确认作废授权码 ${record.codeHint}？作废后不能恢复。`)) return
  revokingId.value = record.id
  errorMessage.value = ''
  try {
    const updated = await revokeLicenseCode(record.id)
    const index = codes.value.findIndex((item) => item.id === updated.id)
    if (index >= 0) codes.value.splice(index, 1, updated)
    successMessage.value = `授权码 ${updated.codeHint} 已作废。`
  } catch (error) {
    errorMessage.value = errorText(error, '授权码作废失败，请稍后重试。')
  } finally {
    revokingId.value = null
  }
}

async function goToPage(nextPage: number) {
  if (nextPage < 1 || nextPage > pageCount.value || nextPage === page.value) return
  page.value = nextPage
  await load(false)
}

function statusText(value: LicenseCodeStatus) {
  return { active: '未使用', redeemed: '已兑换', revoked: '已作废' }[value]
}

function formatTime(timestamp?: number) {
  if (!timestamp) return '-'
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
    if (typeof message === 'string' && message) return message
  }
  return fallback
}
</script>

<template>
  <section class="license-page">
    <header class="section-heading">
      <div>
        <h2>授权码库存</h2>
        <p>共 {{ total }} 个记录，授权码明文仅在生成完成时展示一次。</p>
      </div>
      <div class="heading-actions">
        <button class="secondary-action" type="button" :disabled="refreshing" @click="load(true)">
          <Refresh aria-hidden="true" />
          {{ refreshing ? '刷新中' : '刷新' }}
        </button>
        <button class="primary-action" type="button" :disabled="tools.length === 0" @click="openGenerator">
          <Plus aria-hidden="true" />
          生成授权码
        </button>
      </div>
    </header>

    <p v-if="errorMessage" class="page-message error-message" role="alert">{{ errorMessage }}</p>
    <p v-if="successMessage" class="page-message success-message" role="status">{{ successMessage }}</p>

    <form class="filter-bar" @submit.prevent="applyFilters">
      <label>
        <span>状态</span>
        <select v-model="status">
          <option value="">全部状态</option>
          <option value="active">未使用</option>
          <option value="redeemed">已兑换</option>
          <option value="revoked">已作废</option>
        </select>
      </label>
      <label>
        <span>工具</span>
        <select v-model="toolCode">
          <option value="">全部工具</option>
          <option v-for="tool in tools" :key="tool.code" :value="tool.code">{{ tool.name }}</option>
        </select>
      </label>
      <label class="search-field">
        <span>查询</span>
        <div>
          <Search aria-hidden="true" />
          <input v-model="search" type="search" placeholder="授权码、批次、备注或用户" autocomplete="off">
        </div>
      </label>
      <button class="filter-submit" type="submit">查询</button>
      <button class="filter-reset" type="button" @click="resetFilters">重置</button>
    </form>

    <div class="table-wrap">
      <table>
        <caption class="sr-only">授权码库存列表</caption>
        <thead>
          <tr>
            <th scope="col">授权码</th>
            <th scope="col">工具 / 批次</th>
            <th scope="col">状态</th>
            <th scope="col">兑换用户</th>
            <th scope="col">生成信息</th>
            <th scope="col"><span class="sr-only">操作</span></th>
          </tr>
        </thead>
        <tbody v-if="loading">
          <tr v-for="index in 5" :key="index" class="skeleton-row"><td colspan="6"><span /></td></tr>
        </tbody>
        <tbody v-else-if="codes.length === 0">
          <tr>
            <td colspan="6" class="empty-table">
              <Key aria-hidden="true" />
              <strong>没有符合条件的授权码</strong>
              <span>调整筛选条件，或生成一批新的授权码。</span>
            </td>
          </tr>
        </tbody>
        <tbody v-else>
          <tr v-for="record in codes" :key="record.id">
            <td>
              <code class="code-hint">{{ record.codeHint }}</code>
              <span v-if="record.note" class="record-note" :title="record.note">{{ record.note }}</span>
            </td>
            <td>
              <strong class="tool-name">{{ record.toolName }}</strong>
              <code class="batch-no">{{ record.batchNo }}</code>
            </td>
            <td>
              <span class="status-badge" :class="record.status">{{ statusText(record.status) }}</span>
              <small v-if="record.redeemedAt">{{ formatTime(record.redeemedAt) }}</small>
              <small v-else-if="record.revokedAt">{{ formatTime(record.revokedAt) }}</small>
            </td>
            <td>
              <template v-if="record.redeemedByUserId">
                <strong class="user-name">{{ record.redeemedByUsername || `用户 ${record.redeemedByUserId}` }}</strong>
                <span class="user-email">{{ record.redeemedByEmail || `ID ${record.redeemedByUserId}` }}</span>
              </template>
              <span v-else class="muted-value">-</span>
            </td>
            <td>
              <span class="creator">{{ record.createdByAdminUsername || '已删除管理员' }}</span>
              <small>{{ formatTime(record.createdAt) }}</small>
            </td>
            <td class="operation-cell">
              <button
                v-if="record.status === 'active'"
                type="button"
                :disabled="revokingId === record.id"
                @click="revoke(record)"
              >
                {{ revokingId === record.id ? '处理中' : '作废' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <footer class="pagination">
      <span>第 {{ page }} / {{ pageCount }} 页</span>
      <div>
        <button type="button" :disabled="page <= 1 || loading" @click="goToPage(page - 1)">上一页</button>
        <button type="button" :disabled="page >= pageCount || loading" @click="goToPage(page + 1)">下一页</button>
      </div>
    </footer>

    <CenteredModal :open="generatorOpen" labelledby="license-generator-title" width="640px" @close="closeGenerator">
      <section class="generator-panel">
        <header>
          <div>
            <p>{{ generatedCodes.length ? generatedBatchNo : '一次性兑换凭证' }}</p>
            <h3 id="license-generator-title">{{ generatedCodes.length ? '授权码已生成' : '生成授权码' }}</h3>
          </div>
          <button type="button" aria-label="关闭" :disabled="generating" @click="closeGenerator"><Close /></button>
        </header>

        <template v-if="generatedCodes.length">
          <div class="one-time-notice" role="status">
            <Key aria-hidden="true" />
            <div>
              <strong>请立即保存</strong>
              <span>服务器不保存授权码明文，关闭后无法再次查看。</span>
            </div>
          </div>
          <textarea class="generated-output" :value="generatedText" readonly aria-label="本批授权码" />
          <p v-if="copyState" class="copy-state">{{ copyState }}</p>
          <div class="result-actions">
            <button class="download-action" type="button" @click="downloadGenerated">
              <Download aria-hidden="true" />下载 TXT
            </button>
            <button class="copy-action" type="button" @click="copyGenerated">
              <CopyDocument aria-hidden="true" />复制全部
            </button>
          </div>
        </template>

        <form v-else @submit.prevent="submitGenerator">
          <label for="generator-tool">授权工具</label>
          <select id="generator-tool" v-model="generator.toolCode" :disabled="generating">
            <option value="" disabled>请选择工具</option>
            <option v-for="tool in tools.filter((item) => item.active)" :key="tool.code" :value="tool.code">
              {{ tool.name }}（{{ tool.code }}）
            </option>
          </select>

          <label for="generator-count">生成数量</label>
          <input id="generator-count" v-model.number="generator.count" type="number" min="1" max="100" step="1" :disabled="generating">
          <span class="field-hint">单次最多生成 100 个，每个授权码只能兑换一次。</span>

          <label for="generator-note">批次备注</label>
          <input id="generator-note" v-model="generator.note" type="text" maxlength="200" :disabled="generating" placeholder="例如：闲鱼 2026-08 第一批">

          <p v-if="generatorError" class="form-error" role="alert">{{ generatorError }}</p>
          <div class="form-actions">
            <button class="cancel-action" type="button" :disabled="generating" @click="closeGenerator">取消</button>
            <button class="generate-action" type="submit" :disabled="generating">
              {{ generating ? '生成中' : '确认生成' }}
            </button>
          </div>
        </form>
      </section>
    </CenteredModal>
  </section>
</template>

<style scoped>
.license-page { min-width: 0; padding: 34px clamp(24px, 4vw, 48px) 48px; }
.section-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; margin-bottom: 22px; }
.section-heading h2, .section-heading p { margin: 0; }
.section-heading h2 { color: oklch(27% 0.018 165); font-size: 18px; }
.section-heading p { margin-top: 7px; color: oklch(53% 0.018 165); font-size: 13px; }
.heading-actions, .result-actions, .form-actions, .pagination div { display: flex; gap: 9px; }
.heading-actions button, .result-actions button, .form-actions button, .pagination button { display: inline-flex; align-items: center; justify-content: center; gap: 7px; min-height: 38px; padding: 0 13px; border-radius: 6px; cursor: pointer; font-size: 13px; font-weight: 700; }
.heading-actions svg, .result-actions svg { width: 16px; height: 16px; }
.primary-action, .generate-action, .copy-action { color: oklch(98% 0.004 150); background: oklch(38% 0.075 165); border: 1px solid oklch(38% 0.075 165); }
.secondary-action, .download-action, .cancel-action { color: oklch(42% 0.025 165); background: oklch(99% 0.003 150); border: 1px solid oklch(82% 0.018 158); }
button:disabled { cursor: wait !important; opacity: .58; }
.page-message { margin: 0 0 14px; padding: 10px 12px; border: 1px solid; border-radius: 6px; font-size: 13px; }
.error-message, .form-error { color: oklch(42% 0.13 28); background: oklch(95% 0.035 28); border-color: oklch(82% 0.065 28); }
.success-message { color: oklch(35% 0.08 150); background: oklch(94% 0.035 150); border-color: oklch(80% 0.055 150); }
.filter-bar { display: grid; grid-template-columns: 150px 180px minmax(260px, 1fr) auto auto; align-items: end; gap: 10px; margin-bottom: 14px; padding: 13px 14px; background: oklch(99% 0.003 150); border: 1px solid oklch(87% 0.014 155); border-radius: 8px; }
.filter-bar label { display: grid; gap: 6px; color: oklch(45% 0.018 165); font-size: 11px; font-weight: 700; }
.filter-bar select, .filter-bar input { width: 100%; height: 36px; color: oklch(29% 0.018 165); background: oklch(99% 0.003 150); border: 1px solid oklch(82% 0.018 158); border-radius: 5px; outline: none; font: inherit; font-size: 12px; }
.filter-bar select { padding: 0 9px; }
.search-field div { position: relative; }
.search-field svg { position: absolute; top: 10px; left: 10px; width: 15px; height: 15px; color: oklch(57% 0.014 165); }
.search-field input { padding: 0 10px 0 32px; }
.filter-submit, .filter-reset { height: 36px; padding: 0 14px; border-radius: 5px; cursor: pointer; font-size: 12px; font-weight: 700; }
.filter-submit { color: oklch(98% 0.004 150); background: oklch(38% 0.075 165); border: 1px solid oklch(38% 0.075 165); }
.filter-reset { color: oklch(45% 0.02 165); background: transparent; border: 1px solid oklch(82% 0.018 158); }
.table-wrap { width: 100%; min-width: 0; overflow-x: auto; background: oklch(99% 0.003 150); border: 1px solid oklch(87% 0.014 155); border-radius: 8px; }
table { width: 100%; min-width: 1040px; border-collapse: collapse; font-size: 12px; }
th { height: 43px; padding: 0 15px; color: oklch(51% 0.018 165); background: oklch(96% 0.01 155); border-bottom: 1px solid oklch(87% 0.014 155); font-size: 11px; text-align: left; }
td { height: 67px; padding: 10px 15px; border-bottom: 1px solid oklch(91% 0.01 155); vertical-align: middle; }
tbody tr:last-child td { border-bottom: 0; }
tbody tr:hover { background: oklch(98% 0.008 155); }
.code-hint { display: block; color: oklch(31% 0.045 165); font-size: 12px; font-weight: 800; }
.record-note { display: block; max-width: 180px; margin-top: 5px; overflow: hidden; color: oklch(55% 0.016 165); text-overflow: ellipsis; white-space: nowrap; }
.tool-name, .user-name { display: block; color: oklch(31% 0.018 165); }
.batch-no, .user-email, .creator, td small { display: block; margin-top: 4px; color: oklch(55% 0.016 165); font-size: 10px; }
.status-badge { display: inline-flex; padding: 4px 8px; border-radius: 5px; font-size: 11px; font-weight: 800; }
.status-badge.active { color: oklch(38% 0.085 150); background: oklch(93% 0.038 150); }
.status-badge.redeemed { color: oklch(39% 0.085 250); background: oklch(94% 0.035 250); }
.status-badge.revoked { color: oklch(45% 0.055 30); background: oklch(95% 0.03 30); }
.muted-value { color: oklch(64% 0.012 165); }
.operation-cell { width: 76px; text-align: right; }
.operation-cell button { min-height: 30px; padding: 0 9px; color: oklch(43% 0.09 28); background: transparent; border: 1px solid oklch(84% 0.045 28); border-radius: 5px; cursor: pointer; font-size: 11px; }
.empty-table { height: 210px; color: oklch(53% 0.018 165); text-align: center; }
.empty-table svg { width: 26px; height: 26px; margin-bottom: 8px; }
.empty-table strong, .empty-table span { display: block; }
.empty-table strong { color: oklch(34% 0.018 165); font-size: 14px; }
.empty-table span { margin-top: 6px; }
.skeleton-row span { display: block; height: 17px; background: oklch(92% 0.012 160); border-radius: 4px; animation: pulse 1.2s ease-in-out infinite alternate; }
.pagination { display: flex; align-items: center; justify-content: space-between; margin-top: 12px; color: oklch(53% 0.018 165); font-size: 12px; }
.pagination button { min-height: 32px; color: oklch(43% 0.025 165); background: oklch(99% 0.003 150); border: 1px solid oklch(82% 0.018 158); }
.generator-panel { padding: 22px; }
.generator-panel > header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; padding-bottom: 16px; border-bottom: 1px solid oklch(89% 0.012 155); }
.generator-panel header p, .generator-panel header h3 { margin: 0; }
.generator-panel header p { margin-bottom: 5px; color: oklch(49% 0.07 165); font-size: 11px; font-weight: 800; }
.generator-panel header h3 { color: oklch(27% 0.018 165); font-size: 17px; }
.generator-panel header button { display: grid; place-items: center; width: 30px; height: 30px; padding: 0; color: oklch(52% 0.018 165); background: transparent; border: 0; border-radius: 5px; cursor: pointer; }
.generator-panel header svg { width: 17px; height: 17px; }
.generator-panel form { display: grid; padding-top: 4px; }
.generator-panel form label { margin: 15px 0 7px; color: oklch(35% 0.018 165); font-size: 12px; font-weight: 700; }
.generator-panel form input, .generator-panel form select { width: 100%; height: 40px; padding: 0 10px; color: oklch(27% 0.018 165); background: oklch(99% 0.003 150); border: 1px solid oklch(82% 0.018 158); border-radius: 6px; outline: none; font: inherit; font-size: 13px; }
.field-hint { margin-top: 6px; color: oklch(57% 0.014 165); font-size: 10px; }
.form-error { margin: 15px 0 0; padding: 9px 10px; border: 1px solid; border-radius: 5px; font-size: 11px; }
.form-actions { display: grid; grid-template-columns: 1fr 1.4fr; margin-top: 22px; }
.one-time-notice { display: flex; align-items: center; gap: 12px; margin: 18px 0 12px; padding: 12px 14px; color: oklch(38% 0.085 75); background: oklch(95% 0.04 85); border: 1px solid oklch(82% 0.07 80); border-radius: 6px; }
.one-time-notice > svg { flex: none; width: 20px; height: 20px; }
.one-time-notice strong, .one-time-notice span { display: block; }
.one-time-notice span { margin-top: 3px; font-size: 11px; }
.generated-output { width: 100%; height: 230px; padding: 13px; resize: none; color: oklch(27% 0.025 165); background: oklch(97% 0.008 155); border: 1px solid oklch(82% 0.018 158); border-radius: 6px; outline: none; font: 600 13px/1.7 Consolas, monospace; }
.copy-state { min-height: 18px; margin: 8px 0 0; color: oklch(40% 0.07 155); font-size: 11px; }
.result-actions { justify-content: flex-end; margin-top: 12px; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
button:focus-visible, input:focus-visible, select:focus-visible, textarea:focus-visible { outline: 2px solid oklch(51% 0.085 165); outline-offset: 2px; }
@keyframes pulse { to { opacity: .45; } }
@media (max-width: 960px) { .filter-bar { grid-template-columns: 1fr 1fr; } .search-field { grid-column: 1 / -1; } }
@media (prefers-reduced-motion: reduce) { .skeleton-row span { animation: none; } }
</style>
