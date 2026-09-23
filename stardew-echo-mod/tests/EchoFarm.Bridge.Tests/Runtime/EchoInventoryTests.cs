using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class EchoInventoryTests
{
    [Fact]
    public void NewStackConsumesOneSlot()
    {
        var inventory = new EchoInventory(capacity: 2);

        bool added = inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 1, 0));

        Assert.True(added);
        Assert.Equal(1, inventory.FreeSlots);
        Assert.Equal(new EchoItemStack("(O)24", "Parsnip", 1, 0), Assert.Single(inventory.Snapshot()));
    }

    [Fact]
    public void MatchingItemAndQualityStacksInExistingSlot()
    {
        var inventory = new EchoInventory(capacity: 1);
        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 1, 0));

        bool added = inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 3, 0));

        Assert.True(added);
        Assert.Equal(0, inventory.FreeSlots);
        Assert.Equal(4, Assert.Single(inventory.Snapshot()).Quantity);
    }

    [Fact]
    public void DifferentQualityConsumesAnotherSlot()
    {
        var inventory = new EchoInventory(capacity: 2);
        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 1, 0));

        bool added = inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 1, 2));

        Assert.True(added);
        Assert.Equal(0, inventory.FreeSlots);
        Assert.Equal(2, inventory.Snapshot().Count);
    }

    [Fact]
    public void FullInventoryRejectsEntireNewStack()
    {
        var inventory = new EchoInventory(capacity: 1);
        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 1, 0));

        bool added = inventory.TryAdd(new EchoItemStack("(O)188", "Green Bean", 5, 0));

        Assert.False(added);
        EchoItemStack remaining = Assert.Single(inventory.Snapshot());
        Assert.Equal("(O)24", remaining.ItemId);
        Assert.Equal(1, remaining.Quantity);
    }

    [Fact]
    public void RemoveDeletesOnlyTheRequestedQuantity()
    {
        var inventory = new EchoInventory(capacity: 2);
        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 4, 0));

        Assert.True(inventory.Remove("(O)24", quality: 0, quantity: 3));
        Assert.Equal(1, Assert.Single(inventory.Snapshot()).Quantity);
        Assert.True(inventory.Remove("(O)24", quality: 0, quantity: 1));
        Assert.True(inventory.IsEmpty);
        Assert.Equal(2, inventory.FreeSlots);
    }

    [Fact]
    public void SnapshotDoesNotChangeAfterLaterInventoryMutations()
    {
        var inventory = new EchoInventory(capacity: 1);
        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 1, 0));
        IReadOnlyList<EchoItemStack> before = inventory.Snapshot();

        inventory.TryAdd(new EchoItemStack("(O)24", "Parsnip", 2, 0));

        Assert.Equal(1, Assert.Single(before).Quantity);
        Assert.Equal(3, Assert.Single(inventory.Snapshot()).Quantity);
    }
}
