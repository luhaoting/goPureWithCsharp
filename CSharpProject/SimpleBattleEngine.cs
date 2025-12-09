using System;
using System.Collections.Generic;
using Google.Protobuf;
using GoPureWithCsharp.Battle;

namespace GoPureWithCsharp
{
    /// <summary>
    /// 简单的战斗引擎 Demo
    /// 模拟战斗过程、记录事件、生成回放
    /// </summary>
    public class SimpleBattleEngine
    {
        private static readonly Random _random = new Random();

        /// <summary>
        /// 执行战斗
        /// </summary>
        /// <param name="request">战斗请求</param>
        /// <returns>战斗结果</returns>
        public static BattleResult ExecuteBattle(StartBattle request)
        {
            Console.WriteLine($"[Battle] 开始战斗 ID={request.BattleId}, ATK={request.Atk.TeamId}, DEF={request.Def.TeamId}");

            // 初始化战斗状态
            var events = new List<BattleEvent>();
            long startTime = DateTimeOffset.Now.ToUnixTimeMilliseconds();

            // Demo: 模拟 3 回合战斗
            int atkHealth = 300;
            int defHealth = 300;

            for (int round = 1; round <= 3; round++)
            {
                // ATK 攻击 DEF
                int atkDamage = _random.Next(20, 50);
                defHealth -= atkDamage;
                events.Add(new BattleEvent
                {
                    Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                    EventType = "attack",
                    PerformerId = request.Atk.TeamId,
                    TargetId = request.Def.TeamId,
                    Value = atkDamage,
                });

                // DEF 反击 ATK
                int defDamage = _random.Next(15, 40);
                atkHealth -= defDamage;
                events.Add(new BattleEvent
                {
                    Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                    EventType = "attack",
                    PerformerId = request.Def.TeamId,
                    TargetId = request.Atk.TeamId,
                    Value = defDamage,
                });

                Console.WriteLine($"[Battle] 回合 {round}: ATK={atkHealth} HP, DEF={defHealth} HP");
            }

            // 确定胜负
            long endTime = DateTimeOffset.Now.ToUnixTimeMilliseconds();
            uint winner = atkHealth > defHealth ? request.Atk.TeamId : request.Def.TeamId;
            uint loser = atkHealth > defHealth ? request.Def.TeamId : request.Atk.TeamId;

            events.Add(new BattleEvent
            {
                Timestamp = endTime,
                EventType = "end",
                PerformerId = winner,
                TargetId = loser,
                Value = 1,
            });

            var result = new BattleResult
            {
                Winner = winner,
                Loser = loser,
                AtkDamage = 300 - atkHealth,
                DefDamage = 300 - defHealth,
                Duration = endTime - startTime,
                BattleScore = (300 - defHealth) * 10,
            };

            // 生成回放
            var replay = new BattleReplay
            {
                BattleId = request.BattleId,
                StartTime = startTime,
                EndTime = endTime,
                AtkTeam = request.Atk,
                DefTeam = request.Def,
                Result = result,
                Version = "1.0",
            };

            foreach (var evt in events)
            {
                replay.Events.Add(evt);
            }

            // 通过回调通知 Go (可选)
            NotifyBattleCompleted(replay);

            Console.WriteLine($"[Battle] 战斗结束，胜方={winner}, 积分={result.BattleScore}");

            return result;
        }

        /// <summary>
        /// 通知 Go 战斗完成
        /// </summary>
        private static void NotifyBattleCompleted(BattleReplay replay)
        {
            // 构建通知消息
            var notification = new BattleNotification
            {
                Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                NotificationType = NotificationType.BattleCompleted,
                BattleId = replay.BattleId,
                Payload = replay.ToByteString(), // 序列化回放数据
            };

            // 通过回调发送给 Go
            byte[] notificationBytes = notification.ToByteArray();
            BattleCallbackManager.NotifyBattle(notificationBytes);
        }

        /// <summary>
        /// 执行批量战斗
        /// </summary>
        public static BatchBattleResponse ExecuteBatchBattle(BatchBattleRequest batchRequest)
        {
            Console.WriteLine($"[Battle] 开始批量战斗 ID={batchRequest.BatchId}, 数量={batchRequest.Battles.Count}");

            var response = new BatchBattleResponse
            {
                BatchId = batchRequest.BatchId,
            };

            long totalStartTime = DateTimeOffset.Now.ToUnixTimeMilliseconds();

            foreach (var battleReq in batchRequest.Battles)
            {
                try
                {
                    var result = ExecuteBattle(battleReq);
                    response.Results.Add(result);
                    response.SuccessCount++;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[Battle] 战斗执行失败: {ex.Message}");
                    response.FailureCount++;
                }
            }

            long totalEndTime = DateTimeOffset.Now.ToUnixTimeMilliseconds();
            response.TotalDuration = totalEndTime - totalStartTime;

            Console.WriteLine($"[Battle] 批量战斗完成: 成功={response.SuccessCount}, 失败={response.FailureCount}");

            return response;
        }
    }
}
