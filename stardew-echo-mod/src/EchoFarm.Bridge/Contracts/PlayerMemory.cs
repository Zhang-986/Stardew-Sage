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
    public ActionResult? Result { get; init; }
}

public sealed class EchoSessionMemory : StrictContract
{
    public string SessionId { get; init; } = string.Empty;
    public int Day { get; init; }
    public string Status { get; init; } = string.Empty;
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
}
