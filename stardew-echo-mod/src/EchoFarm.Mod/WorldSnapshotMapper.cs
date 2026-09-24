using System.Globalization;
using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Runtime;
using Microsoft.Xna.Framework;
using StardewValley;
using StardewValley.Objects;
using StardewValley.TerrainFeatures;

using BridgeChest = EchoFarm.Bridge.Contracts.Chest;
using BridgeCrop = EchoFarm.Bridge.Contracts.Crop;
using BridgePosition = EchoFarm.Bridge.Contracts.Position;
using GameCrop = StardewValley.Crop;
using SObject = StardewValley.Object;

namespace EchoFarm.Mod;

internal sealed class WorldSnapshotMapper
{
    private readonly SnapshotVersionTracker snapshotVersions = new();

    public WorldSnapshot Capture(
        string saveId,
        string sessionId,
        EchoAvatarState echo,
        long tick,
        IReadOnlyList<PlayerActivity> recentPlayerActions
    )
    {
        GameLocation location = Game1.currentLocation;
        var crops = new List<BridgeCrop>();
        foreach ((Vector2 tile, TerrainFeature feature) in location.terrainFeatures.Pairs)
        {
            if (feature is not HoeDirt { crop: not null } dirt)
                continue;
            GameCrop crop = dirt.crop;
            bool mature = crop.currentPhase.Value >= crop.phaseDays.Count - 1 &&
                (!crop.fullyGrown.Value || crop.dayOfCurrentPhase.Value <= 0);
            crops.Add(new BridgeCrop
            {
                Id = TargetId(location, "crop", tile),
                Position = Position(tile),
                Kind = crop.indexOfHarvest.Value,
                GrowthStage = crop.currentPhase.Value,
                Mature = mature,
                NeedsWater = dirt.state.Value == HoeDirt.dry
            });
        }
        crops.Sort((left, right) => StringComparer.Ordinal.Compare(left.Id, right.Id));

        var chests = new List<BridgeChest>();
        foreach ((Vector2 tile, SObject item) in location.Objects.Pairs)
        {
            if (item is Chest)
            {
                chests.Add(new BridgeChest
                {
                    Id = TargetId(location, "chest", tile),
                    Label = "chest",
                    Position = Position(tile)
                });
            }
        }
        chests.Sort((left, right) => StringComparer.Ordinal.Compare(left.Id, right.Id));

        IReadOnlyList<WaterSource> waterSources = FindNearbyWater(location, echo.Tile, radius: 24)
            .OrderBy(source => source.Id, StringComparer.Ordinal)
            .ToArray();
        BridgePosition playerPosition = Position(echo.Tile);
        InventoryItem[] inventoryItems = echo.Inventory.Snapshot()
            .OrderBy(stack => stack.ItemId, StringComparer.Ordinal)
            .ThenBy(stack => stack.Quality)
            .Select(stack => new InventoryItem
            {
                ItemId = stack.ItemId,
                Name = stack.Name,
                Quantity = stack.Quantity
            })
            .ToArray();
        var inventory = new InventorySummary
        {
            FreeSlots = echo.Inventory.FreeSlots,
            Items = inventoryItems
        };
        var wateringCan = new ToolState
        {
            Name = "Watering Can",
            Level = 0,
            Water = echo.Water,
            Capacity = echo.WaterCapacity
        };
        int day = Math.Max(1, Game1.Date.TotalDays);
        int timeOfDay = Game1.timeOfDay;
        Weather weather = GetWeather(location);
        string locationName = location.NameOrUniqueName;
        int maxEnergy = Game1.player.MaxStamina;
        string fingerprint = EchoJson.Serialize(new
        {
            saveId,
            sessionId,
            tick,
            day,
            timeOfDay,
            weather,
            locationName,
            playerPosition,
            echo.Energy,
            maxEnergy,
            inventory,
            wateringCan,
            crops,
            waterSources,
            chests,
            recentPlayerActions
        });
        return new WorldSnapshot
        {
            SaveId = saveId,
            SessionId = sessionId,
            SnapshotVersion = snapshotVersions.Next(fingerprint),
            Tick = tick,
            Day = day,
            TimeOfDay = timeOfDay,
            Weather = weather,
            Location = locationName,
            PlayerPosition = playerPosition,
            Energy = echo.Energy,
            MaxEnergy = maxEnergy,
            Inventory = inventory,
            WateringCan = wateringCan,
            Crops = crops,
            WaterSources = waterSources,
            Chests = chests,
            Obstacles = Array.Empty<BridgePosition>(),
            RecentPlayerActions = Array.AsReadOnly(recentPlayerActions.ToArray())
        };
    }

    public static string TargetId(GameLocation location, string kind, Vector2 tile) =>
        string.Create(CultureInfo.InvariantCulture, $"{location.NameOrUniqueName}:{kind}:{(int)tile.X}:{(int)tile.Y}");

    private static IReadOnlyList<WaterSource> FindNearbyWater(GameLocation location, Vector2 center, int radius)
    {
        var result = new List<WaterSource>();
        int minX = Math.Max(0, (int)center.X - radius);
        int minY = Math.Max(0, (int)center.Y - radius);
        int maxX = Math.Min(location.Map.Layers[0].LayerWidth - 1, (int)center.X + radius);
        int maxY = Math.Min(location.Map.Layers[0].LayerHeight - 1, (int)center.Y + radius);
        for (int x = minX; x <= maxX; x++)
        {
            for (int y = minY; y <= maxY; y++)
            {
                if (!location.CanRefillWateringCanOnTile(x, y))
                    continue;
                var tile = new Vector2(x, y);
                result.Add(new WaterSource
                {
                    Id = TargetId(location, "water", tile),
                    Position = Position(tile)
                });
            }
        }
        return result;
    }

    internal static Weather GetWeather(GameLocation location)
    {
        if (location.IsSnowingHere())
            return Weather.Snow;
        if (Game1.isLightning && location.IsRainingHere())
            return Weather.Storm;
        return location.IsRainingHere() ? Weather.Rainy : Weather.Sunny;
    }

    private static BridgePosition Position(Vector2 tile) => new() { X = (int)tile.X, Y = (int)tile.Y };
}
