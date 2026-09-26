using EchoFarm.Bridge.Transport;

namespace EchoFarm.Bridge.Tests.Transport;

public sealed class EchoFarmClientFactoryTests
{
    [Fact]
    public void CreateUsesIndependentCommandAndReadProfiles()
    {
        Uri coreUrl = new("http://127.0.0.1:18471");

        EchoFarmClientSet clients = EchoFarmClientFactory.Create(coreUrl, TimeSpan.FromSeconds(100));

        Assert.NotSame(clients.Commands, clients.Reads);
        Assert.Equal(coreUrl, clients.Commands.BaseAddress);
        Assert.Equal(coreUrl, clients.Reads.BaseAddress);
        Assert.Equal(TimeSpan.FromSeconds(100), clients.Commands.RequestTimeout);
        Assert.Equal(TimeSpan.FromSeconds(2), clients.Reads.RequestTimeout);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(301)]
    public void CreateRejectsUnsafeCommandTimeouts(int seconds)
    {
        Uri coreUrl = new("http://127.0.0.1:18471");

        Assert.Throws<ArgumentOutOfRangeException>(() =>
            EchoFarmClientFactory.Create(coreUrl, TimeSpan.FromSeconds(seconds)));
    }
}
