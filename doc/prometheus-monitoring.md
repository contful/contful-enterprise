# Prometheus + Grafana 监控配置指南

> 版本: v1.3.0 | 企业版专属功能

Contful 企业版的 Admin API (:9080) 和 Open API (:8080) 均暴露 `GET /metrics` 端点，输出 Prometheus 标准格式，无需认证。

---

## 1. 架构

```
┌─────────────┐     scrape /metrics      ┌──────────────┐
│  Admin API  │◄─────────────────────────│  Prometheus  │
│   :9080     │                          │   :9090      │
└─────────────┘                          └──────┬───────┘
                                               │ query
┌─────────────┐     scrape /metrics      ┌──────▼───────┐
│  Open API   │◄─────────────────────────│   Grafana    │
│   :8080     │                          │   :3000      │
└─────────────┘                          └──────────────┘
```

---

## 2. 指标清单

### Admin API (:9080)

| 类别 | 指标 | 类型 |
|------|------|------|
| 运行 | `contful_uptime_seconds` | gauge |
| HTTP | `contful_http_requests_total` / `2xx` / `4xx` / `5xx` | counter |
| DB | `contful_db_pool_open_connections` / `in_use` / `idle` / `wait_count` / `max_idle_closed` | gauge+counter |
| 排期 | `contful_schedule_runs_total` / `published_total` / `unpublished_total` / `skipped_total` / `errors_total` | counter |
| 业务 | `contful_content_published_total` / `contful_api_token_calls_total` | counter |

### Open API (:8080)

| 类别 | 指标 | 类型 |
|------|------|------|
| 运行 | `contful_uptime_seconds` | gauge |
| HTTP | `contful_http_requests_total` / `2xx` / `4xx` / `5xx` | counter |
| DB | `contful_db_pool_open_connections` / `in_use` / `idle` / `wait_count` / `max_idle_closed` | gauge+counter |

---

## 3. Prometheus 配置

### 3.1 安装 Prometheus

```bash
# Docker
docker run -d --name prometheus -p 9090:9090 \
  -v /opt/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

### 3.2 scrape 配置

```yaml
# /opt/prometheus/prometheus.yml
global:
  scrape_interval: 30s
  evaluation_interval: 30s

scrape_configs:
  - job_name: 'contful-admin'
    scrape_interval: 30s
    metrics_path: '/metrics'
    static_configs:
      - targets: ['localhost:9080']
        labels:
          service: 'admin'

  - job_name: 'contful-openapi'
    scrape_interval: 30s
    metrics_path: '/metrics'
    static_configs:
      - targets: ['localhost:8080', 'openapi-2:8080', 'openapi-3:8080']
        labels:
          service: 'openapi'
```

> 💡 OpenAPI 如有多个副本，在 `targets` 中添加所有实例地址。

### 3.3 验证

```bash
# 手动测试指标端点
curl http://localhost:9080/metrics
curl http://localhost:8080/metrics

# 检查 Prometheus targets
curl http://localhost:9090/api/v1/targets | grep contful
```

---

## 4. Grafana 配置

### 4.1 安装 Grafana

```bash
docker run -d --name grafana -p 3000:3000 grafana/grafana
```

### 4.2 添加数据源

1. 登录 Grafana → Connections → Data Sources → Add → Prometheus
2. URL: `http://prometheus:9090`（或宿主机 `http://localhost:9090`）
3. Save & Test

### 4.3 导入面板

1. Dashboards → Import → Upload JSON file
2. 选择 `grafana/contful-enterprise-dashboard.json`
3. 选择 Prometheus 数据源 → Import

---

## 5. Docker Compose 一键部署

```yaml
# docker-compose.monitoring.yaml
services:
  prometheus:
    image: prom/prometheus
    container_name: contful-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'

  grafana:
    image: grafana/grafana
    container_name: contful-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

```bash
docker compose -f docker-compose.monitoring.yaml up -d
```

---

## 6. 报警规则（示例）

```yaml
# prometheus-rules.yml
groups:
  - name: contful
    rules:
      - alert: HighErrorRate
        expr: rate(contful_http_requests_5xx[5m]) / rate(contful_http_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Contful 错误率超过 5%"

      - alert: DBPoolExhausted
        expr: contful_db_pool_wait_count > 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "数据库连接池出现等待"

      - alert: ScheduleErrors
        expr: rate(contful_schedule_errors_total[5m]) > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "排期调度器出现执行错误"
```

---

## 7. 常见问题

**Q: /metrics 返回空？**

检查服务是否为企业版（社区版无此端点），确认 `admin/internal/metrics/` 包已编译。

**Q: 指标数据不更新？**

Prometheus 默认 30s 抓取一次，刷新 Grafana 面板或等待下一个抓取周期。

**Q: 如何自定义抓取间隔？**

修改 `prometheus.yml` 中 `scrape_interval` 值（建议 ≥ 15s）。Contful `/metrics` 端点无性能开销，可以更频繁抓取。
