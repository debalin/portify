import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '127.0.0.1',
    port: 5175,
    strictPort: true,
    proxy: {
      '/converter.v1.ConverterService': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
