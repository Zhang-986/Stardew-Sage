using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Runtime;

public readonly record struct GridBounds(int MinX, int MinY, int Width, int Height)
{
    public bool Contains(Position position) =>
        position.X >= MinX && position.Y >= MinY && position.X < MinX + Width && position.Y < MinY + Height;
}

public sealed class PathNotFoundException : Exception
{
    public PathNotFoundException() : base("No safe path to an adjacent target tile was found.") { }
}

public sealed class GridPathfinder
{
    private static readonly (int X, int Y)[] NeighborOffsets =
    {
        (0, 1),
        (1, 0),
        (0, -1),
        (-1, 0)
    };

    public IReadOnlyList<Position> FindPathToAdjacent(
        Position start,
        Position target,
        GridBounds bounds,
        Func<Position, bool> isWalkable)
    {
        ArgumentNullException.ThrowIfNull(isWalkable);
        if (!bounds.Contains(start))
            throw new ArgumentOutOfRangeException(nameof(start));
        if (!bounds.Contains(target))
            throw new ArgumentOutOfRangeException(nameof(target));
        if (IsAdjacent(start, target))
            return Array.Empty<Position>();

        var startKey = (start.X, start.Y);
        var queue = new Queue<(int X, int Y)>();
        var previous = new Dictionary<(int X, int Y), (int X, int Y)?>
        {
            [startKey] = null
        };
        queue.Enqueue(startKey);

        while (queue.Count > 0)
        {
            (int x, int y) = queue.Dequeue();
            foreach ((int offsetX, int offsetY) in NeighborOffsets)
            {
                var next = new Position { X = x + offsetX, Y = y + offsetY };
                var nextKey = (next.X, next.Y);
                if (!bounds.Contains(next) || previous.ContainsKey(nextKey) || !isWalkable(next))
                    continue;

                previous[nextKey] = (x, y);
                if (IsAdjacent(next, target))
                    return Reconstruct(previous, startKey, nextKey);
                queue.Enqueue(nextKey);
            }
        }

        throw new PathNotFoundException();
    }

    private static IReadOnlyList<Position> Reconstruct(
        IReadOnlyDictionary<(int X, int Y), (int X, int Y)?> previous,
        (int X, int Y) start,
        (int X, int Y) end)
    {
        var reversed = new List<Position>();
        (int X, int Y)? current = end;
        while (current is { } tile && tile != start)
        {
            reversed.Add(new Position { X = tile.X, Y = tile.Y });
            current = previous[tile];
        }
        reversed.Reverse();
        return reversed;
    }

    private static bool IsAdjacent(Position left, Position right) =>
        Math.Abs(left.X - right.X) + Math.Abs(left.Y - right.Y) == 1;
}
