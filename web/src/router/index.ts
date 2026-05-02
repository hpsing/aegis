import { createRouter, createWebHashHistory } from 'vue-router'

// Hash history: routes live at `/#/jobs/14` instead of `/jobs/14`.
// This is required for GitHub Pages (no SPA fallback for unknown
// paths → hard refresh of `/jobs/14` 404s) and is harmless for the
// embedded build served from the Go binary. Keep it for both.
const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'live', component: () => import('@/views/LiveDemo.vue') },
    { path: '/jobs/:id', name: 'job', component: () => import('@/views/JobInspector.vue'), props: true },
    { path: '/verifiers', name: 'verifiers', component: () => import('@/views/VerifierRoster.vue') },
    { path: '/system', name: 'system', component: () => import('@/views/SystemHealth.vue') },
  ],
})

export default router
