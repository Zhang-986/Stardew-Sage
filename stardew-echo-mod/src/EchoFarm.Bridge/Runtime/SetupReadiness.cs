namespace EchoFarm.Bridge.Runtime;

public static class SetupIssueCodes
{
    public const string Ready = "ready";
    public const string DemoMode = "demo_mode";
    public const string MissingModelUrl = "missing_model_url";
    public const string InvalidModelUrl = "invalid_model_url";
    public const string MissingModelName = "missing_model_name";
    public const string MissingApiKey = "missing_api_key";
    public const string MissingCore = "missing_core";
    public const string InvalidCoreUrl = "invalid_core_url";
    public const string UnsupportedConnectionMode = "unsupported_connection_mode";
    public const string InvalidRelayConfiguration = "invalid_relay_configuration";
    public const string UnhealthyCore = "unhealthy_core";
    public const string UnsupportedModelMode = "unsupported_model_mode";
    public const string InvalidStartupTimeout = "invalid_startup_timeout";
    public const string InvalidCommandTimeout = "invalid_command_timeout";
    public const string InvalidModelBudget = "invalid_model_budget";
}

public enum CoreEndpointStatus
{
    Unknown,
    Healthy,
    Unreachable,
    OccupiedUnhealthy
}

public sealed record SetupReadinessInput(
    string? CoreUrl,
    string? ConnectionMode,
    string? ModelMode,
    string? ModelBaseUrl,
    string? ModelName,
    bool ApiKeyPresent,
    bool AutoStartCore,
    string? CoreExecutablePath,
    bool CoreExecutableExists,
    CoreEndpointStatus EndpointStatus,
    int StartupTimeoutSeconds,
    int CommandTimeoutSeconds,
    int MaxModelCallsPerSession,
    int MaxReportedTokensPerSession
);

public sealed record SetupReadinessReport(
    string Code,
    string ModelMode,
    string Message,
    string Correction,
    bool CanAttemptStart,
    bool IsDemo
);

public sealed record SetupStatusView(
    SetupReadinessReport Readiness,
    CoreHostState CoreState,
    string SessionState,
    string DatabasePath,
    bool HarvestEnabled,
    string? LatestSafeError
);

public static class SetupReadiness
{
    public static SetupReadinessReport Evaluate(SetupReadinessInput input)
    {
        ArgumentNullException.ThrowIfNull(input);
        string connectionMode = input.ConnectionMode?.Trim().ToLowerInvariant() ?? string.Empty;
        string mode = input.ModelMode?.Trim().ToLowerInvariant() ?? string.Empty;

        if (!TryLoopbackHttpUri(input.CoreUrl, out _))
        {
            return Failure(
                SetupIssueCodes.InvalidCoreUrl,
                mode,
                "Core URL is not a safe local HTTP endpoint.",
                "Set CoreUrl to http://127.0.0.1:18471."
            );
        }
        if (connectionMode is not ("local" or "relay"))
        {
            return Failure(
                SetupIssueCodes.UnsupportedConnectionMode,
                mode,
                "The configured connection mode is unsupported.",
                "Set ConnectionMode to local or relay."
            );
        }
        if (connectionMode == "relay" && input.AutoStartCore)
        {
            return Failure(
                SetupIssueCodes.InvalidRelayConfiguration,
                "remote",
                "Relay mode cannot start a model process on Windows.",
                "Set AutoStartCore to false when ConnectionMode is relay."
            );
        }
        if (connectionMode == "local" && mode is not ("fixture" or "openai"))
        {
            return Failure(
                SetupIssueCodes.UnsupportedModelMode,
                mode,
                "The configured model mode is unsupported.",
                "Set ModelMode to fixture or openai."
            );
        }
        if (input.StartupTimeoutSeconds is < 1 or > 60)
        {
            return Failure(
                SetupIssueCodes.InvalidStartupTimeout,
                mode,
                "The core startup timeout is outside the supported range.",
                "Set CoreStartupTimeoutSeconds to a value from 1 through 60."
            );
        }
        if (input.CommandTimeoutSeconds is < 1 or > 300)
        {
            return Failure(
                SetupIssueCodes.InvalidCommandTimeout,
                connectionMode == "relay" ? "remote" : mode,
                "The command timeout is outside the supported range.",
                "Set CommandTimeoutSeconds to a value from 1 through 300."
            );
        }
        if (connectionMode == "local" && (input.MaxModelCallsPerSession <= 0 || input.MaxReportedTokensPerSession <= 0))
        {
            return Failure(
                SetupIssueCodes.InvalidModelBudget,
                mode,
                "The configured model budget is invalid.",
                "Set both model budget values to positive integers."
            );
        }
        if (connectionMode == "local" && mode == "openai")
        {
            if (string.IsNullOrWhiteSpace(input.ModelBaseUrl))
            {
                return Failure(
                    SetupIssueCodes.MissingModelUrl,
                    mode,
                    "The OpenAI-compatible model endpoint is missing.",
                    "Set ModelBaseUrl to the provider's HTTP(S) /v1 endpoint."
                );
            }
            if (!TryHttpUri(input.ModelBaseUrl, out _))
            {
                return Failure(
                    SetupIssueCodes.InvalidModelUrl,
                    mode,
                    "The OpenAI-compatible model endpoint is invalid.",
                    "Set ModelBaseUrl to an absolute HTTP(S) URL."
                );
            }
            if (string.IsNullOrWhiteSpace(input.ModelName))
            {
                return Failure(
                    SetupIssueCodes.MissingModelName,
                    mode,
                    "The model name is missing.",
                    "Set ModelName in config.json."
                );
            }
            if (!input.ApiKeyPresent)
            {
                return Failure(
                    SetupIssueCodes.MissingApiKey,
                    mode,
                    "The model API key is not present in the process environment.",
                    "Set ECHOFARM_MODEL_API_KEY before launching SMAPI; never put the key in config.json."
                );
            }
        }
        if (input.EndpointStatus == CoreEndpointStatus.OccupiedUnhealthy)
        {
            return Failure(
                SetupIssueCodes.UnhealthyCore,
                mode,
                "The configured core port is occupied by an unhealthy service.",
                "Stop the process using this port or choose another loopback port in CoreUrl."
            );
        }
        if (input.EndpointStatus == CoreEndpointStatus.Unreachable && input.AutoStartCore && !input.CoreExecutableExists)
        {
            return Failure(
                SetupIssueCodes.MissingCore,
                mode,
                "The bundled EchoFarm core executable is missing.",
                $"Reinstall EchoFarm so '{input.CoreExecutablePath}' exists."
            );
        }
        if (input.EndpointStatus == CoreEndpointStatus.Unreachable && !input.AutoStartCore)
        {
            string correction = connectionMode == "relay"
                ? "Start echofarm-relay.exe on Windows, then retry."
                : "Start the core service or enable AutoStartCore.";
            return Failure(
                SetupIssueCodes.UnhealthyCore,
                connectionMode == "relay" ? "remote" : mode,
                connectionMode == "relay"
                    ? "No healthy EchoFarm relay is listening on Windows loopback."
                    : "No healthy EchoFarm core is listening and auto-start is disabled.",
                correction
            );
        }

        if (connectionMode == "relay")
        {
            return new SetupReadinessReport(
                SetupIssueCodes.Ready,
                "remote",
                "EchoFarm secure relay is ready; AI and memory stay on the Mac server.",
                string.Empty,
                CanAttemptStart: true,
                IsDemo: false
            );
        }

        bool demo = mode == "fixture";
        return new SetupReadinessReport(
            demo ? SetupIssueCodes.DemoMode : SetupIssueCodes.Ready,
            mode,
            demo ? "EchoFarm is ready in deterministic DEMO mode." : "EchoFarm AI configuration is ready.",
            demo ? "Set ModelMode to openai and provide model settings when you want real model calls." : string.Empty,
            CanAttemptStart: true,
            IsDemo: demo
        );
    }

    public static SetupReadinessReport CoreFailure(SetupReadinessReport current, string correction)
    {
        ArgumentNullException.ThrowIfNull(current);
        return new SetupReadinessReport(
            SetupIssueCodes.UnhealthyCore,
            current.ModelMode,
            "EchoFarm core did not become healthy.",
            correction,
            CanAttemptStart: true,
            current.IsDemo
        );
    }

    public static IReadOnlyList<string> BuildStatusLines(SetupStatusView status)
    {
        ArgumentNullException.ThrowIfNull(status);
        string marker = status.Readiness.IsDemo ? " DEMO" : string.Empty;
        var lines = new List<string>(7)
        {
            $"EchoFarm{marker} · setup={status.Readiness.Code}",
            $"Model: {status.Readiness.ModelMode} · Core: {status.CoreState}",
            $"Session: {status.SessionState}",
            $"Database: {status.DatabasePath}",
            $"Harvest: {(status.HarvestEnabled ? "enabled (experimental)" : "disabled")}",
            "Learn-only: chopping, mining, mine floors, fishing",
            $"Next: {status.Readiness.Correction}"
        };
        if (!string.IsNullOrWhiteSpace(status.LatestSafeError))
            lines.Add($"Latest safe error: {status.LatestSafeError}");
        return Array.AsReadOnly(lines.ToArray());
    }

    private static SetupReadinessReport Failure(string code, string mode, string message, string correction) =>
        new(code, mode, message, correction, CanAttemptStart: false, IsDemo: mode == "fixture");

    private static bool TryLoopbackHttpUri(string? value, out Uri? uri) =>
        TryHttpUri(value, out uri) && uri!.IsLoopback;

    private static bool TryHttpUri(string? value, out Uri? uri)
    {
        if (!Uri.TryCreate(value, UriKind.Absolute, out uri))
            return false;
        return uri.Scheme == Uri.UriSchemeHttp || uri.Scheme == Uri.UriSchemeHttps;
    }
}
