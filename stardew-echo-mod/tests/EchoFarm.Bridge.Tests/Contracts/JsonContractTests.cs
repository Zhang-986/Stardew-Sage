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
        Assert.Equal(EventKind.WaterTarget, demonstration.Events[2].Kind);
        Assert.True(document.RootElement.TryGetProperty("saveId", out _));
        Assert.True(document.RootElement.GetProperty("events")[2].TryGetProperty("targetId", out _));
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
