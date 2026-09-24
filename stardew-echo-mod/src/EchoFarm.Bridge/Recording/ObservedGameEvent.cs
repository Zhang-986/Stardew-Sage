using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Recording;

public sealed class GameStateSample
{
    public int Energy { get; init; }
    public int Health { get; init; }
    public int Water { get; init; }
    public int InventoryCount { get; init; }
    public int MineFloor { get; init; }
}

public sealed class ObservedGameEvent
{
    public EventKind Kind { get; init; }
    public long Tick { get; init; }
    public Position Position { get; init; } = new();
    public string? Location { get; init; }
    public int TimeOfDay { get; init; }
    public string? TargetId { get; init; }
    public string? TargetKind { get; init; }
    public string? Tool { get; init; }
    public long DurationTicks { get; init; }
    public IReadOnlyList<ItemDelta> ItemDeltas { get; init; } = Array.Empty<ItemDelta>();
    public GameStateSample Before { get; init; } = new();
    public GameStateSample After { get; init; } = new();
    public bool Success { get; init; }
    public string? ErrorCode { get; init; }
}
