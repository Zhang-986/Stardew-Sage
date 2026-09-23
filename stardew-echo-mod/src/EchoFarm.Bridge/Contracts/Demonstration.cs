namespace EchoFarm.Bridge.Contracts;

public enum EventKind
{
    Move,
    EquipTool,
    WaterTarget,
    RefillCan,
    HarvestTarget,
    DepositItems
}

public sealed class StateDelta : StrictContract
{
    public int EnergyDelta { get; init; }
    public int WaterDelta { get; init; }
    public int InventoryDelta { get; init; }
}

public sealed class DemonstrationEvent : StrictContract
{
    public string Id { get; init; } = string.Empty;
    public EventKind Kind { get; init; }
    public long Tick { get; init; }
    public Position Position { get; init; } = new();
    public string? TargetId { get; init; }
    public string? Tool { get; init; }
    public StateDelta Delta { get; init; } = new();
    public bool Success { get; init; }
    public string? ErrorCode { get; init; }
}

public sealed class Demonstration : StrictContract
{
    public string Id { get; init; } = string.Empty;
    public string SaveId { get; init; } = string.Empty;
    public string SessionId { get; init; } = string.Empty;
    public long StartedAt { get; init; }
    public long EndedAt { get; init; }
    public IReadOnlyList<DemonstrationEvent> Events { get; init; } = Array.Empty<DemonstrationEvent>();
}
