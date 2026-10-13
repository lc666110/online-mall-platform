import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
export default defineConfig({plugins:[vue()],server:{port:5173,strictPort:true,proxy:{'/api':{target:'http://127.0.0.1:8080',changeOrigin:true}}},build:{target:'es2022'}})
