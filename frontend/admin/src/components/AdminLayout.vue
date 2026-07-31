<script setup lang="ts">
import {
  Avatar,
  Clock,
  Document,
  Goods,
  Key,
  Setting,
  SwitchButton,
  User,
} from '@element-plus/icons-vue'
import { ElIcon } from 'element-plus'
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useAdminAuthStore } from '../stores/adminAuth'
import 'element-plus/es/components/icon/style/css'

const route = useRoute()
const router = useRouter()
const auth = useAdminAuthStore()

const navigation = [
  { label: '概览', icon: Document, to: '/' },
  { label: '用户', icon: User, to: '/users' },
  { label: '管理员', icon: Avatar, to: '/admins' },
  { label: '工具', icon: Goods, to: '/tools' },
  { label: '订单', icon: Clock },
  { label: '授权', icon: Key },
  { label: '审计日志', icon: Setting },
]

const pageTitle = computed(() => String(route.meta.title || '管理后台'))

async function logout() {
  auth.logout()
  await router.replace('/login')
}
</script>

<template>
  <div class="admin-shell">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-mark">A</span>
        <div>
          <strong>AutomaticTools</strong>
          <span>管理后台</span>
        </div>
      </div>

      <nav aria-label="管理导航">
        <template v-for="item in navigation" :key="item.label">
          <RouterLink
            v-if="item.to"
            class="nav-item nav-link"
            :to="item.to"
            :title="item.label"
          >
            <el-icon :size="18"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </RouterLink>
          <button
            v-else
            class="nav-item"
            :title="`${item.label}功能正在接入`"
            type="button"
            aria-disabled="true"
          >
            <el-icon :size="18"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </button>
        </template>
      </nav>
    </aside>

    <main class="workspace">
      <header class="topbar">
        <div>
          <p>平台管理</p>
          <h1>{{ pageTitle }}</h1>
        </div>
        <div class="topbar-actions">
          <span class="environment">生产环境</span>
          <span class="admin-account">{{ auth.admin?.username }}</span>
          <button class="logout-button" type="button" @click="logout">
            <el-icon :size="16"><SwitchButton /></el-icon>
            <span>退出登录</span>
          </button>
        </div>
      </header>

      <RouterView />
    </main>
  </div>
</template>
