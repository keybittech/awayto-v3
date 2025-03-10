import path from 'path';
import fs from 'fs';
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import viteCompression from 'vite-plugin-compression';
import circleDeps from 'vite-plugin-circular-dependency';
import cssInjected from 'vite-plugin-css-injected-by-js';
import { viteStaticCopy } from 'vite-plugin-static-copy'
import dotenv from 'dotenv';

dotenv.config({ path: './.env.local' });

const { VITE_AWAYTO_WEBAPP } = process.env;

const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = (relativePath: string) => path.resolve(appDirectory, relativePath);

export default defineConfig(_ => {
  return {
    base: '/app',
    server: {
      port: 3000
    },
    build: {
      outDir: 'build',
      rollupOptions: {
        output: {
          manualChunks: a => a.includes('mui') ? 'mui' : a.includes('node_modules') ? 'pkg' : 'y',
        }
      }
    },
    resolve: {
      alias: {
        '@mui/material/Grid': path.resolve(appDirectory, './node_modules/@mui/material/Grid2'),
        'awayto/hooks': resolveApp('.' + VITE_AWAYTO_WEBAPP + '/hooks/index.ts'),
      }
    },
    plugins: [
      react(),
      cssInjected(),
      circleDeps({
        ignoreDynamicImport: true
      }),
      viteStaticCopy({
        targets: [
          {
            src: 'node_modules/pdfjs-dist/build/pdf.worker.min.mjs',
            dest: ''
          }
        ]
      }),
      viteCompression({
        algorithm: 'gzip',
        filter: /\.(js|mjs|json|css|html)$/i,
        threshold: 1024,
      })
    ]
  };
});

