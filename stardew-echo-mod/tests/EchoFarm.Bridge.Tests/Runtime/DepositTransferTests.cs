using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class DepositTransferTests
{
    [Fact]
    public void EmptyInventoryReturnsExplicitFailureWithoutCallingChest()
    {
        var inventory = new EchoInventory(capacity: 12);
        bool called = false;

        DepositTransferResult result = DepositTransfer.MoveAll(inventory, stack =>
        {
            called = true;
            return stack.Quantity;
        });

        Assert.False(result.Success);
        Assert.Equal("inventory_empty", result.ErrorCode);
        Assert.Equal(0, result.MovedQuantity);
        Assert.False(called);
    }

    [Fact]
    public void ChestAcceptingEverythingClearsEchoInventory()
    {
        var inventory = new EchoInventory(capacity: 12);
        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 3, 0));

        DepositTransferResult result = DepositTransfer.MoveAll(inventory, stack => stack.Quantity);

        Assert.True(result.Success);
        Assert.Null(result.ErrorCode);
        Assert.Equal(3, result.MovedQuantity);
        Assert.True(inventory.IsEmpty);
    }

    [Fact]
    public void FullChestRemovesOnlyTheQuantityActuallyAccepted()
    {
        var inventory = new EchoInventory(capacity: 12);
        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 3, 0));

        DepositTransferResult result = DepositTransfer.MoveAll(inventory, _ => 1);

        Assert.False(result.Success);
        Assert.Equal("chest_full", result.ErrorCode);
        Assert.Equal(1, result.MovedQuantity);
        Assert.Equal(2, Assert.Single(inventory.Snapshot()).Quantity);
    }

    [Fact]
    public void RejectsInvalidAcceptedQuantityWithoutChangingInventory()
    {
        var inventory = new EchoInventory(capacity: 12);
        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 3, 0));

        Assert.Throws<InvalidOperationException>(() => DepositTransfer.MoveAll(inventory, _ => 4));

        Assert.Equal(3, Assert.Single(inventory.Snapshot()).Quantity);
    }
}
