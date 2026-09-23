using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class ActionSafetyGateTests
{
    [Fact]
    public void AcceptsCurrentDryCrop()
    {
        WorldSnapshot snapshot = Snapshot();
        var action = Action(snapshot, ActionKind.WaterTarget, "crop-dry");

        new ActionSafetyGate().EnsureSafe(action, snapshot);
    }

    [Theory]
    [InlineData("stale_snapshot")]
    [InlineData("unknown_target")]
    [InlineData("rain")]
    [InlineData("empty_can")]
    [InlineData("immature_crop")]
    [InlineData("full_can")]
    [InlineData("missing_chest")]
    public void RejectsUnsafeActions(string scenario)
    {
        WorldSnapshot snapshot = Snapshot();
        HighLevelAction action;
        switch (scenario)
        {
            case "stale_snapshot":
                action = Action(snapshot, ActionKind.WaterTarget, "crop-dry", snapshot.SnapshotVersion - 1);
                break;
            case "unknown_target":
                action = Action(snapshot, ActionKind.WaterTarget, "crop-missing");
                break;
            case "rain":
                snapshot = Copy(snapshot, weather: Weather.Rainy);
                action = Action(snapshot, ActionKind.WaterTarget, "crop-dry");
                break;
            case "empty_can":
                snapshot = Copy(snapshot, water: 0);
                action = Action(snapshot, ActionKind.WaterTarget, "crop-dry");
                break;
            case "immature_crop":
                action = Action(snapshot, ActionKind.HarvestTarget, "crop-dry");
                break;
            case "full_can":
                snapshot = Copy(snapshot, water: 40);
                action = Action(snapshot, ActionKind.RefillCan, "pond-1");
                break;
            case "missing_chest":
                action = Action(snapshot, ActionKind.DepositItems, "chest-missing");
                break;
            default:
                throw new ArgumentOutOfRangeException(nameof(scenario));
        }

        Assert.Throws<UnsafeActionException>(() => new ActionSafetyGate().EnsureSafe(action, snapshot));
    }

    internal static WorldSnapshot Snapshot() => new()
    {
        SaveId = "farm-1", SessionId = "echo-day-2", SnapshotVersion = 4, Day = 2, TimeOfDay = 620,
        Weather = Weather.Sunny, Location = "Farm", Energy = 200, MaxEnergy = 270,
        WateringCan = new ToolState { Name = "Watering Can", Water = 10, Capacity = 40 },
        Crops = new[]
        {
            new Crop { Id = "crop-dry", NeedsWater = true },
            new Crop { Id = "crop-ripe", Mature = true }
        },
        WaterSources = new[] { new WaterSource { Id = "pond-1" } },
        Chests = new[] { new Chest { Id = "chest-1" } }
    };

    internal static HighLevelAction Action(WorldSnapshot snapshot, ActionKind kind, string? targetId, long? version = null) => new()
    {
        SaveId = snapshot.SaveId,
        SessionId = snapshot.SessionId,
        SnapshotVersion = version ?? snapshot.SnapshotVersion,
        Kind = kind,
        TargetId = targetId,
        Reason = "test"
    };

    private static WorldSnapshot Copy(WorldSnapshot source, Weather? weather = null, int? water = null) => new()
    {
        SaveId = source.SaveId,
        SessionId = source.SessionId,
        SnapshotVersion = source.SnapshotVersion,
        Day = source.Day,
        TimeOfDay = source.TimeOfDay,
        Weather = weather ?? source.Weather,
        Location = source.Location,
        PlayerPosition = source.PlayerPosition,
        Energy = source.Energy,
        MaxEnergy = source.MaxEnergy,
        Inventory = source.Inventory,
        WateringCan = new ToolState
        {
            Name = source.WateringCan.Name,
            Level = source.WateringCan.Level,
            Water = water ?? source.WateringCan.Water,
            Capacity = source.WateringCan.Capacity
        },
        Crops = source.Crops,
        WaterSources = source.WaterSources,
        Chests = source.Chests,
        Obstacles = source.Obstacles
    };
}
