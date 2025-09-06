package psl

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	leaderElector     *LeaderElector
	onceLeaderElector sync.Once
)

type LeaderElector struct {
	locker     *DistributedLock
	lockKey    string
	nodeID     string
	ttl        int
	renewalInt time.Duration

	mu               sync.RWMutex
	isLeader         bool
	stopCh           chan struct{}
	stopped          bool
	callbacks        []func(bool)
	consecutiveFails int
	maxFails         int
	retryCh          chan struct{}
}

type LeaderChangedCallback func(bool)

func InitLeaderElector(ctx context.Context, nodeID, lockKey string, opts ...func(*LeaderElector)) error {
	var initErr error
	onceLeaderElector.Do(func() {
		if distributedLock == nil {
			initErr = fmt.Errorf("dislocker not initialized")
			return
		}

		elector := &LeaderElector{
			locker:     distributedLock,
			lockKey:    lockKey,
			nodeID:     nodeID,
			ttl:        30,
			renewalInt: 15 * time.Second,
			stopCh:     make(chan struct{}),
			retryCh:    make(chan struct{}, 1),
			maxFails:   3,
		}

		for _, opt := range opts {
			opt(elector)
		}

		leaderElector = elector
		go elector.runElectionLoop(ctx)
	})
	return initErr
}

func GetLeaderElector() *LeaderElector {
	return leaderElector
}

func (e *LeaderElector) IsLeader() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isLeader
}

func (e *LeaderElector) OnLeaderChanged(callback func(bool)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.callbacks = append(e.callbacks, callback)
}

func (e *LeaderElector) Stop() {
	e.mu.Lock()
	if e.stopped {
		e.mu.Unlock()
		return
	}
	e.stopped = true
	close(e.stopCh)
	isLeader := e.isLeader
	e.mu.Unlock()

	if isLeader && e.locker != nil {
		GetLogger().Infof("[leader_elector] releasing leadership for key=%s node=%s", e.lockKey, e.nodeID)
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		e.locker.Release(releaseCtx, e.lockKey, e.nodeID)
	}

	e.mu.Lock()
	e.isLeader = false
	callbacks := e.callbacks
	e.callbacks = nil
	e.mu.Unlock()

	for _, cb := range callbacks {
		cb(false)
	}
}

func (e *LeaderElector) runElectionLoop(ctx context.Context) {
	electionTicker := time.NewTicker(e.renewalInt)
	defer electionTicker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ctx.Done():
			e.Stop()
			return
		case <-electionTicker.C:
			e.tryElectOrRenew()
		case <-e.retryCh:
			e.tryElectOrRenew()
		}
	}
}

func (e *LeaderElector) tryElectOrRenew() {
	e.mu.RLock()
	wasLeader := e.isLeader
	e.mu.RUnlock()

	if e.locker == nil {
		GetLogger().Errorf("[leader_elector] locker is nil, cannot try elect or renew")
		return
	}

	lockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var acquired bool
	var err error

	if wasLeader {
		acquired, err = e.locker.Renew(lockCtx, e.lockKey, e.nodeID, e.ttl)
		if err != nil {
			e.mu.Lock()
			e.consecutiveFails++
			fails := e.consecutiveFails
			e.mu.Unlock()
			GetLogger().Errorf("[leader_elector] renew lock failed (attempt %d/%d): %v", fails, e.maxFails, err)
			if fails >= e.maxFails {
				GetLogger().Warnf("[leader_elector] too many consecutive renew failures (%d), releasing leadership", fails)
				e.loseLeadership()
			}
			return
		}

		if !acquired {
			e.mu.Lock()
			e.consecutiveFails++
			fails := e.consecutiveFails
			e.mu.Unlock()
			GetLogger().Warnf("[leader_elector] renew returned false (attempt %d/%d), another node likely acquired lock", fails, e.maxFails)
			if fails >= e.maxFails {
				GetLogger().Warnf("[leader_elector] lost leadership for key=%s, releasing", e.lockKey)
				e.loseLeadership()
			}
			return
		}

		e.mu.Lock()
		e.consecutiveFails = 0
		e.mu.Unlock()
	} else {
		acquired, err = e.locker.TryAcquire(lockCtx, e.lockKey, e.nodeID, e.ttl)
		if err != nil {
			GetLogger().Errorf("[leader_elector] acquire lock failed: %v", err)
			return
		}

		if acquired {
			GetLogger().Infof("[leader_elector] acquired leadership for key=%s node=%s", e.lockKey, e.nodeID)
			e.mu.Lock()
			e.isLeader = true
			e.consecutiveFails = 0
			callbacks := e.callbacks
			e.mu.Unlock()
			for _, cb := range callbacks {
				cb(true)
			}
		} else {
			select {
			case e.retryCh <- struct{}{}:
				GetLogger().Debugf("[leader_elector] lock unavailable, scheduling retry in 2s")
			default:
			}
			go func() {
				time.Sleep(2 * time.Second)
				select {
				case e.retryCh <- struct{}{}:
				default:
				}
			}()
		}
	}
}

func (e *LeaderElector) loseLeadership() {
	e.mu.Lock()
	e.isLeader = false
	e.consecutiveFails = 0
	callbacks := e.callbacks
	e.mu.Unlock()
	for _, cb := range callbacks {
		cb(false)
	}
}
