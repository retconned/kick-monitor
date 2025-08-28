import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import path from "path";
import react from "@vitejs/plugin-react";
// https://vite.dev/config/

export default defineConfig({
    plugins: [react(), tailwindcss()],
    base: "./",
    build: {
        outDir: "dist", // Default, but ensures the build output goes here
    },
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
});
