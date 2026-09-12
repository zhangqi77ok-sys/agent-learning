import fs from 'fs'

const store = fs.readFileSync(new URL('../src/stores/workbench.ts', import.meta.url), 'utf8')
if (store.includes('watch(') && !store.includes('import {') && !/import\s*\{[^}]*\bwatch\b/.test(store)) {
  console.error('workbench.ts calls watch() but does not import it — desktop boot will throw')
  process.exit(1)
}
if (/\bstore\.gitVersion\b/.test(store)) {
  console.error('workbench.ts references undefined store.gitVersion')
  process.exit(1)
}

const html = fs.readFileSync(new URL('../dist/index.html', import.meta.url), 'utf8')
if (html.includes('src="/assets') || html.includes("src='/assets") || html.includes('href="/assets')) {
  console.error('dist/index.html uses absolute /assets paths — Wails WebView will white-screen. Set vite base: "./"')
  process.exit(1)
}
if (!html.includes('./assets/') && !html.includes('assets/')) {
  console.error('dist/index.html missing asset links')
  process.exit(1)
}
console.log('ok: frontend assets are relative')
