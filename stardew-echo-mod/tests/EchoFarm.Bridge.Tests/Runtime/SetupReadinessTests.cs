using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class SetupReadinessTests
{
    [Fact]
    public void FixtureModeIsReadyButExplicitlyMarkedAsDemo()
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(modelMode: "fixture"));

        Assert.Equal(SetupIssueCodes.DemoMode, report.Code);
        Assert.True(report.CanAttemptStart);
        Assert.True(report.IsDemo);
        Assert.Contains("DEMO", report.Message, StringComparison.Ordinal);
    }

    [Fact]
    public void OpenAiModeIsReadyWhenProviderSettingsAndEnvironmentKeyExist()
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(
            modelMode: "openai",
            modelBaseUrl: "https://api.openai.com/v1",
            modelName: "gpt-5-mini",
            apiKeyPresent: true));

        Assert.Equal(SetupIssueCodes.Ready, report.Code);
        Assert.True(report.CanAttemptStart);
        Assert.False(report.IsDemo);
    }

    [Theory]
    [InlineData(null, "gpt-5-mini", true, SetupIssueCodes.MissingModelUrl)]
    [InlineData("https://api.openai.com/v1", null, true, SetupIssueCodes.MissingModelName)]
    [InlineData("https://api.openai.com/v1", "gpt-5-mini", false, SetupIssueCodes.MissingApiKey)]
    public void OpenAiModeReportsOneActionableMissingSetting(
        string? modelBaseUrl,
        string? modelName,
        bool apiKeyPresent,
        string expectedCode)
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(
            modelMode: "openai",
            modelBaseUrl: modelBaseUrl,
            modelName: modelName,
            apiKeyPresent: apiKeyPresent));

        Assert.Equal(expectedCode, report.Code);
        Assert.False(report.CanAttemptStart);
        Assert.False(string.IsNullOrWhiteSpace(report.Correction));
    }

    [Theory]
    [InlineData("not-a-url")]
    [InlineData("https://example.com:18471")]
    [InlineData("ftp://127.0.0.1:18471")]
    public void InvalidOrNonLoopbackCoreUrlFailsClosed(string coreUrl)
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(coreUrl: coreUrl));

        Assert.Equal(SetupIssueCodes.InvalidCoreUrl, report.Code);
        Assert.False(report.CanAttemptStart);
        Assert.Contains("127.0.0.1", report.Correction, StringComparison.Ordinal);
    }

    [Fact]
    public void MissingBundledCoreIsReportedBeforeProcessLaunch()
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(
            coreExecutableExists: false,
            endpointStatus: CoreEndpointStatus.Unreachable));

        Assert.Equal(SetupIssueCodes.MissingCore, report.Code);
        Assert.False(report.CanAttemptStart);
        Assert.Contains("install", report.Correction, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public void UnknownEndpointDefersMissingExecutableDecisionUntilInspection()
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(
            coreExecutableExists: false,
            endpointStatus: CoreEndpointStatus.Unknown));

        Assert.True(report.CanAttemptStart);
        Assert.Equal(SetupIssueCodes.DemoMode, report.Code);
    }

    [Fact]
    public void ExistingExternalCoreDoesNotRequireBundledExecutable()
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(
            autoStartCore: false,
            coreExecutableExists: false,
            endpointStatus: CoreEndpointStatus.Healthy));

        Assert.Equal(SetupIssueCodes.DemoMode, report.Code);
        Assert.True(report.CanAttemptStart);
    }

    [Fact]
    public void OccupiedButUnhealthyEndpointHasDistinctIssueCode()
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(
            endpointStatus: CoreEndpointStatus.OccupiedUnhealthy));

        Assert.Equal(SetupIssueCodes.UnhealthyCore, report.Code);
        Assert.False(report.CanAttemptStart);
        Assert.Contains("port", report.Correction, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public void UnsupportedModelModeFailsClosed()
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(modelMode: "magic"));

        Assert.Equal(SetupIssueCodes.UnsupportedModelMode, report.Code);
        Assert.False(report.CanAttemptStart);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(61)]
    public void InvalidStartupTimeoutFailsClosed(int seconds)
    {
        SetupReadinessReport report = SetupReadiness.Evaluate(Settings(startupTimeoutSeconds: seconds));

        Assert.Equal(SetupIssueCodes.InvalidStartupTimeout, report.Code);
        Assert.False(report.CanAttemptStart);
    }

    [Fact]
    public void StatusLinesExposeOperationalStateWithoutSecrets()
    {
        const string secret = "sk-do-not-display";
        SetupReadinessReport readiness = SetupReadiness.Evaluate(Settings(
            modelMode: "openai",
            modelBaseUrl: "https://api.openai.com/v1",
            modelName: "gpt-5-mini",
            apiKeyPresent: !string.IsNullOrWhiteSpace(secret)));
        var status = new SetupStatusView(
            readiness,
            CoreHostState.Owned,
            "Acting",
            "C:/Users/test/AppData/Local/EchoFarm/echofarm.db",
            HarvestEnabled: false,
            LatestSafeError: "path_blocked");

        string rendered = string.Join("\n", SetupReadiness.BuildStatusLines(status));

        Assert.Contains("openai", rendered, StringComparison.OrdinalIgnoreCase);
        Assert.Contains("Owned", rendered, StringComparison.Ordinal);
        Assert.Contains("Acting", rendered, StringComparison.Ordinal);
        Assert.Contains("echofarm.db", rendered, StringComparison.Ordinal);
        Assert.Contains("disabled", rendered, StringComparison.OrdinalIgnoreCase);
        Assert.Contains("path_blocked", rendered, StringComparison.Ordinal);
        Assert.DoesNotContain(secret, rendered, StringComparison.Ordinal);
    }

    private static SetupReadinessInput Settings(
        string coreUrl = "http://127.0.0.1:18471",
        string modelMode = "fixture",
        string? modelBaseUrl = null,
        string? modelName = null,
        bool apiKeyPresent = false,
        bool autoStartCore = true,
        bool coreExecutableExists = true,
        CoreEndpointStatus endpointStatus = CoreEndpointStatus.Unknown,
        int startupTimeoutSeconds = 10) => new(
            CoreUrl: coreUrl,
            ModelMode: modelMode,
            ModelBaseUrl: modelBaseUrl,
            ModelName: modelName,
            ApiKeyPresent: apiKeyPresent,
            AutoStartCore: autoStartCore,
            CoreExecutablePath: "C:/Games/Stardew Valley/Mods/EchoFarm/core/echofarm-core.exe",
            CoreExecutableExists: coreExecutableExists,
            EndpointStatus: endpointStatus,
            StartupTimeoutSeconds: startupTimeoutSeconds);
}
