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
}

public sealed class ActionResponse : StrictContract
{
    public HighLevelAction Action { get; init; } = new();
}

public sealed class ActionResultRequest : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public WorldSnapshot Snapshot { get; init; } = new();
    public ActionResult Result { get; init; } = new();
}
