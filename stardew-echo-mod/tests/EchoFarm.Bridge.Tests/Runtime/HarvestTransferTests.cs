using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class HarvestTransferTests
{
    [Theory]
    [InlineData(1, 4, false, 0, false)]
    [InlineData(3, 4, false, 0, true)]
    [InlineData(3, 4, true, 2, false)]
    public void RejectsCropThatIsNotHarvestable(
        int currentPhase,
        int phaseCount,
        bool fullyGrown,
        int dayOfCurrentPhase,
        bool dead)
    {
        var inventory = new EchoInventory(capacity: 12);
        var crop = new CropGrowthState(currentPhase, phaseCount, fullyGrown, dayOfCurrentPhase, dead, RegrowDays: -1);

        HarvestTransferResult result = HarvestTransfer.TryCollect(inventory, Parsnip(), crop);

        Assert.False(result.Success);
        Assert.Equal("target_changed", result.ErrorCode);
        Assert.Null(result.Transition);
        Assert.True(inventory.IsEmpty);
    }

    [Fact]
    public void FullInventoryRejectsHarvestWithoutChangingExistingItems()
    {
        var inventory = new EchoInventory(capacity: 1);
        inventory.TryAdd(new EchoItemStack("(O)188", "Green Bean", 2, 0));

        HarvestTransferResult result = HarvestTransfer.TryCollect(inventory, Parsnip(), Mature(regrowDays: -1));

        Assert.False(result.Success);
        Assert.Equal("inventory_full", result.ErrorCode);
        Assert.Null(result.Transition);
        EchoItemStack existing = Assert.Single(inventory.Snapshot());
        Assert.Equal("(O)188", existing.ItemId);
        Assert.Equal(2, existing.Quantity);
    }

    [Fact]
    public void OneShotCropMovesYieldAndRequestsCropRemoval()
    {
        var inventory = new EchoInventory(capacity: 12);

        HarvestTransferResult result = HarvestTransfer.TryCollect(inventory, Parsnip(quantity: 2), Mature(regrowDays: -1));

        Assert.True(result.Success);
        Assert.Null(result.ErrorCode);
        Assert.Equal(new CropGrowthTransition(RemoveCrop: true, FullyGrown: false, DayOfCurrentPhase: 0), result.Transition);
        Assert.Equal(2, Assert.Single(inventory.Snapshot()).Quantity);
    }

    [Fact]
    public void RegrowingCropMovesYieldAndStartsRegrowCountdown()
    {
        var inventory = new EchoInventory(capacity: 12);

        HarvestTransferResult result = HarvestTransfer.TryCollect(inventory, Parsnip(), Mature(regrowDays: 4));

        Assert.True(result.Success);
        Assert.Equal(new CropGrowthTransition(RemoveCrop: false, FullyGrown: true, DayOfCurrentPhase: 4), result.Transition);
        Assert.Single(inventory.Snapshot());
    }

    private static EchoItemStack Parsnip(int quantity = 1) => new("(O)24", "Parsnip", quantity, 0);

    private static CropGrowthState Mature(int regrowDays) => new(
        CurrentPhase: 3,
        PhaseCount: 4,
        FullyGrown: false,
        DayOfCurrentPhase: 0,
        Dead: false,
        RegrowDays: regrowDays
    );
}
