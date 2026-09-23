using System.Net;
using System.Text;
using System.Text.Json;
using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Transport;

namespace EchoFarm.Bridge.Tests.Transport;

public sealed class EchoFarmClientTests
{
    [Fact]
    public async Task LearnPostsExactGoContract()
    {
        string? path = null;
        string? body = null;
        var handler = new StubHttpHandler(async request =>
        {
            path = request.RequestUri?.AbsolutePath;
            body = await request.Content!.ReadAsStringAsync();
            return Json(HttpStatusCode.OK, EchoJson.Serialize(new LearnResponse
            {
                PlayerModel = ValidModel(),
                Skill = ValidSkill()
            }));
        });
        var client = CreateClient(handler);

        LearnResponse response = await client.LearnAsync(ValidDemonstration(), CancellationToken.None);

        Assert.Equal("/v1/demonstrations/learn", path);
        using JsonDocument document = JsonDocument.Parse(body!);
        Assert.Equal("farm-1", document.RootElement.GetProperty("saveId").GetString());
        Assert.Equal("farm-1", response.PlayerModel.SaveId);
    }

    [Fact]
    public async Task NextActionRejectsStaleCorrelation()
    {
        WorldSnapshot snapshot = ValidSnapshot();
        var stale = new ActionResponse
        {
            Action = new HighLevelAction
            {
                SaveId = snapshot.SaveId,
                SessionId = snapshot.SessionId,
                SnapshotVersion = snapshot.SnapshotVersion - 1,
                Kind = ActionKind.WaterTarget,
                TargetId = "crop-new",
                Reason = "stale"
            }
        };
        var client = CreateClient(new StubHttpHandler(_ => Task.FromResult(Json(HttpStatusCode.OK, EchoJson.Serialize(stale)))));

        await Assert.ThrowsAsync<StaleActionException>(() => client.NextActionAsync(snapshot, CancellationToken.None));
    }

    [Fact]
    public async Task ServiceUnavailableDoesNotLeakResponseBody()
    {
        const string secretBody = "provider response with api-key-secret";
        var client = CreateClient(new StubHttpHandler(_ => Task.FromResult(Json(HttpStatusCode.ServiceUnavailable, secretBody))));

        ModelUnavailableException error = await Assert.ThrowsAsync<ModelUnavailableException>(
            () => client.NextActionAsync(ValidSnapshot(), CancellationToken.None));

        Assert.DoesNotContain("api-key-secret", error.Message, StringComparison.Ordinal);
    }

    [Fact]
    public async Task MalformedActionJSONIsRejected()
    {
        var client = CreateClient(new StubHttpHandler(_ => Task.FromResult(Json(HttpStatusCode.OK, "not-json"))));

        await Assert.ThrowsAsync<EchoFarmProtocolException>(
            () => client.NextActionAsync(ValidSnapshot(), CancellationToken.None));
    }

    [Fact]
    public async Task ReportActionResultUsesLatestSnapshotCorrelation()
    {
        WorldSnapshot snapshot = ValidSnapshot();
        var action = ValidAction(snapshot);
        var request = new ActionResultRequest
        {
            SaveId = snapshot.SaveId,
            Snapshot = snapshot,
            Result = new ActionResult
            {
                SaveId = snapshot.SaveId,
                SessionId = snapshot.SessionId,
                SnapshotVersion = snapshot.SnapshotVersion - 1,
                Action = action,
                Status = ActionStatus.Failed,
                ErrorCode = "path_blocked"
            }
        };
        var next = new ActionResponse { Action = ValidAction(snapshot) };
        var client = CreateClient(new StubHttpHandler(message =>
        {
            Assert.Equal("/v1/echo/action-result", message.RequestUri?.AbsolutePath);
            return Task.FromResult(Json(HttpStatusCode.OK, EchoJson.Serialize(next)));
        }));

        HighLevelAction result = await client.ReportActionResultAsync(request, CancellationToken.None);

        Assert.Equal(snapshot.SnapshotVersion, result.SnapshotVersion);
    }

    private static EchoFarmClient CreateClient(HttpMessageHandler handler) =>
        new(new HttpClient(handler) { BaseAddress = new Uri("http://127.0.0.1:18471") }, TimeSpan.FromSeconds(2));

    private static HttpResponseMessage Json(HttpStatusCode status, string content) => new(status)
    {
        Content = new StringContent(content, Encoding.UTF8, "application/json")
    };

    private static Demonstration ValidDemonstration() => new()
    {
        Id = "demo-1", SaveId = "farm-1", SessionId = "teaching-1", StartedAt = 1, EndedAt = 2,
        Events = new[]
        {
            new DemonstrationEvent
            {
                Id = "water-1", Kind = EventKind.WaterTarget, Tick = 1, TargetId = "crop-old",
                Position = new Position(), Delta = new StateDelta(), Success = true
            }
        }
    };

    private static WorldSnapshot ValidSnapshot() => new()
    {
        SaveId = "farm-1", SessionId = "day-2", SnapshotVersion = 7, Day = 2, TimeOfDay = 620,
        Weather = Weather.Sunny, Location = "Farm", Energy = 200, MaxEnergy = 270,
        WateringCan = new ToolState { Name = "Watering Can", Water = 5, Capacity = 40 },
        Crops = new[] { new Crop { Id = "crop-new", NeedsWater = true } }
    };

    private static PlayerModel ValidModel() => new()
    {
        SaveId = "farm-1", Revision = 1, EnergyReserve = 40
    };

    private static SkillProgram ValidSkill() => new()
    {
        Name = "morning-farm-routine", Revision = 1, Goal = "care for crops", TargetSelector = "actionable_crops",
        Steps = new[] { new SkillStep { Action = ActionKind.WaterTarget, TargetSelector = "dry_crops" } },
        SuccessConditions = new[] { "all crops cared for" }, EvidenceEventIds = new[] { "water-1" }
    };

    private static HighLevelAction ValidAction(WorldSnapshot snapshot) => new()
    {
        SaveId = snapshot.SaveId,
        SessionId = snapshot.SessionId,
        SnapshotVersion = snapshot.SnapshotVersion,
        Kind = ActionKind.WaterTarget,
        TargetId = "crop-new",
        Reason = "current dry crop"
    };

    private sealed class StubHttpHandler : HttpMessageHandler
    {
        private readonly Func<HttpRequestMessage, Task<HttpResponseMessage>> responder;

        public StubHttpHandler(Func<HttpRequestMessage, Task<HttpResponseMessage>> responder)
        {
            this.responder = responder;
        }

        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken) =>
            responder(request);
    }
}
