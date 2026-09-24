using System.Diagnostics;
using System.Net;
using System.Net.Sockets;
using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;
using EchoFarm.Bridge.Runtime;
using EchoFarm.Bridge.Transport;

namespace EchoFarm.Bridge.Tests.Integration;

public sealed class GoCoreContractTests
{
    [Fact]
    public async Task DotNetBridgeLearnsAndActsAgainstGoCore()
    {
        string repository = FindRepositoryRoot();
        int port = ReservePort();
        string database = Path.Combine(Path.GetTempPath(), $"echofarm-integration-{Guid.NewGuid():N}.db");
        using Process process = StartCore(repository, port, database);
        try
        {
            using var http = new HttpClient { BaseAddress = new Uri($"http://127.0.0.1:{port}") };
            await WaitUntilHealthy(http, process, TimeSpan.FromSeconds(45));
            var client = new EchoFarmClient(http, TimeSpan.FromSeconds(5));

            Demonstration teaching = EchoJson.Deserialize<Demonstration>(
                File.ReadAllText(Path.Combine(repository, "demo", "fixtures", "morning-teaching.json")));
            LearnResponse learned = await client.LearnAsync(teaching, CancellationToken.None);
            Assert.Equal("demo-farm", learned.PlayerModel.SaveId);
            Assert.Equal("morning-farm-routine", learned.Skill.Name);

            WorldSnapshot rainy = EchoJson.Deserialize<WorldSnapshot>(
                File.ReadAllText(Path.Combine(repository, "demo", "fixtures", "changed-rainy-farm.json")));
            HighLevelAction rainyAction = (await client.NextActionAsync(rainy, CancellationToken.None)).Action;
            Assert.Equal(ActionKind.HarvestTarget, rainyAction.Kind);
            Assert.Equal("crop-new-ripe", rainyAction.TargetId);

            WorldSnapshot emptyCan = EchoJson.Deserialize<WorldSnapshot>(
                File.ReadAllText(Path.Combine(repository, "demo", "fixtures", "empty-can-farm.json")));
            HighLevelAction refillAction = (await client.NextActionAsync(emptyCan, CancellationToken.None)).Action;
            Assert.Equal(ActionKind.RefillCan, refillAction.Kind);
            Assert.Equal("pond-south", refillAction.TargetId);

            ActionResultRequest fullInventory = EchoJson.Deserialize<ActionResultRequest>(
                File.ReadAllText(Path.Combine(repository, "demo", "fixtures", "full-inventory-result.json")));
            WorldSnapshot beforeFullInventory = WithVersion(fullInventory.Snapshot, fullInventory.Result.SnapshotVersion);
            HighLevelAction attemptedHarvest = (await client.NextActionAsync(beforeFullInventory, CancellationToken.None)).Action;
            Assert.Equal(fullInventory.Result.Action.Kind, attemptedHarvest.Kind);
            Assert.Equal(fullInventory.Result.Action.TargetId, attemptedHarvest.TargetId);
            HighLevelAction depositAction = (await client.ReportActionResultAsync(fullInventory, CancellationToken.None)).Action;
            Assert.Equal(ActionKind.DepositItems, depositAction.Kind);
            Assert.Equal("shipping-chest", depositAction.TargetId);
        }
        finally
        {
            if (!process.HasExited)
                process.Kill(entireProcessTree: true);
            await process.WaitForExitAsync();
            File.Delete(database);
        }
    }

    [Fact]
    public async Task SemanticClassifierFeedsStableLifestyleMemoryThroughRealGoCore()
    {
        string repository = FindRepositoryRoot();
        int port = ReservePort();
        string database = Path.Combine(Path.GetTempPath(), $"echofarm-activity-integration-{Guid.NewGuid():N}.db");
        using Process process = StartCore(repository, port, database);
        try
        {
            using var http = new HttpClient { BaseAddress = new Uri($"http://127.0.0.1:{port}") };
            await WaitUntilHealthy(http, process, TimeSpan.FromSeconds(45));
            var client = new EchoFarmClient(http, TimeSpan.FromSeconds(5));
            var recorder = new TeachingRecorder(() => Guid.NewGuid().ToString("N"));
            LearnResponse? learned = null;

            foreach (int day in new[] { 1, 2 })
            {
                recorder.Start("classified-farm", day * 1000, day, Weather.Sunny);
                foreach (ObservedGameEvent observed in ClassifiedRoutine(day * 1000))
                    Assert.True(recorder.Observe(observed));
                Demonstration demonstration = recorder.Stop(day * 1000 + 500);
                learned = await client.LearnAsync(demonstration, CancellationToken.None);
            }

            Assert.NotNull(learned);
            Assert.All(learned.Skill.Steps, step => Assert.Equal(ActionKind.StopSession, step.Action));
            EchoMemoryView memory = await client.GetMemoryAsync("classified-farm", CancellationToken.None);
            Assert.Equal(2, memory.ModelRevision);
            Assert.Contains(memory.StableTraits, trait => trait.Key == PreferenceKey.ActivityOrder);
            Assert.Contains(memory.StableTraits, trait => trait.Key == PreferenceKey.ResourcePriority && trait.Value == "Wood");
            Assert.Contains(memory.StableTraits, trait => trait.Key == PreferenceKey.MineExitPolicy);
            Assert.Contains(memory.StableTraits, trait => trait.Key == PreferenceKey.FishingContext);
            IReadOnlyList<string> f9Lines = EchoMemoryPresenter.BuildLines(memory);
            Assert.Contains(f9Lines, line => line.Contains("活动顺序", StringComparison.Ordinal));
            Assert.Contains(f9Lines, line => line.Contains("资源偏好", StringComparison.Ordinal));
            Assert.DoesNotContain(f9Lines, line => line.Contains("tree-12-8", StringComparison.Ordinal));
        }
        finally
        {
            if (!process.HasExited)
                process.Kill(entireProcessTree: true);
            await process.WaitForExitAsync();
            File.Delete(database);
        }
    }

    private static IReadOnlyList<ObservedGameEvent> ClassifiedRoutine(long offset)
    {
        var tracker = new SemanticActivityTracker();
        var result = new List<ObservedGameEvent>();

        Assert.True(tracker.TryBegin(SemanticActivityFamily.TreeChopping,
            Activity(offset + 10, "Farm", "tree-12-8", "tree", true, energy: 100)));
        result.Add(Assert.IsType<ObservedGameEvent>(tracker.Observe(
            Activity(offset + 80, "Farm", "tree-12-8", "tree", false, energy: 90,
                items: new[] { new ItemDelta { ItemId = "388", Name = "Wood", Quantity = 12 } }))));

        Assert.True(tracker.TryBegin(SemanticActivityFamily.RockBreaking,
            Activity(offset + 100, "UndergroundMine20", "rock-4-6", "rock", true, energy: 90, mineFloor: 20)));
        result.Add(Assert.IsType<ObservedGameEvent>(tracker.Observe(
            Activity(offset + 140, "UndergroundMine20", "rock-4-6", "rock", false, energy: 86, mineFloor: 20,
                items: new[] { new ItemDelta { ItemId = "390", Name = "Stone", Quantity = 2 } }))));

        result.Add(Assert.IsType<ObservedGameEvent>(tracker.RecordMineTransition(
            Activity(offset + 160, "UndergroundMine20", "mine-floor-20", "mine_floor", true, energy: 86, mineFloor: 20),
            Activity(offset + 170, "UndergroundMine21", "mine-floor-21", "mine_floor", true, energy: 86, mineFloor: 21),
            floorDelta: 1)));

        Assert.True(tracker.TryBegin(SemanticActivityFamily.Fishing,
            Activity(offset + 200, "Beach", "sea", "fish", true, energy: 86)));
        result.Add(Assert.IsType<ObservedGameEvent>(tracker.CompleteFishing(
            Activity(offset + 320, "Beach", "sea", "fish", true, energy: 78,
                items: new[] { new ItemDelta { ItemId = "128", Name = "Pufferfish", Quantity = 1 } }),
            caught: true)));

        return result;
    }

    private static ActivitySample Activity(
        long tick,
        string location,
        string targetId,
        string targetKind,
        bool targetPresent,
        int energy,
        int mineFloor = 0,
        IReadOnlyList<ItemDelta>? items = null) => new(
            tick,
            location,
            900,
            new Position { X = 12, Y = 8 },
            targetId,
            targetKind,
            targetKind == "tree" ? "Axe" : targetKind == "rock" ? "Pickaxe" : "Fishing Rod",
            new GameStateSample { Energy = energy, Health = 100, MineFloor = mineFloor },
            items ?? Array.Empty<ItemDelta>(),
            targetPresent
        );

    private static WorldSnapshot WithVersion(WorldSnapshot source, long snapshotVersion) => new()
    {
        SaveId = source.SaveId,
        SessionId = source.SessionId,
        SnapshotVersion = snapshotVersion,
        Tick = source.Tick,
        Day = source.Day,
        TimeOfDay = source.TimeOfDay,
        Weather = source.Weather,
        Location = source.Location,
        PlayerPosition = source.PlayerPosition,
        Energy = source.Energy,
        MaxEnergy = source.MaxEnergy,
        Inventory = source.Inventory,
        WateringCan = source.WateringCan,
        Crops = source.Crops,
        WaterSources = source.WaterSources,
        Chests = source.Chests,
        Obstacles = source.Obstacles,
        RecentPlayerActions = source.RecentPlayerActions
    };

    private static Process StartCore(string repository, int port, string database)
    {
        var start = new ProcessStartInfo("go", "run ./cmd/echofarm")
        {
            WorkingDirectory = Path.Combine(repository, "echofarm-core"),
            UseShellExecute = false,
            RedirectStandardOutput = true,
            RedirectStandardError = true
        };
        start.Environment["ECHOFARM_MODEL_MODE"] = "fixture";
        start.Environment["ECHOFARM_ADDRESS"] = $"127.0.0.1:{port}";
        start.Environment["ECHOFARM_DATABASE_PATH"] = database;
        return Process.Start(start) ?? throw new InvalidOperationException("Could not start Go core.");
    }

    private static async Task WaitUntilHealthy(HttpClient client, Process process, TimeSpan timeout)
    {
        using var cancellation = new CancellationTokenSource(timeout);
        while (!cancellation.IsCancellationRequested)
        {
            if (process.HasExited)
            {
                string stderr = await process.StandardError.ReadToEndAsync();
                throw new InvalidOperationException($"Go core exited before health check: {stderr}");
            }
            try
            {
                using HttpResponseMessage response = await client.GetAsync("/healthz", cancellation.Token);
                if (response.IsSuccessStatusCode)
                    return;
            }
            catch (HttpRequestException)
            {
            }
            await Task.Delay(100, cancellation.Token);
        }
        throw new TimeoutException("Go core did not become healthy.");
    }

    private static int ReservePort()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int port = ((IPEndPoint)listener.LocalEndpoint).Port;
        listener.Stop();
        return port;
    }

    private static string FindRepositoryRoot()
    {
        DirectoryInfo? current = new(AppContext.BaseDirectory);
        while (current is not null)
        {
            if (File.Exists(Path.Combine(current.FullName, "echofarm-core", "go.mod")))
                return current.FullName;
            current = current.Parent;
        }
        throw new DirectoryNotFoundException("Could not locate the EchoFarm repository root.");
    }
}
