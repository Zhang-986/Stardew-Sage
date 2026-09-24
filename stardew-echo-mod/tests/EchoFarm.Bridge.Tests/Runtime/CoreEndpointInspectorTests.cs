using System.Net;
using System.Net.Sockets;
using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class CoreEndpointInspectorTests
{
    [Fact]
    public async Task HealthyEndpointWinsWithoutPortFallback()
    {
        var inspector = new TcpCoreEndpointInspector(new ProbeStub(healthy: true), TimeSpan.FromMilliseconds(250));

        CoreEndpointStatus status = await inspector.InspectAsync(
            new Uri("http://127.0.0.1:18471"),
            CancellationToken.None);

        Assert.Equal(CoreEndpointStatus.Healthy, status);
    }

    [Fact]
    public async Task OpenPortWithoutHealthyCoreIsReportedAsOccupied()
    {
        using var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int port = ((IPEndPoint)listener.LocalEndpoint).Port;
        var inspector = new TcpCoreEndpointInspector(new ProbeStub(healthy: false), TimeSpan.FromMilliseconds(250));

        CoreEndpointStatus status = await inspector.InspectAsync(
            new Uri($"http://127.0.0.1:{port}"),
            CancellationToken.None);

        Assert.Equal(CoreEndpointStatus.OccupiedUnhealthy, status);
    }

    [Fact]
    public async Task ClosedPortIsReportedAsUnreachable()
    {
        int port;
        using (var listener = new TcpListener(IPAddress.Loopback, 0))
        {
            listener.Start();
            port = ((IPEndPoint)listener.LocalEndpoint).Port;
        }
        var inspector = new TcpCoreEndpointInspector(new ProbeStub(healthy: false), TimeSpan.FromMilliseconds(250));

        CoreEndpointStatus status = await inspector.InspectAsync(
            new Uri($"http://127.0.0.1:{port}"),
            CancellationToken.None);

        Assert.Equal(CoreEndpointStatus.Unreachable, status);
    }

    private sealed class ProbeStub : ICoreHealthProbe
    {
        private readonly bool healthy;

        public ProbeStub(bool healthy)
        {
            this.healthy = healthy;
        }

        public Task<bool> IsHealthyAsync(Uri baseUri, CancellationToken cancellationToken) =>
            Task.FromResult(healthy);
    }
}
