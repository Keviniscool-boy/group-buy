package config

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RDB 全局 Redis 客户端
var RDB *redis.Client

type localCacheEntry struct {
	value     string
	expiresAt time.Time
}

var localCache sync.Map

// InitRedis 初始化 Redis 连线
func InitRedis() *redis.Client {
	RDB = redis.NewClient(&redis.Options{
		Addr:     RedisHost + ":" + RedisPort,
		Password: RedisPassword,
		DB:       RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Printf("Redis 连线失败: %v", err)
		log.Println("系统将以无缓存模式继续运行")
		RDB = nil
		return nil
	}

	log.Println("Redis 连线成功")
	return RDB
}

// CacheGet 从缓存取值，Redis 不可用时使用本机内存兜底
func CacheGet(ctx context.Context, key string) (string, bool) {
	if RDB != nil {
		val, err := RDB.Get(ctx, key).Result()
		if err == nil {
			log.Printf("[Redis] 🟢 缓存命中 (Hit)，Key: %s", key)
			return val, true
		}
		log.Printf("[Redis] 🔴 缓存未命中 (Miss)，Key: %s", key)
	} else {
		log.Printf("[Redis] ❌ Redis未连接，Key: %s", key)
	}

	return localCacheGet(key)
}

// CacheSet 写入缓存，Redis 写入失败时仍保留本机内存副本
func CacheSet(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	localCacheSet(key, fmt.Sprint(value), ttl)
	if RDB == nil {
		return
	}
	if err := RDB.Set(ctx, key, value, ttl).Err(); err != nil {
		log.Printf("[Redis] 写入失败，Key: %s, err: %v", key, err)
	}
}

// CacheDel 删除 Redis 键，失败忽略
func CacheDel(ctx context.Context, keys ...string) {
	for _, key := range keys {
		localCache.Delete(key)
	}
	if RDB == nil {
		return
	}
	RDB.Del(ctx, keys...)
}

// CacheDelPattern 按前缀删除 Redis 键（用于清除某一类缓存）
func CacheDelPattern(ctx context.Context, pattern string) {
	localCacheDelPattern(pattern)
	if RDB == nil {
		return
	}
	iter := RDB.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		RDB.Del(ctx, iter.Val())
	}
}

// CacheKey 生成缓存键
func CacheKey(parts ...string) string {
	return fmt.Sprintf("zhixiang:%s", joinParts(parts...))
}

func joinParts(parts ...string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ":"
		}
		result += p
	}
	return result
}

func localCacheSet(key, value string, ttl time.Duration) {
	entry := localCacheEntry{value: value}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}
	localCache.Store(key, entry)
}

func localCacheGet(key string) (string, bool) {
	raw, ok := localCache.Load(key)
	if !ok {
		return "", false
	}

	entry, ok := raw.(localCacheEntry)
	if !ok {
		localCache.Delete(key)
		return "", false
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		localCache.Delete(key)
		return "", false
	}

	return entry.value, true
}

func localCacheDelPattern(pattern string) {
	prefix := strings.TrimSuffix(pattern, "*")
	localCache.Range(func(key, _ interface{}) bool {
		cacheKey, ok := key.(string)
		if ok && strings.HasPrefix(cacheKey, prefix) {
			localCache.Delete(cacheKey)
		}
		return true
	})
}
