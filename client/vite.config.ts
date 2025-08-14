import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: true, // これでネットワークURLが表示される
    open: true  // 起動時に自動でブラウザを開く
  }
})
