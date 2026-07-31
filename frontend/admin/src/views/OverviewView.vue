<script setup lang="ts">
import { Setting } from '@element-plus/icons-vue'
import { ElIcon } from 'element-plus'
import 'element-plus/es/components/icon/style/css'
</script>

<template>
  <section class="empty-state" aria-labelledby="empty-title">
    <div class="empty-icon"><el-icon :size="28"><Setting /></el-icon></div>
    <h2 id="empty-title">运营模块正在接入</h2>
    <p>用户、管理员与工具目录管理已经启用，订单和授权数据将在对应模块中逐步开放。</p>
  </section>
</template>
