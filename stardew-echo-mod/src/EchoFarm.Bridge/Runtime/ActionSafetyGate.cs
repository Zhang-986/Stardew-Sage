using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Runtime;

public sealed class UnsafeActionException : Exception
{
    public UnsafeActionException(string message) : base(message) { }
}

public sealed class ActionSafetyGate
{
    public void EnsureSafe(HighLevelAction action, WorldSnapshot snapshot)
    {
        if (action.SaveId != snapshot.SaveId ||
            action.SessionId != snapshot.SessionId ||
            action.SnapshotVersion != snapshot.SnapshotVersion)
        {
            throw new UnsafeActionException("Action correlation is stale.");
        }
        if (!Enum.IsDefined(typeof(ActionKind), action.Kind))
            throw new UnsafeActionException("Action kind is not allowed.");

        switch (action.Kind)
        {
            case ActionKind.StopSession:
                return;
            case ActionKind.WaterTarget:
                {
                    Crop crop = FindCrop(snapshot, action.TargetId);
                    if (snapshot.Weather is Weather.Rainy or Weather.Storm)
                        throw new UnsafeActionException("Watering is invalid in rain.");
                    if (snapshot.WateringCan.Water <= 0)
                        throw new UnsafeActionException("Watering can is empty.");
                    if (!crop.NeedsWater)
                        throw new UnsafeActionException("Crop does not need water.");
                    return;
                }
            case ActionKind.HarvestTarget:
                if (!FindCrop(snapshot, action.TargetId).Mature)
                    throw new UnsafeActionException("Crop is not mature.");
                return;
            case ActionKind.RefillCan:
                EnsureTarget(snapshot.WaterSources.Select(item => item.Id), action.TargetId, "water source");
                if (snapshot.WateringCan.Water >= snapshot.WateringCan.Capacity)
                    throw new UnsafeActionException("Watering can is already full.");
                return;
            case ActionKind.DepositItems:
                EnsureTarget(snapshot.Chests.Select(item => item.Id), action.TargetId, "chest");
                return;
            case ActionKind.MoveTo:
                if (action.Destination is null)
                {
                    IEnumerable<string> ids = snapshot.Crops.Select(item => item.Id)
                        .Concat(snapshot.WaterSources.Select(item => item.Id))
                        .Concat(snapshot.Chests.Select(item => item.Id));
                    EnsureTarget(ids, action.TargetId, "move");
                }
                return;
            case ActionKind.EquipTool:
                if (string.IsNullOrWhiteSpace(action.TargetId))
                    throw new UnsafeActionException("Tool target is required.");
                return;
            default:
                throw new UnsafeActionException("Action kind is not allowed.");
        }
    }

    private static Crop FindCrop(WorldSnapshot snapshot, string? targetId)
    {
        Crop? crop = snapshot.Crops.FirstOrDefault(item => item.Id == targetId);
        return crop ?? throw new UnsafeActionException($"Crop target '{targetId}' is unknown.");
    }

    private static void EnsureTarget(IEnumerable<string> ids, string? targetId, string kind)
    {
        if (string.IsNullOrWhiteSpace(targetId) || !ids.Contains(targetId, StringComparer.Ordinal))
            throw new UnsafeActionException($"The {kind} target '{targetId}' is unknown.");
    }
}
