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
      path: '/sign-in',
      name: 'sign-in',
      component: () => import('../views/SignInView.vue'),
    },
    {
      path: '/sign-up',
      name: 'sign-up',
      component: () => import('../views/SignUpView.vue'),
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/UserDashboard.vue'),
    },
    {
      path: '/features',
      name: 'features',
      component: () => import('../views/FeaturesView.vue'),
    },
    {
      path: '/clusters/:id',
      name: 'manage-cluster',
      component: () => import('../views/ManageClusterView.vue'),
      props: true
    },
    {
      path: '/deploy',
      name: 'deploy-cluster',
      component: () => import('../views/DeployClusterView.vue'),
    },
    {
      path: '/reserve',
      name: 'reserve',
      component: () => import('../views/ReserveView.vue'),
    },
    {
      path: '/pricing',
      name: 'pricing',
      component: () => import('../views/PricingView.vue'),
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
    },
  ],
})

// router.beforeEach((to, from, next) => {
//   if (to.path === '/admin') {
//     // Use Pinia store to check admin status
//     const userStore = useUserStore()
//     if (!userStore.isLoggedIn) {
//       // Not logged in, redirect to sign-in
//       return next('/sign-in')
//     }
//     if (!userStore.isAdmin) {
//       // Not an admin, redirect to dashboard
//       return next('/dashboard')
//     }
//   }
//   next()
// })

export default router
