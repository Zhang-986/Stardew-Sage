using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class CoreLaunchOptionsFactoryTests
{
    [Fact]
    public void BuildsUnixBundledPathAndPrivateDatabasePath()
    {
        CoreLaunchOptions options = CoreLaunchOptionsFactory.Create(new CoreLaunchSettings(
            ModDirectory: "/game/Mods/EchoFarm",
            ApplicationDataDirectory: "/users/test/.local/share",
            CoreUri: new Uri("http://127.0.0.1:18471"),
            AutoStart: true,
            ExecutablePath: null,
            StartupTimeout: TimeSpan.FromSeconds(10),
            ModelMode: "openai",
            ModelBaseUrl: "http://127.0.0.1:11434/v1",
            ModelName: "qwen",
            DatabasePath: null,
            MaxModelCallsPerSession: 32,
            MaxReportedTokensPerSession: 100000,
            IsWindows: false
        ));

        Assert.Equal(Path.Combine("/game/Mods/EchoFarm", "core", "echofarm-core"), options.ExecutablePath);
        Assert.Equal(Path.Combine("/users/test/.local/share", "EchoFarm"), options.WorkingDirectory);
        Assert.Equal("127.0.0.1:18471", options.Environment["ECHOFARM_ADDRESS"]);
        Assert.Equal("openai", options.Environment["ECHOFARM_MODEL_MODE"]);
        Assert.Equal("http://127.0.0.1:11434/v1", options.Environment["ECHOFARM_MODEL_BASE_URL"]);
        Assert.Equal("qwen", options.Environment["ECHOFARM_MODEL_NAME"]);
        Assert.Equal("32", options.Environment["ECHOFARM_MAX_MODEL_CALLS_PER_SESSION"]);
        Assert.Equal("100000", options.Environment["ECHOFARM_MAX_REPORTED_TOKENS_PER_SESSION"]);
        Assert.Equal(Path.Combine(options.WorkingDirectory, "echofarm.db"), options.Environment["ECHOFARM_DATABASE_PATH"]);
        Assert.DoesNotContain("ECHOFARM_MODEL_API_KEY", options.Environment.Keys);
        Assert.Equal(40, options.HealthCheckAttempts);
        Assert.Equal(TimeSpan.FromMilliseconds(250), options.HealthCheckInterval);
    }

    [Fact]
    public void BuildsWindowsExecutableNameAndResolvesRelativeOverride()
    {
        CoreLaunchSettings defaults = Settings() with { IsWindows = true };
        CoreLaunchSettings overridden = defaults with { ExecutablePath = "tools/custom-core.exe" };

        CoreLaunchOptions defaultOptions = CoreLaunchOptionsFactory.Create(defaults);
        CoreLaunchOptions overriddenOptions = CoreLaunchOptionsFactory.Create(overridden);

        Assert.EndsWith(Path.Combine("core", "echofarm-core.exe"), defaultOptions.ExecutablePath, StringComparison.Ordinal);
        Assert.Equal(Path.Combine(defaults.ModDirectory, "tools", "custom-core.exe"), overriddenOptions.ExecutablePath);
    }

    [Theory]
    [InlineData("http://0.0.0.0:18471", "openai", 10)]
    [InlineData("http://127.0.0.1:18471", "scripted", 10)]
    [InlineData("http://127.0.0.1:18471", "openai", 0)]
    [InlineData("http://127.0.0.1:18471", "openai", 61)]
    public void RejectsUnsafeSettings(string uri, string mode, int timeoutSeconds)
    {
        CoreLaunchSettings settings = Settings() with
        {
            CoreUri = new Uri(uri),
            ModelMode = mode,
            StartupTimeout = TimeSpan.FromSeconds(timeoutSeconds)
        };

        Assert.ThrowsAny<ArgumentException>(() => CoreLaunchOptionsFactory.Create(settings));
    }

    [Theory]
    [InlineData(0, 100000)]
    [InlineData(32, 0)]
    public void RejectsNonPositiveModelBudgets(int calls, int tokens)
    {
        CoreLaunchSettings settings = Settings() with
        {
            MaxModelCallsPerSession = calls,
            MaxReportedTokensPerSession = tokens
        };

        Assert.Throws<ArgumentOutOfRangeException>(() => CoreLaunchOptionsFactory.Create(settings));
    }

    private static CoreLaunchSettings Settings() => new(
        ModDirectory: "/game/Mods/EchoFarm",
        ApplicationDataDirectory: "/users/test/.local/share",
        CoreUri: new Uri("http://127.0.0.1:18471"),
        AutoStart: true,
        ExecutablePath: null,
        StartupTimeout: TimeSpan.FromSeconds(10),
        ModelMode: "openai",
        ModelBaseUrl: null,
        ModelName: null,
        DatabasePath: null,
        MaxModelCallsPerSession: 32,
        MaxReportedTokensPerSession: 100000,
        IsWindows: false
    );
}
