// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package metrics

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Registry 指标注册中心（原子操作，协程安全）
type Registry struct {
	startTime time.Time

	// HTTP 请求计数
	httpTotal   atomic.Int64
	http2xx     atomic.Int64
	http4xx     atomic.Int64
	http5xx     atomic.Int64

	// 排期执行
	scheduleRuns      atomic.Int64
	schedulePublished atomic.Int64
	scheduleUnpublished atomic.Int64
	scheduleSkipped   atomic.Int64
	scheduleErrors    atomic.Int64

	// 业务
	contentPublished atomic.Int64
	apiTokenCalls    atomic.Int64
}

// 全局单例
var global = &Registry{startTime: time.Now()}

// G returns the global metrics registry.
func G() *Registry { return global }

// RecordHTTP records an HTTP request.
func (r *Registry) RecordHTTP(status int, latency time.Duration) {
	r.httpTotal.Add(1)
	switch {
	case status >= 200 && status < 300:
		r.http2xx.Add(1)
	case status >= 400 && status < 500:
		r.http4xx.Add(1)
	case status >= 500:
		r.http5xx.Add(1)
	}
}

// RecordSchedule records a schedule execution.
func (r *Registry) RecordSchedule(published, unpublished, skipped, errors int64) {
	r.scheduleRuns.Add(1)
	r.schedulePublished.Add(published)
	r.scheduleUnpublished.Add(unpublished)
	r.scheduleSkipped.Add(skipped)
	r.scheduleErrors.Add(errors)
}

// RecordContentPublish increments content published counter.
func (r *Registry) RecordContentPublish() { r.contentPublished.Add(1) }

// RecordAPITokenCall increments API token call counter.
func (r *Registry) RecordAPITokenCall() { r.apiTokenCalls.Add(1) }

// PrometheusText renders all metrics in Prometheus exposition format.
func (r *Registry) PrometheusText() string {
	var b strings.Builder

	// HELP/TYPE + value
	emit := func(name, help, typ, value string) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s %s\n%s %s\n", name, help, name, typ, name, value)
	}
	emitGauge := func(name, help string, value int64) {
		emit(name, help, "gauge", fmt.Sprintf("%d", value))
	}
	emitCounter := func(name, help string, value int64) {
		emit(name, help, "counter", fmt.Sprintf("%d", value))
	}

	// Uptime
	uptime := int64(time.Since(r.startTime).Seconds())
	emitGauge("contful_uptime_seconds", "Process uptime in seconds.", uptime)

	// HTTP
	emitCounter("contful_http_requests_total", "Total HTTP requests.", r.httpTotal.Load())
	emitCounter("contful_http_requests_2xx", "HTTP 2xx responses.", r.http2xx.Load())
	emitCounter("contful_http_requests_4xx", "HTTP 4xx responses.", r.http4xx.Load())
	emitCounter("contful_http_requests_5xx", "HTTP 5xx responses.", r.http5xx.Load())

	// Schedule
	emitCounter("contful_schedule_runs_total", "Schedule cron executions.", r.scheduleRuns.Load())
	emitCounter("contful_schedule_published_total", "Entries published by schedule.", r.schedulePublished.Load())
	emitCounter("contful_schedule_unpublished_total", "Entries unpublished by schedule.", r.scheduleUnpublished.Load())
	emitCounter("contful_schedule_skipped_total", "Schedule executions skipped (status mismatch).", r.scheduleSkipped.Load())
	emitCounter("contful_schedule_errors_total", "Schedule execution errors.", r.scheduleErrors.Load())

	// Business
	emitCounter("contful_content_published_total", "Content entries published.", r.contentPublished.Load())
	emitCounter("contful_api_token_calls_total", "OpenAPI token calls.", r.apiTokenCalls.Load())

	return b.String()
}

// Middleware returns a Gin middleware that records HTTP metrics.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		global.RecordHTTP(c.Writer.Status(), time.Since(start))
	}
}
