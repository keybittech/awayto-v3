import path from 'path';
import fs from 'fs';
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import viteCompression from 'vite-plugin-compression';
import circleDeps from 'vite-plugin-circular-dependency';
import dotenv from 'dotenv';

dotenv.config({ path: './.env.local' });

const { VITE_AWAYTO_WEBAPP } = process.env;

const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = (relativePath: string) => path.resolve(appDirectory, relativePath);

const manualChunks = (_: string) => {
  return 'pkg';
}

export default defineConfig(async _ => {
  const { viteStaticCopy } = await import('vite-plugin-static-copy');
  return {
    base: '/app',
    server: {
      port: 3000
    },
    html: {
      cspNonce: 'VITE_NONCE'
    },
    esbuild: {
      define: {
        'println': 'console.log'
      }
    },
    build: {
      outDir: 'build',
      rollupOptions: {
        output: {
          manualChunks
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

