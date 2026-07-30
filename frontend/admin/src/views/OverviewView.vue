<script setup lang="ts">
import {
  Clock,
  Document,
  Goods,
  Key,
  Setting,
  User,
} from '@element-plus/icons-vue'
import { ElIcon } from 'element-plus'
import 'element-plus/es/components/icon/style/css'

const navigation = [
  { label: '概览', icon: Document, active: true },
  { label: '用户', icon: User },
  { label: '工具', icon: Goods },
  { label: '订单', icon: Clock },
  { label: '授权', icon: Key },
  { label: '审计日志', icon: Setting },
]
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
        <button
          v-for="item in navigation"
          :key="item.label"
          class="nav-item"
          :class="{ active: item.active }"
          :title="item.label"
          type="button"
        >
          <el-icon :size="18"><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <main class="workspace">
      <header class="topbar">
        <div>
          <p>平台管理</p>
          <h1>概览</h1>
        </div>
        <span class="environment">生产环境</span>
      </header>

      <section class="empty-state" aria-labelledby="empty-title">
        <div class="empty-icon"><el-icon :size="28"><Setting /></el-icon></div>
        <h2 id="empty-title">管理接口尚未接入</h2>
        <p>管理员认证和权限接口完成后，这里将显示平台的真实运营数据。</p>
      </section>
    </main>
  </div>
</template>
