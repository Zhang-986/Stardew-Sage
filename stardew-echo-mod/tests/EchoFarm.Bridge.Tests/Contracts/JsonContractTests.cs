using System.Text.Json;
using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Tests.Contracts;

public sealed class JsonContractTests
{
    [Fact]
    public void DemonstrationFixtureRoundTripsWithGoFieldNames()
    {
        string json = ReadFixture("morning-teaching.json");

        Demonstration demonstration = EchoJson.Deserialize<Demonstration>(json);
        string serialized = EchoJson.Serialize(demonstration);
        using JsonDocument document = JsonDocument.Parse(serialized);

        Assert.Equal("demo-farm", demonstration.SaveId);
        Assert.Equal(1, demonstration.Day);
        Assert.Equal(Weather.Sunny, demonstration.Weather);
        Assert.Equal(EventKind.WaterTarget, demonstration.Events[2].Kind);
        Assert.True(document.RootElement.TryGetProperty("saveId", out _));
        Assert.True(document.RootElement.GetProperty("events")[2].TryGetProperty("targetId", out _));
    }

    [Fact]
    public void ContinuumSnapshotRoundTripsPlayerActivity()
    {
        string json = ReadFixture("day-4-coplay-farm.json");

        WorldSnapshot snapshot = EchoJson.Deserialize<WorldSnapshot>(json);
        string serialized = EchoJson.Serialize(snapshot);
        using JsonDocument document = JsonDocument.Parse(serialized);

        Assert.Equal(3000, snapshot.Tick);
        Assert.Equal(2, snapshot.RecentPlayerActions.Count);
        Assert.Equal(EventKind.WaterTarget, snapshot.RecentPlayerActions[0].Kind);
        Assert.Equal("crop-north-1", snapshot.RecentPlayerActions[0].TargetId);
        Assert.True(document.RootElement.TryGetProperty("recentPlayerActions", out _));
    }

    [Fact]
    public void EchoMemoryViewUsesStrictSnakeCaseEnums()
    {
        const string json = """
            {
              "saveId":"farm-1","modelRevision":3,"learnedThroughDay":3,
              "stableTraits":[{
                "key":"task_order","value":"watering,harvesting","context":"sunny",
                "confidence":0.82,"observationCount":2,"contradictionCount":0,
                "firstSeenDay":1,"lastSeenDay":2,"evidenceRefs":["demo-1:water-1"]
              }],
              "recentLearningChange":{
                "modelRevision":3,"kind":"strengthened","key":"task_order",
                "value":"watering,harvesting","confidence":0.82,"summary":"strengthened"
              },
              "experiences":[{
                "id":"exp-1","saveId":"farm-1","trigger":"inventory_full","context":"sunny",
                "whenSignals":["inventory_full"],"avoidAction":"harvest_target","preferAction":"deposit_items",
                "summary":"deposit first","confidence":0.65,"effectiveConfidence":0.78,
                "successCount":3,"failureCount":1,"neutralCount":2,
                "observationCount":2,"contradictionCount":0,"firstSeenDay":1,"lastSeenDay":3,
                "evidenceRefs":["decision:day-1:1"],"source":"failure"
              }]
            }
            """;

        EchoMemoryView view = EchoJson.Deserialize<EchoMemoryView>(json);

        Assert.Equal(TraitContext.Sunny, view.StableTraits[0].Context);
        Assert.Equal(LearningChangeKind.Strengthened, view.RecentLearningChange!.Kind);
        Assert.Equal(0.78, view.Experiences[0].EffectiveConfidence);
        Assert.Equal(3, view.Experiences[0].SuccessCount);
        Assert.Equal(1, view.Experiences[0].FailureCount);
        Assert.Equal(2, view.Experiences[0].NeutralCount);
    }

    [Fact]
    public void ReflectiveDecisionAndCorrectionContractsMatchGoFields()
    {
        const string responseJson = """
            {
              "action":{"saveId":"farm-1","sessionId":"day-2","snapshotVersion":7,"kind":"deposit_items","targetId":"chest-west","reason":"learned correction"},
              "confidence":0.93,
              "alternatives":[{"saveId":"farm-1","sessionId":"day-2","snapshotVersion":7,"kind":"stop_session","reason":"safe fallback"}],
              "appliedExperiences":["exp-corrected-chest"]
            }
            """;

        ActionResponse response = EchoJson.Deserialize<ActionResponse>(responseJson);

        Assert.Equal(0.93, response.Confidence);
        Assert.Equal(ActionKind.StopSession, response.Alternatives[0].Kind);
        Assert.Equal("exp-corrected-chest", response.AppliedExperiences[0]);

        var correction = new PlayerCorrection
        {
            Id = "correction-1",
            SaveId = "farm-1",
            SessionId = "day-2",
            RejectedDecisionSnapshotVersion = 6,
            RejectedAction = new HighLevelAction
            {
                SaveId = "farm-1",
                SessionId = "day-2",
                SnapshotVersion = 6,
                Kind = ActionKind.DepositItems,
                TargetId = "chest-east",
                Reason = "initial choice"
            },
            Snapshot = new WorldSnapshot
            {
                SaveId = "farm-1",
                SessionId = "day-2",
                SnapshotVersion = 7,
                Day = 2,
                Weather = Weather.Sunny,
                MaxEnergy = 270,
                Chests = new[] { new Chest { Id = "chest-west" } }
            },
            PreferredAction = new HighLevelAction
            {
                SaveId = "farm-1",
                SessionId = "day-2",
                SnapshotVersion = 7,
                Kind = ActionKind.DepositItems,
                TargetId = "chest-west",
                Reason = "player demonstration"
            },
            ObservedAtTick = 240
        };
        using JsonDocument correctionDocument = JsonDocument.Parse(EchoJson.Serialize(correction));

        Assert.Equal(6, correctionDocument.RootElement.GetProperty("rejectedDecisionSnapshotVersion").GetInt64());
        Assert.Equal("chest-west", correctionDocument.RootElement.GetProperty("preferredAction").GetProperty("targetId").GetString());
    }

    [Theory]
    [InlineData("changed-rainy-farm.json", Weather.Rainy, 18)]
    [InlineData("empty-can-farm.json", Weather.Sunny, 0)]
    public void WorldSnapshotFixturesMatchGoContracts(string fixture, Weather weather, int water)
    {
        WorldSnapshot snapshot = EchoJson.Deserialize<WorldSnapshot>(ReadFixture(fixture));
        string serialized = EchoJson.Serialize(snapshot);
        using JsonDocument document = JsonDocument.Parse(serialized);

        Assert.Equal(weather, snapshot.Weather);
        Assert.Equal(water, snapshot.WateringCan.Water);
        Assert.True(document.RootElement.TryGetProperty("snapshotVersion", out _));
        Assert.True(document.RootElement.TryGetProperty("wateringCan", out _));
    }

    [Fact]
    public void WorldSnapshotRoundTripsExplicitActionCapabilities()
    {
        WorldSnapshot snapshot = new()
        {
            SaveId = "farm-1",
            SessionId = "echo-1",
            Day = 1,
            MaxEnergy = 270,
            Location = "Farm",
            Capabilities = new ActionCapabilities { Harvest = false }
        };

        string json = EchoJson.Serialize(snapshot);
        WorldSnapshot roundTrip = EchoJson.Deserialize<WorldSnapshot>(json);

        Assert.NotNull(roundTrip.Capabilities);
        Assert.False(roundTrip.Capabilities.Harvest);
    }

    [Fact]
    public void ModelUsageRoundTripsKnownAndUnknownTokenState()
    {
        const string json = """
            {
              "saveId":"farm-1","sessionId":"echo-4","day":4,
              "session":{"calls":3,"succeeded":2,"failed":1,"reportedTokenCalls":2,"promptTokens":80,"completionTokens":20,"totalTokens":100,"tokensKnown":false,"lastLatencyMs":250,"averageLatencyMs":200},
              "dayTotals":{"calls":5,"succeeded":4,"failed":1,"reportedTokenCalls":4,"promptTokens":160,"completionTokens":40,"totalTokens":200,"tokensKnown":false,"lastLatencyMs":250,"averageLatencyMs":180},
              "callBudget":32,"tokenBudget":100000,"budgetExhausted":false
            }
            """;

        ModelUsageSummary usage = EchoJson.Deserialize<ModelUsageSummary>(json);

        Assert.Equal(3, usage.Session.Calls);
        Assert.Equal(100, usage.Session.TotalTokens);
        Assert.False(usage.Session.TokensKnown);
        Assert.Equal(5, usage.DayTotals.Calls);
        Assert.Equal(json.Replace("\r", string.Empty).Replace("\n", string.Empty).Replace(" ", string.Empty), EchoJson.Serialize(usage));
    }

    [Fact]
    public void UnknownActionKindIsRejected()
    {
        const string json = """
            {
              "saveId":"farm-1","sessionId":"day-2","snapshotVersion":1,
              "kind":"teleport","targetId":"crop-1","reason":"invalid"
            }
            """;

        Assert.Throws<JsonException>(() => EchoJson.Deserialize<HighLevelAction>(json));
    }

    [Fact]
    public void UnknownContractPropertyIsRejected()
    {
        string json = ReadFixture("empty-can-farm.json").Replace(
            "\"obstacles\": []",
            "\"obstacles\": [], \"modelOverride\": true",
            StringComparison.Ordinal);

        JsonException error = Assert.Throws<JsonException>(() => EchoJson.Deserialize<WorldSnapshot>(json));
        Assert.Contains("modelOverride", error.Message, StringComparison.Ordinal);
    }

    private static string ReadFixture(string name) =>
        File.ReadAllText(Path.Combine(AppContext.BaseDirectory, "Fixtures", name));
}
