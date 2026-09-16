import { defineConfig } from "vite";
import { resolve } from "path";

export default defineConfig({
  root: "./",
  build: {
    // Output directly into Go's embedded folder
    outDir: resolve(__dirname, "../web-static"),
    emptyOutDir: true, // Cleans web/ before building fresh assets
    rollupOptions: {
      input: {
        main: resolve(__dirname, "index.html"),
      },
    },
  },
  server: {
    // Proxy local Go server API calls during development (pnpm dev)
    proxy: {
      "/api": "http://localhost:8081",
      "/media": "http://localhost:8081",
    },
  },
});
