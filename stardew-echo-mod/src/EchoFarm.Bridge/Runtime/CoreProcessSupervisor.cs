namespace EchoFarm.Bridge.Runtime;

public enum CoreHostState
{
    Stopped,
    External,
    Owned,
    Failed
}

public sealed record CoreLaunchOptions(
    Uri BaseUri,
    bool AutoStart,
    string ExecutablePath,
    string WorkingDirectory,
    IReadOnlyDictionary<string, string> Environment,
    int HealthCheckAttempts,
    TimeSpan HealthCheckInterval
);

public sealed record CoreProcessStartInfo(
    string ExecutablePath,
    string WorkingDirectory,
    IReadOnlyDictionary<string, string> Environment
);

public interface ICoreHealthProbe
{
    Task<bool> IsHealthyAsync(Uri baseUri, CancellationToken cancellationToken);
}

public interface ICoreProcess : IDisposable
{
    bool HasExited { get; }
    int? ExitCode { get; }
    void Terminate();
}

public interface ICoreProcessLauncher
{
    ICoreProcess Start(CoreProcessStartInfo startInfo);
}

public interface IAsyncDelay
{
    Task WaitAsync(TimeSpan delay, CancellationToken cancellationToken);
}

public sealed class CoreUnavailableException : Exception
{
    public CoreUnavailableException(string message) : base(message) { }
}

public sealed class CoreProcessSupervisor : IDisposable
{
    private readonly ICoreHealthProbe healthProbe;
    private readonly ICoreProcessLauncher processLauncher;
    private readonly IAsyncDelay delay;
    private readonly SemaphoreSlim startGate = new(1, 1);
    private ICoreProcess? ownedProcess;

    public CoreProcessSupervisor(
        ICoreHealthProbe healthProbe,
        ICoreProcessLauncher processLauncher,
        IAsyncDelay delay)
    {
        this.healthProbe = healthProbe ?? throw new ArgumentNullException(nameof(healthProbe));
        this.processLauncher = processLauncher ?? throw new ArgumentNullException(nameof(processLauncher));
        this.delay = delay ?? throw new ArgumentNullException(nameof(delay));
    }

    public CoreHostState State { get; private set; } = CoreHostState.Stopped;

    public async Task<CoreHostState> StartAsync(CoreLaunchOptions options, CancellationToken cancellationToken)
    {
        Validate(options);
        await startGate.WaitAsync(cancellationToken).ConfigureAwait(false);
        try
        {
            if (State is CoreHostState.External or CoreHostState.Owned)
                return State;
            if (await healthProbe.IsHealthyAsync(options.BaseUri, cancellationToken).ConfigureAwait(false))
            {
                State = CoreHostState.External;
                return State;
            }
            if (!options.AutoStart)
            {
                State = CoreHostState.Failed;
                throw new CoreUnavailableException("EchoFarm core auto-start is disabled and no healthy service is running.");
            }

            try
            {
                ownedProcess = processLauncher.Start(new CoreProcessStartInfo(
                    options.ExecutablePath,
                    options.WorkingDirectory,
                    options.Environment
                ));
            }
            catch
            {
                State = CoreHostState.Failed;
                throw;
            }
            try
            {
                for (int attempt = 0; attempt < options.HealthCheckAttempts; attempt++)
                {
                    if (ownedProcess.HasExited)
                    {
                        int? exitCode = ownedProcess.ExitCode;
                        DisposeOwnedProcess(terminate: false);
                        State = CoreHostState.Failed;
                        throw new CoreUnavailableException($"EchoFarm core exited before becoming healthy (exit code {exitCode?.ToString() ?? "unknown"}).");
                    }
                    if (await healthProbe.IsHealthyAsync(options.BaseUri, cancellationToken).ConfigureAwait(false))
                    {
                        State = CoreHostState.Owned;
                        return State;
                    }
                    if (attempt + 1 < options.HealthCheckAttempts)
                        await delay.WaitAsync(options.HealthCheckInterval, cancellationToken).ConfigureAwait(false);
                }
            }
            catch
            {
                if (ownedProcess is not null)
                    DisposeOwnedProcess(terminate: !ownedProcess.HasExited);
                State = CoreHostState.Failed;
                throw;
            }

            DisposeOwnedProcess(terminate: true);
            State = CoreHostState.Failed;
            throw new TimeoutException("EchoFarm core did not become healthy before the startup timeout.");
        }
        finally
        {
            startGate.Release();
        }
    }

    public void Stop()
    {
        if (ownedProcess is not null)
            DisposeOwnedProcess(terminate: !ownedProcess.HasExited);
        State = CoreHostState.Stopped;
    }

    public void Dispose()
    {
        Stop();
        startGate.Dispose();
    }

    private void DisposeOwnedProcess(bool terminate)
    {
        ICoreProcess process = ownedProcess!;
        ownedProcess = null;
        if (terminate)
            process.Terminate();
        process.Dispose();
    }

    private static void Validate(CoreLaunchOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        if (!options.BaseUri.IsAbsoluteUri || !options.BaseUri.IsLoopback)
            throw new ArgumentException("Core base URI must be an absolute loopback URI.", nameof(options));
        if (string.IsNullOrWhiteSpace(options.ExecutablePath))
            throw new ArgumentException("Core executable path is required.", nameof(options));
        if (string.IsNullOrWhiteSpace(options.WorkingDirectory))
            throw new ArgumentException("Core working directory is required.", nameof(options));
        if (options.HealthCheckAttempts <= 0)
            throw new ArgumentOutOfRangeException(nameof(options));
        if (options.HealthCheckInterval < TimeSpan.Zero)
            throw new ArgumentOutOfRangeException(nameof(options));
    }
}
