namespace EchoFarm.Bridge.Contracts;

public enum ActionKind
{
    MoveTo,
    EquipTool,
    WaterTarget,
    RefillCan,
    HarvestTarget,
    DepositItems,
    StopSession
}

public enum ActionStatus
{
    Succeeded,
    Failed
}

public enum UncertaintyCode
{
    MissingExperience,
    ConflictingEvidence,
    NovelContext,
    AmbiguousTarget
}

public sealed class HighLevelAction : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public string SessionId { get; init; } = string.Empty;
    public long SnapshotVersion { get; init; }
    public ActionKind Kind { get; init; }
    public string? TargetId { get; init; }
    public Position? Destination { get; init; }
    public string Reason { get; init; } = string.Empty;
}

public sealed class ActionResult : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public string SessionId { get; init; } = string.Empty;
    public long SnapshotVersion { get; init; }
    public HighLevelAction Action { get; init; } = new();
    public ActionStatus Status { get; init; }
    public string? ErrorCode { get; init; }
}

public sealed class LearnResponse : StrictContract
{
    public PlayerModel PlayerModel { get; init; } = new();
    public SkillProgram Skill { get; init; } = new();
    public LearningChange LearningChange { get; init; } = new();
}

public sealed class ActionResponse : StrictContract
{
    public HighLevelAction Action { get; init; } = new();
    public double Confidence { get; init; }
    public IReadOnlyList<HighLevelAction> Alternatives { get; init; } = Array.Empty<HighLevelAction>();
    public IReadOnlyList<string> AppliedExperiences { get; init; } = Array.Empty<string>();
}

public sealed class ActionProposal : StrictContract
{
    public HighLevelAction Primary { get; init; } = new();
    public IReadOnlyList<HighLevelAction> Alternatives { get; init; } = Array.Empty<HighLevelAction>();
    public double ModelConfidence { get; init; }
    public IReadOnlyList<UncertaintyCode> UncertaintyCodes { get; init; } = Array.Empty<UncertaintyCode>();
    public IReadOnlyList<string> AppliedExperienceIds { get; init; } = Array.Empty<string>();
}

public sealed class ActionResultRequest : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public WorldSnapshot Snapshot { get; init; } = new();
    public ActionResult Result { get; init; } = new();
}

public sealed class PlayerCorrection : StrictContract
{
    public string Id { get; init; } = string.Empty;
    public string SaveId { get; init; } = string.Empty;
    public string SessionId { get; init; } = string.Empty;
    public long RejectedDecisionSnapshotVersion { get; init; }
    public HighLevelAction RejectedAction { get; init; } = new();
    public WorldSnapshot Snapshot { get; init; } = new();
    public HighLevelAction PreferredAction { get; init; } = new();
    public long ObservedAtTick { get; init; }
}

public sealed class CorrectionResponse : StrictContract
{
    public PolicyExperience Experience { get; init; } = new();
}
