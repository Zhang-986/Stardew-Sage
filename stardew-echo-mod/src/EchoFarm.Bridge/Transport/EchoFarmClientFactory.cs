namespace EchoFarm.Bridge.Transport;

public sealed record EchoFarmClientSet(EchoFarmClient Commands, EchoFarmClient Reads);

public static class EchoFarmClientFactory
{
    private static readonly TimeSpan CommandTimeout = TimeSpan.FromSeconds(35);
    private static readonly TimeSpan ReadTimeout = TimeSpan.FromSeconds(2);

    public static EchoFarmClientSet Create(Uri baseAddress)
    {
        if (baseAddress is null)
            throw new ArgumentNullException(nameof(baseAddress));
        if (!baseAddress.IsAbsoluteUri || !baseAddress.IsLoopback)
            throw new ArgumentException("EchoFarm client requires an absolute loopback base address.", nameof(baseAddress));

        return new EchoFarmClientSet(
            CreateClient(baseAddress, CommandTimeout),
            CreateClient(baseAddress, ReadTimeout)
        );
    }

    private static EchoFarmClient CreateClient(Uri baseAddress, TimeSpan timeout) =>
        new(new HttpClient { BaseAddress = baseAddress }, timeout);
}
