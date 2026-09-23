using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class GridPathfinderTests
{
    [Fact]
    public void FindsShortestPathToAdjacentTileAroundObstacle()
    {
        var pathfinder = new GridPathfinder();
        var blocked = new HashSet<(int X, int Y)> { (2, 1) };

        IReadOnlyList<Position> path = pathfinder.FindPathToAdjacent(
            new Position { X = 1, Y = 1 },
            new Position { X = 3, Y = 1 },
            new GridBounds(0, 0, 5, 5),
            tile => !blocked.Contains((tile.X, tile.Y))
        );

        Assert.Equal(3, path.Count);
        Assert.Equal((1, 2), (path[0].X, path[0].Y));
        Assert.Equal((3, 2), (path[^1].X, path[^1].Y));
    }

    [Fact]
    public void ReturnsEmptyPathWhenAlreadyAdjacent()
    {
        IReadOnlyList<Position> path = new GridPathfinder().FindPathToAdjacent(
            new Position { X = 4, Y = 4 },
            new Position { X = 5, Y = 4 },
            new GridBounds(0, 0, 10, 10),
            _ => true
        );

        Assert.Empty(path);
    }

    [Fact]
    public void ReportsNoPathWhenTargetIsEnclosed()
    {
        var blocked = new HashSet<(int X, int Y)> { (2, 1), (3, 2), (2, 3), (1, 2) };

        Assert.Throws<PathNotFoundException>(() => new GridPathfinder().FindPathToAdjacent(
            new Position { X = 0, Y = 0 },
            new Position { X = 2, Y = 2 },
            new GridBounds(0, 0, 5, 5),
            tile => !blocked.Contains((tile.X, tile.Y))
        ));
    }

    [Fact]
    public void RefusesToSearchOutsideBoundedMap()
    {
        Assert.Throws<ArgumentOutOfRangeException>(() => new GridPathfinder().FindPathToAdjacent(
            new Position { X = -1, Y = 0 },
            new Position { X = 2, Y = 2 },
            new GridBounds(0, 0, 5, 5),
            _ => true
        ));
    }
}
