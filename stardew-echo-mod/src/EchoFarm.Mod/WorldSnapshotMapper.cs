using System.Globalization;
using EchoFarm.Bridge.Contracts;
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
    private long snapshotVersion;

    public WorldSnapshot Capture(string saveId, string sessionId, EchoAvatarState echo)
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

        var waterSources = FindNearbyWater(location, echo.Tile, radius: 24);
        return new WorldSnapshot
        {
            SaveId = saveId,
            SessionId = sessionId,
            SnapshotVersion = Interlocked.Increment(ref snapshotVersion),
            Day = Math.Max(1, Game1.Date.TotalDays),
            TimeOfDay = Game1.timeOfDay,
            Weather = GetWeather(location),
            Location = location.NameOrUniqueName,
            PlayerPosition = Position(echo.Tile),
            Energy = echo.Energy,
            MaxEnergy = Game1.player.MaxStamina,
            Inventory = new InventorySummary
            {
                FreeSlots = Math.Max(0, 12 - echo.Inventory.Count),
                Items = echo.Inventory.Select(pair => new InventoryItem
                {
                    ItemId = pair.Key,
                    Name = pair.Key,
                    Quantity = pair.Value
                }).ToArray()
            },
            WateringCan = new ToolState
            {
                Name = "Watering Can",
                Level = 0,
                Water = echo.Water,
                Capacity = echo.WaterCapacity
            },
            Crops = crops,
            WaterSources = waterSources,
            Chests = chests,
            Obstacles = Array.Empty<BridgePosition>()
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

    private static Weather GetWeather(GameLocation location)
    {
        if (location.IsSnowingHere())
            return Weather.Snow;
        if (Game1.isLightning && location.IsRainingHere())
            return Weather.Storm;
        return location.IsRainingHere() ? Weather.Rainy : Weather.Sunny;
    }

    private static BridgePosition Position(Vector2 tile) => new() { X = (int)tile.X, Y = (int)tile.Y };
}
