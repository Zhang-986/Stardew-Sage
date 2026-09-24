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
    private EchoSession session = null!;
    private EchoRenderer renderer = null!;
    private EchoMemoryOverlay memoryOverlay = null!;
    private EchoFarmClient commandClient = null!;
    private EchoFarmClient readClient = null!;
    private CoreProcessSupervisor coreHost = null!;
    private CoreLaunchOptions coreLaunchOptions = null!;
    private Task<bool>? coreStartup;
    private readonly CancellationTokenSource modLifetime = new();
    private CancellationTokenSource saveLifetime = new();
    private ObservationProbe? pendingObservation;
    private bool memoryRefreshInFlight;

    public override void Entry(IModHelper helper)
    {
        config = helper.ReadConfig<ModConfig>();
        if (!Uri.TryCreate(config.CoreUrl, UriKind.Absolute, out Uri? coreUrl) || !coreUrl.IsLoopback)
            throw new InvalidOperationException("EchoFarm CoreUrl must be an absolute loopback URL.");

        string applicationData = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
        if (string.IsNullOrWhiteSpace(applicationData))
            applicationData = helper.DirectoryPath;
        coreLaunchOptions = CoreLaunchOptionsFactory.Create(new CoreLaunchSettings(
            helper.DirectoryPath,
            applicationData,
            coreUrl,
            config.AutoStartCore,
            config.CoreExecutablePath,
            TimeSpan.FromSeconds(config.CoreStartupTimeoutSeconds),
            config.ModelMode,
            config.ModelBaseUrl,
            config.ModelName,
            config.DatabasePath,
            OperatingSystem.IsWindows()
        ));
        coreHost = new CoreProcessSupervisor(
            new HttpCoreHealthProbe(new HttpClient(), TimeSpan.FromMilliseconds(500)),
            new SystemCoreProcessLauncher((message, isError) =>
                Monitor.Log($"[EchoFarm Core] {message}", isError ? LogLevel.Warn : LogLevel.Trace)),
            new TaskAsyncDelay()
        );

        recorder = new TeachingRecorder();
        gamePort = new StardewGamePort(Monitor);
        renderer = new EchoRenderer(gamePort.Echo);
        memoryOverlay = new EchoMemoryOverlay();
        EchoFarmClientSet clients = EchoFarmClientFactory.Create(coreUrl);
        commandClient = clients.Commands;
        readClient = clients.Reads;
        session = new EchoSession(recorder, commandClient, gamePort, new ActionSafetyGate());

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
        if (session.State is EchoSessionState.Acting or EchoSessionState.AwaitingResult or EchoSessionState.Correcting)
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
            if (memoryOverlay.Visible)
                await RefreshMemoryAsync();
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
        gamePort?.ResetPlayerActivity();
        memoryOverlay?.Hide();
        saveLifetime.Cancel();
        session?.Abort();
    }

    private async Task RefreshMemoryAsync()
    {
        if (memoryRefreshInFlight)
            return;
        memoryRefreshInFlight = true;
        try
        {
            if (!await EnsureCoreStartedAsync())
            {
                memoryOverlay.ShowError("Core unavailable");
                return;
            }
            EchoFarm.Bridge.Contracts.EchoMemoryView view = await readClient.GetMemoryAsync(SaveId(), saveLifetime.Token);
            memoryOverlay.Update(view);
        }
        catch (Exception error) when (error is EchoFarmException or OperationCanceledException)
        {
            memoryOverlay.ShowError(error is OperationCanceledException ? "Request cancelled" : "Core unavailable");
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
        if (!await EnsureCoreStartedAsync())
            return;
        try
        {
            await readClient.GetPlayerModelAsync(SaveId(), saveLifetime.Token);
            session.MarkReady(SaveId());
            Monitor.Log($"EchoFarm memory loaded. Press {config.SummonKey} to summon Echo.", LogLevel.Info);
        }
        catch (EchoMemoryNotFoundException)
        {
            Monitor.Log($"No Echo memory yet. Press {config.RecordKey} to teach a morning routine.", LogLevel.Info);
        }
        catch (Exception error) when (error is EchoFarmException or OperationCanceledException)
        {
            Monitor.Log($"Echo core is unavailable: {error.Message}", LogLevel.Warn);
        }
    }

    private async Task<bool> EnsureCoreStartedAsync()
    {
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
            CoreHostState state = await coreHost.StartAsync(coreLaunchOptions, modLifetime.Token);
            string ownership = state == CoreHostState.Owned ? "bundled process" : "existing process";
            Monitor.Log($"EchoFarm core is ready ({ownership}).", LogLevel.Info);
            return true;
        }
        catch (Exception error)
        {
            Monitor.Log($"EchoFarm core could not start: {error.Message}", LogLevel.Error);
            return false;
        }
    }

    private void OnProcessExit(object? sender, EventArgs e)
    {
        modLifetime.Cancel();
        if (coreHost?.State is CoreHostState.Owned or CoreHostState.External)
            coreHost.Stop();
    }

    private static string SaveId() => Game1.uniqueIDForThisGame.ToString(CultureInfo.InvariantCulture);
}
