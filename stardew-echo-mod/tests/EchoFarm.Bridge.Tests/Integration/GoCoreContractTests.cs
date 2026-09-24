using System.Diagnostics;
using System.Net;
using System.Net.Sockets;
using EchoFarm.Bridge.Contracts;
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
            HighLevelAction rainyAction = await client.NextActionAsync(rainy, CancellationToken.None);
            Assert.Equal(ActionKind.HarvestTarget, rainyAction.Kind);
            Assert.Equal("crop-new-ripe", rainyAction.TargetId);

            WorldSnapshot emptyCan = EchoJson.Deserialize<WorldSnapshot>(
                File.ReadAllText(Path.Combine(repository, "demo", "fixtures", "empty-can-farm.json")));
            HighLevelAction refillAction = await client.NextActionAsync(emptyCan, CancellationToken.None);
            Assert.Equal(ActionKind.RefillCan, refillAction.Kind);
            Assert.Equal("pond-south", refillAction.TargetId);

            ActionResultRequest fullInventory = EchoJson.Deserialize<ActionResultRequest>(
                File.ReadAllText(Path.Combine(repository, "demo", "fixtures", "full-inventory-result.json")));
            WorldSnapshot beforeFullInventory = WithVersion(fullInventory.Snapshot, fullInventory.Result.SnapshotVersion);
            HighLevelAction attemptedHarvest = await client.NextActionAsync(beforeFullInventory, CancellationToken.None);
            Assert.Equal(fullInventory.Result.Action.Kind, attemptedHarvest.Kind);
            Assert.Equal(fullInventory.Result.Action.TargetId, attemptedHarvest.TargetId);
            HighLevelAction depositAction = await client.ReportActionResultAsync(fullInventory, CancellationToken.None);
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
