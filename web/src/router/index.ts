import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'live', component: () => import('@/views/LiveDemo.vue') },
    { path: '/jobs/:id', name: 'job', component: () => import('@/views/JobInspector.vue'), props: true },
    { path: '/verifiers', name: 'verifiers', component: () => import('@/views/VerifierRoster.vue') },
    { path: '/system', name: 'system', component: () => import('@/views/SystemHealth.vue') },
  ],
})

export default router
