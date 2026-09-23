namespace EchoFarm.Bridge.Runtime;

public sealed record CropGrowthState(
    int CurrentPhase,
    int PhaseCount,
    bool FullyGrown,
    int DayOfCurrentPhase,
    bool Dead,
    int RegrowDays
);

public sealed record CropGrowthTransition(bool RemoveCrop, bool FullyGrown, int DayOfCurrentPhase);

public sealed record HarvestTransferResult(
    bool Success,
    string? ErrorCode,
    CropGrowthTransition? Transition
);

public static class HarvestTransfer
{
    public static HarvestTransferResult TryCollect(
        EchoInventory inventory,
        EchoItemStack yield,
        CropGrowthState crop)
    {
        ArgumentNullException.ThrowIfNull(inventory);
        ArgumentNullException.ThrowIfNull(yield);
        ArgumentNullException.ThrowIfNull(crop);

        bool mature = !crop.Dead &&
            crop.PhaseCount > 0 &&
            crop.CurrentPhase >= crop.PhaseCount - 1 &&
            (!crop.FullyGrown || crop.DayOfCurrentPhase <= 0);
        if (!mature)
            return new HarvestTransferResult(false, "target_changed", null);
        if (!inventory.TryAdd(yield))
            return new HarvestTransferResult(false, "inventory_full", null);

        CropGrowthTransition transition = crop.RegrowDays > 0
            ? new CropGrowthTransition(false, true, crop.RegrowDays)
            : new CropGrowthTransition(true, false, 0);
        return new HarvestTransferResult(true, null, transition);
    }
}
