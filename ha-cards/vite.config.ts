import { resolve } from 'node:path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      input: {
        app: resolve(__dirname, 'index.html'),
        'ajaxbridge-lovelace': resolve(__dirname, 'src/ha/register.tsx'),
      },
      output: {
        entryFileNames: (chunkInfo) =>
          chunkInfo.name === 'ajaxbridge-lovelace' ? 'ajaxbridge-lovelace.js' : 'assets/[name].js',
        chunkFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash][extname]',
      },
    },
  },
  server: { port: 5173 },
});
