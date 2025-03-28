import path from 'path';
import fs from 'fs';
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import viteCompression from 'vite-plugin-compression';
import circleDeps from 'vite-plugin-circular-dependency';
import { viteStaticCopy } from 'vite-plugin-static-copy'
import dotenv from 'dotenv';

dotenv.config({ path: './.env.local' });

const { VITE_AWAYTO_WEBAPP } = process.env;

const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = (relativePath: string) => path.resolve(appDirectory, relativePath);

const chunks = [
  ['data-grid', 'dg'],
  ['date-pickers', 'dp'],
  ['icons-material', 'im'],
  ['material', 'm'],
];

const manualChunks = (a: string) => {
  for (let i = 0; i < chunks.length; i++) {
    if (a.includes(chunks[i][0])) {
      return chunks[i][1];
    }
  }
  const mr = Math.ceil(Math.random() * 4);
  return 'x' + mr;
}

export default defineConfig(_ => {
  return {
    base: '/app',
    server: {
      port: 3000
    },
    html: {
      cspNonce: 'VITE_NONCE'
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

