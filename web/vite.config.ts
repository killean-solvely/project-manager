import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // Proxy API calls to the Go backend so there's no CORS in dev.
    proxy: {
      '/api': 'http://localhost:4523',
    },
  },
})
