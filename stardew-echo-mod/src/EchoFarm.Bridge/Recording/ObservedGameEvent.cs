using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Recording;

public sealed class GameStateSample
{
    public int Energy { get; init; }
    public int Water { get; init; }
    public int InventoryCount { get; init; }
}

public sealed class ObservedGameEvent
{
    public EventKind Kind { get; init; }
    public long Tick { get; init; }
    public Position Position { get; init; } = new();
    public string? TargetId { get; init; }
    public string? Tool { get; init; }
    public GameStateSample Before { get; init; } = new();
    public GameStateSample After { get; init; } = new();
    public bool Success { get; init; }
    public string? ErrorCode { get; init; }
}
