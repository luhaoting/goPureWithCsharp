using System;
using GoPureWithCsharp.Battle;

namespace GoPureWithCsharp
{
    /// <summary>
    /// 战斗输入 - 玩家或 AI 的操作指令
    /// </summary>
    public class BattleInput
    {
        public uint BattleId { get; set; }
        public uint TeamId { get; set; }
        public byte ActionType { get; set; } // 0=attack, 1=defend, 2=skill
        public int ActionValue { get; set; } // 技能参数或伤害修饰
    }

    /// <summary>
    /// 战斗输入处理器
    /// </summary>
    public static class BattleInputHandler
    {
        /// <summary>
        /// 处理战斗输入
        /// </summary>
        public static int ProcessBattleInput(uint battleId, uint teamId, byte actionType, int actionValue)
        {
            var battle = BattleManager.GetBattle(battleId);
            if (battle == null)
            {
                return -1; // 战斗不存在
            }

            if (battle.IsFinished)
            {
                return -2; // 战斗已结束
            }

            // 根据操作类型处理
            switch (actionType)
            {
                case 0: // Attack
                    // 直接攻击不需要额外处理，ExecuteRound 已经随机生成伤害
                    return 0;
                case 1: // Defend
                    // 防守逻辑：降低本回合伤害
                    return 0;
                case 2: // Skill
                    // 技能逻辑：自定义伤害或效果
                    return 0;
                default:
                    return -3; // 无效操作类型
            }
        }
    }
}
