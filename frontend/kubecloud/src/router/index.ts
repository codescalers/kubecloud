import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import { useUserStore } from '../stores/user'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: () => import('../views/ResetPasswordView.vue'),
    },
    {
      path: '/sign-in',
      name: 'sign-in',
      component: () => import('../views/SignInView.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/sign-up',
      name: 'sign-up',
      component: () => import('../views/SignUpView.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/UserDashboard.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/features',
      name: 'features',
      component: () => import('../views/FeaturesView.vue'),
    },
    {
      path: '/clusters/:id',
      name: 'manage-cluster',
      component: () => import('../components/dashboard/ManageClusterView.vue'),
      props: true,
      meta: { requiresAuth: true }
    },
    {
      path: '/deploy',
      name: 'deploy-cluster',
      component: () => import('../views/DeployClusterView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/nodes',
      name: 'reserve',
      component: () => import('../views/ReserveView.vue'),
    },
    {
      path: '/use-cases',
      name: 'usecases',
      component: () => import('../views/UseCasesView.vue'),
    },
    {
      path: '/docs',
      name: 'docs',
      component: () => import('../views/DocsView.vue'),
    },
    {
      path: '/admin',
      name: 'admin-dashboard',
      component: () => import('../views/AdminDashboard.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/register/verify',
      name: 'register-verify',
      component: () => import('../views/RegisterVerifyView.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('../views/ForgotPasswordView.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/pending-requests',
      name: 'pending-requests',
      component: () => import('../views/PendingRecordsView.vue')
    },
  ],
  scrollBehavior() {
    // Always scroll to top on route change
    return { top: 0 }
  }
})

router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()

  // Check if user is authenticated
  const isAuthenticated = userStore.isLoggedIn

  // Note: Password reset cleanup is handled in the components themselves
  // to avoid interfering with multi-tab usage

  // If route requires guest (sign-in, sign-up) and user is authenticated
  if (to.meta.requiresGuest && isAuthenticated) {
    // Redirect authenticated users to home
    return next('/')
  }

  // If route requires authentication and user is not authenticated
  if (to.meta.requiresAuth && !isAuthenticated) {
    // Redirect to sign-in
    return next('/sign-in')
  }

  // If route requires admin access and user is not admin
  if (to.meta.requiresAdmin && !userStore.isAdmin) {
    // Not an admin, redirect to dashboard
    return next('/dashboard')
  }  // For public routes (home, features, etc.), allow access to everyone
  next()
})

export default router
