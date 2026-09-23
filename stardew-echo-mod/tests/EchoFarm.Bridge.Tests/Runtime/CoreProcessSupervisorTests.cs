using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class CoreProcessSupervisorTests
{
    [Fact]
    public async Task ReusesHealthyExternalCoreWithoutLaunchingOrKillingIt()
    {
        var probe = new ProbeStub(true);
        var launcher = new LauncherStub();
        var supervisor = new CoreProcessSupervisor(probe, launcher, new DelayStub());

        CoreHostState state = await supervisor.StartAsync(Options(), CancellationToken.None);
        supervisor.Stop();

        Assert.Equal(CoreHostState.External, state);
        Assert.Equal(0, launcher.StartCalls);
        Assert.Null(launcher.Process);
    }

    [Fact]
    public async Task DisabledAutoStartReportsUnavailable()
    {
        var launcher = new LauncherStub();
        var supervisor = new CoreProcessSupervisor(new ProbeStub(false), launcher, new DelayStub());

        CoreUnavailableException error = await Assert.ThrowsAsync<CoreUnavailableException>(
            () => supervisor.StartAsync(Options(autoStart: false), CancellationToken.None));

        Assert.Contains("disabled", error.Message, StringComparison.OrdinalIgnoreCase);
        Assert.Equal(CoreHostState.Failed, supervisor.State);
        Assert.Equal(0, launcher.StartCalls);
    }

    [Fact]
    public async Task LaunchedChildBecomesOwnedAfterHealthCheckSucceeds()
    {
        var probe = new ProbeStub(false, false, true);
        var launcher = new LauncherStub();
        var delay = new DelayStub();
        var supervisor = new CoreProcessSupervisor(probe, launcher, delay);

        CoreHostState state = await supervisor.StartAsync(Options(), CancellationToken.None);
        CoreHostState repeated = await supervisor.StartAsync(Options(), CancellationToken.None);

        Assert.Equal(CoreHostState.Owned, state);
        Assert.Equal(CoreHostState.Owned, repeated);
        Assert.Equal(1, launcher.StartCalls);
        Assert.Equal(1, delay.Calls);
    }

    [Fact]
    public async Task ChildExitBeforeHealthReportsExitCode()
    {
        var process = new ProcessStub { HasExited = true, ExitCode = 17 };
        var launcher = new LauncherStub(process);
        var supervisor = new CoreProcessSupervisor(new ProbeStub(false), launcher, new DelayStub());

        CoreUnavailableException error = await Assert.ThrowsAsync<CoreUnavailableException>(
            () => supervisor.StartAsync(Options(), CancellationToken.None));

        Assert.Contains("17", error.Message, StringComparison.Ordinal);
        Assert.Equal(CoreHostState.Failed, supervisor.State);
        Assert.True(process.Disposed);
        Assert.False(process.Terminated);
    }

    [Fact]
    public async Task StartupTimeoutTerminatesOnlyTheOwnedChild()
    {
        var process = new ProcessStub();
        var launcher = new LauncherStub(process);
        var supervisor = new CoreProcessSupervisor(new ProbeStub(false, false, false), launcher, new DelayStub());

        await Assert.ThrowsAsync<TimeoutException>(
            () => supervisor.StartAsync(Options(healthCheckAttempts: 2), CancellationToken.None));

        Assert.True(process.Terminated);
        Assert.True(process.Disposed);
        Assert.Equal(CoreHostState.Failed, supervisor.State);
    }

    [Fact]
    public async Task LaunchFailureMarksSupervisorFailed()
    {
        var supervisor = new CoreProcessSupervisor(
            new ProbeStub(false),
            new ThrowingLauncher(),
            new DelayStub()
        );

        await Assert.ThrowsAsync<CoreUnavailableException>(
            () => supervisor.StartAsync(Options(), CancellationToken.None));

        Assert.Equal(CoreHostState.Failed, supervisor.State);
    }

    [Fact]
    public async Task StopTerminatesAnOwnedChildOnce()
    {
        var process = new ProcessStub();
        var supervisor = new CoreProcessSupervisor(
            new ProbeStub(false, true),
            new LauncherStub(process),
            new DelayStub()
        );
        await supervisor.StartAsync(Options(), CancellationToken.None);

        supervisor.Stop();
        supervisor.Stop();

        Assert.Equal(1, process.TerminateCalls);
        Assert.True(process.Disposed);
        Assert.Equal(CoreHostState.Stopped, supervisor.State);
    }

    private static CoreLaunchOptions Options(bool autoStart = true, int healthCheckAttempts = 3) => new(
        BaseUri: new Uri("http://127.0.0.1:18471"),
        AutoStart: autoStart,
        ExecutablePath: "/mods/EchoFarm/core/echofarm-core",
        WorkingDirectory: "/mods/EchoFarm",
        Environment: new Dictionary<string, string>(),
        HealthCheckAttempts: healthCheckAttempts,
        HealthCheckInterval: TimeSpan.FromMilliseconds(1)
    );

    private sealed class ProbeStub : ICoreHealthProbe
    {
        private readonly Queue<bool> results;
        private bool last;

        public ProbeStub(params bool[] results)
        {
            this.results = new Queue<bool>(results);
            last = results.LastOrDefault();
        }

        public Task<bool> IsHealthyAsync(Uri baseUri, CancellationToken cancellationToken)
        {
            if (results.TryDequeue(out bool result))
                last = result;
            return Task.FromResult(last);
        }
    }

    private sealed class LauncherStub : ICoreProcessLauncher
    {
        public LauncherStub(ProcessStub? process = null)
        {
            Process = process;
        }

        public ProcessStub? Process { get; private set; }
        public int StartCalls { get; private set; }

        public ICoreProcess Start(CoreProcessStartInfo startInfo)
        {
            StartCalls++;
            return Process ??= new ProcessStub();
        }
    }

    private sealed class ProcessStub : ICoreProcess
    {
        public bool HasExited { get; set; }
        public int? ExitCode { get; set; }
        public int TerminateCalls { get; private set; }
        public bool Terminated => TerminateCalls > 0;
        public bool Disposed { get; private set; }

        public void Terminate()
        {
            TerminateCalls++;
            HasExited = true;
        }

        public void Dispose() => Disposed = true;
    }

    private sealed class ThrowingLauncher : ICoreProcessLauncher
    {
        public ICoreProcess Start(CoreProcessStartInfo startInfo) =>
            throw new CoreUnavailableException("missing executable");
    }

    private sealed class DelayStub : IAsyncDelay
    {
        public int Calls { get; private set; }

        public Task WaitAsync(TimeSpan delay, CancellationToken cancellationToken)
        {
            Calls++;
            return Task.CompletedTask;
        }
    }
}
