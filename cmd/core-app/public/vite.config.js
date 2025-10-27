import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
  // Set the root of the project to the 'html' directory.
  // Vite will look for index.html in this folder.
  root: 'html',

  build: {
    // Set the output directory for the build.
    // We need to go up one level from 'html' to place the 'dist' folder
    // in the correct location for Wails.
    outDir: '../dist',
    // Ensure the output directory is emptied before each build.
    emptyOutDir: true,
  },
});
