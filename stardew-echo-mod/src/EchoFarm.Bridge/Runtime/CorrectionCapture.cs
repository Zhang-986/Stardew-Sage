using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;

namespace EchoFarm.Bridge.Runtime;

public sealed class CorrectionCapture
{
    public const long DefaultCorrectionWindowTicks = 20 * 60;

    private readonly Func<string> idFactory;
    private readonly long correctionWindowTicks;
    private HighLevelAction? rejectedAction;
    private long deadlineTick;

    public CorrectionCapture(Func<string>? idFactory = null, long correctionWindowTicks = DefaultCorrectionWindowTicks)
    {
        if (correctionWindowTicks <= 0)
            throw new ArgumentOutOfRangeException(nameof(correctionWindowTicks));
        this.idFactory = idFactory ?? (() => $"correction-{Guid.NewGuid():N}");
        this.correctionWindowTicks = correctionWindowTicks;
    }

    public bool IsArmed => rejectedAction is not null;
    public PlayerCorrection? PendingCorrection { get; private set; }

    public void Arm(HighLevelAction action, long currentTick)
    {
        ArgumentNullException.ThrowIfNull(action);
        if (string.IsNullOrWhiteSpace(action.SaveId) || string.IsNullOrWhiteSpace(action.SessionId) || currentTick < 0)
            throw new ArgumentException("A correlated rejected action and current tick are required.", nameof(action));
        rejectedAction = action;
        deadlineTick = checked(currentTick + correctionWindowTicks);
        PendingCorrection = null;
    }

    public bool TryCapture(WorldSnapshot snapshot, ObservedGameEvent observed, out PlayerCorrection? correction)
    {
        ArgumentNullException.ThrowIfNull(snapshot);
        ArgumentNullException.ThrowIfNull(observed);
        correction = null;
        if (rejectedAction is null || PendingCorrection is not null)
            return false;
        if (Expire(observed.Tick) || !observed.Success || !TryMap(observed.Kind, out ActionKind preferredKind) || string.IsNullOrWhiteSpace(observed.TargetId))
            return false;
        if (snapshot.SaveId != rejectedAction.SaveId || snapshot.SessionId != rejectedAction.SessionId || snapshot.SnapshotVersion < rejectedAction.SnapshotVersion)
            throw new InvalidOperationException("Correction snapshot does not follow the rejected decision.");

        correction = new PlayerCorrection
        {
            Id = idFactory(),
            SaveId = snapshot.SaveId,
            SessionId = snapshot.SessionId,
            RejectedDecisionSnapshotVersion = rejectedAction.SnapshotVersion,
            RejectedAction = rejectedAction,
            Snapshot = snapshot,
            PreferredAction = new HighLevelAction
            {
                SaveId = snapshot.SaveId,
                SessionId = snapshot.SessionId,
                SnapshotVersion = snapshot.SnapshotVersion,
                Kind = preferredKind,
                TargetId = observed.TargetId,
                Reason = "player demonstrated a better action"
            },
            ObservedAtTick = observed.Tick
        };
        PendingCorrection = correction;
        return true;
    }

    public bool Expire(long currentTick)
    {
        if (rejectedAction is null || PendingCorrection is not null || currentTick < deadlineTick)
            return false;
        Cancel();
        return true;
    }

    public void Commit()
    {
        if (PendingCorrection is null)
            throw new InvalidOperationException("No correction is ready to commit.");
        Cancel();
    }

    public void Cancel()
    {
        rejectedAction = null;
        PendingCorrection = null;
        deadlineTick = 0;
    }

    private static bool TryMap(EventKind kind, out ActionKind action)
    {
        switch (kind)
        {
            case EventKind.WaterTarget:
                action = ActionKind.WaterTarget;
                return true;
            case EventKind.RefillCan:
                action = ActionKind.RefillCan;
                return true;
            case EventKind.HarvestTarget:
                action = ActionKind.HarvestTarget;
                return true;
            case EventKind.DepositItems:
                action = ActionKind.DepositItems;
                return true;
            default:
                action = default;
                return false;
        }
    }
}
