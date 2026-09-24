using System.Globalization;
using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Runtime;

public static class EchoMemoryPresenter
{
    public static IReadOnlyList<string> BuildLines(EchoMemoryView view)
    {
        ArgumentNullException.ThrowIfNull(view);
        var lines = new List<string>(8)
        {
            $"Echo v{view.ModelRevision} · 学习到第 {view.LearnedThroughDay} 天"
        };
        foreach (TraitMemory trait in view.StableTraits.Take(4))
        {
            int confidence = (int)Math.Round(trait.Confidence * 100, MidpointRounding.AwayFromZero);
            lines.Add($"习惯 {TraitName(trait.Key)}：{trait.Value}（{confidence}%）");
        }
        if (view.RecentLearningChange is not null)
            lines.Add($"最近学习：{view.RecentLearningChange.Summary}");
        if (view.LastDecision is not null)
        {
            lines.Add($"你正在：{IntentName(view.LastDecision.InferredIntent)}");
            string target = view.LastDecision.FinalAction.TargetId ?? "结束本轮";
            lines.Add(string.Create(CultureInfo.InvariantCulture,
                $"分工：Echo -> {view.LastDecision.FinalAction.Kind} {target}（避开 {view.LastDecision.PlayerClaimedTargets.Count} 个目标）"));
        }
        return Array.AsReadOnly(lines.Take(8).ToArray());
    }

    private static string TraitName(PreferenceKey key) => key switch
    {
        PreferenceKey.TaskOrder => "任务顺序",
        PreferenceKey.PreferredChest => "常用箱子",
        PreferenceKey.EnergyReserve => "体力保留",
        PreferenceKey.RouteStyle => "行动风格",
        _ => key.ToString()
    };

    private static string IntentName(PlayerIntent intent) => intent switch
    {
        PlayerIntent.Watering => "浇水",
        PlayerIntent.Harvesting => "收获",
        PlayerIntent.Depositing => "存箱",
        _ => "观察农场"
    };
}
