import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const clientRoot = path.resolve(__dirname, 'client')

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendUrl = env.VITE_BACKEND_URL || `http://127.0.0.1:${env.VITE_BACKEND_PORT || process.env.WEBAPIPORT || '31000'}`

  return {
    root: clientRoot,
    base: '/apps/spam/',
    plugins: [vue()],
    server: {
      host: '127.0.0.1',
      port: Number(env.VITE_DEV_PORT || 5175),
      proxy: {
        '/api': { target: backendUrl, changeOrigin: true, secure: false },
        '/spam': { target: backendUrl, changeOrigin: true, secure: false }
      }
    },
    build: {
      outDir: path.resolve(__dirname, 'dist'),
      emptyOutDir: true,
      rollupOptions: {
        output: {
          hashCharacters: 'hex',
          entryFileNames: 'assets/[name]-[hash].js',
          chunkFileNames: 'assets/[name]-[hash].js',
          assetFileNames: 'assets/[name]-[hash][extname]'
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
