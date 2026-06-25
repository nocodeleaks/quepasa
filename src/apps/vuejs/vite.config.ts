import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const clientRoot = path.resolve(__dirname, 'client')
const appBase = '/apps/vuejs/'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  // load .env variables for current mode
  const env = loadEnv(mode, process.cwd(), '')
  const backendUrl = env.VITE_BACKEND_URL || `http://127.0.0.1:${env.VITE_BACKEND_PORT || process.env.WEBAPIPORT || '32000'}`

  return {
    root: clientRoot,
    plugins: [vue()],
    base: appBase,
    server: {
      port: Number(env.VITE_DEV_PORT || 5173),
      proxy: {
        '/api': {
          target: backendUrl,
          changeOrigin: true,
          secure: false,
        },
        '/form': {
          target: backendUrl,
          changeOrigin: true,
          secure: false,
        },
        '/login': {
          target: backendUrl,
          changeOrigin: true,
          secure: false,
        },
        '/logout': {
          target: backendUrl,
          changeOrigin: true,
          secure: false,
        },
        '/session': {
          target: backendUrl,
          changeOrigin: true,
          secure: false,
        }
      }
    },
    build: {
      outDir: path.resolve(__dirname, 'dist'),
      emptyOutDir: true,
      rollupOptions: {
        output: {
          // Content-hashed filenames so every deploy emits new asset URLs.
          // index.html is served with Cache-Control: no-store (see webserver.go),
          // so the browser always reads the fresh HTML, which then points at the
          // new hashed assets — no hard refresh needed after a deploy.
          //
          // Hashes use lowercase hex (not the default mixed-case base64) and the
          // name parts are lowercased, because the server normalizes every
          // request path to lowercase (MiddlewareForNormalizePaths). A mixed-case
          // asset name would never match on disk and would fall back to index.html
          // (served as text/html), breaking module script loading.
          hashCharacters: 'hex',
          entryFileNames: 'assets/[name]-[hash].js',
          chunkFileNames: (chunkInfo) => `assets/${chunkInfo.name.toLowerCase()}-[hash].js`,
          assetFileNames: (assetInfo) => {
            const name = assetInfo.name || 'asset'
            const ext = name.substring(name.lastIndexOf('.'))
            const baseName = name.substring(0, name.lastIndexOf('.'))
            return `assets/${baseName.toLowerCase()}-[hash]${ext}`
          }
        }
      }
    },
    resolve: {
      alias: {
        '@': path.resolve(clientRoot, 'src')
      }
    }
  }
})
