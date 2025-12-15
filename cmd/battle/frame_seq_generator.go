package main

import (
	"sync"
	"time"
)

// 这是一个逻辑帧序列生成器，只需要配置每秒多少帧，会根据时间获得启动后当前是第几帧

// TimeProvider 时间提供者接口，用于依赖注入，便于测试
type TimeProvider interface {
	// Now 返回当前时间
	Now() time.Time
	// Since 返回从指定时间以来经过的时间
	Since(t time.Time) time.Duration
	// Sleep 暂停执行指定的时间
	Sleep(d time.Duration)
}

// DefaultTimeProvider 默认时间提供者（使用标准库 time）
type DefaultTimeProvider struct{}

// Now 实现 TimeProvider 接口
func (DefaultTimeProvider) Now() time.Time {
	return time.Now()
}

// Since 实现 TimeProvider 接口
func (DefaultTimeProvider) Since(t time.Time) time.Duration {
	return time.Since(t)
}

// Sleep 实现 TimeProvider 接口
func (DefaultTimeProvider) Sleep(d time.Duration) {
	time.Sleep(d)
}

// FrameSeqGenerator 逻辑帧序列生成器
type FrameSeqGenerator struct {
	fps          int64         // 每秒帧数
	frameTime    time.Duration // 每一帧的时间间隔（纳秒）
	startTime    time.Time     // 启动时间
	timeProvider TimeProvider  // 时间提供者，可注入
	mu           sync.RWMutex  // 读写锁
}

// NewFrameSeqGenerator 创建一个帧序列生成器，使用默认时间提供者
// fps: 每秒帧数（例如 30 表示 30fps）
func NewFrameSeqGenerator(fps int64) *FrameSeqGenerator {
	return NewFrameSeqGeneratorWithTimeProvider(fps, DefaultTimeProvider{})
}

// NewFrameSeqGeneratorWithTimeProvider 创建一个帧序列生成器，注入自定义时间提供者
// fps: 每秒帧数（例如 30 表示 30fps）
// tp: 时间提供者（用于测试可以注入 mock 时间）
func NewFrameSeqGeneratorWithTimeProvider(fps int64, tp TimeProvider) *FrameSeqGenerator {
	if fps <= 0 {
		fps = 30 // 默认 30fps
	}
	if tp == nil {
		tp = DefaultTimeProvider{}
	}

	return &FrameSeqGenerator{
		fps:          fps,
		frameTime:    time.Duration(time.Second.Nanoseconds() / fps),
		startTime:    tp.Now(),
		timeProvider: tp,
	}
}

// GetCurrentFrame 获取当前帧序列号（从 0 开始）
// 返回值是自启动以来经历的帧数
func (fg *FrameSeqGenerator) GetCurrentFrame() uint64 {
	fg.mu.RLock()
	defer fg.mu.RUnlock()

	elapsed := fg.timeProvider.Since(fg.startTime)
	// 计算经过了多少帧
	frame := elapsed.Nanoseconds() / int64(fg.frameTime)

	return uint64(frame)
}

// Reset 重置生成器（重新启动计时）
func (fg *FrameSeqGenerator) Reset() {
	fg.mu.Lock()
	defer fg.mu.Unlock()

	fg.startTime = fg.timeProvider.Now()
}

// GetFPS 获取设定的每秒帧数
func (fg *FrameSeqGenerator) GetFPS() int64 {
	fg.mu.RLock()
	defer fg.mu.RUnlock()

	return fg.fps
}

// SetFPS 设置每秒帧数
func (fg *FrameSeqGenerator) SetFPS(fps int64) {
	fg.mu.Lock()
	defer fg.mu.Unlock()

	if fps <= 0 {
		fps = 30
	}

	fg.fps = fps
	fg.frameTime = time.Duration(time.Second.Nanoseconds() / fps)
}

// SleepUntilNextFrame 睡眠到下一帧
func (fg *FrameSeqGenerator) SleepUntilNextFrame() {
	fg.mu.RLock()
	frameTime := fg.frameTime
	startTime := fg.startTime
	tp := fg.timeProvider
	fg.mu.RUnlock()

	// 计算应该睡眠的时间
	elapsed := tp.Since(startTime)
	nextFrameTime := frameTime * time.Duration(elapsed.Nanoseconds()/int64(frameTime)+1)
	sleepTime := nextFrameTime - elapsed

	if sleepTime > 0 {
		tp.Sleep(sleepTime)
	}
}
