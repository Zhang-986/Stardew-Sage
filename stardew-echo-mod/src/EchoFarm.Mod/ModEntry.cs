using System.Globalization;
using EchoFarm.Bridge.Recording;
using EchoFarm.Bridge.Runtime;
using EchoFarm.Bridge.Transport;
using StardewModdingAPI;
using StardewModdingAPI.Events;
using StardewValley;

namespace EchoFarm.Mod;

public sealed class ModEntry : Mod
{
    private ModConfig config = null!;
    private TeachingRecorder recorder = null!;
    private StardewGamePort gamePort = null!;
    private EchoSession? session;
    private EchoRenderer renderer = null!;
    private EchoMemoryOverlay memoryOverlay = null!;
    private EchoFarmClient? readClient;
    private CoreProcessSupervisor? coreHost;
    private CoreLaunchOptions? coreLaunchOptions;
    private ICoreEndpointInspector? coreEndpointInspector;
    private SetupReadinessInput setupInput = null!;
    private SetupReadinessReport setupReadiness = null!;
    private string databasePath = string.Empty;
    private string? latestSafeError;
    private Task<bool>? coreStartup;
    private readonly CancellationTokenSource modLifetime = new();
    private CancellationTokenSource saveLifetime = new();
    private ObservationProbe? pendingObservation;
    private bool memoryRefreshInFlight;

    public override void Entry(IModHelper helper)
    {
        config = helper.ReadConfig<ModConfig>();
        recorder = new TeachingRecorder();
        gamePort = new StardewGamePort(Monitor, config.EnableExperimentalHarvest);
        renderer = new EchoRenderer(gamePort.Echo);
        memoryOverlay = new EchoMemoryOverlay();

        string applicationData = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
        if (string.IsNullOrWhiteSpace(applicationData))
            applicationData = helper.DirectoryPath;
        string workingDirectory = Path.Combine(applicationData, "EchoFarm");
        string executablePath = ResolveCoreExecutablePath(helper.DirectoryPath);
        databasePath = ResolveDatabasePath(workingDirectory);
        setupInput = new SetupReadinessInput(
            config.CoreUrl,
            config.ModelMode,
            config.ModelBaseUrl,
            config.ModelName,
            ApiKeyPresent: !string.IsNullOrWhiteSpace(Environment.GetEnvironmentVariable("ECHOFARM_MODEL_API_KEY")),
            config.AutoStartCore,
            executablePath,
            File.Exists(executablePath),
            CoreEndpointStatus.Unknown,
            config.CoreStartupTimeoutSeconds
        );
        setupReadiness = SetupReadiness.Evaluate(setupInput);
        UpdateOperationalStatus();

        if (setupReadiness.CanAttemptStart && Uri.TryCreate(config.CoreUrl, UriKind.Absolute, out Uri? coreUrl))
        {
            coreLaunchOptions = CoreLaunchOptionsFactory.Create(new CoreLaunchSettings(
                helper.DirectoryPath,
                applicationData,
                coreUrl,
                config.AutoStartCore,
                config.CoreExecutablePath,
                TimeSpan.FromSeconds(config.CoreStartupTimeoutSeconds),
                setupReadiness.ModelMode,
                config.ModelBaseUrl,
                config.ModelName,
                config.DatabasePath,
                OperatingSystem.IsWindows()
            ));
            databasePath = coreLaunchOptions.Environment["ECHOFARM_DATABASE_PATH"];
            var healthProbe = new HttpCoreHealthProbe(new HttpClient(), TimeSpan.FromMilliseconds(500));
            coreEndpointInspector = new TcpCoreEndpointInspector(healthProbe, TimeSpan.FromMilliseconds(500));
            coreHost = new CoreProcessSupervisor(
                healthProbe,
                new SystemCoreProcessLauncher((message, isError) =>
                    Monitor.Log($"[EchoFarm Core] {message}", isError ? LogLevel.Warn : LogLevel.Trace)),
                new TaskAsyncDelay()
            );
            EchoFarmClientSet clients = EchoFarmClientFactory.Create(coreUrl);
            readClient = clients.Reads;
            session = new EchoSession(recorder, clients.Commands, gamePort, new ActionSafetyGate());
            UpdateOperationalStatus();
        }
        else
        {
            LogSetupIssue();
        }

        helper.Events.GameLoop.GameLaunched += OnGameLaunched;
        helper.Events.GameLoop.SaveLoaded += OnSaveLoaded;
        helper.Events.GameLoop.DayStarted += OnDayStarted;
        helper.Events.GameLoop.Saving += OnSaving;
        helper.Events.GameLoop.ReturnedToTitle += OnReturnedToTitle;
        helper.Events.GameLoop.UpdateTicked += OnUpdateTicked;
        helper.Events.Input.ButtonPressed += OnButtonPressed;
        helper.Events.Display.RenderedWorld += OnRenderedWorld;
        helper.Events.Display.RenderedHud += OnRenderedHud;
        AppDomain.CurrentDomain.ProcessExit += OnProcessExit;
    }

    private async void OnGameLaunched(object? sender, GameLaunchedEventArgs e) =>
        await EnsureCoreStartedAsync();

    private async void OnSaveLoaded(object? sender, SaveLoadedEventArgs e)
    {
        ResetSaveLifetime();
        await RestoreLearnedStateAsync();
    }

    private async void OnDayStarted(object? sender, DayStartedEventArgs e)
    {
        if (session?.State is EchoSessionState.Acting or EchoSessionState.AwaitingResult or EchoSessionState.Correcting)
            session.Abort();
        ResetSaveLifetime();
        await RestoreLearnedStateAsync();
    }

    private void OnSaving(object? sender, SavingEventArgs e) => StopForWorldChange();

    private void OnReturnedToTitle(object? sender, ReturnedToTitleEventArgs e) => StopForWorldChange();

    private async void OnButtonPressed(object? sender, ButtonPressedEventArgs e)
    {
        if (!Context.IsWorldReady)
            return;

        if (e.Button == config.MemoryKey)
        {
            memoryOverlay.Toggle();
            UpdateOperationalStatus();
            if (memoryOverlay.Visible)
                await RefreshMemoryAsync();
            return;
        }

        if (session is null)
        {
            LogSetupIssue();
            return;
        }

        if (e.Button == config.CorrectionKey)
        {
            if (session.State == EchoSessionState.Correcting)
            {
                if (session.HasPendingCorrection)
                {
                    bool learned = await session.RetryCorrectionAsync(saveLifetime.Token);
                    Monitor.Log(
                        learned ? "Echo accepted the correction and will replan." : "Echo is still paused; press the correction key to retry.",
                        learned ? LogLevel.Info : LogLevel.Warn
                    );
                }
                else
                {
                    session.CancelCorrection();
                    Monitor.Log("Echo correction cancelled.", LogLevel.Info);
                }
            }
            else if (session.BeginCorrection(Game1.ticks))
            {
                Monitor.Log("Echo paused. Perform one successful farm action within 20 seconds to teach the better choice.", LogLevel.Info);
            }
            else
            {
                Monitor.Log("Echo has no pending decision to correct.", LogLevel.Info);
            }
            return;
        }

        if (e.Button == config.RecordKey)
        {
            if (session.State == EchoSessionState.Recording)
            {
                try
                {
                    EchoSessionState state = await session.CompleteTeachingAsync(Game1.ticks, saveLifetime.Token);
                    Monitor.Log($"Echo learned the morning routine ({state}). Press {config.SummonKey} to summon it.", LogLevel.Info);
                }
                catch (Exception error)
                {
                    Monitor.Log($"Echo could not learn this demonstration: {error.Message}", LogLevel.Warn);
                }
            }
            else if (session.State is EchoSessionState.Idle or EchoSessionState.Ready)
            {
                session.BeginTeaching(
                    SaveId(),
                    Game1.ticks,
                    Math.Max(1, Game1.Date.TotalDays),
                    WorldSnapshotMapper.GetWeather(Game1.currentLocation)
                );
                Monitor.Log("Echo recording started. Play the morning routine, then press the record key again.", LogLevel.Info);
            }
            return;
        }

        if (e.Button == config.SummonKey)
        {
            if (session.State != EchoSessionState.Ready)
            {
                Monitor.Log("Teach Echo a routine before summoning it.", LogLevel.Info);
                return;
            }
            gamePort.ResetPlayerActivity();
            session.StartEcho(SaveId(), $"echo-{Game1.Date.TotalDays}-{Guid.NewGuid():N}");
            Monitor.Log("Echo joined the farm.", LogLevel.Info);
            return;
        }

        if ((session.State is EchoSessionState.Recording or EchoSessionState.Acting or EchoSessionState.AwaitingResult or EchoSessionState.Correcting) &&
            (e.Button.IsUseToolButton() || e.Button.IsActionButton()))
            pendingObservation = gamePort.BeginObservation(e.Button, Game1.ticks);
    }

    private async void OnUpdateTicked(object? sender, UpdateTickedEventArgs e)
    {
        if (!Context.IsWorldReady)
            return;

        gamePort.Pump();
        if (session is null)
            return;
        if (pendingObservation is not null && e.Ticks > (ulong)pendingObservation.Tick)
        {
            ObservedGameEvent observed = gamePort.CompleteObservation(pendingObservation, Game1.ticks);
            if (session.State == EchoSessionState.Recording)
                session.Observe(observed);
            else if (session.State == EchoSessionState.Correcting)
            {
                bool learned = await session.ObserveCorrectionAsync(observed, saveLifetime.Token);
                if (learned)
                {
                    Monitor.Log("Echo learned the corrected action and will use it in similar situations.", LogLevel.Info);
                    if (memoryOverlay.Visible)
                        await RefreshMemoryAsync();
                }
                else if (session.HasPendingCorrection && session.LastError is not null)
                {
                    Monitor.Log("Echo could not save the correction and remains paused. Press the correction key to retry.", LogLevel.Warn);
                }
            }
            else if (session.State is EchoSessionState.Acting or EchoSessionState.AwaitingResult)
                gamePort.RecordPlayerActivity(observed);
            pendingObservation = null;
        }
        if (session.ExpireCorrection(Game1.ticks))
            Monitor.Log("Echo correction timed out; normal planning resumed.", LogLevel.Info);
        if (session.State == EchoSessionState.Recording)
        {
            ObservedGameEvent? movement = gamePort.ObserveMovement(Game1.ticks);
            if (movement is not null)
                session.Observe(movement);
        }
        if (session.State == EchoSessionState.Acting && e.IsMultipleOf(15))
            await session.TickAsync(saveLifetime.Token);
        if (memoryOverlay.Visible && e.IsMultipleOf(60) && !memoryRefreshInFlight)
            await RefreshMemoryAsync();
    }

    private void OnRenderedWorld(object? sender, RenderedWorldEventArgs e) => renderer.Draw(e.SpriteBatch);

    private void OnRenderedHud(object? sender, RenderedHudEventArgs e) => memoryOverlay.Draw(e.SpriteBatch);

    private void StopForWorldChange()
    {
        pendingObservation = null;
        gamePort.ResetPlayerActivity();
        memoryOverlay.Hide();
        saveLifetime.Cancel();
        session?.Abort();
        UpdateOperationalStatus();
    }

    private async Task RefreshMemoryAsync()
    {
        if (memoryRefreshInFlight)
            return;
        memoryRefreshInFlight = true;
        try
        {
            UpdateOperationalStatus();
            if (readClient is null || session is null || !await EnsureCoreStartedAsync())
            {
                memoryOverlay.ShowError($"Setup blocked [{setupReadiness.Code}]. {setupReadiness.Correction}");
                return;
            }
            EchoFarm.Bridge.Contracts.EchoMemoryView view = await readClient.GetMemoryAsync(SaveId(), saveLifetime.Token);
            memoryOverlay.Update(view);
            latestSafeError = null;
            UpdateOperationalStatus();
        }
        catch (EchoMemoryNotFoundException)
        {
            memoryOverlay.ShowError($"No learned memory yet. Press {config.RecordKey} to teach Echo.");
        }
        catch (Exception error) when (error is EchoFarmException or OperationCanceledException)
        {
            latestSafeError = error is OperationCanceledException ? "request_cancelled" : "core_request_failed";
            UpdateOperationalStatus();
            memoryOverlay.ShowError(error is OperationCanceledException
                ? "Request cancelled safely."
                : "Core request failed safely; check the setup action above.");
        }
        finally
        {
            memoryRefreshInFlight = false;
        }
    }

    private void ResetSaveLifetime()
    {
        saveLifetime.Dispose();
        saveLifetime = new CancellationTokenSource();
    }

    private async Task RestoreLearnedStateAsync()
    {
        if (readClient is null || session is null || !await EnsureCoreStartedAsync())
            return;
        try
        {
            await readClient.GetPlayerModelAsync(SaveId(), saveLifetime.Token);
            session.MarkReady(SaveId());
            latestSafeError = null;
            UpdateOperationalStatus();
            Monitor.Log($"EchoFarm memory loaded. Press {config.SummonKey} to summon Echo.", LogLevel.Info);
        }
        catch (EchoMemoryNotFoundException)
        {
            Monitor.Log($"No Echo memory yet. Press {config.RecordKey} to teach a morning routine.", LogLevel.Info);
        }
        catch (Exception error) when (error is EchoFarmException or OperationCanceledException)
        {
            latestSafeError = error is OperationCanceledException ? "request_cancelled" : "core_request_failed";
            UpdateOperationalStatus();
            Monitor.Log($"Echo core is unavailable [{latestSafeError}]. Press {config.MemoryKey} for setup status.", LogLevel.Warn);
        }
    }

    private async Task<bool> EnsureCoreStartedAsync()
    {
        if (!setupReadiness.CanAttemptStart || coreHost is null || coreLaunchOptions is null || coreEndpointInspector is null)
        {
            UpdateOperationalStatus();
            return false;
        }
        coreStartup ??= StartCoreAsync();
        bool ready = await coreStartup;
        if (!ready)
            coreStartup = null;
        return ready;
    }

    private async Task<bool> StartCoreAsync()
    {
        try
        {
            CoreEndpointStatus endpoint = await coreEndpointInspector!.InspectAsync(
                coreLaunchOptions!.BaseUri,
                modLifetime.Token
            );
            setupReadiness = SetupReadiness.Evaluate(setupInput with { EndpointStatus = endpoint });
            UpdateOperationalStatus();
            if (!setupReadiness.CanAttemptStart)
            {
                LogSetupIssue();
                return false;
            }
            CoreHostState state = await coreHost.StartAsync(coreLaunchOptions, modLifetime.Token);
            string ownership = state == CoreHostState.Owned ? "bundled process" : "existing process";
            latestSafeError = null;
            UpdateOperationalStatus();
            Monitor.Log($"EchoFarm core is ready ({ownership}).", LogLevel.Info);
            return true;
        }
        catch (OperationCanceledException) when (modLifetime.IsCancellationRequested)
        {
            latestSafeError = "shutdown_cancelled";
            UpdateOperationalStatus();
            return false;
        }
        catch (Exception error) when (error is CoreUnavailableException or TimeoutException)
        {
            setupReadiness = SetupReadiness.CoreFailure(
                setupReadiness,
                $"Run the Windows doctor or stop the process using {coreLaunchOptions!.BaseUri.Authority}, then retry."
            );
            latestSafeError = error is TimeoutException ? "core_start_timeout" : "core_start_failed";
            UpdateOperationalStatus();
            Monitor.Log($"EchoFarm core could not start [{latestSafeError}]. Press {config.MemoryKey} for the corrective action.", LogLevel.Error);
            return false;
        }
    }

    private void UpdateOperationalStatus()
    {
        string sessionState = session?.State.ToString() ?? "Unavailable";
        string? safeError = latestSafeError ?? session?.LastError?.GetType().Name;
        memoryOverlay.UpdateStatus(SetupReadiness.BuildStatusLines(new SetupStatusView(
            setupReadiness,
            coreHost?.State ?? CoreHostState.Stopped,
            sessionState,
            databasePath,
            config.EnableExperimentalHarvest,
            safeError
        )));
    }

    private void LogSetupIssue() => Monitor.Log(
        $"EchoFarm setup [{setupReadiness.Code}]: {setupReadiness.Message} {setupReadiness.Correction}",
        setupReadiness.CanAttemptStart ? LogLevel.Info : LogLevel.Error
    );

    private string ResolveCoreExecutablePath(string modDirectory)
    {
        string configured = string.IsNullOrWhiteSpace(config.CoreExecutablePath)
            ? Path.Combine("core", OperatingSystem.IsWindows() ? "echofarm-core.exe" : "echofarm-core")
            : config.CoreExecutablePath;
        return Path.GetFullPath(Path.IsPathRooted(configured) ? configured : Path.Combine(modDirectory, configured));
    }

    private string ResolveDatabasePath(string workingDirectory)
    {
        string configured = string.IsNullOrWhiteSpace(config.DatabasePath) ? "echofarm.db" : config.DatabasePath;
        return Path.GetFullPath(Path.IsPathRooted(configured) ? configured : Path.Combine(workingDirectory, configured));
    }

    private void OnProcessExit(object? sender, EventArgs e)
    {
        modLifetime.Cancel();
        if (coreHost?.State is CoreHostState.Owned or CoreHostState.External)
            coreHost.Stop();
    }

    private static string SaveId() => Game1.uniqueIDForThisGame.ToString(CultureInfo.InvariantCulture);
}
