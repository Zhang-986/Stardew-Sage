namespace EchoFarm.Bridge.Contracts;

public enum BehaviorKind
{
    Watering,
    Refilling,
    Harvesting,
    Depositing
}

public enum PreferenceKey
{
    TaskOrder,
    PreferredChest,
    EnergyReserve,
    RouteStyle
}

public sealed class ObservedPreference : StrictContract
{
    public PreferenceKey Key { get; init; }
    public string Value { get; init; } = string.Empty;
    public IReadOnlyList<string> EvidenceEventIds { get; init; } = Array.Empty<string>();
    public int ObservationCount { get; init; }
    public double Confidence { get; init; }
}

public enum TraitContext
{
    Any,
    Sunny,
    Rainy,
    Storm,
    Snow
}

public sealed class TraitMemory : StrictContract
{
    public PreferenceKey Key { get; init; }
    public string Value { get; init; } = string.Empty;
    public TraitContext Context { get; init; }
    public double Confidence { get; init; }
    public int ObservationCount { get; init; }
    public int ContradictionCount { get; init; }
    public int FirstSeenDay { get; init; }
    public int LastSeenDay { get; init; }
    public IReadOnlyList<string> EvidenceRefs { get; init; } = Array.Empty<string>();
}

public enum LearningChangeKind
{
    Added,
    Strengthened,
    Weakened,
    Unchanged
}

public sealed class LearningChange : StrictContract
{
    public int ModelRevision { get; init; }
    public LearningChangeKind Kind { get; init; }
    public PreferenceKey? Key { get; init; }
    public string? Value { get; init; }
    public double PreviousConfidence { get; init; }
    public double Confidence { get; init; }
    public string Summary { get; init; } = string.Empty;
}

public sealed class PlayerModel : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public int Revision { get; init; }
    public int LearnedThroughDay { get; init; }
    public IReadOnlyList<BehaviorKind> CommonTaskOrder { get; init; } = Array.Empty<BehaviorKind>();
    public string? PreferredChestId { get; init; }
    public int EnergyReserve { get; init; }
    public string? RouteStyle { get; init; }
    public IReadOnlyList<ObservedPreference> Preferences { get; init; } = Array.Empty<ObservedPreference>();
    public IReadOnlyList<TraitMemory> Traits { get; init; } = Array.Empty<TraitMemory>();
}

public sealed class SkillStep : StrictContract
{
    public ActionKind Action { get; init; }
    public string? TargetSelector { get; init; }
}

public sealed class RecoveryStrategy : StrictContract
{
    public string FailureCode { get; init; } = string.Empty;
    public IReadOnlyList<ActionKind> Actions { get; init; } = Array.Empty<ActionKind>();
}

public sealed class SkillProgram : StrictContract
{
    public string Name { get; init; } = string.Empty;
    public int Revision { get; init; }
    public string Goal { get; init; } = string.Empty;
    public IReadOnlyList<string> Preconditions { get; init; } = Array.Empty<string>();
    public string TargetSelector { get; init; } = string.Empty;
    public IReadOnlyList<BehaviorKind> PreferredOrder { get; init; } = Array.Empty<BehaviorKind>();
    public IReadOnlyList<SkillStep> Steps { get; init; } = Array.Empty<SkillStep>();
    public IReadOnlyList<string> SuccessConditions { get; init; } = Array.Empty<string>();
    public IReadOnlyList<string> StopConditions { get; init; } = Array.Empty<string>();
    public IReadOnlyList<RecoveryStrategy> RecoveryStrategies { get; init; } = Array.Empty<RecoveryStrategy>();
    public IReadOnlyList<string> EvidenceEventIds { get; init; } = Array.Empty<string>();
}

public enum PlayerIntent
{
    Unknown,
    Watering,
    Harvesting,
    Depositing
}

public enum ExperienceTrigger
{
    InventoryFull,
    OutOfWater,
    PathBlocked,
    ChestFull,
    PlayerCorrection
}

public enum SituationSignal
{
    InventoryFull,
    InventoryHasItems,
    CanEmpty,
    Raining,
    TargetBlocked
}

public enum ExperienceSource
{
    Failure,
    Correction
}

public sealed class PolicyExperience : StrictContract
{
    public string Id { get; init; } = string.Empty;
    public string SaveId { get; init; } = string.Empty;
    public ExperienceTrigger Trigger { get; init; }
    public TraitContext Context { get; init; }
    public IReadOnlyList<SituationSignal> WhenSignals { get; init; } = Array.Empty<SituationSignal>();
    public ActionKind? AvoidAction { get; init; }
    public ActionKind PreferAction { get; init; }
    public string? PreferredTargetId { get; init; }
    public string Summary { get; init; } = string.Empty;
    public double Confidence { get; init; }
    public double EffectiveConfidence { get; init; }
    public int SuccessCount { get; init; }
    public int FailureCount { get; init; }
    public int NeutralCount { get; init; }
    public int ObservationCount { get; init; }
    public int ContradictionCount { get; init; }
    public int FirstSeenDay { get; init; }
    public int LastSeenDay { get; init; }
    public IReadOnlyList<string> EvidenceRefs { get; init; } = Array.Empty<string>();
    public ExperienceSource Source { get; init; }
}

public sealed class DecisionRecord : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public string SessionId { get; init; } = string.Empty;
    public long SnapshotVersion { get; init; }
    public int Day { get; init; }
    public int ModelRevision { get; init; }
    public PlayerIntent InferredIntent { get; init; }
    public IReadOnlyList<string> PlayerClaimedTargets { get; init; } = Array.Empty<string>();
    public HighLevelAction CandidateAction { get; init; } = new();
    public HighLevelAction FinalAction { get; init; } = new();
    public ActionProposal? Proposal { get; init; }
    public double PolicyConfidence { get; init; }
    public int SelectedCandidate { get; init; }
    public IReadOnlyList<HighLevelAction> SafeAlternatives { get; init; } = Array.Empty<HighLevelAction>();
    public ActionResult? Result { get; init; }
}

public sealed class EchoSessionMemory : StrictContract
{
    public string SessionId { get; init; } = string.Empty;
    public int Day { get; init; }
    public string Status { get; init; } = string.Empty;
}

public sealed class ModelUsageTotals : StrictContract
{
    public int Calls { get; init; }
    public int Succeeded { get; init; }
    public int Failed { get; init; }
    public int ReportedTokenCalls { get; init; }
    public int PromptTokens { get; init; }
    public int CompletionTokens { get; init; }
    public int TotalTokens { get; init; }
    public bool TokensKnown { get; init; }
    public long LastLatencyMs { get; init; }
    public long AverageLatencyMs { get; init; }
}

public sealed class ModelUsageSummary : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public string SessionId { get; init; } = string.Empty;
    public int Day { get; init; }
    public ModelUsageTotals Session { get; init; } = new();
    public ModelUsageTotals DayTotals { get; init; } = new();
    public int CallBudget { get; init; }
    public int TokenBudget { get; init; }
    public bool BudgetExhausted { get; init; }
}

public sealed class EchoMemoryView : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public int ModelRevision { get; init; }
    public int LearnedThroughDay { get; init; }
    public IReadOnlyList<TraitMemory> StableTraits { get; init; } = Array.Empty<TraitMemory>();
    public LearningChange? RecentLearningChange { get; init; }
    public EchoSessionMemory? ActiveSession { get; init; }
    public DecisionRecord? LastDecision { get; init; }
    public IReadOnlyList<PolicyExperience> Experiences { get; init; } = Array.Empty<PolicyExperience>();
    public ModelUsageSummary? ModelUsage { get; init; }
}
