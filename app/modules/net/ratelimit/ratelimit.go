package ratelimit

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"pichost.io/app/utils/base"
	"pichost.io/internal/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type KeyFunc func(ctx *gin.Context) string

func IPKeyFunc(ctx *gin.Context) string {
	ip := ctx.ClientIP()
	if ip == "" {
		ip = "127.0.0.1"
	}
	return ip
}

func UserOrIPKeyFunc(ctx *gin.Context) string {
	if rawID, exists := ctx.Get("auth_user_id"); exists {
		if userID, ok := rawID.(uuid.UUID); ok && userID != uuid.Nil {
			return "user:" + userID.String()
		}
	}
	return "ip:" + IPKeyFunc(ctx)
}

type memoryEntry struct {
	count     int
	expiresAt time.Time
}

type inMemoryStore struct {
	mu    sync.Mutex
	items map[string]*memoryEntry
}

var globalMemoryStore = &inMemoryStore{
	items: make(map[string]*memoryEntry),
}

func (m *inMemoryStore) allow(key string, limit int, window time.Duration) (count int, remainingTTL time.Duration, allowed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	entry, exists := m.items[key]
	if !exists || now.After(entry.expiresAt) {
		entry = &memoryEntry{
			count:     1,
			expiresAt: now.Add(window),
		}
		m.items[key] = entry
		return 1, window, true
	}

	entry.count++
	remaining := entry.expiresAt.Sub(now)
	if remaining < 0 {
		remaining = time.Second
	}

	if entry.count > limit {
		return entry.count, remaining, false
	}
	return entry.count, remaining, true
}

func New(redisSvc *redis.RedisService, prefix string, limit int, window time.Duration, keyFunc KeyFunc) gin.HandlerFunc {
	if keyFunc == nil {
		keyFunc = IPKeyFunc
	}

	return func(ctx *gin.Context) {
		keyTarget := keyFunc(ctx)
		key := fmt.Sprintf("ratelimit:%s:%s", prefix, keyTarget)

		var count int
		var remainingTTL time.Duration
		var allowed bool
		usedRedis := false

		if redisSvc != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						usedRedis = false
					}
				}()
				client := redisSvc.DB()
				if client != nil {
					c := ctx.Request.Context()
					pipe := client.Pipeline()
					incrCmd := pipe.Incr(c, key)
					ttlCmd := pipe.TTL(c, key)
					_, err := pipe.Exec(c)
					if err == nil {
						val := int(incrCmd.Val())
						if val == 1 {
							client.Expire(c, key, window)
							ttlCmd = client.TTL(c, key)
						}
						ttl := ttlCmd.Val()
						if ttl <= 0 {
							ttl = window
						}
						count = val
						remainingTTL = ttl
						allowed = (count <= limit)
						usedRedis = true
					}
				}
			}()
		}

		if !usedRedis {
			count, remainingTTL, allowed = globalMemoryStore.allow(key, limit, window)
		}

		retryAfterSeconds := int(remainingTTL.Seconds())
		if retryAfterSeconds <= 0 {
			retryAfterSeconds = 1
		}

		ctx.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}
		ctx.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			ctx.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
			_ = base.JSON(ctx, http.StatusTooManyRequests, "too_many_requests", nil, nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
