package game

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"game-sandbox/server/utils"
)

// デッドロック検出用の構造体
type lockTracker struct {
	holders map[string]*lockInfo
	mu      sync.Mutex
}

type lockInfo struct {
	goroutineID uint64
	stackTrace  string
	acquiredAt  time.Time
}

var globalLockTracker = &lockTracker{
	holders: make(map[string]*lockInfo),
}

// ゴルーチンIDを取得する関数
func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	var id uint64
	fmt.Sscanf(string(b), "goroutine %d ", &id)
	return id
}

// スタックトレースを取得
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	trace := ""
	for {
		frame, more := frames.Next()
		trace += fmt.Sprintf("  %s\n    %s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return trace
}

// ロック取得を記録
func (g *Game) trackLockAcquire(lockName string) {
	globalLockTracker.mu.Lock()
	defer globalLockTracker.mu.Unlock()

	info := &lockInfo{
		goroutineID: getGoroutineID(),
		stackTrace:  getStackTrace(),
		acquiredAt:  time.Now(),
	}

	// 既に別のゴルーチンが保持している場合は警告
	if existing, exists := globalLockTracker.holders[lockName]; exists {
		if existing.goroutineID != info.goroutineID {
			utils.Warn("Lock contention detected", map[string]interface{}{
				"lock":              lockName,
				"holder_goroutine":  existing.goroutineID,
				"waiting_goroutine": info.goroutineID,
				"held_duration":     time.Since(existing.acquiredAt).String(),
				"holder_trace":      existing.stackTrace,
				"waiter_trace":      info.stackTrace,
			})
		}
	}

	globalLockTracker.holders[lockName] = info
}

// ロック解放を記録
func (g *Game) trackLockRelease(lockName string) {
	globalLockTracker.mu.Lock()
	defer globalLockTracker.mu.Unlock()

	delete(globalLockTracker.holders, lockName)
}

// デッドロック検出器を起動
func (g *Game) StartDeadlockDetector() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			g.checkForDeadlock()
		}
	}()
}

// デッドロックをチェック
func (g *Game) checkForDeadlock() {
	globalLockTracker.mu.Lock()
	defer globalLockTracker.mu.Unlock()

	now := time.Now()
	for lockName, info := range globalLockTracker.holders {
		duration := now.Sub(info.acquiredAt)
		if duration > 3*time.Second {
			utils.Error("Potential deadlock detected", map[string]interface{}{
				"lock":         lockName,
				"goroutine_id": info.goroutineID,
				"held_for":     duration.String(),
				"stack_trace":  info.stackTrace,
			})
		}
	}
}

// デバッグ用のLock/Unlock
func (g *Game) LockWithDebug(name string) {
	g.trackLockAcquire(name)
	g.mu.Lock()
}

func (g *Game) UnlockWithDebug(name string) {
	g.mu.Unlock()
	g.trackLockRelease(name)
}

func (g *Game) RLockWithDebug(name string) {
	g.trackLockAcquire(name + "_read")
	g.mu.RLock()
}

func (g *Game) RUnlockWithDebug(name string) {
	g.mu.RUnlock()
	g.trackLockRelease(name + "_read")
}
