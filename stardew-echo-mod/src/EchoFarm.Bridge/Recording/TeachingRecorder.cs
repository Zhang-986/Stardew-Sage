using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Recording;

public sealed class TeachingRecorder
{
    private readonly Func<string> idFactory;
    private readonly List<DemonstrationEvent> events = new();
    private string? saveId;
    private string? sessionId;
    private long startedAt;
    private Position? lastMovePosition;

    public TeachingRecorder(Func<string>? idFactory = null)
    {
        this.idFactory = idFactory ?? (() => Guid.NewGuid().ToString("N"));
    }

    public bool IsRecording { get; private set; }

    public string Start(string saveId, long tick)
    {
        if (IsRecording)
            throw new InvalidOperationException("A teaching session is already active.");
        if (string.IsNullOrWhiteSpace(saveId))
            throw new ArgumentException("Save ID is required.", nameof(saveId));

        this.saveId = saveId;
        sessionId = idFactory();
        startedAt = tick;
        events.Clear();
        lastMovePosition = null;
        IsRecording = true;
        return sessionId;
    }

    public bool Observe(ObservedGameEvent observed)
    {
        if (!IsRecording)
            return false;

        if (observed.Kind == EventKind.Move && SamePosition(lastMovePosition, observed.Position))
            return false;

        var captured = new DemonstrationEvent
        {
            Id = idFactory(),
            Kind = observed.Kind,
            Tick = observed.Tick,
            Position = Copy(observed.Position),
            TargetId = observed.TargetId,
            Tool = observed.Tool,
            Delta = new StateDelta
            {
                EnergyDelta = observed.After.Energy - observed.Before.Energy,
                WaterDelta = observed.After.Water - observed.Before.Water,
                InventoryDelta = observed.After.InventoryCount - observed.Before.InventoryCount
            },
            Success = observed.Success,
            ErrorCode = observed.ErrorCode
        };
        events.Add(captured);
        lastMovePosition = observed.Kind == EventKind.Move ? Copy(observed.Position) : null;
        return true;
    }

    public Demonstration Stop(long tick)
    {
        if (!IsRecording || saveId is null || sessionId is null)
            throw new InvalidOperationException("No teaching session is active.");

        var demonstration = new Demonstration
        {
            Id = idFactory(),
            SaveId = saveId,
            SessionId = sessionId,
            StartedAt = startedAt,
            EndedAt = tick,
            Events = Array.AsReadOnly(events.ToArray())
        };
        Reset();
        return demonstration;
    }

    public void Cancel()
    {
        Reset();
    }

    private void Reset()
    {
        IsRecording = false;
        saveId = null;
        sessionId = null;
        startedAt = 0;
        lastMovePosition = null;
        events.Clear();
    }

    private static bool SamePosition(Position? left, Position right) =>
        left is not null && left.X == right.X && left.Y == right.Y;

    private static Position Copy(Position position) => new() { X = position.X, Y = position.Y };
}
