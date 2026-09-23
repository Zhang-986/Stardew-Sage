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
    private CancellationTokenSource saveLifetime = new();
    private ObservationProbe? pendingObservation;

    public override void Entry(IModHelper helper)
    {
        config = helper.ReadConfig<ModConfig>();
        if (!Uri.TryCreate(config.CoreUrl, UriKind.Absolute, out Uri? coreUrl) || !coreUrl.IsLoopback)
            throw new InvalidOperationException("EchoFarm CoreUrl must be an absolute loopback URL.");

        recorder = new TeachingRecorder();
        gamePort = new StardewGamePort(Monitor);
        renderer = new EchoRenderer(gamePort.Echo);
        var httpClient = new HttpClient { BaseAddress = coreUrl };
        var coreClient = new EchoFarmClient(httpClient, TimeSpan.FromSeconds(35));
        session = new EchoSession(recorder, coreClient, gamePort, new ActionSafetyGate());

        helper.Events.GameLoop.SaveLoaded += OnSaveLoaded;
        helper.Events.GameLoop.DayStarted += OnDayStarted;
        helper.Events.GameLoop.Saving += OnSaving;
        helper.Events.GameLoop.ReturnedToTitle += OnReturnedToTitle;
        helper.Events.GameLoop.UpdateTicked += OnUpdateTicked;
        helper.Events.Input.ButtonPressed += OnButtonPressed;
        helper.Events.Display.RenderedWorld += OnRenderedWorld;
    }

    private void OnSaveLoaded(object? sender, SaveLoadedEventArgs e)
    {
        ResetSaveLifetime();
        Monitor.Log("EchoFarm ready. Press F7 to teach a morning routine.", LogLevel.Info);
    }

    private void OnDayStarted(object? sender, DayStartedEventArgs e)
    {
        if (session.State is EchoSessionState.Acting or EchoSessionState.AwaitingResult)
            session.Abort();
    }

    private void OnSaving(object? sender, SavingEventArgs e) => StopForWorldChange();

    private void OnReturnedToTitle(object? sender, ReturnedToTitleEventArgs e) => StopForWorldChange();

    private async void OnButtonPressed(object? sender, ButtonPressedEventArgs e)
    {
        if (!Context.IsWorldReady)
            return;

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
                session.BeginTeaching(SaveId(), Game1.ticks);
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
            session.StartEcho(SaveId(), $"echo-{Game1.Date.TotalDays}-{Guid.NewGuid():N}");
            Monitor.Log("Echo joined the farm.", LogLevel.Info);
            return;
        }

        if (session.State == EchoSessionState.Recording && (e.Button.IsUseToolButton() || e.Button.IsActionButton()))
            pendingObservation = gamePort.BeginObservation(e.Button, Game1.ticks);
    }

    private async void OnUpdateTicked(object? sender, UpdateTickedEventArgs e)
    {
        if (!Context.IsWorldReady)
            return;

        gamePort.Pump();
        if (session.State == EchoSessionState.Recording)
        {
            if (pendingObservation is not null && e.Ticks > (ulong)pendingObservation.Tick)
            {
                session.Observe(gamePort.CompleteObservation(pendingObservation, Game1.ticks));
                pendingObservation = null;
            }
            ObservedGameEvent? movement = gamePort.ObserveMovement(Game1.ticks);
            if (movement is not null)
                session.Observe(movement);
        }
        if (session.State == EchoSessionState.Acting && e.IsMultipleOf(15))
            await session.TickAsync(saveLifetime.Token);
    }

    private void OnRenderedWorld(object? sender, RenderedWorldEventArgs e) => renderer.Draw(e.SpriteBatch);

    private void StopForWorldChange()
    {
        pendingObservation = null;
        saveLifetime.Cancel();
        session?.Abort();
    }

    private void ResetSaveLifetime()
    {
        saveLifetime.Dispose();
        saveLifetime = new CancellationTokenSource();
    }

    private static string SaveId() => Game1.uniqueIDForThisGame.ToString(CultureInfo.InvariantCulture);
}
