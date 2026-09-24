using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Recording;

public enum SemanticActivityFamily
{
    TreeChopping,
    RockBreaking,
    Fishing
}

public sealed record ActivitySample(
    long Tick,
    string Location,
    int TimeOfDay,
    Position Position,
    string TargetId,
    string TargetKind,
    string Tool,
    GameStateSample State,
    IReadOnlyList<ItemDelta> ItemDeltas,
    bool TargetPresent
);

public sealed class SemanticActivityTracker
{
    private const long ToolTimeoutTicks = 600;
    private const long FishingTimeoutTicks = 7200;
    private PendingActivity? pending;

    public bool HasPendingActivity => pending is not null;

    public bool TryBegin(SemanticActivityFamily family, ActivitySample sample)
    {
        ArgumentNullException.ThrowIfNull(sample);
        ValidateSample(sample);
        if (pending is not null)
            return false;
        pending = new PendingActivity(family, sample);
        return true;
    }

    public ObservedGameEvent? Observe(ActivitySample sample)
    {
        ArgumentNullException.ThrowIfNull(sample);
        ValidateSample(sample);
        if (pending is null)
            return null;
        if (sample.Tick < pending.Start.Tick)
            throw new ArgumentOutOfRangeException(nameof(sample), "Activity tick cannot move backwards.");
        if (!StringComparer.Ordinal.Equals(sample.Location, pending.Start.Location))
            return Finish(sample, success: false, "location_changed");
        if (!StringComparer.Ordinal.Equals(sample.TargetId, pending.Start.TargetId))
            return Finish(sample, success: false, "target_changed");
        long timeout = pending.Family == SemanticActivityFamily.Fishing ? FishingTimeoutTicks : ToolTimeoutTicks;
        if (sample.Tick - pending.Start.Tick > timeout)
            return Finish(sample, success: false, "activity_timeout");
        if (pending.Family is SemanticActivityFamily.TreeChopping or SemanticActivityFamily.RockBreaking && !sample.TargetPresent)
            return Finish(sample, success: true, errorCode: null);
        return null;
    }

    public ObservedGameEvent? CompleteFishing(ActivitySample sample, bool caught)
    {
        ArgumentNullException.ThrowIfNull(sample);
        if (pending?.Family != SemanticActivityFamily.Fishing)
            return null;
        if (caught && !sample.ItemDeltas.Any(item => item.Quantity > 0))
            throw new ArgumentException("A caught fish requires a positive item delta.", nameof(sample));
        return Finish(sample, caught, caught ? null : "fish_escaped",
            caught ? EventKind.FishCaught : EventKind.FishEscaped);
    }

    public ObservedGameEvent? RecordMineTransition(ActivitySample before, ActivitySample after, int floorDelta)
    {
        ArgumentNullException.ThrowIfNull(before);
        ArgumentNullException.ThrowIfNull(after);
        ValidateSample(before);
        ValidateSample(after);
        if (floorDelta == 0 || after.State.MineFloor <= 0 || after.Tick < before.Tick)
            return null;
        return BuildEvent(
            EventKind.EnterMineFloor,
            before,
            after,
            success: true,
            errorCode: null,
            itemDeltas: NormalizeItemDeltas(after.ItemDeltas)
        );
    }

    public ObservedGameEvent? Cancel(long tick, string errorCode)
    {
        if (pending is null)
            return null;
        if (errorCode is not ("activity_cancelled" or "location_changed" or "target_changed" or "activity_timeout"))
            throw new ArgumentException("Unsupported activity cancellation code.", nameof(errorCode));
        ActivitySample end = pending.Start with { Tick = Math.Max(tick, pending.Start.Tick) };
        return Finish(end, success: false, errorCode);
    }

    public void Reset() => pending = null;

    private ObservedGameEvent Finish(ActivitySample sample, bool success, string? errorCode, EventKind? kind = null)
    {
        PendingActivity active = pending!;
        pending = null;
        EventKind resolvedKind = kind ?? active.Family switch
        {
            SemanticActivityFamily.TreeChopping => EventKind.ChopTree,
            SemanticActivityFamily.RockBreaking => EventKind.BreakRock,
            SemanticActivityFamily.Fishing => EventKind.FishEscaped,
            _ => throw new InvalidOperationException("Unsupported semantic activity family.")
        };
        return BuildEvent(
            resolvedKind,
            active.Start,
            sample,
            success,
            errorCode,
            NormalizeItemDeltas(sample.ItemDeltas)
        );
    }

    private static ObservedGameEvent BuildEvent(
        EventKind kind,
        ActivitySample start,
        ActivitySample end,
        bool success,
        string? errorCode,
        IReadOnlyList<ItemDelta> itemDeltas) => new()
    {
        Kind = kind,
        Tick = end.Tick,
        Position = new Position { X = start.Position.X, Y = start.Position.Y },
        Location = end.Location,
        TimeOfDay = end.TimeOfDay,
        TargetId = start.TargetId,
        TargetKind = start.TargetKind,
        Tool = start.Tool,
        DurationTicks = end.Tick - start.Tick,
        ItemDeltas = itemDeltas,
        Before = start.State,
        After = end.State,
        Success = success,
        ErrorCode = errorCode
    };

    private static IReadOnlyList<ItemDelta> NormalizeItemDeltas(IReadOnlyList<ItemDelta> values)
    {
        ItemDelta[] result = values
            .Where(item => !string.IsNullOrWhiteSpace(item.ItemId))
            .GroupBy(item => item.ItemId, StringComparer.Ordinal)
            .Select(group => new ItemDelta
            {
                ItemId = group.Key,
                Name = group.Select(item => item.Name).FirstOrDefault(name => !string.IsNullOrWhiteSpace(name)),
                Quantity = group.Sum(item => item.Quantity)
            })
            .Where(item => item.Quantity != 0)
            .OrderBy(item => item.ItemId, StringComparer.Ordinal)
            .ToArray();
        if (result.Length > 32)
            throw new ArgumentException("Activity item deltas cannot exceed 32 entries.", nameof(values));
        return Array.AsReadOnly(result);
    }

    private static void ValidateSample(ActivitySample sample)
    {
        if (sample.Tick < 0 || string.IsNullOrWhiteSpace(sample.Location) ||
            string.IsNullOrWhiteSpace(sample.TargetId) || string.IsNullOrWhiteSpace(sample.TargetKind))
            throw new ArgumentException("Activity sample identity and non-negative tick are required.", nameof(sample));
        if (sample.TimeOfDay is < 0 or > 2600)
            throw new ArgumentOutOfRangeException(nameof(sample), "Activity time of day is out of range.");
    }

    private sealed record PendingActivity(SemanticActivityFamily Family, ActivitySample Start);
}
