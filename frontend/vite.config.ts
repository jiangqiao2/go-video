import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
      port: 3000,
      host: true,
      proxy: {
        // 用户服务API代理
        '/api/v1/open/users': {
          target: 'http://localhost:8081',
          changeOrigin: true,
          secure: false,
        },
        '/api/v1/inner/users': {
          target: 'http://localhost:8081',
          changeOrigin: true,
          secure: false,
        },
        // 上传服务API代理
        '/api/v1/inner/upload': {
          target: 'http://localhost:8082',
          changeOrigin: true,
          secure: false,
        },
        // 健康检查等其他API
        '/api': {
          target: 'http://localhost:8081',
          changeOrigin: true,
          secure: false,
        },
      },
    },
  build: {
    outDir: 'dist',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          antd: ['antd'],
        },
      },
    },
  },
})