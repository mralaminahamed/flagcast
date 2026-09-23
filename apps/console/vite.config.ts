import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

const target = process.env.VITE_API_PROXY || "http://localhost:8201";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: true,
    port: 8200,
    proxy: {
      "/api": { target, changeOrigin: true },
    },
  },
});
