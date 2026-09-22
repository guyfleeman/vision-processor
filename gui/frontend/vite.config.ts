import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    // The Go host serves /api and /ws on :8085; everything else (this dev
    // server) is same-origin so no CORS handling is needed on either side.
    proxy: {
      "/api": "http://localhost:8085",
      "/ws": { target: "ws://localhost:8085", ws: true },
    },
  },
});
