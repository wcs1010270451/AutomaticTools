<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import axios from 'axios'
import { Close, Delete, EditPen, Plus, Refresh } from '@element-plus/icons-vue'
import {
  createAdmin,
  deleteAdmin,
  listAdmins,
  updateAdmin,
  type AdminAccount,
} from '../api/admin'
import CenteredModal from '../components/CenteredModal.vue'
import { useAdminAuthStore } from '../stores/adminAuth'

type EditorMode = 'create' | 'edit'

const auth = useAdminAuthStore()
const admins = ref<AdminAccount[]>([])
const loading = ref(true)
const refreshing = ref(false)
const saving = ref(false)
const deletingId = ref<number | null>(null)
const deleteConfirmId = ref<number | null>(null)
const editorMode = ref<EditorMode | null>(null)
const errorMessage = ref('')
const successMessage = ref('')
const formError = ref('')

const form = reactive({
  id: 0,
  username: '',
  password: '',
  status: 'active' as 'active' | 'disabled',
})

const editorTitle = computed(() => editorMode.value === 'create' ? '新增管理员' : '修改管理员')
const editingSelf = computed(() => editorMode.value === 'edit' && form.id === auth.admin?.id)

onMounted(() => load(false))

async function load(isRefresh: boolean) {
  if (isRefresh) {
    refreshing.value = true
  } else {
    loading.value = true
  }
  errorMessage.value = ''

  try {
    admins.value = await listAdmins()
  } catch (error) {
    errorMessage.value = errorText(error, '管理员列表加载失败，请稍后重试。')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function openCreate() {
  editorMode.value = 'create'
  deleteConfirmId.value = null
  form.id = 0
  form.username = ''
  form.password = ''
  form.status = 'active'
  formError.value = ''
  successMessage.value = ''
}

function openEdit(admin: AdminAccount) {
  editorMode.value = 'edit'
  deleteConfirmId.value = null
  form.id = admin.id
  form.username = admin.username
  form.password = ''
  form.status = admin.status as 'active' | 'disabled'
  formError.value = ''
  successMessage.value = ''
}

function closeEditor() {
  if (saving.value) return
  editorMode.value = null
  formError.value = ''
}

async function submitEditor() {
  const username = form.username.trim()
  if (!/^[a-zA-Z0-9_.-]{3,32}$/.test(username)) {
    formError.value = '用户名须为 3 到 32 位，仅可包含字母、数字、下划线、横线或点。'
    return
  }
  if (editorMode.value === 'create' && !form.password) {
    formError.value = '请输入管理员密码。'
    return
  }
  if (form.password && form.password.length < 6) {
    formError.value = '密码至少需要 6 个字符。'
    return
  }

  saving.value = true
  formError.value = ''
  errorMessage.value = ''

  try {
    let saved: AdminAccount
    if (editorMode.value === 'create') {
      saved = await createAdmin({
        username,
        password: form.password,
        status: form.status,
      })
      admins.value = [...admins.value, saved].sort((left, right) => left.id - right.id)
      successMessage.value = `管理员 ${saved.username} 已创建。`
    } else {
      saved = await updateAdmin(form.id, {
        username,
        ...(form.password ? { password: form.password } : {}),
        status: form.status,
      })
      const index = admins.value.findIndex((admin) => admin.id === saved.id)
      if (index >= 0) {
        admins.value.splice(index, 1, saved)
      }
      if (saved.id === auth.admin?.id) {
        auth.updateAccount(saved)
      }
      successMessage.value = `管理员 ${saved.username} 已更新。`
    }
    closeEditor()
  } catch (error) {
    formError.value = errorText(error, '保存管理员失败，请稍后重试。')
  } finally {
    saving.value = false
  }
}

function askDelete(adminId: number) {
  deleteConfirmId.value = adminId
  editorMode.value = null
  successMessage.value = ''
}

async function confirmDelete(admin: AdminAccount) {
  deletingId.value = admin.id
  errorMessage.value = ''

  try {
    await deleteAdmin(admin.id)
    admins.value = admins.value.filter((item) => item.id !== admin.id)
    deleteConfirmId.value = null
    successMessage.value = `管理员 ${admin.username} 已删除。`
  } catch (error) {
    errorMessage.value = errorText(error, '删除管理员失败，请稍后重试。')
  } finally {
    deletingId.value = null
  }
}

function formatTime(timestamp?: number) {
  if (!timestamp) return '从未登录'
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
  <section class="admins-page">
    <header class="section-heading">
      <div>
        <h2>管理员账号</h2>
        <p>管理可登录后台的账号、状态与密码。</p>
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
          新增管理员
        </button>
      </div>
    </header>

    <p v-if="errorMessage" class="page-message error-message" role="alert">
      {{ errorMessage }}
    </p>
    <p v-if="successMessage" class="page-message success-message" role="status">
      {{ successMessage }}
    </p>

    <div class="admin-management">
      <div class="admin-table-wrap">
        <table class="admin-table">
          <caption class="sr-only">管理员账号列表</caption>
          <thead>
            <tr>
              <th scope="col">账号</th>
              <th scope="col">状态</th>
              <th scope="col">最后登录</th>
              <th scope="col">创建时间</th>
              <th scope="col"><span class="sr-only">操作</span></th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr v-for="index in 4" :key="index" class="skeleton-row">
              <td colspan="5"><span /></td>
            </tr>
          </tbody>
          <tbody v-else-if="admins.length === 0">
            <tr>
              <td colspan="5" class="empty-table">
                <strong>还没有管理员账号</strong>
                <span>创建第一个管理员后即可在此维护。</span>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr v-for="admin in admins" :key="admin.id">
              <td>
                <div class="admin-identity">
                  <span class="admin-avatar">{{ admin.username.slice(0, 1).toUpperCase() }}</span>
                  <div>
                    <strong>{{ admin.username }}</strong>
                    <span>
                      ID {{ admin.id }}
                      <em v-if="admin.id === auth.admin?.id">当前账号</em>
                    </span>
                  </div>
                </div>
              </td>
              <td>
                <span class="status-badge" :class="admin.status">
                  {{ admin.status === 'active' ? '启用' : '停用' }}
                </span>
              </td>
              <td class="date-cell">{{ formatTime(admin.lastLoginAt) }}</td>
              <td class="date-cell">{{ formatTime(admin.createdAt) }}</td>
              <td class="operation-cell">
                <div v-if="deleteConfirmId === admin.id" class="delete-confirm">
                  <span>确认删除？</span>
                  <button type="button" @click="deleteConfirmId = null">取消</button>
                  <button
                    class="confirm-delete"
                    type="button"
                    :disabled="deletingId === admin.id"
                    @click="confirmDelete(admin)"
                  >
                    {{ deletingId === admin.id ? '删除中' : '删除' }}
                  </button>
                </div>
                <div v-else class="row-actions">
                  <button type="button" title="修改管理员" @click="openEdit(admin)">
                    <EditPen aria-hidden="true" />
                    <span>修改</span>
                  </button>
                  <button
                    class="delete-action"
                    type="button"
                    :disabled="admin.id === auth.admin?.id"
                    :title="admin.id === auth.admin?.id ? '不能删除当前登录账号' : '删除管理员'"
                    @click="askDelete(admin.id)"
                  >
                    <Delete aria-hidden="true" />
                    <span>删除</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </div>

    <CenteredModal
      :open="Boolean(editorMode)"
      labelledby="editor-title"
      width="440px"
      @close="closeEditor"
    >
      <aside v-if="editorMode" class="admin-editor">
          <header>
            <div>
              <p>{{ editorMode === 'create' ? '新增账号' : `管理员 #${form.id}` }}</p>
              <h3 id="editor-title">{{ editorTitle }}</h3>
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
            <label for="editor-username">管理员账号</label>
            <input
              id="editor-username"
              v-model="form.username"
              type="text"
              autocomplete="off"
              maxlength="32"
              :disabled="saving"
              placeholder="例如 admin_ops"
            >
            <span class="field-hint">3 到 32 位，可使用字母、数字、下划线、横线和点。</span>

            <label for="editor-password">
              密码
              <span v-if="editorMode === 'edit'">（选填）</span>
            </label>
            <input
              id="editor-password"
              v-model="form.password"
              type="password"
              :autocomplete="editorMode === 'create' ? 'new-password' : 'off'"
              :disabled="saving"
              :placeholder="editorMode === 'create' ? '至少 6 个字符' : '留空则不修改密码'"
            >

            <label for="editor-status">账号状态</label>
            <select id="editor-status" v-model="form.status" :disabled="saving">
              <option value="active">启用</option>
              <option value="disabled" :disabled="editingSelf">停用</option>
            </select>
            <span v-if="editingSelf" class="field-hint">当前登录账号不能设为停用。</span>

            <p v-if="formError" class="form-error" role="alert">{{ formError }}</p>

            <div class="editor-actions">
              <button class="cancel-action" type="button" :disabled="saving" @click="closeEditor">
                取消
              </button>
              <button class="save-action" type="submit" :disabled="saving">
                {{ saving ? '保存中' : '保存管理员' }}
              </button>
            </div>
          </form>
      </aside>
    </CenteredModal>
  </section>
</template>

<style scoped>
.admins-page {
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
  transition:
    background 180ms ease-out,
    border-color 180ms ease-out,
    color 180ms ease-out;
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
  border-color: oklch(72% 0.04 160);
}

.primary-action {
  color: oklch(98% 0.004 150);
  background: oklch(38% 0.075 165);
  border: 1px solid oklch(38% 0.075 165);
}

.primary-action:hover {
  background: oklch(33% 0.075 165);
  border-color: oklch(33% 0.075 165);
}

.section-actions button:focus-visible,
.row-actions button:focus-visible,
.delete-confirm button:focus-visible,
.admin-editor button:focus-visible,
.admin-editor input:focus-visible,
.admin-editor select:focus-visible {
  outline: 2px solid oklch(51% 0.085 165);
  outline-offset: 2px;
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

.admin-management {
  min-width: 0;
}

.admin-table-wrap {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  overflow-x: auto;
  background: oklch(99% 0.003 150);
  border: 1px solid oklch(87% 0.014 155);
  border-radius: 8px;
}

.admin-table {
  width: 100%;
  min-width: 760px;
  border-collapse: collapse;
  font-size: 13px;
}

.admin-table th {
  height: 44px;
  padding: 0 16px;
  color: oklch(51% 0.018 165);
  background: oklch(96% 0.01 155);
  border-bottom: 1px solid oklch(87% 0.014 155);
  font-size: 12px;
  font-weight: 700;
  text-align: left;
}

.admin-table td {
  height: 66px;
  padding: 10px 16px;
  border-bottom: 1px solid oklch(91% 0.01 155);
  vertical-align: middle;
}

.admin-table tbody tr:last-child td {
  border-bottom: 0;
}

.admin-table tbody tr:hover {
  background: oklch(98% 0.008 155);
}

.admin-identity {
  display: flex;
  align-items: center;
  gap: 11px;
}

.admin-avatar {
  display: grid;
  flex: none;
  place-items: center;
  width: 34px;
  height: 34px;
  color: oklch(36% 0.07 165);
  background: oklch(91% 0.045 158);
  border-radius: 6px;
  font-size: 13px;
  font-weight: 800;
}

.admin-identity div {
  display: grid;
  gap: 4px;
}

.admin-identity strong {
  color: oklch(29% 0.018 165);
  font-size: 13px;
}

.admin-identity span {
  color: oklch(58% 0.014 165);
  font-size: 11px;
  font-style: normal;
}

.admin-identity em {
  margin-left: 6px;
  padding: 2px 5px;
  color: oklch(39% 0.075 165);
  background: oklch(92% 0.04 158);
  border-radius: 4px;
  font-size: 10px;
  font-style: normal;
  font-weight: 700;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 8px;
  border-radius: 5px;
  font-size: 11px;
  font-weight: 700;
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
  color: oklch(48% 0.018 165);
  background: oklch(93% 0.012 160);
}

.status-badge.disabled::before {
  background: oklch(64% 0.018 165);
}

.date-cell {
  color: oklch(48% 0.016 165);
  white-space: nowrap;
}

.operation-cell {
  width: 180px;
  text-align: right;
}

.row-actions,
.delete-confirm {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
}

.row-actions button,
.delete-confirm button {
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
}

.row-actions button:hover:not(:disabled) {
  color: oklch(34% 0.07 165);
  background: oklch(93% 0.03 158);
}

.row-actions button svg {
  width: 14px;
  height: 14px;
}

.row-actions .delete-action {
  color: oklch(47% 0.09 28);
}

.row-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.38;
}

.delete-confirm span {
  color: oklch(43% 0.06 28);
  font-size: 11px;
}

.delete-confirm button {
  border-color: oklch(82% 0.025 160);
}

.delete-confirm .confirm-delete {
  color: oklch(98% 0.004 25);
  background: oklch(48% 0.13 28);
  border-color: oklch(48% 0.13 28);
}

.delete-confirm button:disabled {
  cursor: wait;
  opacity: 0.65;
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

.admin-editor {
  padding: 22px;
}

.admin-editor > header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid oklch(89% 0.012 155);
}

.admin-editor header p,
.admin-editor header h3 {
  margin: 0;
}

.admin-editor header p {
  margin-bottom: 5px;
  color: oklch(49% 0.07 165);
  font-size: 10px;
  font-weight: 800;
}

.admin-editor header h3 {
  color: oklch(27% 0.018 165);
  font-size: 17px;
}

.admin-editor header button {
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

.admin-editor header button:hover {
  background: oklch(93% 0.018 158);
}

.admin-editor header button:disabled {
  cursor: wait;
  opacity: 0.5;
}

.admin-editor header svg {
  width: 17px;
  height: 17px;
}

.admin-editor form {
  display: grid;
  padding-top: 18px;
}

.admin-editor label {
  margin: 18px 0 8px;
  color: oklch(35% 0.018 165);
  font-size: 12px;
  font-weight: 700;
}

.admin-editor label:first-child {
  margin-top: 0;
}

.admin-editor label span {
  color: oklch(57% 0.014 165);
  font-weight: 400;
}

.admin-editor input,
.admin-editor select {
  width: 100%;
  height: 42px;
  padding: 0 11px;
  color: oklch(27% 0.018 165);
  background: oklch(99% 0.003 150);
  border: 1px solid oklch(82% 0.018 158);
  border-radius: 6px;
  outline: none;
  font: inherit;
  font-size: 13px;
}

.admin-editor input:focus,
.admin-editor select:focus {
  border-color: oklch(51% 0.085 165);
  box-shadow: 0 0 0 3px oklch(74% 0.08 160 / 0.16);
}

.admin-editor input:disabled,
.admin-editor select:disabled {
  cursor: wait;
  opacity: 0.68;
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

@media (max-width: 720px) {
  .admins-page {
    padding: 24px 16px 36px;
  }

  .section-heading {
    align-items: stretch;
    flex-direction: column;
    gap: 16px;
  }

  .section-actions {
    display: grid;
    grid-template-columns: 1fr 1.4fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .section-actions button,
  .skeleton-row span {
    animation: none;
    transition: none;
  }
}
</style>
