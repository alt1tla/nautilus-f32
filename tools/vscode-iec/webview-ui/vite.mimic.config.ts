import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

// Second webview bundle: the mimic editor. Separate from fbd-flow so the
// P&ID editor doesn't carry Svelte Flow (and vice versa). Built AFTER the
// fbd bundle (which owns emptyOutDir) — see the package build script.
export default defineConfig({
  plugins: [svelte()],
  define: { "process.env.NODE_ENV": JSON.stringify("production") },
  build: {
    outDir: "../media/dist",
    emptyOutDir: false,
    lib: { entry: "src/mimic/main.ts", formats: ["iife"], name: "MimicEditor", fileName: () => "mimic-editor.js" },
    rollupOptions: { output: { assetFileNames: "mimic-editor.[ext]" } },
  },
});
