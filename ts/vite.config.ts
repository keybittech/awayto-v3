import path from 'path';
import fs from 'fs';
import crypto from 'crypto';
import { sync } from 'glob';
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import viteCompression from 'vite-plugin-compression';
import copy from 'rollup-plugin-copy';
import circular from 'circular-dependency-plugin';
import { fileURLToPath } from 'url';
import dotenv from 'dotenv';

dotenv.config({ path: './.env.local' });

const { VITE_AWAYTO_WEBAPP, VITE_AWAYTO_WEBAPP_MODULES } = process.env;

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * 
 * @param {string} n A path name returned from glob.sync
 * @returns An object like `{ 'MyComponent': 'common/views/MyComponent' }`
 */
const buildPathObject = (n: string[]) => ({ [`${n[n.length - 1].split('.')[0]}`]: `${n[n.length - 3]}/${n[n.length - 2]}/${n[n.length - 1].split('.')[0]}` })

const appBuildOutputPath = path.resolve(__dirname + VITE_AWAYTO_WEBAPP + '/build.json');
const appRolesOutputPath = path.resolve(__dirname + VITE_AWAYTO_WEBAPP + '/roles.json');

try {
  if (!fs.existsSync(appBuildOutputPath)) fs.closeSync(fs.openSync(appBuildOutputPath, 'w'));
  if (!fs.existsSync(appRolesOutputPath)) fs.closeSync(fs.openSync(appRolesOutputPath, 'w'));
} catch (error) { }

/**
 * 
 * @param {string} path A file path to a set of globbable files
 * @returns An object containing file names as keys and values as file paths
 * ```
 * {
 *   "views": {
 *     "Home": "common/views/Home",
 *     "Login": "common/views/Login",
 *     "Secure": "common/views/Secure",
 *   },
 *   "reducers": {
 *     "login": "common/reducers/login",
 *     "util": "common/reducers/util",
 *   }
 * }
 * ```
 */
function parseResource(path: string) {
  return sync(path).map((m) => buildPathObject(m.split('/'))).reduce((a, b) => ({ ...a, ...b }), {});
}

/**
 * <p>We keep a reference to the old hash of files</p>.
 */
let oldAppOutputHash: string;
let oldRolesOutputHash: string;

/**
 * <p>This function runs on build and when webpack dev server receives a request.</p>
 * <p>Scan the file system for views and reducers and parse them into something we can use in the app.</p>
 * <p>Check against a hash of existing file structure to see if we need to update the build file. The build file is used later in the app to load the views and reducers.</p>
 * 
 */
function checkWriteBuildFile() {
  try {
    // all files are placed into this object as views
    const files = {
      views: parseResource('.' + VITE_AWAYTO_WEBAPP_MODULES + '/**/*.tsx')
    };

    // search all files for role exporting, which will later be used to limit the fetching of that file based on roles
    const roles: Record<string, string> = {};
    Object.values(files.views).forEach(file => {
      const fileString = fs.readFileSync('.' + VITE_AWAYTO_WEBAPP + '/' + file + '.tsx').toString();
      const found = fileString.match(/(?<=export const roles = )(\[.*\])/igm);
      if (found?.length) {
        roles[file] = JSON.parse(found[0].replace(/\'/g, '"'));
      }
    });
    const rolesString = JSON.stringify({ roles });
    const newRolesOutputHash = crypto.createHash('sha1').update(Buffer.from(rolesString)).digest('base64');
    if (oldRolesOutputHash !== newRolesOutputHash) {
      oldRolesOutputHash = newRolesOutputHash;
      fs.writeFileSync(appRolesOutputPath, rolesString);
    }

    const filesString = JSON.stringify(files);
    const newAppOutputHash = crypto.createHash('sha1').update(Buffer.from(filesString)).digest('base64');
    if (oldAppOutputHash !== newAppOutputHash) {
      oldAppOutputHash = newAppOutputHash;
      fs.writeFile(appBuildOutputPath, filesString, () => { })
    }
  } catch (error) {
    console.log('error!', error)
  }
}

checkWriteBuildFile();

// Replicating the resolveApp functionality
const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = (relativePath: string) => path.resolve(appDirectory, relativePath);

export default defineConfig(_ => {
  return {
    server: {
      port: 3000
    },
    base: '/app',
    resolve: {
      alias: {
        '@mui/material/Grid': path.resolve(appDirectory, './node_modules/@mui/material/Grid2'),
        'awayto/hooks': resolveApp('.' + VITE_AWAYTO_WEBAPP + '/hooks/index.ts'),
      }
    },
    plugins: [
      react(),
      // Circular dependency checking
      {
        name: 'circular-dependency',
        ...(new circular({
          exclude: /a\.js|node_modules/,
          include: /src/,
          failOnError: true,
          allowAsyncCycles: false,
          cwd: process.cwd()
        })),
        apply: 'build'
      },

      // File copying (equivalent to CopyWebpackPlugin)
      copy({
        targets: [
          {
            src: 'node_modules/pdfjs-dist/build/pdf.worker.min.mjs',
            dest: 'dist/static/js'
          }
        ],
        hook: 'writeBundle'
      }),

      // Compression (equivalent to CompressionWebpackPlugin)
      viteCompression({
        algorithm: 'gzip',
        filter: /\.(js|css|html|svg)$/i,
        threshold: 10240,
      })
    ]
  };
});

