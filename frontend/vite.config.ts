import path from "node:path";

import solid from "vite-plugin-solid";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [solid()],
  base: ".",
  resolve: {
    alias: {
      "~": path.resolve(__dirname, "./src"),
    },
  },
});
