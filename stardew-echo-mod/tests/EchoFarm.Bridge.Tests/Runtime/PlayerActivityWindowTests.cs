using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;
using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class PlayerActivityWindowTests
{
    [Fact]
    public void SnapshotDropsStaleAndFailedActivityAndDeduplicatesTargets()
    {
        var window = new PlayerActivityWindow(windowTicks: 1_200);
        window.Add(Activity(EventKind.WaterTarget, "stale", 700, success: true));
        window.Add(Activity(EventKind.WaterTarget, "crop-1", 1_900, success: true));
        window.Add(Activity(EventKind.WaterTarget, "crop-1", 1_950, success: true));
        window.Add(Activity(EventKind.HarvestTarget, "failed", 1_980, success: false));

        IReadOnlyList<PlayerActivity> got = window.Snapshot(2_000);

        PlayerActivity activity = Assert.Single(got);
        Assert.Equal("crop-1", activity.TargetId);
        Assert.Equal(1_950, activity.Tick);
    }

    [Fact]
    public void ResetRemovesActivityBetweenSaves()
    {
        var window = new PlayerActivityWindow();
        window.Add(Activity(EventKind.HarvestTarget, "crop-1", 10, success: true));

        window.Reset();

        Assert.Empty(window.Snapshot(10));
    }

    private static ObservedGameEvent Activity(EventKind kind, string targetId, long tick, bool success) => new()
    {
        Kind = kind,
        TargetId = targetId,
        Tick = tick,
        Position = new Position(),
        Before = new GameStateSample(),
        After = new GameStateSample(),
        Success = success
    };
}
