// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/admin/api': {
        target: 'http://localhost:9080',
        changeOrigin: true,
        configure(proxy) {
          proxy.on('proxyRes', (proxyRes) => {
            const origin = (proxyRes.headers['origin'] as string) || '*'
            proxyRes.headers['access-control-allow-origin'] = origin
            proxyRes.headers['access-control-allow-credentials'] = 'true'
            proxyRes.headers['access-control-allow-headers'] = 'Authorization, Content-Type, X-Requested-With'
            proxyRes.headers['access-control-allow-methods'] = 'GET, POST, PUT, DELETE, PATCH, OPTIONS'
          })

          proxy.on('proxyReq', (proxyReq, req) => {
            const origin = (req.headers['origin'] as string) || '*'
            proxyReq.setHeader('origin', origin)
          })
        },
      },
    },
  },
})
