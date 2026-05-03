// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import http from 'http'
import { type ViteDevServer } from 'vite'

// 轻量代理中间件：用 Node.js 内置 http 转发请求，并附加 CORS 头
function proxyMiddleware(server: ViteDevServer) {
  server.middlewares.use('/admin/api', (req, res) => {
    const origin = (req.headers.origin as string) || '*'

    // OPTIONS 预检请求直接返回 CORS 头，不转发
    if (req.method === 'OPTIONS') {
      res.setHeader('Access-Control-Allow-Origin', origin)
      res.setHeader('Access-Control-Allow-Credentials', 'true')
      res.setHeader('Access-Control-Allow-Headers', 'Authorization, Content-Type, X-Requested-With')
      res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, PATCH, OPTIONS')
      res.setHeader('Access-Control-Max-Age', '86400')
      res.writeHead(204)
      res.end()
      return
    }

    // 转发请求到 Admin API
    const url = new URL(req.url!, 'http://localhost:9080')
    const options: http.RequestOptions = {
      hostname: 'localhost',
      port: 9080,
      path: url.pathname + url.search,
      method: req.method,
      headers: {
        ...req.headers,
        host: 'localhost:9080',
      },
    }

    const proxyReq = http.request(options, (proxyRes) => {
      // 追加 CORS 头后返回
      const headers = proxyRes.headers as Record<string, string | string[] | undefined>
      headers['access-control-allow-origin'] = origin
      headers['access-control-allow-credentials'] = 'true'
      headers['access-control-allow-headers'] = 'Authorization, Content-Type, X-Requested-With'
      headers['access-control-allow-methods'] = 'GET, POST, PUT, DELETE, PATCH, OPTIONS'
      res.writeHead(proxyRes.statusCode!, proxyRes.headers)
      proxyRes.pipe(res, { end: true })
    })

    proxyReq.on('error', (err) => {
      console.error('[proxy] request failed:', err.message)
      if (!res.headersSent) {
        res.writeHead(502, { 'content-type': 'application/json' })
        res.end(JSON.stringify({ code: 502, message: 'Proxy error: ' + err.message }))
      }
    })

    req.pipe(proxyReq, { end: true })
  })
}

export default defineConfig({
  plugins: [
    vue(),
    {
      name: 'contful-admin-proxy',
      configureServer: (server) => {
        proxyMiddleware(server)
      },
    },
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    // 关闭 Vite 内置代理（由我们的中间件接管）
    proxy: {},
  },
})
