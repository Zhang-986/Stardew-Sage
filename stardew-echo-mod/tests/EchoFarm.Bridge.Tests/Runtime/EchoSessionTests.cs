using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;
using EchoFarm.Bridge.Runtime;
using EchoFarm.Bridge.Transport;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class EchoSessionTests
{
    [Fact]
    public async Task TeachingThenEchoRunsOneActionAtATime()
    {
        var ids = new Queue<string>(new[] { "teaching-1", "event-1", "demo-1" });
        var recorder = new TeachingRecorder(() => ids.Dequeue());
        WorldSnapshot snapshot = ActionSafetyGateTests.Snapshot();
        var client = new ClientStub
        {
            LearnResponse = new LearnResponse { PlayerModel = Model(), Skill = Skill() },
            NextAction = ActionSafetyGateTests.Action(snapshot, ActionKind.WaterTarget, "crop-dry"),
            AfterResultAction = ActionSafetyGateTests.Action(snapshot, ActionKind.StopSession, null)
        };
        var game = new GamePortStub(snapshot);
        var session = new EchoSession(recorder, client, game, new ActionSafetyGate());

        Assert.Equal(EchoSessionState.Recording, session.BeginTeaching("farm-1", 1));
        session.Observe(new ObservedGameEvent
        {
            Kind = EventKind.WaterTarget,
            Tick = 2,
            TargetId = "crop-old",
            Position = new Position(),
            Before = new GameStateSample(),
            After = new GameStateSample(),
            Success = true
        });
        Assert.Equal(EchoSessionState.Ready, await session.CompleteTeachingAsync(3, CancellationToken.None));
        Assert.Equal(EchoSessionState.Acting, session.StartEcho("farm-1", "echo-day-2"));

        Assert.True(await session.TickAsync(CancellationToken.None));
        Assert.Equal(1, game.ExecuteCalls);
        Assert.Equal(EchoSessionState.Acting, session.State);
        Assert.False(await session.TickAsync(CancellationToken.None));
        Assert.Equal(EchoSessionState.Ready, session.State);
        Assert.Equal(1, game.HideCalls);
    }

    [Fact]
    public async Task ConcurrentTickDoesNotStartSecondAction()
    {
        WorldSnapshot snapshot = ActionSafetyGateTests.Snapshot();
        var client = new ClientStub { NextAction = ActionSafetyGateTests.Action(snapshot, ActionKind.WaterTarget, "crop-dry") };
        var game = new GamePortStub(snapshot) { BlockExecution = true };
        var session = ReadySession(client, game);
        session.StartEcho("farm-1", "echo-day-2");

        Task<bool> first = session.TickAsync(CancellationToken.None);
        await game.ExecutionStarted.Task.WaitAsync(TimeSpan.FromSeconds(2));
        bool second = await session.TickAsync(CancellationToken.None);
        game.ReleaseExecution.SetResult();
        await first;

        Assert.False(second);
        Assert.Equal(1, game.ExecuteCalls);
    }

    [Fact]
    public async Task ModelFailureStopsEchoWithoutThrowingIntoGameLoop()
    {
        var client = new ClientStub { Failure = new ModelUnavailableException() };
        var game = new GamePortStub(ActionSafetyGateTests.Snapshot());
        var session = ReadySession(client, game);
        session.StartEcho("farm-1", "echo-day-2");

        bool progressed = await session.TickAsync(CancellationToken.None);

        Assert.False(progressed);
        Assert.Equal(EchoSessionState.Idle, session.State);
        Assert.IsType<ModelUnavailableException>(session.LastError);
        Assert.Equal(1, game.HideCalls);
    }

    [Fact]
    public void AbortDropsRecordingAndHidesEcho()
    {
        var game = new GamePortStub(ActionSafetyGateTests.Snapshot());
        var session = ReadySession(new ClientStub(), game);
        session.StartEcho("farm-1", "echo-day-2");

        session.Abort();

        Assert.Equal(EchoSessionState.Idle, session.State);
        Assert.Equal(1, game.HideCalls);
    }

    private static EchoSession ReadySession(ClientStub client, GamePortStub game)
    {
        var session = new EchoSession(new TeachingRecorder(), client, game, new ActionSafetyGate());
        session.MarkReady("farm-1");
        return session;
    }

    private static PlayerModel Model() => new() { SaveId = "farm-1", Revision = 1, EnergyReserve = 40 };

    private static SkillProgram Skill() => new()
    {
        Name = "morning-farm-routine",
        Revision = 1,
        Goal = "care for crops",
        TargetSelector = "actionable_crops",
        Steps = new[] { new SkillStep { Action = ActionKind.WaterTarget } },
        SuccessConditions = new[] { "done" },
        EvidenceEventIds = new[] { "event-1" }
    };

    private sealed class ClientStub : IEchoFarmClient
    {
        public LearnResponse LearnResponse { get; init; } = new();
        public HighLevelAction NextAction { get; init; } = new();
        public HighLevelAction AfterResultAction { get; init; } = new();
        public Exception? Failure { get; init; }

        public Task<LearnResponse> LearnAsync(Demonstration demonstration, CancellationToken cancellationToken) =>
            Failure is null ? Task.FromResult(LearnResponse) : Task.FromException<LearnResponse>(Failure);

        public Task<HighLevelAction> NextActionAsync(WorldSnapshot snapshot, CancellationToken cancellationToken) =>
            Failure is null ? Task.FromResult(NextAction) : Task.FromException<HighLevelAction>(Failure);

        public Task<HighLevelAction> ReportActionResultAsync(ActionResultRequest request, CancellationToken cancellationToken) =>
            Failure is null ? Task.FromResult(AfterResultAction) : Task.FromException<HighLevelAction>(Failure);

        public Task<PlayerModel> GetPlayerModelAsync(string saveId, CancellationToken cancellationToken) =>
            Failure is null ? Task.FromResult(LearnResponse.PlayerModel) : Task.FromException<PlayerModel>(Failure);
    }

    private sealed class GamePortStub : IGamePort
    {
        private readonly WorldSnapshot snapshot;

        public GamePortStub(WorldSnapshot snapshot)
        {
            this.snapshot = snapshot;
        }

        public bool BlockExecution { get; init; }
        public int ExecuteCalls { get; private set; }
        public int HideCalls { get; private set; }
        public TaskCompletionSource ExecutionStarted { get; } = new(TaskCreationOptions.RunContinuationsAsynchronously);
        public TaskCompletionSource ReleaseExecution { get; } = new(TaskCreationOptions.RunContinuationsAsynchronously);

        public void ShowEcho() { }

        public void HideEcho() => HideCalls++;

        public Task<WorldSnapshot> CaptureSnapshotAsync(string saveId, string sessionId, CancellationToken cancellationToken) =>
            Task.FromResult(snapshot);

        public async Task<ActionResult> ExecuteAsync(HighLevelAction action, CancellationToken cancellationToken)
        {
            ExecuteCalls++;
            ExecutionStarted.TrySetResult();
            if (BlockExecution)
                await ReleaseExecution.Task.WaitAsync(cancellationToken);
            return new ActionResult
            {
                SaveId = action.SaveId,
                SessionId = action.SessionId,
                SnapshotVersion = action.SnapshotVersion,
                Action = action,
                Status = ActionStatus.Succeeded
            };
        }
    }
}
