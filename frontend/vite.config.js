import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: "../../backend/template/static/dist",
    emptyOutDir: true,
    manifest: true,
    rollupOptions: {
      input: {
        stockTake: "./src/stock-take/main.js"
      }
    }
  }
})
