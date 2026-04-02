import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: () => import('../views/LiveHallPage.vue'),
    },
    {
      path: '/live/:id',
      name: 'LiveView',
      component: () => import('../views/LiveViewPage.vue'),
    },
    {
      path: '/go-live',
      name: 'GoLive',
      component: () => import('../views/GoLivePage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/LoginPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('../views/RegisterPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/profile',
      name: 'Profile',
      component: () => import('../views/ProfilePage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/user/:id',
      name: 'PublicProfile',
      component: () => import('../views/ProfilePage.vue'),
    },
  ],
})

// 路由守衛
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('access_token')
  const isAuthenticated = !!token

  if (to.meta.requiresAuth && !isAuthenticated) {
    // 未登入用戶訪問需認證頁面 → 跳轉登入
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else if (to.meta.guest && isAuthenticated) {
    // 已登入用戶訪問登入/註冊頁 → 跳轉首頁
    next('/')
  } else {
    next()
  }
})

export default router
