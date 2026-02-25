import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// https://vitejs.dev/config/
export default defineConfig(({ command, mode }) => {
  // load .env variables for current mode
  const env = loadEnv(mode, process.cwd(), '')
  const backendUrl = env.VITE_BACKEND_URL || `http://127.0.0.1:${env.VITE_BACKEND_PORT || process.env.WEBAPIPORT || '32000'}`

  return {
    plugins: [vue()],
    base: command === 'serve' ? '/' : '/assets/frontend/',
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
      outDir: path.resolve(__dirname, '../assets/frontend'),
      emptyOutDir: true,
      rollupOptions: {
        output: {
          entryFileNames: 'assets/index.js',
          chunkFileNames: (chunkInfo) => `assets/${chunkInfo.name.toLowerCase()}.js`,
          assetFileNames: (assetInfo) => {
            const name = assetInfo.name || 'asset'
            const ext = name.substring(name.lastIndexOf('.'))
            const baseName = name.substring(0, name.lastIndexOf('.'))
            return `assets/${baseName.toLowerCase()}${ext}`
          }
        }
      }
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src')
      }
    }
  }
})
