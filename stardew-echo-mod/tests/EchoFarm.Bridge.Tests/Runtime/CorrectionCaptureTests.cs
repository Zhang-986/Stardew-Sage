using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;
using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class CorrectionCaptureTests
{
    [Fact]
    public void CapturesOnlyTheFirstSuccessfulSupportedPlayerAction()
    {
        WorldSnapshot snapshot = Snapshot();
        HighLevelAction rejected = ActionSafetyGateTests.Action(snapshot, ActionKind.DepositItems, "chest-east", 3);
        var capture = new CorrectionCapture(() => "correction-1", correctionWindowTicks: 1200);
        capture.Arm(rejected, currentTick: 100);

        Assert.False(capture.TryCapture(snapshot, Observed(EventKind.Move, "tile-1", 101, success: true), out _));
        Assert.False(capture.TryCapture(snapshot, Observed(EventKind.DepositItems, "chest-west", 102, success: false), out _));
        Assert.True(capture.TryCapture(snapshot, Observed(EventKind.DepositItems, "chest-west", 103, success: true), out PlayerCorrection? correction));

        Assert.NotNull(correction);
        Assert.Equal("correction-1", correction!.Id);
        Assert.Equal(3, correction.RejectedDecisionSnapshotVersion);
        Assert.Equal(ActionKind.DepositItems, correction.PreferredAction.Kind);
        Assert.Equal("chest-west", correction.PreferredAction.TargetId);
        Assert.False(capture.TryCapture(snapshot, Observed(EventKind.WaterTarget, "crop-dry", 104, success: true), out _));
    }

    [Fact]
    public void ExpiresAndCancelsWithoutCreatingEvidence()
    {
        WorldSnapshot snapshot = Snapshot();
        var capture = new CorrectionCapture(() => "correction-1", correctionWindowTicks: 1200);
        capture.Arm(ActionSafetyGateTests.Action(snapshot, ActionKind.DepositItems, "chest-east", 3), currentTick: 100);

        Assert.False(capture.Expire(currentTick: 1299));
        Assert.True(capture.Expire(currentTick: 1300));
        Assert.False(capture.IsArmed);
        Assert.Null(capture.PendingCorrection);
    }

    private static WorldSnapshot Snapshot() => new()
    {
        SaveId = "farm-1",
        SessionId = "echo-day-2",
        SnapshotVersion = 4,
        Tick = 103,
        Day = 2,
        TimeOfDay = 700,
        Weather = Weather.Sunny,
        Location = "Farm",
        Energy = 200,
        MaxEnergy = 270,
        Inventory = new InventorySummary
        {
            FreeSlots = 0,
            Items = new[] { new InventoryItem { ItemId = "parsnip", Name = "Parsnip", Quantity = 1 } }
        },
        WateringCan = new ToolState { Name = "Watering Can", Water = 10, Capacity = 40 },
        Crops = new[] { new Crop { Id = "crop-dry", NeedsWater = true } },
        Chests = new[] { new Chest { Id = "chest-east" }, new Chest { Id = "chest-west" } }
    };

    private static ObservedGameEvent Observed(EventKind kind, string targetId, long tick, bool success) => new()
    {
        Kind = kind,
        TargetId = targetId,
        Tick = tick,
        Success = success,
        Before = new GameStateSample(),
        After = new GameStateSample()
    };
}
