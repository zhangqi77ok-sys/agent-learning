import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './index.css'

function showBootError(err: unknown) {
  const el = document.getElementById('app')
  if (!el) return
  const msg = err instanceof Error ? `${err.message}\n${err.stack || ''}` : String(err)
  el.innerHTML = `<pre style="padding:24px;color:#b91c1c;white-space:pre-wrap;font:12px/1.5 ui-monospace,monospace">湉码启动失败（不是空窗口）:\n\n${msg}</pre>`
}

window.addEventListener('error', (e) => showBootError(e.error || e.message))
window.addEventListener('unhandledrejection', (e) => showBootError(e.reason))

try {
  const app = createApp(App)
  app.use(createPinia())
  app.config.errorHandler = (err) => showBootError(err)
  app.mount('#app')
} catch (err) {
  showBootError(err)
}
