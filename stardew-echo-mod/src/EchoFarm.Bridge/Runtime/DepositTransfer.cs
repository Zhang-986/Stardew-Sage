namespace EchoFarm.Bridge.Runtime;

public sealed record DepositTransferResult(bool Success, string? ErrorCode, int MovedQuantity);

public static class DepositTransfer
{
    public static DepositTransferResult MoveAll(
        EchoInventory inventory,
        Func<EchoItemStack, int> deposit)
    {
        ArgumentNullException.ThrowIfNull(inventory);
        ArgumentNullException.ThrowIfNull(deposit);
        if (inventory.IsEmpty)
            return new DepositTransferResult(false, "inventory_empty", 0);

        int moved = 0;
        foreach (EchoItemStack stack in inventory.Snapshot())
        {
            int accepted = deposit(stack);
            if (accepted < 0 || accepted > stack.Quantity)
                throw new InvalidOperationException("The chest accepted an invalid item quantity.");
            if (accepted == 0)
                continue;
            if (!inventory.Remove(stack.ItemId, stack.Quality, accepted))
                throw new InvalidOperationException("Echo inventory changed during deposit.");
            moved = checked(moved + accepted);
        }

        return inventory.IsEmpty
            ? new DepositTransferResult(true, null, moved)
            : new DepositTransferResult(false, "chest_full", moved);
    }
}
