using EchoFarm.Bridge.Transport;

namespace EchoFarm.Bridge.Tests.Transport;

public sealed class EchoFarmClientFactoryTests
{
    [Fact]
    public void CreateUsesIndependentCommandAndReadProfiles()
    {
        Uri coreUrl = new("http://127.0.0.1:18471");

        EchoFarmClientSet clients = EchoFarmClientFactory.Create(coreUrl);

        Assert.NotSame(clients.Commands, clients.Reads);
        Assert.Equal(coreUrl, clients.Commands.BaseAddress);
        Assert.Equal(coreUrl, clients.Reads.BaseAddress);
        Assert.Equal(TimeSpan.FromSeconds(35), clients.Commands.RequestTimeout);
        Assert.Equal(TimeSpan.FromSeconds(2), clients.Reads.RequestTimeout);
    }
}
