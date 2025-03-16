package middleware

import (
	"sync"
	"time"

	"github.com/beego/beego/v2/server/web/context"
)

// RateLimiter 简单的限流实现
type IPRateLimiter struct {
	sync.Mutex
	ipRequestCount map[string]int
	ipLastRequest  map[string]time.Time
}

var limiter = &IPRateLimiter{
	ipRequestCount: make(map[string]int),
	ipLastRequest:  make(map[string]time.Time),
}

// 清理过期的IP请求记录
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			limiter.Lock()
			for ip, lastTime := range limiter.ipLastRequest {
				if time.Since(lastTime) > 5*time.Minute {
					delete(limiter.ipRequestCount, ip)
					delete(limiter.ipLastRequest, ip)
				}
			}
			limiter.Unlock()
		}
	}()
}

// RateLimiter 限流中间件
func RateLimiter(ctx *context.Context) {
	ip := ctx.Input.IP()
	
	limiter.Lock()
	defer limiter.Unlock()
	
	// 检查IP的请求频率
	now := time.Now()
	if lastTime, exists := limiter.ipLastRequest[ip]; exists {
		if now.Sub(lastTime) < time.Second { // 1秒内
			count := limiter.ipRequestCount[ip]
			if count > 10 { // 单个IP 1秒内最多10个请求
				ctx.Output.SetStatus(429)
				ctx.Output.JSON(map[string]interface{}{
					"code":    429,
					"message": "请求过于频繁，请稍后再试",
				}, true, false)
				return
			}
			limiter.ipRequestCount[ip] = count + 1
		} else {
			// 重置计数
			limiter.ipRequestCount[ip] = 1
		}
	} else {
		limiter.ipRequestCount[ip] = 1
	}
	
	limiter.ipLastRequest[ip] = now
} 