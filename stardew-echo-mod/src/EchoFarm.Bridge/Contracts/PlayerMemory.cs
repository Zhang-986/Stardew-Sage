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

public sealed class PlayerModel : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public int Revision { get; init; }
    public IReadOnlyList<BehaviorKind> CommonTaskOrder { get; init; } = Array.Empty<BehaviorKind>();
    public string? PreferredChestId { get; init; }
    public int EnergyReserve { get; init; }
    public string? RouteStyle { get; init; }
    public IReadOnlyList<ObservedPreference> Preferences { get; init; } = Array.Empty<ObservedPreference>();
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
