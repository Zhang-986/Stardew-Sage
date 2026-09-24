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
        if (view.ModelUsage is not null)
        {
            ModelUsageSummary usage = view.ModelUsage;
            string tokens = usage.Session.TokensKnown
                ? usage.Session.TotalTokens.ToString(CultureInfo.InvariantCulture)
                : $"unknown ({usage.Session.ReportedTokenCalls}/{usage.Session.Calls} reported)";
            lines.Add($"AI session: calls {usage.Session.Calls}/{usage.CallBudget} · tokens {tokens}/{usage.TokenBudget}");
            lines.Add($"AI health: failure {usage.Session.Failed} · last {usage.Session.LastLatencyMs}ms · day calls {usage.DayTotals.Calls}");
        }
        foreach (TraitMemory trait in view.StableTraits.Take(2))
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
            int confidence = (int)Math.Round(view.LastDecision.PolicyConfidence * 100, MidpointRounding.AwayFromZero);
            lines.Add(string.Create(CultureInfo.InvariantCulture,
                $"决策：Echo -> {view.LastDecision.FinalAction.Kind} {target}（{confidence}%，备选 {view.LastDecision.SafeAlternatives.Count}，避开 {view.LastDecision.PlayerClaimedTargets.Count} 个目标）"));
        }
        foreach (PolicyExperience experience in view.Experiences.Take(2))
        {
            string source = experience.Source == ExperienceSource.Correction ? "玩家纠正" : "失败反思";
            string target = string.IsNullOrWhiteSpace(experience.PreferredTargetId) ? experience.PreferAction.ToString() : experience.PreferredTargetId;
            string effectiveness = string.Empty;
            if (experience.EffectiveConfidence > 0)
            {
                int confidence = (int)Math.Round(experience.EffectiveConfidence * 100, MidpointRounding.AwayFromZero);
                effectiveness = $"（有效 {confidence}% · 成功 {experience.SuccessCount} / 失败 {experience.FailureCount}）";
            }
            lines.Add($"经验·{source}：{experience.Summary} -> {target}{effectiveness}");
        }
        return Array.AsReadOnly(lines.Take(10).ToArray());
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
