import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  build: {
    emptyOutDir: true,
    outDir: "../../cmd/remote/_client/build",
  },
  plugins: [react(), tsconfigPaths()],
});
