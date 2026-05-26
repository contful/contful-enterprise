# Contful Enterprise — Grafana 监控面板

配套 `GET /metrics` Prometheus 端点，提供 5 个监控分组共 19 个面板。

## 快速导入

1. Grafana → Dashboards → Import → Upload JSON file
2. 选择 `contful-enterprise-dashboard.json`
3. 选择 Prometheus 数据源

## 面板分组

| 分组 | 面板数 | 指标 |
|------|--------|------|
| 概览 | 4 | 运行时间、QPS、错误率、总请求数 |
| HTTP 请求 | 2 | 状态码分布趋势 (2xx/4xx/5xx) |
| 数据库连接池 | 5 | 打开/使用中/空闲连接、等待次数、趋势图 |
| 排期调度 | 5 | 执行次数、已发布/下架/跳过/错误、趋势图 |
| 业务指标 | 3 | 内容发布量、API 调用量、趋势图 |

## Prometheus 配置

```yaml
scrape_configs:
  - job_name: 'contful-admin'
    scrape_interval: 30s
    static_configs:
      - targets: ['localhost:9080']
    metrics_path: '/metrics'

  - job_name: 'contful-openapi'
    scrape_interval: 30s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```
