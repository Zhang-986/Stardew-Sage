namespace EchoFarm.Bridge.Runtime;

public sealed record CoreLaunchSettings(
    string ModDirectory,
    string ApplicationDataDirectory,
    Uri CoreUri,
    bool AutoStart,
    string? ExecutablePath,
    TimeSpan StartupTimeout,
    string ModelMode,
    string? ModelBaseUrl,
    string? ModelName,
    string? DatabasePath,
    bool IsWindows
);

public static class CoreLaunchOptionsFactory
{
    private static readonly TimeSpan PollInterval = TimeSpan.FromMilliseconds(250);

    public static CoreLaunchOptions Create(CoreLaunchSettings settings)
    {
        ArgumentNullException.ThrowIfNull(settings);
        if (string.IsNullOrWhiteSpace(settings.ModDirectory))
            throw new ArgumentException("Mod directory is required.", nameof(settings));
        if (string.IsNullOrWhiteSpace(settings.ApplicationDataDirectory))
            throw new ArgumentException("Application data directory is required.", nameof(settings));
        if (!settings.CoreUri.IsAbsoluteUri || !settings.CoreUri.IsLoopback)
            throw new ArgumentException("Core URI must be an absolute loopback URI.", nameof(settings));
        if (settings.StartupTimeout <= TimeSpan.Zero || settings.StartupTimeout > TimeSpan.FromSeconds(60))
            throw new ArgumentOutOfRangeException(nameof(settings), "Core startup timeout must be between 1 millisecond and 60 seconds.");
        if (settings.ModelMode is not ("openai" or "fixture"))
            throw new ArgumentException("Model mode must be openai or fixture.", nameof(settings));

        string executable = string.IsNullOrWhiteSpace(settings.ExecutablePath)
            ? Path.Combine(settings.ModDirectory, "core", settings.IsWindows ? "echofarm-core.exe" : "echofarm-core")
            : ResolvePath(settings.ModDirectory, settings.ExecutablePath);
        string workingDirectory = Path.Combine(settings.ApplicationDataDirectory, "EchoFarm");
        string databasePath = string.IsNullOrWhiteSpace(settings.DatabasePath)
            ? Path.Combine(workingDirectory, "echofarm.db")
            : ResolvePath(workingDirectory, settings.DatabasePath);

        var environment = new Dictionary<string, string>(StringComparer.Ordinal)
        {
            ["ECHOFARM_ADDRESS"] = settings.CoreUri.Authority,
            ["ECHOFARM_DATABASE_PATH"] = databasePath,
            ["ECHOFARM_MODEL_MODE"] = settings.ModelMode
        };
        AddIfPresent(environment, "ECHOFARM_MODEL_BASE_URL", settings.ModelBaseUrl);
        AddIfPresent(environment, "ECHOFARM_MODEL_NAME", settings.ModelName);

        int attempts = Math.Max(1, (int)Math.Ceiling(settings.StartupTimeout.TotalMilliseconds / PollInterval.TotalMilliseconds));
        return new CoreLaunchOptions(
            settings.CoreUri,
            settings.AutoStart,
            executable,
            workingDirectory,
            environment,
            attempts,
            PollInterval
        );
    }

    private static string ResolvePath(string root, string path) =>
        Path.GetFullPath(Path.IsPathRooted(path) ? path : Path.Combine(root, path));

    private static void AddIfPresent(IDictionary<string, string> environment, string key, string? value)
    {
        if (!string.IsNullOrWhiteSpace(value))
            environment[key] = value;
    }
}
