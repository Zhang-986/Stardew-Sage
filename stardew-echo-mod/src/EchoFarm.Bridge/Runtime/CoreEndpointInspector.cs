using System.Net.Sockets;

namespace EchoFarm.Bridge.Runtime;

public interface ICoreEndpointInspector
{
    Task<CoreEndpointStatus> InspectAsync(Uri baseUri, CancellationToken cancellationToken);
}

public sealed class TcpCoreEndpointInspector : ICoreEndpointInspector
{
    private readonly ICoreHealthProbe healthProbe;
    private readonly TimeSpan timeout;

    public TcpCoreEndpointInspector(ICoreHealthProbe healthProbe, TimeSpan timeout)
    {
        this.healthProbe = healthProbe ?? throw new ArgumentNullException(nameof(healthProbe));
        if (timeout <= TimeSpan.Zero)
            throw new ArgumentOutOfRangeException(nameof(timeout));
        this.timeout = timeout;
    }

    public async Task<CoreEndpointStatus> InspectAsync(Uri baseUri, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(baseUri);
        if (!baseUri.IsAbsoluteUri || !baseUri.IsLoopback)
            throw new ArgumentException("Core base URI must be an absolute loopback URI.", nameof(baseUri));
        if (await healthProbe.IsHealthyAsync(baseUri, cancellationToken).ConfigureAwait(false))
            return CoreEndpointStatus.Healthy;

        using var client = new TcpClient();
        using var timeoutSource = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
        timeoutSource.CancelAfter(timeout);
        try
        {
            await client.ConnectAsync(baseUri.Host, baseUri.Port, timeoutSource.Token).ConfigureAwait(false);
            return CoreEndpointStatus.OccupiedUnhealthy;
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            return CoreEndpointStatus.Unreachable;
        }
        catch (SocketException)
        {
            return CoreEndpointStatus.Unreachable;
        }
    }
}
