using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;
using EchoFarm.Bridge.Transport;

namespace EchoFarm.Bridge.Runtime;

public enum EchoSessionState
{
    Idle,
    Recording,
    Learning,
    Ready,
    Acting,
    AwaitingResult
}

public sealed class EchoSession
{
    private readonly TeachingRecorder recorder;
    private readonly IEchoFarmClient client;
    private readonly IGamePort game;
    private readonly ActionSafetyGate safetyGate;
    private string? saveId;
    private string? sessionId;
    private HighLevelAction? pendingAction;
    private int actionInFlight;

    public EchoSession(TeachingRecorder recorder, IEchoFarmClient client, IGamePort game, ActionSafetyGate safetyGate)
    {
        this.recorder = recorder ?? throw new ArgumentNullException(nameof(recorder));
        this.client = client ?? throw new ArgumentNullException(nameof(client));
        this.game = game ?? throw new ArgumentNullException(nameof(game));
        this.safetyGate = safetyGate ?? throw new ArgumentNullException(nameof(safetyGate));
    }

    public EchoSessionState State { get; private set; } = EchoSessionState.Idle;
    public Exception? LastError { get; private set; }

    public EchoSessionState BeginTeaching(string saveId, long tick, int day = 0, Weather weather = Weather.Sunny)
    {
        if (State != EchoSessionState.Idle && State != EchoSessionState.Ready)
            throw new InvalidOperationException($"Cannot record while Echo is {State}.");
        this.saveId = saveId;
        recorder.Start(saveId, tick, day, weather);
        LastError = null;
        State = EchoSessionState.Recording;
        return State;
    }

    public bool Observe(ObservedGameEvent observed) =>
        State == EchoSessionState.Recording && recorder.Observe(observed);

    public async Task<EchoSessionState> CompleteTeachingAsync(long tick, CancellationToken cancellationToken)
    {
        if (State != EchoSessionState.Recording)
            throw new InvalidOperationException("No teaching session is active.");
        Demonstration demonstration = recorder.Stop(tick);
        State = EchoSessionState.Learning;
        try
        {
            await client.LearnAsync(demonstration, cancellationToken).ConfigureAwait(false);
            State = EchoSessionState.Ready;
            return State;
        }
        catch
        {
            State = EchoSessionState.Idle;
            throw;
        }
    }

    public void MarkReady(string saveId)
    {
        if (State is EchoSessionState.Recording or EchoSessionState.Learning or EchoSessionState.Acting or EchoSessionState.AwaitingResult)
            throw new InvalidOperationException($"Cannot mark ready while Echo is {State}.");
        this.saveId = string.IsNullOrWhiteSpace(saveId)
            ? throw new ArgumentException("Save ID is required.", nameof(saveId))
            : saveId;
        State = EchoSessionState.Ready;
    }

    public EchoSessionState StartEcho(string saveId, string sessionId)
    {
        if (State != EchoSessionState.Ready)
            throw new InvalidOperationException("Echo has no learned routine ready.");
        if (this.saveId != saveId)
            throw new InvalidOperationException("Echo memory belongs to another save.");
        if (string.IsNullOrWhiteSpace(sessionId))
            throw new ArgumentException("Session ID is required.", nameof(sessionId));

        this.sessionId = sessionId;
        pendingAction = null;
        LastError = null;
        game.ShowEcho();
        State = EchoSessionState.Acting;
        return State;
    }

    public async Task<bool> TickAsync(CancellationToken cancellationToken)
    {
        if (State != EchoSessionState.Acting || Interlocked.CompareExchange(ref actionInFlight, 1, 0) != 0)
            return false;

        try
        {
            WorldSnapshot snapshot = await game.CaptureSnapshotAsync(saveId!, sessionId!, cancellationToken).ConfigureAwait(false);
            HighLevelAction action = pendingAction ?? await client.NextActionAsync(snapshot, cancellationToken).ConfigureAwait(false);
            pendingAction = null;
            safetyGate.EnsureSafe(action, snapshot);
            if (action.Kind == ActionKind.StopSession)
            {
                game.HideEcho();
                State = EchoSessionState.Ready;
                return false;
            }

            State = EchoSessionState.AwaitingResult;
            ActionResult result = await game.ExecuteAsync(action, cancellationToken).ConfigureAwait(false);
            WorldSnapshot latest = await game.CaptureSnapshotAsync(saveId!, sessionId!, cancellationToken).ConfigureAwait(false);
            pendingAction = await client.ReportActionResultAsync(new ActionResultRequest
            {
                SaveId = saveId!,
                Snapshot = latest,
                Result = result
            }, cancellationToken).ConfigureAwait(false);
            State = EchoSessionState.Acting;
            return true;
        }
        catch (Exception error) when (error is EchoFarmException or UnsafeActionException or OperationCanceledException)
        {
            LastError = error;
            Abort();
            return false;
        }
        finally
        {
            Interlocked.Exchange(ref actionInFlight, 0);
        }
    }

    public void Abort()
    {
        recorder.Cancel();
        pendingAction = null;
        sessionId = null;
        game.HideEcho();
        State = EchoSessionState.Idle;
    }
}
