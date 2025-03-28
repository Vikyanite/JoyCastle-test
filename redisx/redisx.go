package redisx

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"sync/atomic"
	"time"
)

type Client struct {
	*redis.Client
	ctx            context.Context
	recoverFuncs   []RecoverFunc
	redisAlive     atomic.Bool
	redisFailCount int32
}

type RecoverFunc func(rdb *redis.Client)

func NewRedisClient(rdb *redis.Client, InCtx context.Context) *Client {
	ret := &Client{
		Client: rdb,
		ctx:    InCtx,
	}
	ret.redisAlive.Store(true)
	return ret
}

func (r *Client) RegisterRecoverFunc(rf RecoverFunc) {
	r.recoverFuncs = append(r.recoverFuncs, rf)
}

func (r *Client) IsAlive() bool {
	return r.redisAlive.Load()
}

// MonitorRedisHealth 每隔 interval 秒 Ping Redis
func (r *Client) MonitorRedisHealth(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			err := r.Client.Ping(r.ctx).Err()
			if err != nil {
				log.Println("[WARN] Redis ping failed:", err)
				if atomic.AddInt32(&r.redisFailCount, 1) >= 3 {
					r.redisAlive.Store(false)
					log.Println("[ALERT] Redis marked as DOWN")
				}
				continue
			}

			atomic.StoreInt32(&r.redisFailCount, 0)
			if !r.redisAlive.Load() {
				for _, recoverFunc := range r.recoverFuncs {
					recoverFunc(r.Client)
				}
				log.Println("[INFO] Redis recovered")
			}
			r.redisAlive.Store(true)
		}
	}
}
