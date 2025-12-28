import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: true, 
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://backend:8000', // Docker service name
        changeOrigin: true,
        secure: false,
        // Optional: If backend expects /start-test but you call /api/start-test
        // rewrite: (path) => path.replace(/^\/api/, '') 
      },
      '/ws': {
        target: 'ws://backend:8000',
        ws: true,
        changeOrigin: true
      }
    }
  }
})
