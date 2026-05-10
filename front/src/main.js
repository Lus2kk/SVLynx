import { createApp } from 'vue'
import App from './App.vue'

document.addEventListener('touchmove', (e) => {
  if (e.target.closest('.messages-area, .sidebar-list')) return
  e.preventDefault()
}, { passive: false })

createApp(App).mount('#app')