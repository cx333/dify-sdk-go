// Package ratelimit 提供基于令牌桶的出站 HTTP 限流。
// 限制发往 Dify 的请求速率，防止打爆上游。
package ratelimit

import (
	"context"
	"net/http"

	"golang.org/x/time/rate"
)

// Transport 包装 http.RoundTripper，在每次出站请求前从令牌桶获取令牌。
// rps=0 时自动降级为无限制透传。
type Transport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

// NewTransport 创建限流传输层。
// rps: 每秒允许的请求数，0 表示不限制。
// burst: 突发容量，允许瞬间超过 rps 的请求数。
func NewTransport(base http.RoundTripper, rps float64, burst int) *Transport {
	t := &Transport{base: base}
	if rps > 0 {
		t.limiter = rate.NewLimiter(rate.Limit(rps), burst)
	}
	return t
}

// RoundTrip 实现 http.RoundTripper 接口。
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.limiter != nil {
		if err := t.limiter.Wait(req.Context()); err != nil {
			return nil, err
		}
	}
	return t.base.RoundTrip(req)
}

// Limiter 返回底层 *rate.Limiter，供监控查询状态。
func (t *Transport) Limiter() *rate.Limiter {
	return t.limiter
}

// Wait 阻塞等待令牌可用，调用方可自行控制等待策略。
func (t *Transport) Wait(ctx context.Context) error {
	if t.limiter == nil {
		return nil
	}
	return t.limiter.Wait(ctx)
}

// Allow 非阻塞检查令牌是否可用。
func (t *Transport) Allow() bool {
	if t.limiter == nil {
		return true
	}
	return t.limiter.Allow()
}
