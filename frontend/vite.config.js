import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: "../backend/template/static/dist",
    emptyOutDir: true,
    manifest: true,
    rollupOptions: {
      input: {
        stockCounting: "./src/stock-taking/counting/main.js",
        stockDashboard: "./src/stock-taking/dashboard/main.js",
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"), // <-- this makes @/ point to src/
    },
  },
});
