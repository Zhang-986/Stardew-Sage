namespace EchoFarm.Bridge.Transport;

public sealed record EchoFarmClientSet(EchoFarmClient Commands, EchoFarmClient Reads);

public static class EchoFarmClientFactory
{
    private static readonly TimeSpan ReadTimeout = TimeSpan.FromSeconds(2);

    public static EchoFarmClientSet Create(Uri baseAddress, TimeSpan commandTimeout)
    {
        if (baseAddress is null)
            throw new ArgumentNullException(nameof(baseAddress));
        if (!baseAddress.IsAbsoluteUri || !baseAddress.IsLoopback)
            throw new ArgumentException("EchoFarm client requires an absolute loopback base address.", nameof(baseAddress));
        if (commandTimeout <= TimeSpan.Zero || commandTimeout > TimeSpan.FromMinutes(5))
            throw new ArgumentOutOfRangeException(nameof(commandTimeout), "Command timeout must be between 1 millisecond and 5 minutes.");

        return new EchoFarmClientSet(
            CreateClient(baseAddress, commandTimeout),
            CreateClient(baseAddress, ReadTimeout)
        );
    }

    private static EchoFarmClient CreateClient(Uri baseAddress, TimeSpan timeout) =>
        new(new HttpClient { BaseAddress = baseAddress }, timeout);
}
