package middleware

import (
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"online-mall-platform/server/model"
	"online-mall-platform/server/service"
	"strings"
	"time"
)

func Auth(a *service.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "请先登录"})
			return
		}
		u, sid, e := a.Authenticate(c.Request.Context(), strings.TrimPrefix(raw, "Bearer "))
		if e != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "登录已失效"})
			return
		}
		c.Set("user", u)
		c.Set("sid", sid)
		c.Next()
	}
}
func Admin(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.MustGet("user").(model.User)
		for _, r := range roles {
			if u.Role == r {
				c.Set("admin", true)
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "无权执行此操作"})
	}
}

var counter = redis.NewScript(`local n=redis.call('INCR',KEYS[1]); if n==1 then redis.call('EXPIRE',KEYS[1],ARGV[1]) end; return n`)

func RateLimit(a *service.App, scope string, limit int64, seconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := sha256.Sum256([]byte(c.ClientIP()))
		key := fmt.Sprintf("rate:%s:%x", scope, hash[:12])
		n, e := counter.Run(c.Request.Context(), a.Redis, []string{key}, seconds).Int64()
		if e != nil {
			c.AbortWithStatusJSON(503, gin.H{"error": "服务暂不可用"})
			return
		}
		if n > limit {
			c.Header("Retry-After", fmt.Sprint(seconds))
			c.AbortWithStatusJSON(429, gin.H{"error": "请求过于频繁"})
			return
		}
		c.Next()
	}
}
func Visit(a *service.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := sha256.Sum256([]byte(c.ClientIP()))
		key := fmt.Sprintf("visits:%s:%x", time.Now().UTC().Format("2006-01-02"), hash[:12])
		counter.Run(c.Request.Context(), a.Redis, []string{key}, 7*24*3600)
		c.Next()
	}
}
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
		c.Next()
	}
}
