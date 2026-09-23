using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class SnapshotVersionTrackerTests
{
    [Fact]
    public void RepeatedEquivalentStateKeepsTheSameVersion()
    {
        var versions = new SnapshotVersionTracker();

        long first = versions.Next("day=2;crop=dry");
        long second = versions.Next("day=2;crop=dry");

        Assert.Equal(1, first);
        Assert.Equal(first, second);
    }

    [Fact]
    public void EachStateChangeAdvancesTheVersion()
    {
        var versions = new SnapshotVersionTracker();

        Assert.Equal(1, versions.Next("day=2;crop=dry"));
        Assert.Equal(2, versions.Next("day=2;crop=watered"));
        Assert.Equal(3, versions.Next("day=2;crop=dry"));
    }
}
