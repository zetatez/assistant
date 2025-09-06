package psl

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDistributedLock_NormalizeTTL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectPing()

	cfg := &Config{
		Dislock: DislockConfig{
			DefaultTTL: 30,
			MaxTTL:     300,
		},
	}

	locker := &DistributedLock{
		db:         db,
		cfg:        cfg,
		defaultTTL: cfg.Dislock.DefaultTTL,
		maxTTL:     cfg.Dislock.MaxTTL,
	}

	tests := []struct {
		name     string
		inputTTL int
		expected int
	}{
		{"zero ttl uses default", 0, 30},
		{"negative ttl uses default", -1, 30},
		{"valid ttl unchanged", 50, 50},
		{"ttl exceeds max uses max", 500, 300},
		{"ttl at max unchanged", 300, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := locker.normalizeTTL(tt.inputTTL)
			if result != tt.expected {
				t.Errorf("normalizeTTL(%d) = %d, want %d", tt.inputTTL, result, tt.expected)
			}
		})
	}
}

func TestDistributedLock_IsHeldRequiresQueries(t *testing.T) {
	t.Skip("IsHeld requires repo.Queries initialization which needs proper sqlc setup")
}

func TestDistributedLock_TryAcquire(t *testing.T) {
	t.Skip("TryAcquire requires repo.Queries initialization")
}

func TestDistributedLock_Release(t *testing.T) {
	t.Skip("Release requires repo.Queries initialization")
}

func TestDistributedLock_Renew(t *testing.T) {
	t.Skip("Renew requires repo.Queries initialization")
}

func TestDistributedLock_CountActive(t *testing.T) {
	t.Skip("CountActive requires repo.Queries initialization")
}

func TestLockInfo_Fields(t *testing.T) {
	now := time.Now()
	info := LockInfo{
		ID:         1,
		LockKey:    "testkey",
		LockHolder: "holder1",
		LockTTL:    30,
		ExpireTime: now.Add(time.Hour),
		IsActive:   true,
	}

	if info.ID != 1 {
		t.Errorf("expected ID 1, got %d", info.ID)
	}
	if info.LockKey != "testkey" {
		t.Errorf("expected LockKey testkey, got %s", info.LockKey)
	}
	if !info.IsActive {
		t.Error("expected IsActive to be true")
	}
}
