using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;

namespace EchoFarm.Bridge.Tests.Recording;

public sealed class TeachingRecorderTests
{
    [Fact]
    public void RecordsChangedTilesAndSemanticStateDeltas()
    {
        var ids = new Queue<string>(new[] { "session-1", "event-1", "event-2", "event-3", "demo-1" });
        var recorder = new TeachingRecorder(() => ids.Dequeue());
        string sessionId = recorder.Start("farm-1", 100);

        Assert.Equal("session-1", sessionId);
        Assert.True(recorder.Observe(Move(110, 4, 8)));
        Assert.False(recorder.Observe(Move(111, 4, 8)));
        Assert.True(recorder.Observe(Move(120, 5, 8)));
        Assert.True(recorder.Observe(new ObservedGameEvent
        {
            Kind = EventKind.WaterTarget,
            Tick = 130,
            Position = new Position { X = 6, Y = 8 },
            TargetId = "crop-1",
            Tool = "Watering Can",
            Before = new GameStateSample { Energy = 100, Water = 10, InventoryCount = 3 },
            After = new GameStateSample { Energy = 98, Water = 9, InventoryCount = 3 },
            Success = true
        }));

        Demonstration demonstration = recorder.Stop(200);

        Assert.Equal("demo-1", demonstration.Id);
        Assert.Equal("farm-1", demonstration.SaveId);
        Assert.Equal("session-1", demonstration.SessionId);
        Assert.Equal(3, demonstration.Events.Count);
        DemonstrationEvent water = demonstration.Events[2];
        Assert.Equal("event-3", water.Id);
        Assert.Equal(-2, water.Delta.EnergyDelta);
        Assert.Equal(-1, water.Delta.WaterDelta);
        Assert.Equal(0, water.Delta.InventoryDelta);
    }

    [Fact]
    public void PreservesFailedSemanticAction()
    {
        var ids = new Queue<string>(new[] { "session-1", "event-1", "demo-1" });
        var recorder = new TeachingRecorder(() => ids.Dequeue());
        recorder.Start("farm-1", 100);

        recorder.Observe(new ObservedGameEvent
        {
            Kind = EventKind.WaterTarget,
            Tick = 130,
            Position = new Position { X = 6, Y = 8 },
            TargetId = "crop-1",
            Before = new GameStateSample(),
            After = new GameStateSample(),
            Success = false,
            ErrorCode = "out_of_water"
        });

        Demonstration demonstration = recorder.Stop(200);

        Assert.False(demonstration.Events[0].Success);
        Assert.Equal("out_of_water", demonstration.Events[0].ErrorCode);
    }

    [Fact]
    public void RejectsNestedStartAndStopOutsideRecording()
    {
        var recorder = new TeachingRecorder(() => Guid.NewGuid().ToString("N"));

        Assert.Throws<InvalidOperationException>(() => recorder.Stop(1));
        Assert.False(recorder.Observe(Move(1, 1, 1)));
        recorder.Start("farm-1", 1);
        Assert.Throws<InvalidOperationException>(() => recorder.Start("farm-1", 2));
    }

    [Fact]
    public void CancelDropsCapturedEvents()
    {
        var ids = new Queue<string>(new[] { "session-1", "event-1", "session-2", "demo-2" });
        var recorder = new TeachingRecorder(() => ids.Dequeue());
        recorder.Start("farm-1", 1);
        recorder.Observe(Move(2, 1, 1));

        recorder.Cancel();
        recorder.Start("farm-1", 3);
        Demonstration demonstration = recorder.Stop(4);

        Assert.Empty(demonstration.Events);
        Assert.Equal("session-2", demonstration.SessionId);
    }

    private static ObservedGameEvent Move(long tick, int x, int y) => new()
    {
        Kind = EventKind.Move,
        Tick = tick,
        Position = new Position { X = x, Y = y },
        Before = new GameStateSample(),
        After = new GameStateSample(),
        Success = true
    };
}
