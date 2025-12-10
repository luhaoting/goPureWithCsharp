using System;
using System.Collections.Generic;
using GoPureWithCsharp.Battle;

namespace GoPureWithCsharp
{
    /// <summary>
    /// Protobuf 序列化缓冲池
    /// 用于缓存 protobuf 序列化产生的字节数据，减少内存申请
    /// </summary>
    public class ProtoBufferPool
    {
        private readonly Queue<byte[]> _buffers;
        private readonly int _bufferSize;
        private readonly object _lock = new object();

        public ProtoBufferPool(int initialCapacity = 10, int bufferSize = 4096)
        {
            _bufferSize = bufferSize;
            _buffers = new Queue<byte[]>(initialCapacity);

            // 预先创建缓冲区
            for (int i = 0; i < initialCapacity; i++)
            {
                _buffers.Enqueue(new byte[bufferSize]);
            }
        }

        /// <summary>
        /// 从池中获取一个缓冲区
        /// </summary>
        public byte[] Get()
        {
            lock (_lock)
            {
                if (_buffers.Count > 0)
                {
                    return _buffers.Dequeue();
                }
            }

            // 如果池中没有可用缓冲区，创建新的
            return new byte[_bufferSize];
        }

        /// <summary>
        /// 将缓冲区放回池中
        /// </summary>
        public void Put(byte[] buffer)
        {
            if (buffer == null || buffer.Length != _bufferSize)
                return;

            lock (_lock)
            {
                _buffers.Enqueue(buffer);
            }
        }

        /// <summary>
        /// 获取池中当前缓冲区数量
        /// </summary>
        public int Count => _buffers.Count;
    }

    /// <summary>
    /// BattleNotification 对象池
    /// </summary>
    public class BattleNotificationPool
    {
        private readonly Queue<BattleNotification> _objects;
        private readonly int _capacity;
        private readonly object _lock = new object();

        public BattleNotificationPool(int initialCapacity = 10)
        {
            _capacity = initialCapacity;
            _objects = new Queue<BattleNotification>(initialCapacity);

            for (int i = 0; i < initialCapacity; i++)
            {
                _objects.Enqueue(new BattleNotification());
            }
        }

        public BattleNotification Get()
        {
            lock (_lock)
            {
                if (_objects.Count > 0)
                {
                    return _objects.Dequeue();
                }
            }
            return new BattleNotification();
        }

        public void Put(BattleNotification obj)
        {
            if (obj == null)
                return;

            lock (_lock)
            {
                if (_objects.Count < _capacity)
                {
                    _objects.Enqueue(obj);
                }
            }
        }

        public int Count => _objects.Count;
    }

    /// <summary>
    /// ProgressReport 对象池
    /// </summary>
    public class ProgressReportPool
    {
        private readonly Queue<ProgressReport> _objects;
        private readonly int _capacity;
        private readonly object _lock = new object();

        public ProgressReportPool(int initialCapacity = 10)
        {
            _capacity = initialCapacity;
            _objects = new Queue<ProgressReport>(initialCapacity);

            for (int i = 0; i < initialCapacity; i++)
            {
                _objects.Enqueue(new ProgressReport());
            }
        }

        public ProgressReport Get()
        {
            lock (_lock)
            {
                if (_objects.Count > 0)
                {
                    return _objects.Dequeue();
                }
            }
            return new ProgressReport();
        }

        public void Put(ProgressReport obj)
        {
            if (obj == null)
                return;

            lock (_lock)
            {
                if (_objects.Count < _capacity)
                {
                    _objects.Enqueue(obj);
                }
            }
        }

        public int Count => _objects.Count;
    }

    /// <summary>
    /// 全局 Protobuf 对象池管理器
    /// 提供单例的对象池访问
    /// </summary>
    public static class ProtoPoolManager
    {
        private static ProtoBufferPool? _bufferPool;
        private static BattleNotificationPool? _notificationPool;
        private static ProgressReportPool? _progressReportPool;

        public static ProtoBufferPool GetBufferPool()
        {
            if (_bufferPool is null)
            {
                _bufferPool = new ProtoBufferPool(20, 4096);
            }
            return _bufferPool;
        }

        public static BattleNotificationPool GetNotificationPool()
        {
            if (_notificationPool is null)
            {
                _notificationPool = new BattleNotificationPool(20);
            }
            return _notificationPool;
        }

        public static ProgressReportPool GetProgressReportPool()
        {
            if (_progressReportPool is null)
            {
                _progressReportPool = new ProgressReportPool(20);
            }
            return _progressReportPool;
        }
    }
}
