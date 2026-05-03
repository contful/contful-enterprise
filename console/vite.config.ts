// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// 内联 CORS 插件：开发环境下为代理响应添加 CORS 头
function viteCorsPlugin() {
  return {
    name: 'vite-cors',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        // 只对 /admin/api 路径添加 CORS 头
        if (req.url?.startsWith('/admin/api')) {
          const origin = req.headers.origin || ''
          res.setHeader('Access-Control-Allow-Origin', origin)
          res.setHeader('Access-Control-Allow-Credentials', 'true')
          res.setHeader('Access-Control-Allow-Headers', 'Authorization, Content-Type, X-Requested-With')
          res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS, PATCH')
          res.setHeader('Access-Control-Max-Age', '86400')

          // 预检请求直接返回 204
          if (req.method === 'OPTIONS') {
            res.writeHead(204)
            res.end()
            return
          }
        }
        next()
      })
    },
  }
}

export default defineConfig({
  plugins: [vue(), viteCorsPlugin()],
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
        // 在代理响应中追加 CORS 头，确保浏览器能通过预检
        onProxyRes(proxyRes, req) {
          const origin = (req.headers.origin as string) || ''
          proxyRes.headers['access-control-allow-origin'] = origin
          proxyRes.headers['access-control-allow-credentials'] = 'true'
          proxyRes.headers['access-control-allow-headers'] = 'Authorization, Content-Type, X-Requested-With'
          proxyRes.headers['access-control-allow-methods'] = 'GET, POST, PUT, DELETE, OPTIONS, PATCH'
        },
      },
    },
  },
})
