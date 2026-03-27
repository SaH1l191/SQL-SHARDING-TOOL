import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: true,
  },
  build: {
    // esbuild is the default and much faster than tsc for transpiling
    minify: 'esbuild',
    target: 'esnext',
    sourcemap: false,
  },
})
