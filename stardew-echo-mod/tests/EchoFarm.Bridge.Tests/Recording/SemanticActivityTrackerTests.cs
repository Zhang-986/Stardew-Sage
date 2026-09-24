using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;

namespace EchoFarm.Bridge.Tests.Recording;

public sealed class SemanticActivityTrackerTests
{
    [Theory]
    [InlineData(SemanticActivityFamily.TreeChopping, EventKind.ChopTree, "tree")]
    [InlineData(SemanticActivityFamily.RockBreaking, EventKind.BreakRock, "rock")]
    public void ToolEpisodeAggregatesProgressIntoOneSuccessfulEvent(
        SemanticActivityFamily family,
        EventKind expectedKind,
        string targetKind)
    {
        var tracker = new SemanticActivityTracker();
        ActivitySample start = Sample(100, "Farm", "target-1", targetKind, targetPresent: true, energy: 100);

        Assert.True(tracker.TryBegin(family, start));
        Assert.Null(tracker.Observe(Sample(130, "Farm", "target-1", targetKind, targetPresent: true, energy: 98)));
        ObservedGameEvent completed = Assert.IsType<ObservedGameEvent>(tracker.Observe(
            Sample(180, "Farm", "target-1", targetKind, targetPresent: false, energy: 90,
                items: new[] { new ItemDelta { ItemId = "388", Name = "Wood", Quantity = 14 } })));

        Assert.Equal(expectedKind, completed.Kind);
        Assert.True(completed.Success);
        Assert.Equal(80, completed.DurationTicks);
        Assert.Equal(100, completed.Before.Energy);
        Assert.Equal(90, completed.After.Energy);
        Assert.Equal(14, completed.ItemDeltas[0].Quantity);
        Assert.False(tracker.HasPendingActivity);
    }

    [Fact]
    public void TrackerRejectsConcurrentEpisodeAndBoundsTimeout()
    {
        var tracker = new SemanticActivityTracker();
        ActivitySample start = Sample(100, "Farm", "tree-1", "tree", targetPresent: true);

        Assert.True(tracker.TryBegin(SemanticActivityFamily.TreeChopping, start));
        Assert.False(tracker.TryBegin(SemanticActivityFamily.RockBreaking,
            Sample(110, "Farm", "rock-1", "rock", targetPresent: true)));
        ObservedGameEvent timeout = Assert.IsType<ObservedGameEvent>(tracker.Observe(
            Sample(701, "Farm", "tree-1", "tree", targetPresent: true)));

        Assert.False(timeout.Success);
        Assert.Equal("activity_timeout", timeout.ErrorCode);
        Assert.Equal(601, timeout.DurationTicks);
    }

    [Theory]
    [InlineData("Mine", "tree-2", "target_changed")]
    [InlineData("Forest", "tree-1", "location_changed")]
    public void ToolEpisodeFailsOnIdentityChange(string location, string targetId, string expectedError)
    {
        var tracker = new SemanticActivityTracker();
        tracker.TryBegin(SemanticActivityFamily.TreeChopping,
            Sample(100, "Mine", "tree-1", "tree", targetPresent: true));

        ObservedGameEvent failed = Assert.IsType<ObservedGameEvent>(tracker.Observe(
            Sample(120, location, targetId, "tree", targetPresent: true)));

        Assert.False(failed.Success);
        Assert.Equal(expectedError, failed.ErrorCode);
    }

    [Fact]
    public void FishingRequiresConfirmedCatchEvidenceAndRepresentsEscape()
    {
        var tracker = new SemanticActivityTracker();
        tracker.TryBegin(SemanticActivityFamily.Fishing,
            Sample(200, "Beach", "sea", "fish", targetPresent: true));

        ObservedGameEvent caught = Assert.IsType<ObservedGameEvent>(tracker.CompleteFishing(
            Sample(320, "Beach", "sea", "fish", targetPresent: true,
                items: new[] { new ItemDelta { ItemId = "128", Name = "Pufferfish", Quantity = 1 } }), caught: true));
        Assert.Equal(EventKind.FishCaught, caught.Kind);
        Assert.True(caught.Success);

        tracker.TryBegin(SemanticActivityFamily.Fishing,
            Sample(400, "Beach", "sea", "fish", targetPresent: true));
        ObservedGameEvent escaped = Assert.IsType<ObservedGameEvent>(tracker.CompleteFishing(
            Sample(500, "Beach", "sea", "fish", targetPresent: true), caught: false));
        Assert.Equal(EventKind.FishEscaped, escaped.Kind);
        Assert.False(escaped.Success);
        Assert.Equal("fish_escaped", escaped.ErrorCode);
    }

    [Fact]
    public void MineTransitionAndCancellationAreExplicitEvents()
    {
        var tracker = new SemanticActivityTracker();
        ObservedGameEvent transition = Assert.IsType<ObservedGameEvent>(tracker.RecordMineTransition(
            Sample(100, "UndergroundMine20", "mine-floor-20", "mine_floor", true, mineFloor: 20),
            Sample(120, "UndergroundMine21", "mine-floor-21", "mine_floor", true, mineFloor: 21),
            floorDelta: 1));
        Assert.Equal(EventKind.EnterMineFloor, transition.Kind);
        Assert.Equal(1, transition.After.MineFloor - transition.Before.MineFloor);

        tracker.TryBegin(SemanticActivityFamily.RockBreaking,
            Sample(200, "UndergroundMine21", "rock-1", "rock", true));
        ObservedGameEvent cancelled = Assert.IsType<ObservedGameEvent>(tracker.Cancel(230, "activity_cancelled"));
        Assert.Equal("activity_cancelled", cancelled.ErrorCode);
        tracker.Reset();
        Assert.False(tracker.HasPendingActivity);
    }

    private static ActivitySample Sample(
        long tick,
        string location,
        string targetId,
        string targetKind,
        bool targetPresent,
        int energy = 100,
        int mineFloor = 0,
        IReadOnlyList<ItemDelta>? items = null) => new(
            tick,
            location,
            900,
            new Position { X = 12, Y = 8 },
            targetId,
            targetKind,
            targetKind == "tree" ? "Axe" : targetKind == "rock" ? "Pickaxe" : "Fishing Rod",
            new GameStateSample { Energy = energy, Health = 100, MineFloor = mineFloor },
            items ?? Array.Empty<ItemDelta>(),
            targetPresent
        );
}
