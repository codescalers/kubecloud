import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

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
      path: '/usecases',
      name: 'usecases',
      component: () => import('../views/UseCasesView.vue'),
    },
    {
      path: '/docs',
      name: 'docs',
      component: () => import('../views/DocsView.vue'),
    },
  ],
})

export default router
