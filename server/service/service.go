package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log"
	"online-mall-platform/server/config"
	"time"
)

type App struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Config config.Config
}
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }
func Bad(s string) error       { return &Error{400, s} }
func Forbidden() error         { return &Error{403, "无权执行此操作"} }
func NotFound() error          { return &Error{404, "记录不存在"} }
func Normalize(e error) error {
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return NotFound()
	}
	return e
}
func (a *App) Cache(ctx context.Context, key string, out any) bool {
	s, e := a.Redis.Get(ctx, key).Result()
	return e == nil && json.Unmarshal([]byte(s), out) == nil
}
func (a *App) Store(ctx context.Context, key string, v any, ttl time.Duration) {
	if b, e := json.Marshal(v); e == nil {
		if e := a.Redis.Set(ctx, key, b, ttl).Err(); e != nil {
			log.Printf("cache write failed: %v", e)
		}
	}
}
func (a *App) Invalidate(ctx context.Context, keys ...string) {
	if e := a.Redis.Del(ctx, keys...).Err(); e != nil {
		log.Printf("cache delete failed: %v", e)
	}
}
func cacheVersion(v []uint64) uint64 {
	if len(v) > 0 {
		return v[0]
	}
	return 0
}
func CartKey(uid uint, version ...uint64) string {
	return fmt.Sprintf("cart:%d:v%d", uid, cacheVersion(version))
}
func ProductKey(id uint, version ...uint64) string {
	return fmt.Sprintf("product:%d:v%d", id, cacheVersion(version))
}
