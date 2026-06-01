import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import Components from "unplugin-vue-components/vite";
import { VantResolver } from "@vant/auto-import-resolver";

export default defineConfig({
  plugins: [
    vue(),
    Components({
      resolvers: [VantResolver()],
    }),
  ],
  server: {
    host: "0.0.0.0",
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:80",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://localhost:8888",
        ws: true,
      },
      "/videos": {
        target: "http://localhost:80",
        changeOrigin: true,
      },
    },
  },
});
