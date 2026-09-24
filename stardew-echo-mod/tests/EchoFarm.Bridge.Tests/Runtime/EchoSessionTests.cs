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
            NextDecision = new ActionResponse { Action = ActionSafetyGateTests.Action(snapshot, ActionKind.WaterTarget, "crop-dry") },
            AfterResultDecision = new ActionResponse { Action = ActionSafetyGateTests.Action(snapshot, ActionKind.StopSession, null) }
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
        var client = new ClientStub { NextDecision = new ActionResponse { Action = ActionSafetyGateTests.Action(snapshot, ActionKind.WaterTarget, "crop-dry") } };
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

    [Fact]
    public async Task CorrectionPausesEchoAndLearnsFromOneSuccessfulPlayerAction()
    {
        WorldSnapshot snapshot = ActionSafetyGateTests.Snapshot();
        var client = new ClientStub
        {
            NextDecision = new ActionResponse { Action = ActionSafetyGateTests.Action(snapshot, ActionKind.WaterTarget, "crop-dry") },
            AfterResultDecision = new ActionResponse
            {
                Action = ActionSafetyGateTests.Action(snapshot, ActionKind.HarvestTarget, "crop-ripe"),
                Confidence = 0.72
            }
        };
        var session = ReadySession(client, new GamePortStub(snapshot));
        session.StartEcho("farm-1", "echo-day-2");
        await session.TickAsync(CancellationToken.None);

        Assert.True(session.BeginCorrection(currentTick: 100));
        Assert.Equal(EchoSessionState.Correcting, session.State);
        Assert.False(await session.ObserveCorrectionAsync(Observed(EventKind.WaterTarget, "crop-dry", 101, success: false), CancellationToken.None));
        Assert.True(await session.ObserveCorrectionAsync(Observed(EventKind.WaterTarget, "crop-dry", 102, success: true), CancellationToken.None));

        Assert.Equal(EchoSessionState.Acting, session.State);
        Assert.Equal(1, client.CorrectionCalls);
        Assert.Equal(ActionKind.HarvestTarget, client.LastCorrection!.RejectedAction.Kind);
        Assert.Equal(ActionKind.WaterTarget, client.LastCorrection.PreferredAction.Kind);
        Assert.Equal("crop-dry", client.LastCorrection.PreferredAction.TargetId);
        Assert.False(await session.ObserveCorrectionAsync(Observed(EventKind.HarvestTarget, "crop-ripe", 103, success: true), CancellationToken.None));
        Assert.Equal(1, client.CorrectionCalls);
    }

    [Fact]
    public async Task CorrectionExpiresOrCanBeCancelledWithoutExecutingEcho()
    {
        WorldSnapshot snapshot = ActionSafetyGateTests.Snapshot();
        var client = new ClientStub
        {
            NextDecision = new ActionResponse { Action = ActionSafetyGateTests.Action(snapshot, ActionKind.WaterTarget, "crop-dry") },
            AfterResultDecision = new ActionResponse { Action = ActionSafetyGateTests.Action(snapshot, ActionKind.HarvestTarget, "crop-ripe") }
        };
        var game = new GamePortStub(snapshot);
        var session = ReadySession(client, game);
        session.StartEcho("farm-1", "echo-day-2");
        await session.TickAsync(CancellationToken.None);

        Assert.True(session.BeginCorrection(currentTick: 100));
        Assert.True(session.ExpireCorrection(currentTick: 1300));
        Assert.Equal(EchoSessionState.Acting, session.State);
        Assert.Equal(1, game.ExecuteCalls);

        await session.TickAsync(CancellationToken.None);
        Assert.True(session.BeginCorrection(currentTick: 1400));
        session.CancelCorrection();
        Assert.Equal(EchoSessionState.Acting, session.State);
        Assert.False(session.HasPendingCorrection);
    }

    [Fact]
    public async Task FailedCorrectionSubmissionStaysPausedAndCanRetrySameEvidence()
    {
        WorldSnapshot snapshot = ActionSafetyGateTests.Snapshot();
        var client = new ClientStub
        {
            NextDecision = new ActionResponse { Action = ActionSafetyGateTests.Action(snapshot, ActionKind.WaterTarget, "crop-dry") },
            AfterResultDecision = new ActionResponse { Action = ActionSafetyGateTests.Action(snapshot, ActionKind.HarvestTarget, "crop-ripe") },
            CorrectionFailure = new ModelUnavailableException()
        };
        var session = ReadySession(client, new GamePortStub(snapshot));
        session.StartEcho("farm-1", "echo-day-2");
        await session.TickAsync(CancellationToken.None);
        Assert.True(session.BeginCorrection(currentTick: 100));

        Assert.False(await session.ObserveCorrectionAsync(Observed(EventKind.WaterTarget, "crop-dry", 102, success: true), CancellationToken.None));
        Assert.Equal(EchoSessionState.Correcting, session.State);
        Assert.True(session.HasPendingCorrection);
        PlayerCorrection first = client.LastCorrection!;

        client.CorrectionFailure = null;
        Assert.True(await session.RetryCorrectionAsync(CancellationToken.None));
        Assert.Equal(EchoSessionState.Acting, session.State);
        Assert.Equal(2, client.CorrectionCalls);
        Assert.Same(first, client.LastCorrection);
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
        public ActionResponse NextDecision { get; init; } = new();
        public ActionResponse AfterResultDecision { get; init; } = new();
        public Exception? Failure { get; init; }
        public Exception? CorrectionFailure { get; set; }
        public int CorrectionCalls { get; private set; }
        public PlayerCorrection? LastCorrection { get; private set; }

        public Task<LearnResponse> LearnAsync(Demonstration demonstration, CancellationToken cancellationToken) =>
            Failure is null ? Task.FromResult(LearnResponse) : Task.FromException<LearnResponse>(Failure);

        public Task<ActionResponse> NextActionAsync(WorldSnapshot snapshot, CancellationToken cancellationToken) =>
            Failure is null ? Task.FromResult(NextDecision) : Task.FromException<ActionResponse>(Failure);

        public Task<ActionResponse> ReportActionResultAsync(ActionResultRequest request, CancellationToken cancellationToken) =>
            Failure is null ? Task.FromResult(AfterResultDecision) : Task.FromException<ActionResponse>(Failure);

        public Task<CorrectionResponse> CorrectAsync(PlayerCorrection correction, CancellationToken cancellationToken)
        {
            CorrectionCalls++;
            LastCorrection = correction;
            return CorrectionFailure is null
                ? Task.FromResult(new CorrectionResponse())
                : Task.FromException<CorrectionResponse>(CorrectionFailure);
        }

        public Task<PlayerModel> GetPlayerModelAsync(string saveId, CancellationToken cancellationToken) =>
            Failure is null ? Task.FromResult(LearnResponse.PlayerModel) : Task.FromException<PlayerModel>(Failure);

        public Task<EchoMemoryView> GetMemoryAsync(string saveId, CancellationToken cancellationToken) =>
            Failure is null
                ? Task.FromResult(new EchoMemoryView { SaveId = saveId, ModelRevision = LearnResponse.PlayerModel.Revision })
                : Task.FromException<EchoMemoryView>(Failure);
    }

    private static ObservedGameEvent Observed(EventKind kind, string targetId, long tick, bool success) => new()
    {
        Kind = kind,
        TargetId = targetId,
        Tick = tick,
        Success = success,
        Before = new GameStateSample(),
        After = new GameStateSample()
    };

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
