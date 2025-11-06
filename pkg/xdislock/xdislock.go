package xdislock

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

type DisLock struct {
	client    *api.Client
	sessionID string
	key       string
	stopRenew chan struct{}
	cancelFn  context.CancelFunc
}

// 创建新锁
func NewDisLock(address, key string, ttl time.Duration) (*DisLock, error) {
	cfg := api.DefaultConfig()
	cfg.Address = address

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// 创建 Session
	session := &api.SessionEntry{
		Name:      "lock-" + key,
		TTL:       fmt.Sprintf("%ds", int(ttl.Seconds())),
		Behavior:  api.SessionBehaviorDelete,
		LockDelay: 1 * time.Second,
	}

	sessionID, _, err := client.Session().Create(session, nil)
	if err != nil {
		return nil, err
	}

	return &DisLock{
		client:    client,
		sessionID: sessionID,
		key:       "locks/" + key,
		stopRenew: make(chan struct{}),
	}, nil
}

// 尝试获取锁
func (l *DisLock) Acquire(ctx context.Context) (bool, error) {
	p := &api.KVPair{
		Key:     l.key,
		Value:   []byte(time.Now().String()),
		Session: l.sessionID,
	}
	acquired, _, err := l.client.KV().Acquire(p, nil)
	if err != nil {
		return false, err
	}
	if acquired {
		// 启动自动续租 goroutine
		ctx, cancel := context.WithCancel(ctx)
		l.cancelFn = cancel
		go l.autoRenew(ctx)
	}
	return acquired, nil
}

// 自动续租 Session
func (l *DisLock) autoRenew(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // 每隔 5s 续租
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, _, err := l.client.Session().Renew(l.sessionID, nil)
			if err != nil {
				log.Printf("Session renew failed: %v", err)
				return
			}
			log.Println("Session renewed successfully")
		case <-ctx.Done():
			return
		case <-l.stopRenew:
			return
		}
	}
}

// 释放锁
func (l *DisLock) Release() error {
	// 停止续租
	if l.cancelFn != nil {
		l.cancelFn()
	}
	close(l.stopRenew)

	p := &api.KVPair{
		Key:     l.key,
		Session: l.sessionID,
	}
	_, _, err := l.client.KV().Release(p, nil)
	if err != nil {
		return err
	}
	_, err = l.client.Session().Destroy(l.sessionID, nil)
	return err
}
