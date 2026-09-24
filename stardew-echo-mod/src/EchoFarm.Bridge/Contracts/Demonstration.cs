namespace EchoFarm.Bridge.Contracts;

public enum EventKind
{
    Move,
    EquipTool,
    WaterTarget,
    RefillCan,
    HarvestTarget,
    DepositItems,
    ChopTree,
    BreakRock,
    EnterMineFloor,
    FishCaught,
    FishEscaped
}

public sealed class StateDelta : StrictContract
{
    public int EnergyDelta { get; init; }
    public int WaterDelta { get; init; }
    public int InventoryDelta { get; init; }
    public int HealthDelta { get; init; }
    public int MineFloorDelta { get; init; }
}

public sealed class ItemDelta : StrictContract
{
    public string ItemId { get; init; } = string.Empty;
    public string? Name { get; init; }
    public int Quantity { get; init; }
}

public sealed class DemonstrationEvent : StrictContract
{
    public string Id { get; init; } = string.Empty;
    public EventKind Kind { get; init; }
    public long Tick { get; init; }
    public Position Position { get; init; } = new();
    public string? Location { get; init; }
    public int TimeOfDay { get; init; }
    public string? TargetId { get; init; }
    public string? TargetKind { get; init; }
    public string? Tool { get; init; }
    public long DurationTicks { get; init; }
    public StateDelta Delta { get; init; } = new();
    public IReadOnlyList<ItemDelta> ItemDeltas { get; init; } = Array.Empty<ItemDelta>();
    public bool Success { get; init; }
    public string? ErrorCode { get; init; }
}

public sealed class Demonstration : StrictContract
{
    public int SchemaVersion { get; init; }
    public string Id { get; init; } = string.Empty;
    public string SaveId { get; init; } = string.Empty;
    public string SessionId { get; init; } = string.Empty;
    public int Day { get; init; }
    public Weather Weather { get; init; }
    public long StartedAt { get; init; }
    public long EndedAt { get; init; }
    public IReadOnlyList<DemonstrationEvent> Events { get; init; } = Array.Empty<DemonstrationEvent>();
}
