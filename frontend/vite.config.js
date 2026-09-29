import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";

// https://vite.dev/config/
export default defineConfig({
  base: "/static/dist/",
  plugins: [vue()],
  build: {
    outDir: "../backend/template/static/dist",
    emptyOutDir: true,
    manifest: true,
    rollupOptions: {
      input: {
        next: "./src/next/main.js",
        login: "./src/login/main.js",
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"), // <-- this makes @/ point to src/
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/sales": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/inventory": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/admin": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/stock-taking": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
    },
  },
});
