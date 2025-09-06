package xdislock

import (
	"context"
	"fmt"
	"time"
)

func Example() {
	// 连接本地 Consul Agent
	lock, err := NewDisLock("127.0.0.1:8500", "task-job-123", 10*time.Second)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	ok, err := lock.Acquire(ctx)
	if err != nil {
		panic(err)
	}
	if !ok {
		fmt.Println("锁已被其他节点持有")
		return
	}
	fmt.Println("✅ 已获得锁，执行任务中...")

	// 模拟长任务（超过TTL）
	time.Sleep(25 * time.Second)

	fmt.Println("🧹 任务完成，释放锁...")
	if err := lock.Release(); err != nil {
		fmt.Println("释放锁失败：", err)
	} else {
		fmt.Println("✅ 锁已释放")
	}
}
