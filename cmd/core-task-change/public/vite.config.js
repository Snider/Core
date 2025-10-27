import { defineConfig } from 'vite';
import tailwindcss from 'tailwindcss';
import autoprefixer from 'autoprefixer';

// https://vitejs.dev/config/
export default defineConfig({
  // Set the root of the project to the 'html' directory.
  // Vite will look for index.html in this folder.
  root: 'html',

  css: {
    postcss: {
      plugins: [
        tailwindcss('./tailwind.config.js'), // Specify the path to your tailwind.config.js
        autoprefixer(),
      ],
    },
  },

  build: {
    // Set the output directory for the build.
    // We need to go up one level from 'html' to place the 'dist' folder
    // in the correct location for Wails.
    outDir: '../dist',
    // Ensure the output directory is emptied before each build.
    emptyOutDir: true,
    rollupOptions: {
      input: {
        main: 'html/index.html',
        // Add a CSS entry point for Tailwind
        'main.css': 'assets/main.css', 
      },
      output: {
        assetFileNames: (assetInfo) => {
          if (assetInfo.name === 'main.css') return 'assets/main.css';
          return assetInfo.name;
        },
      },
    },
  },
});
