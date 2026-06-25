import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'
import './styles.css'
import { initializeTheme } from './composables/useTheme'

initializeTheme()
createApp(App).use(router).mount('#app')
