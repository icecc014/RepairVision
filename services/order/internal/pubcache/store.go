package pubcache

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// V7 公共报修用的轻量键值缓存：优先使用 Redis；
// Redis 不可用时自动降级为进程内 TTL 缓存，保证公共报修页在单实例部署下依然可用。
type Store struct {
	client *redis.Redis
	mu     sync.Mutex
	mem    map[string]memEntry
}

type memEntry struct {
	value  string
	expire time.Time
}

func New(addr string) *Store {
	s := &Store{mem: make(map[string]memEntry)}
	if addr != "" {
		s.client = redis.New(addr)
	}
	return s
}

// UsingRedis 当前是否走 Redis（Redis 掉线时降级为内存实现）。
func (s *Store) UsingRedis() bool { return s.client != nil }

// Setex 写入带过期时间的键。
func (s *Store) Setex(ctx context.Context, key, value string, ttl time.Duration) {
	if ttl <= 0 {
		ttl = time.Minute
	}
	if s.client != nil {
		if err := s.client.SetexCtx(ctx, key, value, int(ttl.Seconds())); err == nil {
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mem[key] = memEntry{value: value, expire: time.Now().Add(ttl)}
}

// Get 读取键，返回是否命中。
func (s *Store) Get(ctx context.Context, key string) (string, bool) {
	if s.client != nil {
		if value, err := s.client.GetCtx(ctx, key); err == nil {
			return value, true
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.mem[key]
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expire) {
		delete(s.mem, key)
		return "", false
	}
	return entry.value, true
}

// Del 删除键（验证码一次性校验后调用）。
func (s *Store) Del(ctx context.Context, key string) {
	if s.client != nil {
		_, _ = s.client.DelCtx(ctx, key)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.mem, key)
}

// Incr 计数并在首次计数时设置过期时间。
func (s *Store) Incr(ctx context.Context, key string, ttl time.Duration) int64 {
	if ttl <= 0 {
		ttl = time.Minute
	}
	seconds := int(ttl.Seconds())
	if seconds <= 0 {
		seconds = 1
	}
	if s.client != nil {
		if count, err := s.client.IncrCtx(ctx, key); err == nil {
			if count == 1 {
				_ = s.client.ExpireCtx(ctx, key, seconds)
			}
			return count
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	entry, ok := s.mem[key]
	if !ok || now.After(entry.expire) {
		s.mem[key] = memEntry{value: "1", expire: now.Add(ttl)}
		return 1
	}
	count, _ := strconv.ParseInt(entry.value, 10, 64)
	count++
	s.mem[key] = memEntry{value: strconv.FormatInt(count, 10), expire: entry.expire}
	return count
}
