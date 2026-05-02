import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// During `vite dev` the Go backend runs separately on :3000.
// Vite serves the SPA on :5173 and proxies /api to :3000.
// In production the Go binary embeds web/dist/ and serves both itself.
//
// GH Pages mode: when VITE_GH_PAGES_BASE is set (e.g. "/aegis/" by the
// deploy-pages workflow), build with that as the asset base path and
// expect to be served under that prefix. Empty / unset = root.
const ghPagesBase = process.env.VITE_GH_PAGES_BASE || ''

export default defineConfig({
  base: ghPagesBase || '/',
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:3000',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
