import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const API_PORT = process.env.GHDASH_API_PORT ?? '19080'
const WEB_PORT = Number(process.env.GHDASH_WEB_PORT ?? '19090')

export default defineConfig({
  plugins: [vue()],
  server: {
    port: WEB_PORT,
    strictPort: true,
    proxy: {
      '/api': {
        target: `http://127.0.0.1:${API_PORT}`,
        changeOrigin: false,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
