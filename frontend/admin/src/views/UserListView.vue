<script setup lang="ts">
import axios from 'axios'
import { computed, onMounted, ref } from 'vue'
import { Refresh, Search } from '@element-plus/icons-vue'
import { listUsers, type PlatformUser } from '../api/admin'

const pageSizeOptions = [20, 50, 100]

const users = ref<PlatformUser[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const searchInput = ref('')
const appliedSearch = ref('')
const status = ref<'' | 'active' | 'disabled'>('')
const loading = ref(true)
const refreshing = ref(false)
const errorMessage = ref('')
let requestSequence = 0

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const resultStart = computed(() => total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1)
const resultEnd = computed(() => Math.min(page.value * pageSize.value, total.value))
const hasFilters = computed(() => appliedSearch.value !== '' || status.value !== '')

onMounted(() => loadUsers(false))

async function loadUsers(isRefresh: boolean) {
  const sequence = ++requestSequence
  if (isRefresh) {
    refreshing.value = true
  } else {
    loading.value = true
  }
  errorMessage.value = ''

  try {
    const result = await listUsers({
      page: page.value,
      pageSize: pageSize.value,
      ...(appliedSearch.value ? { search: appliedSearch.value } : {}),
      ...(status.value ? { status: status.value } : {}),
    })
    if (sequence !== requestSequence) return

    users.value = result.users
    total.value = result.total
    page.value = result.page
    pageSize.value = result.pageSize
  } catch (error) {
    if (sequence !== requestSequence) return
    errorMessage.value = errorText(error, '用户列表加载失败，请稍后重试。')
  } finally {
    if (sequence === requestSequence) {
      loading.value = false
      refreshing.value = false
    }
  }
}

function applyFilters() {
  appliedSearch.value = searchInput.value.trim()
  page.value = 1
  loadUsers(false)
}

function clearFilters() {
  searchInput.value = ''
  appliedSearch.value = ''
  status.value = ''
  page.value = 1
  loadUsers(false)
}

function changeStatus() {
  page.value = 1
  loadUsers(false)
}

function changePageSize() {
  page.value = 1
  loadUsers(false)
}

function goToPage(nextPage: number) {
  if (nextPage < 1 || nextPage > totalPages.value || nextPage === page.value) return
  page.value = nextPage
  loadUsers(false)
}

function formatTime(timestamp?: number, emptyText = '从未登录') {
  if (!timestamp) return emptyText
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
  <section class="users-page">
    <header class="section-heading">
      <div>
        <h2>平台用户</h2>
        <p>查看已注册用户的联系方式、账号状态与登录时间。</p>
      </div>
      <button
        class="refresh-action"
        type="button"
        :disabled="refreshing"
        @click="loadUsers(true)"
      >
        <Refresh aria-hidden="true" />
        {{ refreshing ? '刷新中' : '刷新' }}
      </button>
    </header>

    <form class="filter-bar" role="search" @submit.prevent="applyFilters">
      <div class="search-field">
        <Search aria-hidden="true" />
        <label class="sr-only" for="user-search">搜索用户</label>
        <input
          id="user-search"
          v-model="searchInput"
          type="search"
          maxlength="100"
          autocomplete="off"
          placeholder="搜索用户名、邮箱或手机号"
        >
      </div>

      <label class="status-filter">
        <span>账号状态</span>
        <select v-model="status" @change="changeStatus">
          <option value="">全部状态</option>
          <option value="active">启用</option>
          <option value="disabled">停用</option>
        </select>
      </label>

      <button class="search-action" type="submit">查询</button>
      <button
        v-if="hasFilters || searchInput"
        class="clear-action"
        type="button"
        @click="clearFilters"
      >
        清除筛选
      </button>
    </form>

    <p v-if="errorMessage" class="page-message error-message" role="alert">
      {{ errorMessage }}
      <button type="button" @click="loadUsers(false)">重新加载</button>
    </p>

    <div class="table-panel">
      <div class="table-summary">
        <p>
          共 <strong>{{ total }}</strong> 位用户
          <span v-if="hasFilters">，当前为筛选结果</span>
        </p>
        <span v-if="total > 0">显示 {{ resultStart }}–{{ resultEnd }} 条</span>
      </div>

      <div class="table-scroll">
        <table class="users-table">
          <caption class="sr-only">平台用户列表</caption>
          <thead>
            <tr>
              <th scope="col">用户</th>
              <th scope="col">联系方式</th>
              <th scope="col">账号状态</th>
              <th scope="col">最后登录</th>
              <th scope="col">注册时间</th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr v-for="index in 6" :key="index" class="skeleton-row">
              <td colspan="5"><span /></td>
            </tr>
          </tbody>
          <tbody v-else-if="users.length === 0">
            <tr>
              <td colspan="5" class="empty-table">
                <strong>{{ hasFilters ? '没有符合条件的用户' : '还没有注册用户' }}</strong>
                <span>
                  {{ hasFilters ? '调整搜索内容或状态筛选后再试。' : '新用户注册后会显示在这里。' }}
                </span>
                <button v-if="hasFilters" type="button" @click="clearFilters">清除筛选</button>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr v-for="user in users" :key="user.id">
              <td>
                <div class="user-identity">
                  <span class="user-avatar">{{ user.username.slice(0, 1).toUpperCase() }}</span>
                  <div>
                    <strong>{{ user.username }}</strong>
                    <span>ID {{ user.id }}</span>
                  </div>
                </div>
              </td>
              <td>
                <div class="contact-details">
                  <span :class="{ muted: !user.email }">{{ user.email || '未填写邮箱' }}</span>
                  <span :class="{ muted: !user.phone }">{{ user.phone || '未填写手机号' }}</span>
                </div>
              </td>
              <td>
                <span class="status-badge" :class="user.status">
                  {{ user.status === 'active' ? '启用' : '停用' }}
                </span>
              </td>
              <td class="date-cell">{{ formatTime(user.lastLoginAt) }}</td>
              <td class="date-cell">{{ formatTime(user.createdAt, '未知') }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <footer class="pagination">
        <label>
          <span>每页</span>
          <select v-model="pageSize" :disabled="loading" @change="changePageSize">
            <option v-for="size in pageSizeOptions" :key="size" :value="size">
              {{ size }} 条
            </option>
          </select>
        </label>

        <div class="page-controls" aria-label="用户列表分页">
          <button
            type="button"
            :disabled="loading || page <= 1"
            @click="goToPage(page - 1)"
          >
            上一页
          </button>
          <span>第 <strong>{{ page }}</strong> / {{ totalPages }} 页</span>
          <button
            type="button"
            :disabled="loading || page >= totalPages"
            @click="goToPage(page + 1)"
          >
            下一页
          </button>
        </div>
      </footer>
    </div>
  </section>
</template>

<style scoped>
.users-page {
  min-width: 980px;
  padding: 34px clamp(24px, 4vw, 48px) 48px;
}

.section-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 26px;
}

.section-heading h2,
.section-heading p {
  margin: 0;
}

.section-heading h2 {
  color: oklch(27% 0.018 165);
  font-size: 22px;
  line-height: 1.2;
}

.section-heading p {
  margin-top: 8px;
  color: oklch(50% 0.018 165);
  font-size: 13px;
}

.refresh-action,
.search-action,
.clear-action,
.page-controls button,
.empty-table button,
.error-message button {
  min-height: 36px;
  border-radius: 6px;
  cursor: pointer;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  transition:
    color 180ms ease-out,
    background 180ms ease-out,
    border-color 180ms ease-out;
}

.refresh-action {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 0 13px;
  color: oklch(38% 0.038 165);
  background: oklch(99% 0.004 150);
  border: 1px solid oklch(82% 0.018 158);
}

.refresh-action svg {
  width: 15px;
  height: 15px;
}

.refresh-action:hover:not(:disabled),
.page-controls button:hover:not(:disabled) {
  color: oklch(32% 0.07 165);
  background: oklch(94% 0.025 158);
  border-color: oklch(70% 0.04 160);
}

.filter-bar {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  padding: 18px;
  background: oklch(99% 0.004 150);
  border: 1px solid oklch(87% 0.014 158);
  border-bottom: 0;
  border-radius: 8px 8px 0 0;
}

.search-field {
  position: relative;
  flex: 1 1 420px;
  max-width: 520px;
}

.search-field svg {
  position: absolute;
  top: 50%;
  left: 13px;
  width: 16px;
  height: 16px;
  color: oklch(56% 0.018 165);
  pointer-events: none;
  transform: translateY(-50%);
}

.search-field input,
.status-filter select,
.pagination select {
  height: 38px;
  color: oklch(29% 0.02 165);
  background: oklch(99% 0.004 150);
  border: 1px solid oklch(80% 0.018 158);
  border-radius: 6px;
  font: inherit;
  font-size: 13px;
}

.search-field input {
  width: 100%;
  padding: 0 13px 0 39px;
}

.search-field input::placeholder {
  color: oklch(61% 0.014 165);
}

.status-filter {
  display: grid;
  gap: 6px;
  color: oklch(45% 0.02 165);
  font-size: 12px;
  font-weight: 700;
}

.status-filter select {
  width: 132px;
  padding: 0 32px 0 11px;
}

.search-field input:focus,
.status-filter select:focus,
.pagination select:focus {
  border-color: oklch(50% 0.09 165);
  outline: 2px solid oklch(78% 0.07 158 / 0.42);
  outline-offset: 1px;
}

.search-action {
  padding: 0 20px;
  color: oklch(98% 0.004 150);
  background: oklch(39% 0.075 165);
  border: 1px solid oklch(39% 0.075 165);
}

.search-action:hover {
  background: oklch(34% 0.075 165);
  border-color: oklch(34% 0.075 165);
}

.clear-action,
.empty-table button,
.error-message button {
  padding: 0 8px;
  color: oklch(42% 0.055 165);
  background: transparent;
  border: 0;
}

.clear-action:hover,
.empty-table button:hover,
.error-message button:hover {
  color: oklch(31% 0.075 165);
  text-decoration: underline;
}

.refresh-action:focus-visible,
.search-action:focus-visible,
.clear-action:focus-visible,
.page-controls button:focus-visible,
.empty-table button:focus-visible,
.error-message button:focus-visible {
  outline: 2px solid oklch(50% 0.09 165);
  outline-offset: 2px;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.page-message {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 12px 0 0;
  padding: 10px 13px;
  border-radius: 6px;
  font-size: 13px;
}

.error-message {
  color: oklch(43% 0.13 28);
  background: oklch(94% 0.035 28);
  border: 1px solid oklch(80% 0.06 28);
}

.table-panel {
  overflow: hidden;
  background: oklch(99% 0.004 150);
  border: 1px solid oklch(87% 0.014 158);
  border-radius: 0 0 8px 8px;
}

.table-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 48px;
  padding: 0 18px;
  color: oklch(53% 0.018 165);
  border-bottom: 1px solid oklch(89% 0.012 158);
  font-size: 12px;
}

.table-summary p {
  margin: 0;
}

.table-summary strong {
  color: oklch(30% 0.03 165);
  font-size: 14px;
}

.table-scroll {
  overflow-x: auto;
}

.users-table {
  width: 100%;
  min-width: 920px;
  border-collapse: collapse;
  table-layout: fixed;
}

.users-table th {
  height: 43px;
  padding: 0 18px;
  color: oklch(50% 0.018 165);
  background: oklch(96.5% 0.009 155);
  border-bottom: 1px solid oklch(87% 0.014 158);
  font-size: 12px;
  font-weight: 700;
  text-align: left;
}

.users-table th:nth-child(1) {
  width: 23%;
}

.users-table th:nth-child(2) {
  width: 29%;
}

.users-table th:nth-child(3) {
  width: 13%;
}

.users-table th:nth-child(4),
.users-table th:nth-child(5) {
  width: 17.5%;
}

.users-table td {
  height: 68px;
  padding: 10px 18px;
  color: oklch(35% 0.02 165);
  border-bottom: 1px solid oklch(91% 0.01 158);
  font-size: 13px;
  vertical-align: middle;
}

.users-table tbody tr:last-child td {
  border-bottom: 0;
}

.users-table tbody tr:not(.skeleton-row):hover td {
  background: oklch(97% 0.012 155);
}

.user-identity {
  display: flex;
  align-items: center;
  gap: 11px;
  min-width: 0;
}

.user-avatar {
  display: grid;
  flex: none;
  place-items: center;
  width: 34px;
  height: 34px;
  color: oklch(34% 0.065 165);
  background: oklch(91% 0.045 158);
  border-radius: 6px;
  font-size: 14px;
  font-weight: 800;
}

.user-identity div,
.contact-details {
  display: grid;
  min-width: 0;
  gap: 4px;
}

.user-identity strong,
.contact-details span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-identity strong {
  color: oklch(27% 0.026 165);
}

.user-identity span,
.contact-details span:last-child {
  color: oklch(56% 0.016 165);
  font-size: 12px;
}

.contact-details .muted {
  color: oklch(65% 0.01 165);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 25px;
  padding: 0 9px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.status-badge::before {
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 50%;
  content: "";
}

.status-badge.active {
  color: oklch(39% 0.095 155);
  background: oklch(93% 0.04 155);
}

.status-badge.disabled {
  color: oklch(48% 0.025 165);
  background: oklch(92% 0.012 165);
}

.date-cell {
  color: oklch(47% 0.014 165);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.skeleton-row td {
  height: 68px;
}

.skeleton-row span {
  display: block;
  width: 100%;
  height: 22px;
  background: oklch(93% 0.012 158);
  border-radius: 4px;
  animation: skeleton-pulse 1.2s ease-in-out infinite alternate;
}

.empty-table {
  height: 260px !important;
  color: oklch(53% 0.018 165) !important;
  text-align: center;
}

.empty-table strong,
.empty-table span {
  display: block;
}

.empty-table strong {
  margin-bottom: 8px;
  color: oklch(31% 0.024 165);
  font-size: 15px;
}

.empty-table span {
  margin-bottom: 8px;
  font-size: 13px;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 60px;
  padding: 0 18px;
  border-top: 1px solid oklch(89% 0.012 158);
}

.pagination label,
.page-controls {
  display: flex;
  align-items: center;
  gap: 9px;
  color: oklch(51% 0.018 165);
  font-size: 12px;
}

.pagination select {
  width: 82px;
  height: 34px;
  padding: 0 8px;
}

.page-controls button {
  min-width: 68px;
  padding: 0 12px;
  color: oklch(40% 0.03 165);
  background: oklch(99% 0.004 150);
  border: 1px solid oklch(82% 0.018 158);
}

.page-controls strong {
  color: oklch(30% 0.03 165);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@keyframes skeleton-pulse {
  from {
    opacity: 0.55;
  }

  to {
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    scroll-behavior: auto !important;
    transition-duration: 0.01ms !important;
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
  }
}
</style>
