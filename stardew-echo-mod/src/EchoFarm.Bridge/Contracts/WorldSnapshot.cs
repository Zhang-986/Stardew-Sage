namespace EchoFarm.Bridge.Contracts;

public enum Weather
{
    Sunny,
    Rainy,
    Storm,
    Snow
}

public sealed class Position : StrictContract
{
    public int X { get; init; }
    public int Y { get; init; }
}

public sealed class Crop : StrictContract
{
    public string Id { get; init; } = string.Empty;
    public Position Position { get; init; } = new();
    public string? Kind { get; init; }
    public int GrowthStage { get; init; }
    public bool Mature { get; init; }
    public bool NeedsWater { get; init; }
}

public sealed class WaterSource : StrictContract
{
    public string Id { get; init; } = string.Empty;
    public Position Position { get; init; } = new();
}

public sealed class Chest : StrictContract
{
    public string Id { get; init; } = string.Empty;
    public string? Label { get; init; }
    public Position Position { get; init; } = new();
}

public sealed class ToolState : StrictContract
{
    public string Name { get; init; } = string.Empty;
    public int Level { get; init; }
    public int Water { get; init; }
    public int Capacity { get; init; }
}

public sealed class InventoryItem : StrictContract
{
    public string ItemId { get; init; } = string.Empty;
    public string Name { get; init; } = string.Empty;
    public int Quantity { get; init; }
}

public sealed class InventorySummary : StrictContract
{
    public int FreeSlots { get; init; }
    public IReadOnlyList<InventoryItem> Items { get; init; } = Array.Empty<InventoryItem>();
}

public sealed class PlayerActivity : StrictContract
{
    public EventKind Kind { get; init; }
    public string TargetId { get; init; } = string.Empty;
    public long Tick { get; init; }
    public bool Success { get; init; }
}

public sealed class WorldSnapshot : StrictContract
{
    public string SaveId { get; init; } = string.Empty;
    public string SessionId { get; init; } = string.Empty;
    public long SnapshotVersion { get; init; }
    public long Tick { get; init; }
    public int Day { get; init; }
    public int TimeOfDay { get; init; }
    public Weather Weather { get; init; }
    public string Location { get; init; } = string.Empty;
    public Position PlayerPosition { get; init; } = new();
    public int Energy { get; init; }
    public int MaxEnergy { get; init; }
    public InventorySummary Inventory { get; init; } = new();
    public ToolState WateringCan { get; init; } = new();
    public IReadOnlyList<Crop> Crops { get; init; } = Array.Empty<Crop>();
    public IReadOnlyList<WaterSource> WaterSources { get; init; } = Array.Empty<WaterSource>();
    public IReadOnlyList<Chest> Chests { get; init; } = Array.Empty<Chest>();
    public IReadOnlyList<Position> Obstacles { get; init; } = Array.Empty<Position>();
    public IReadOnlyList<PlayerActivity> RecentPlayerActions { get; init; } = Array.Empty<PlayerActivity>();
}
