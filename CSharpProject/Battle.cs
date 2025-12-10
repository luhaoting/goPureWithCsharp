using System;
using GoPureWithCsharp.Battle;

namespace GoPureWithCsharp
{
    /// <summary>
    /// 单场战斗实例
    /// </summary>
    public class BattleInstance
    {
        private static readonly Random _random = new Random();

        public uint BattleId { get; private set; }
        public uint AtkTeamId { get; private set; }
        public uint DefTeamId { get; private set; }
        public int AtkHealth { get; private set; }
        public int DefHealth { get; private set; }
        public int CurrentRound { get; private set; }
        public bool IsFinished { get; private set; }
        public uint? Winner { get; private set; }

        /// <summary>
        /// 创建战斗实例
        /// </summary>
        public BattleInstance(uint battleId, uint atkTeamId, uint defTeamId, int initialHealth)
        {
            BattleId = battleId;
            AtkTeamId = atkTeamId;
            DefTeamId = defTeamId;
            AtkHealth = initialHealth;
            DefHealth = initialHealth;
            CurrentRound = 0;
            IsFinished = false;
            Winner = null;
        }

        /// <summary>
        /// 执行一回合战斗
        /// </summary>
        public void ExecuteRound(int minDamage, int maxDamage)
        {
            if (IsFinished) return;

            CurrentRound++;

            // ATK 攻击 DEF
            int atkDamage = _random.Next(minDamage, maxDamage + 1);
            DefHealth -= atkDamage;
            BattleLogger.Debug($"[Battle {BattleId}] Round {CurrentRound}: ATK={AtkTeamId} 攻击 DEF={DefTeamId}, 伤害={atkDamage}, DEF 剩余血量={DefHealth}");

            // 检查 DEF 是否死亡
            if (DefHealth <= 0)
            {
                IsFinished = true;
                Winner = AtkTeamId;
                BattleLogger.Info($"[Battle {BattleId}] DEF={DefTeamId} 死亡, ATK={AtkTeamId} 获胜!");
                return;
            }

            // DEF 反击 ATK
            int defDamage = _random.Next(minDamage, maxDamage + 1);
            AtkHealth -= defDamage;
            BattleLogger.Debug($"[Battle {BattleId}] Round {CurrentRound}: DEF={DefTeamId} 反击 ATK={AtkTeamId}, 伤害={defDamage}, ATK 剩余血量={AtkHealth}");

            // 检查 ATK 是否死亡
            if (AtkHealth <= 0)
            {
                IsFinished = true;
                Winner = DefTeamId;
                BattleLogger.Info($"[Battle {BattleId}] ATK={AtkTeamId} 死亡, DEF={DefTeamId} 获胜!");
            }
        }
    }
}
