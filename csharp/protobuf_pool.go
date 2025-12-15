package csharp

import (
	"sync"

	proto_pb "goPureWithCsharp/csharp/proto"
)

// ============================================================================
// Protobuf 缓冲池 - 缓存序列化后的 proto 字节数据
// ============================================================================

// ProtoBufferPool 用于缓存 protobuf 序列化数据的缓冲池
// 避免频繁的内存申请和 GC 压力
type ProtoBufferPool struct {
	pool *sync.Pool
	size int // 初始大小
}

// NewProtoBufferPool 创建一个 protobuf buffer 池
func NewProtoBufferPool(initialSize int) *ProtoBufferPool {
	return &ProtoBufferPool{
		size: initialSize,
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, initialSize)
			},
		},
	}
}

// Get 从池中获取一个 buffer
func (p *ProtoBufferPool) Get() []byte {
	buf := p.pool.Get().([]byte)
	return buf[:0] // 重置为空，但保留容量
}

// Put 将 buffer 放回池中
func (p *ProtoBufferPool) Put(buf []byte) {
	if cap(buf) >= p.size {
		p.pool.Put(buf[:0]) // 清空但保留容量
	}
	// 容量太小的 buffer 不放回池中，让 GC 回收
}

// ============================================================================
// Protobuf 对象池 - 缓存反序列化后的 proto 对象
// ============================================================================

// BattleNotificationPool 用于缓存 BattleNotification 对象的对象池
type BattleNotificationPool struct {
	pool *sync.Pool
}

// NewBattleNotificationPool 创建一个 BattleNotification 对象池
func NewBattleNotificationPool() *BattleNotificationPool {
	return &BattleNotificationPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &proto_pb.BattleNotification{}
			},
		},
	}
}

// Get 从池中获取一个 BattleNotification 对象
func (p *BattleNotificationPool) Get() *proto_pb.BattleNotification {
	return p.pool.Get().(*proto_pb.BattleNotification)
}

// Put 将 BattleNotification 对象放回池中
func (p *BattleNotificationPool) Put(notif *proto_pb.BattleNotification) {
	if notif != nil {
		// 清空对象字段以帮助 GC
		notif.Reset()
		p.pool.Put(notif)
	}
}

// ProgressReportPool 用于缓存 ProgressReport 对象的对象池
type ProgressReportPool struct {
	pool *sync.Pool
}

// NewProgressReportPool 创建一个 ProgressReport 对象池
func NewProgressReportPool() *ProgressReportPool {
	return &ProgressReportPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &proto_pb.ProgressReport{}
			},
		},
	}
}

// Get 从池中获取一个 ProgressReport 对象
func (p *ProgressReportPool) Get() *proto_pb.ProgressReport {
	return p.pool.Get().(*proto_pb.ProgressReport)
}

// Put 将 ProgressReport 对象放回池中
func (p *ProgressReportPool) Put(report *proto_pb.ProgressReport) {
	if report != nil {
		report.Reset()
		p.pool.Put(report)
	}
}

// ============================================================================
// 全局池实例和单例
// ============================================================================

var (
	// 全局 proto buffer 池（用于序列化数据）
	globalProtoBufferPool *ProtoBufferPool
	bufferPoolOnce        sync.Once

	// 全局 BattleNotification 对象池
	globalBattleNotificationPool *BattleNotificationPool
	notificationPoolOnce         sync.Once

	// 全局 ProgressReport 对象池
	globalProgressReportPool *ProgressReportPool
	progressReportPoolOnce   sync.Once
)

// GetGlobalProtoBufferPool 获取全局 proto buffer 池（单例）
func GetGlobalProtoBufferPool() *ProtoBufferPool {
	bufferPoolOnce.Do(func() {
		// 根据常见 proto 数据大小设置初始容量（约 4KB）
		globalProtoBufferPool = NewProtoBufferPool(4096)
	})
	return globalProtoBufferPool
}

// GetGlobalBattleNotificationPool 获取全局 BattleNotification 对象池（单例）
func GetGlobalBattleNotificationPool() *BattleNotificationPool {
	notificationPoolOnce.Do(func() {
		globalBattleNotificationPool = NewBattleNotificationPool()
	})
	return globalBattleNotificationPool
}

// GetGlobalProgressReportPool 获取全局 ProgressReport 对象池（单例）
func GetGlobalProgressReportPool() *ProgressReportPool {
	progressReportPoolOnce.Do(func() {
		globalProgressReportPool = NewProgressReportPool()
	})
	return globalProgressReportPool
}
