namespace EchoFarm.Bridge.Runtime;

public sealed record EchoItemStack(string ItemId, string Name, int Quantity, int Quality = 0);

public sealed class EchoInventory
{
    private readonly Dictionary<ItemKey, EchoItemStack> stacks = new();

    public EchoInventory(int capacity)
    {
        if (capacity <= 0)
            throw new ArgumentOutOfRangeException(nameof(capacity));
        Capacity = capacity;
    }

    public int Capacity { get; }
    public int FreeSlots => Capacity - stacks.Count;
    public bool IsEmpty => stacks.Count == 0;

    public bool TryAdd(EchoItemStack stack)
    {
        Validate(stack);
        var key = new ItemKey(stack.ItemId, stack.Quality);
        if (stacks.TryGetValue(key, out EchoItemStack? existing))
        {
            stacks[key] = existing with { Quantity = checked(existing.Quantity + stack.Quantity) };
            return true;
        }
        if (stacks.Count >= Capacity)
            return false;
        stacks.Add(key, stack with { });
        return true;
    }

    public bool Remove(string itemId, int quality, int quantity)
    {
        if (string.IsNullOrWhiteSpace(itemId))
            throw new ArgumentException("An item ID is required.", nameof(itemId));
        if (quantity <= 0)
            throw new ArgumentOutOfRangeException(nameof(quantity));

        var key = new ItemKey(itemId, quality);
        if (!stacks.TryGetValue(key, out EchoItemStack? existing) || existing.Quantity < quantity)
            return false;
        if (existing.Quantity == quantity)
            stacks.Remove(key);
        else
            stacks[key] = existing with { Quantity = existing.Quantity - quantity };
        return true;
    }

    public IReadOnlyList<EchoItemStack> Snapshot() =>
        stacks.Values.Select(stack => stack with { }).ToArray();

    private static void Validate(EchoItemStack stack)
    {
        ArgumentNullException.ThrowIfNull(stack);
        if (string.IsNullOrWhiteSpace(stack.ItemId))
            throw new ArgumentException("An item ID is required.", nameof(stack));
        if (string.IsNullOrWhiteSpace(stack.Name))
            throw new ArgumentException("An item name is required.", nameof(stack));
        if (stack.Quantity <= 0)
            throw new ArgumentOutOfRangeException(nameof(stack));
        if (stack.Quality < 0)
            throw new ArgumentOutOfRangeException(nameof(stack));
    }

    private readonly record struct ItemKey(string ItemId, int Quality);
}
