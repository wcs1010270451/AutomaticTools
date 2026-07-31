import { createRouter, createWebHistory } from 'vue-router'
import { useAdminAuthStore } from '../stores/adminAuth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/',
      component: () => import('../components/AdminLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'overview',
          component: () => import('../views/OverviewView.vue'),
          meta: { title: '概览' },
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('../views/UserListView.vue'),
          meta: { title: '用户' },
        },
        {
          path: 'admins',
          name: 'admins',
          component: () => import('../views/AdminListView.vue'),
          meta: { title: '管理员' },
        },
        {
          path: 'tools',
          name: 'tools',
          component: () => import('../views/ToolListView.vue'),
          meta: { title: '工具' },
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAdminAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }
  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'overview' }
  }
})

export default router
