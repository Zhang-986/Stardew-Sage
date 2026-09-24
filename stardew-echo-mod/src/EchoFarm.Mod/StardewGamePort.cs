using System.Collections.Concurrent;
using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;
using EchoFarm.Bridge.Runtime;
using Microsoft.Xna.Framework;
using StardewModdingAPI;
using StardewValley;
using StardewValley.Objects;
using StardewValley.TerrainFeatures;
using BridgePosition = EchoFarm.Bridge.Contracts.Position;
using GameChest = StardewValley.Objects.Chest;
using GameCrop = StardewValley.Crop;
using TileLocation = xTile.Dimensions.Location;

namespace EchoFarm.Mod;

internal sealed class StardewGamePort : IGamePort
{
    private readonly IMonitor monitor;
    private readonly WorldSnapshotMapper snapshots;
    private readonly EchoAvatarState echo = new();
    private readonly ConcurrentQueue<Action> gameThreadWork = new();
    private readonly GridPathfinder pathfinder = new();
    private readonly PlayerActivityWindow playerActivity = new();
    private readonly SemanticActivityTracker semanticActivities = new();
    private BridgePosition? lastObservedPlayerTile;
    private ActiveExecution? activeExecution;
    private SemanticProbe? semanticProbe;

    public StardewGamePort(IMonitor monitor, bool enableExperimentalHarvest)
    {
        this.monitor = monitor;
        snapshots = new WorldSnapshotMapper(enableExperimentalHarvest);
    }

    public EchoAvatarState Echo => echo;

    public bool RecordPlayerActivity(ObservedGameEvent observed) => playerActivity.Add(observed);

    public void ResetPlayerActivity() => playerActivity.Reset();

    public void ShowEcho()
    {
        if (Context.IsWorldReady)
            echo.Tile = Game1.player.Tile;
        echo.Visible = true;
    }

    public void HideEcho() => echo.Visible = false;

    public void Pump()
    {
        if (activeExecution is not null)
        {
            AdvanceExecution();
            return;
        }
        while (gameThreadWork.TryDequeue(out Action? work))
        {
            work();
            if (activeExecution is not null)
                break;
        }
    }

    public ObservedGameEvent? ObserveMovement(long tick)
    {
        if (!Context.IsWorldReady)
            return null;
        var current = new BridgePosition { X = Game1.player.TilePoint.X, Y = Game1.player.TilePoint.Y };
        if (lastObservedPlayerTile is not null && lastObservedPlayerTile.X == current.X && lastObservedPlayerTile.Y == current.Y)
            return null;
        lastObservedPlayerTile = current;
        return new ObservedGameEvent
        {
            Kind = EventKind.Move,
            Tick = tick,
            Position = current,
            Location = Game1.currentLocation.NameOrUniqueName,
            TimeOfDay = Game1.timeOfDay,
            Before = PlayerSample(),
            After = PlayerSample(),
            Success = true
        };
    }

    public ObservationProbe? BeginObservation(SButton button, long tick, bool includeSemanticActivities = false)
    {
        if (!Context.IsWorldReady)
            return null;
        Vector2 target = button.IsUseToolButton()
            ? Game1.player.GetToolLocation() / Game1.tileSize
            : Game1.player.GetGrabTile();
        if (includeSemanticActivities && button.IsUseToolButton() && TryBeginSemanticActivity(target, tick))
            return null;
        EventKind? kind = null;
        if (button.IsUseToolButton() && Game1.player.CurrentTool is StardewValley.Tools.WateringCan)
        {
            kind = Game1.currentLocation.CanRefillWateringCanOnTile((int)target.X, (int)target.Y)
                ? EventKind.RefillCan
                : EventKind.WaterTarget;
        }
        else if (button.IsActionButton())
        {
            if (TryCrop(target, out HoeDirt? dirt) && IsMature(dirt.crop))
                kind = EventKind.HarvestTarget;
            else if (Game1.currentLocation.Objects.TryGetValue(target, out StardewValley.Object? item) && item is GameChest)
                kind = EventKind.DepositItems;
        }
        if (kind is null)
            return null;
        return new ObservationProbe(kind.Value, tick, target, PlayerSample());
    }

    public ObservedGameEvent CompleteObservation(ObservationProbe probe, long tick)
    {
        GameStateSample after = PlayerSample();
        bool success = probe.Kind switch
        {
            EventKind.WaterTarget => TryCrop(probe.TargetTile, out HoeDirt? dirt) && dirt.state.Value == HoeDirt.watered,
            EventKind.RefillCan => after.Water > probe.Before.Water,
            EventKind.HarvestTarget => !TryCrop(probe.TargetTile, out HoeDirt? dirt) || !IsMature(dirt.crop),
            EventKind.DepositItems => after.InventoryCount < probe.Before.InventoryCount,
            _ => false
        };
        return new ObservedGameEvent
        {
            Kind = probe.Kind,
            Tick = tick,
            Position = new BridgePosition { X = (int)probe.TargetTile.X, Y = (int)probe.TargetTile.Y },
            Location = Game1.currentLocation.NameOrUniqueName,
            TimeOfDay = Game1.timeOfDay,
            TargetId = WorldSnapshotMapper.TargetId(Game1.currentLocation, TargetKind(probe.Kind), probe.TargetTile),
            TargetKind = TargetKind(probe.Kind),
            Tool = Game1.player.CurrentTool?.DisplayName,
            Before = probe.Before,
            After = after,
            Success = success,
            ErrorCode = success ? null : "game_action_failed"
        };
    }

    public ObservedGameEvent? PollSemanticActivity(long tick)
    {
        if (semanticProbe is null || !Context.IsWorldReady)
            return null;
        SemanticProbe probe = semanticProbe;
        bool targetPresent = probe.Family switch
        {
            SemanticActivityFamily.TreeChopping =>
                Game1.currentLocation.terrainFeatures.TryGetValue(probe.TargetTile, out TerrainFeature? feature) &&
                feature is StardewValley.TerrainFeatures.Tree,
            SemanticActivityFamily.RockBreaking => Game1.currentLocation.Objects.ContainsKey(probe.TargetTile),
            SemanticActivityFamily.Fishing => true,
            _ => false
        };
        IReadOnlyList<ItemDelta> itemDeltas = PlayerItemDeltas(probe.InventoryBefore);
        ActivitySample sample = ActivitySampleFor(
            tick,
            probe.TargetTile,
            probe.TargetId,
            probe.TargetKind,
            probe.Tool,
            itemDeltas,
            targetPresent
        );
        ObservedGameEvent? observed;
        if (probe.Family == SemanticActivityFamily.Fishing && HasCaughtFish(probe.InventoryBefore))
            observed = semanticActivities.CompleteFishing(sample, caught: true);
        else if (probe.Family == SemanticActivityFamily.Fishing && Game1.player.CurrentTool is not StardewValley.Tools.FishingRod)
            observed = semanticActivities.CompleteFishing(sample, caught: false);
        else
            observed = semanticActivities.Observe(sample);
        if (observed is not null)
            semanticProbe = null;
        return observed;
    }

    public ObservedGameEvent? ObserveMineTransition(GameLocation oldLocation, GameLocation newLocation, long tick)
    {
        int oldFloor = MineFloor(oldLocation.NameOrUniqueName);
        int newFloor = MineFloor(newLocation.NameOrUniqueName);
        if (oldFloor <= 0 || newFloor <= 0 || oldFloor == newFloor)
            return null;
        semanticActivities.Reset();
        semanticProbe = null;
        Vector2 tile = Game1.player.Tile;
        ActivitySample before = ActivitySampleFor(
            Math.Max(0, tick - 1), tile, $"mine-floor-{oldFloor}", "mine_floor", string.Empty,
            Array.Empty<ItemDelta>(), targetPresent: true, location: oldLocation.NameOrUniqueName, mineFloor: oldFloor);
        ActivitySample after = ActivitySampleFor(
            tick, tile, $"mine-floor-{newFloor}", "mine_floor", string.Empty,
            Array.Empty<ItemDelta>(), targetPresent: true, location: newLocation.NameOrUniqueName, mineFloor: newFloor);
        return semanticActivities.RecordMineTransition(before, after, newFloor - oldFloor);
    }

    public void ResetSemanticActivities()
    {
        semanticActivities.Reset();
        semanticProbe = null;
    }

    public Task<WorldSnapshot> CaptureSnapshotAsync(string saveId, string sessionId, CancellationToken cancellationToken) =>
        OnGameThread(() => snapshots.Capture(
            saveId,
            sessionId,
            echo,
            Game1.ticks,
            playerActivity.Snapshot(Game1.ticks)
        ), cancellationToken);

    public Task<ActionResult> ExecuteAsync(HighLevelAction action, CancellationToken cancellationToken) =>
        BeginOnGameThread(action, cancellationToken);

    private Task<ActionResult> BeginOnGameThread(HighLevelAction action, CancellationToken cancellationToken)
    {
        var completion = new TaskCompletionSource<ActionResult>(TaskCreationOptions.RunContinuationsAsynchronously);
        cancellationToken.Register(() => completion.TrySetCanceled(cancellationToken));
        gameThreadWork.Enqueue(() =>
        {
            try
            {
                StartExecution(action, completion);
            }
            catch (Exception error)
            {
                monitor.Log($"Echo action setup failed: {error.Message}", LogLevel.Warn);
                completion.TrySetResult(Result(action, success: false, error: "game_error"));
            }
        });
        return completion.Task;
    }

    private void StartExecution(HighLevelAction action, TaskCompletionSource<ActionResult> completion)
    {
        if (completion.Task.IsCompleted)
            return;
        if (action.Kind is ActionKind.StopSession or ActionKind.EquipTool)
        {
            completion.SetResult(Result(action, success: true, error: null));
            return;
        }
        Vector2 target;
        if (action.Kind == ActionKind.MoveTo && action.Destination is not null)
            target = new Vector2(action.Destination.X, action.Destination.Y);
        else if (!TryResolveTarget(action.TargetId, out target))
        {
            completion.SetResult(Result(action, success: false, error: "target_changed"));
            return;
        }

        try
        {
            int width = Game1.currentLocation.Map.Layers[0].LayerWidth;
            int height = Game1.currentLocation.Map.Layers[0].LayerHeight;
            IReadOnlyList<BridgePosition> path = pathfinder.FindPathToAdjacent(
                new BridgePosition { X = (int)echo.Tile.X, Y = (int)echo.Tile.Y },
                new BridgePosition { X = (int)target.X, Y = (int)target.Y },
                new GridBounds(0, 0, width, height),
                tile => Game1.currentLocation.isTilePassable(new TileLocation(tile.X, tile.Y), Game1.viewport)
            );
            activeExecution = new ActiveExecution(action, target, new Queue<BridgePosition>(path), completion);
            if (path.Count == 0)
                AdvanceExecution();
        }
        catch (Exception error) when (error is PathNotFoundException or ArgumentOutOfRangeException)
        {
            completion.SetResult(Result(action, success: false, error: "path_blocked"));
        }
    }

    private void AdvanceExecution()
    {
        ActiveExecution execution = activeExecution!;
        if (execution.Completion.Task.IsCompleted)
        {
            activeExecution = null;
            return;
        }
        if (execution.Path.TryDequeue(out BridgePosition? next))
        {
            echo.Tile = new Vector2(next.X, next.Y);
            return;
        }

        string? error = null;
        bool success = execution.Action.Kind switch
        {
            ActionKind.MoveTo => true,
            ActionKind.WaterTarget => Water(execution.Target, out error),
            ActionKind.RefillCan => Refill(execution.Target),
            ActionKind.HarvestTarget => Harvest(execution.Target, out error),
            ActionKind.DepositItems => Deposit(execution.Target, out error),
            _ => Unsupported("unsupported_action", out error)
        };
        activeExecution = null;
        execution.Completion.TrySetResult(Result(execution.Action, success, error));
    }

    private bool Water(Vector2 tile, out string? error)
    {
        error = null;
        if (echo.Water <= 0)
            return Unsupported("out_of_water", out error);
        if (!TryCrop(tile, out HoeDirt? dirt) || dirt.state.Value != HoeDirt.dry)
            return Unsupported("target_changed", out error);
        dirt.state.Value = HoeDirt.watered;
        echo.Water--;
        echo.Energy = Math.Max(0, echo.Energy - 2);
        Game1.playSound("wateringCan");
        return true;
    }

    private bool Refill(Vector2 tile)
    {
        if (!Game1.currentLocation.CanRefillWateringCanOnTile((int)tile.X, (int)tile.Y))
            return false;
        echo.Water = echo.WaterCapacity;
        Game1.playSound("slosh");
        return true;
    }

    private bool Harvest(Vector2 tile, out string? error)
    {
        error = null;
        if (!TryCrop(tile, out HoeDirt? dirt) || dirt.crop is not GameCrop crop)
            return Unsupported("target_changed", out error);
        string itemId = crop.indexOfHarvest.Value;
        if (string.IsNullOrWhiteSpace(itemId))
            return Unsupported("unsupported_crop", out error);

        Item harvestedItem = ItemRegistry.Create(itemId);
        int regrowDays = crop.GetData()?.RegrowDays ?? -1;
        var cropState = new CropGrowthState(
            crop.currentPhase.Value,
            crop.phaseDays.Count,
            crop.fullyGrown.Value,
            crop.dayOfCurrentPhase.Value,
            crop.dead.Value,
            regrowDays
        );
        var stack = new EchoItemStack(harvestedItem.QualifiedItemId, harvestedItem.DisplayName, 1, harvestedItem.Quality);
        HarvestTransferResult transfer = HarvestTransfer.TryCollect(echo.Inventory, stack, cropState);
        if (!transfer.Success || transfer.Transition is null)
            return Unsupported(transfer.ErrorCode ?? "game_error", out error);

        CropGrowthTransition transition = transfer.Transition;
        if (transition.RemoveCrop)
        {
            dirt.destroyCrop(showAnimation: false);
        }
        else
        {
            crop.fullyGrown.Value = transition.FullyGrown;
            crop.dayOfCurrentPhase.Value = transition.DayOfCurrentPhase;
            crop.updateDrawMath(tile);
        }
        Game1.playSound("harvest");
        return true;
    }

    private bool Deposit(Vector2 tile, out string? error)
    {
        error = null;
        if (!Game1.currentLocation.Objects.TryGetValue(tile, out StardewValley.Object? item) || item is not GameChest chest)
            return Unsupported("target_changed", out error);

        DepositTransferResult transfer = DepositTransfer.MoveAll(echo.Inventory, stack =>
        {
            Item incoming = ItemRegistry.Create(stack.ItemId, stack.Quantity, stack.Quality);
            Item? remainder = chest.addItem(incoming);
            return stack.Quantity - (remainder?.Stack ?? 0);
        });
        if (transfer.MovedQuantity > 0)
            Game1.playSound("Ship");
        if (!transfer.Success)
            return Unsupported(transfer.ErrorCode ?? "game_error", out error);
        return true;
    }

    private bool TryResolveTarget(string? targetId, out Vector2 tile)
    {
        foreach ((Vector2 candidate, TerrainFeature feature) in Game1.currentLocation.terrainFeatures.Pairs)
        {
            if (feature is HoeDirt && WorldSnapshotMapper.TargetId(Game1.currentLocation, "crop", candidate) == targetId)
            {
                tile = candidate;
                return true;
            }
        }
        foreach ((Vector2 candidate, StardewValley.Object item) in Game1.currentLocation.Objects.Pairs)
        {
            if (item is GameChest && WorldSnapshotMapper.TargetId(Game1.currentLocation, "chest", candidate) == targetId)
            {
                tile = candidate;
                return true;
            }
        }
        int width = Game1.currentLocation.Map.Layers[0].LayerWidth;
        int height = Game1.currentLocation.Map.Layers[0].LayerHeight;
        for (int x = 0; x < width; x++)
        {
            for (int y = 0; y < height; y++)
            {
                var candidate = new Vector2(x, y);
                if (Game1.currentLocation.CanRefillWateringCanOnTile(x, y) &&
                    WorldSnapshotMapper.TargetId(Game1.currentLocation, "water", candidate) == targetId)
                {
                    tile = candidate;
                    return true;
                }
            }
        }
        tile = Vector2.Zero;
        return false;
    }

    private Task<T> OnGameThread<T>(Func<T> action, CancellationToken cancellationToken)
    {
        var completion = new TaskCompletionSource<T>(TaskCreationOptions.RunContinuationsAsynchronously);
        cancellationToken.Register(() => completion.TrySetCanceled(cancellationToken));
        gameThreadWork.Enqueue(() =>
        {
            if (completion.Task.IsCompleted)
                return;
            try
            {
                completion.SetResult(action());
            }
            catch (Exception error)
            {
                monitor.Log($"Echo game operation failed: {error.Message}", LogLevel.Warn);
                completion.SetException(error);
            }
        });
        return completion.Task;
    }

    private static bool TryCrop(Vector2 tile, out HoeDirt? dirt)
    {
        if (Game1.currentLocation.terrainFeatures.TryGetValue(tile, out TerrainFeature? feature) && feature is HoeDirt found)
        {
            dirt = found;
            return true;
        }
        dirt = null;
        return false;
    }

    private static bool IsMature(GameCrop? crop) => crop is not null &&
        !crop.dead.Value &&
        crop.currentPhase.Value >= crop.phaseDays.Count - 1 &&
        (!crop.fullyGrown.Value || crop.dayOfCurrentPhase.Value <= 0);

    private static GameStateSample PlayerSample()
    {
        StardewValley.Tools.WateringCan? can = Game1.player.Items.OfType<StardewValley.Tools.WateringCan>().FirstOrDefault();
        return new GameStateSample
        {
            Energy = (int)Game1.player.Stamina,
            Health = Game1.player.health,
            Water = can?.WaterLeft ?? 0,
            InventoryCount = Game1.player.Items.Count(item => item is not null),
            MineFloor = MineFloor(Game1.currentLocation.NameOrUniqueName)
        };
    }

    private bool TryBeginSemanticActivity(Vector2 target, long tick)
    {
        SemanticActivityFamily family;
        string targetKind;
        if (Game1.player.CurrentTool is StardewValley.Tools.Axe &&
            Game1.currentLocation.terrainFeatures.TryGetValue(target, out TerrainFeature? feature) &&
            feature is StardewValley.TerrainFeatures.Tree)
        {
            family = SemanticActivityFamily.TreeChopping;
            targetKind = "tree";
        }
        else if (Game1.player.CurrentTool is StardewValley.Tools.Pickaxe &&
                 Game1.currentLocation.Objects.ContainsKey(target))
        {
            family = SemanticActivityFamily.RockBreaking;
            targetKind = "rock";
        }
        else if (Game1.player.CurrentTool is StardewValley.Tools.FishingRod)
        {
            family = SemanticActivityFamily.Fishing;
            targetKind = "fish";
        }
        else
        {
            return false;
        }

        if (semanticProbe is not null)
            return semanticProbe.Family == family && semanticProbe.TargetTile == target;
        string targetId = WorldSnapshotMapper.TargetId(Game1.currentLocation, targetKind, target);
        IReadOnlyDictionary<string, PlayerItemState> inventory = PlayerInventorySnapshot();
        ActivitySample sample = ActivitySampleFor(
            tick, target, targetId, targetKind, Game1.player.CurrentTool?.DisplayName ?? string.Empty,
            Array.Empty<ItemDelta>(), targetPresent: true);
        if (!semanticActivities.TryBegin(family, sample))
            return false;
        semanticProbe = new SemanticProbe(family, target, targetId, targetKind, sample.Tool, inventory);
        return true;
    }

    private static ActivitySample ActivitySampleFor(
        long tick,
        Vector2 tile,
        string targetId,
        string targetKind,
        string tool,
        IReadOnlyList<ItemDelta> itemDeltas,
        bool targetPresent,
        string? location = null,
        int? mineFloor = null) => new(
            tick,
            location ?? Game1.currentLocation.NameOrUniqueName,
            Game1.timeOfDay,
            new BridgePosition { X = (int)tile.X, Y = (int)tile.Y },
            targetId,
            targetKind,
            tool,
            mineFloor is null ? PlayerSample() : new GameStateSample
            {
                Energy = (int)Game1.player.Stamina,
                Health = Game1.player.health,
                Water = Game1.player.Items.OfType<StardewValley.Tools.WateringCan>().FirstOrDefault()?.WaterLeft ?? 0,
                InventoryCount = Game1.player.Items.Count(item => item is not null),
                MineFloor = mineFloor.Value
            },
            itemDeltas,
            targetPresent
        );

    private static IReadOnlyDictionary<string, PlayerItemState> PlayerInventorySnapshot() =>
        Game1.player.Items
            .Where(item => item is not null)
            .GroupBy(item => item!.QualifiedItemId, StringComparer.Ordinal)
            .ToDictionary(
                group => group.Key,
                group => new PlayerItemState(group.First()!.DisplayName, group.Sum(item => item!.Stack), group.First()!.Category),
                StringComparer.Ordinal
            );

    private static IReadOnlyList<ItemDelta> PlayerItemDeltas(IReadOnlyDictionary<string, PlayerItemState> before)
    {
        IReadOnlyDictionary<string, PlayerItemState> after = PlayerInventorySnapshot();
        return before.Keys.Concat(after.Keys)
            .Distinct(StringComparer.Ordinal)
            .Select(itemId =>
            {
                before.TryGetValue(itemId, out PlayerItemState? oldValue);
                after.TryGetValue(itemId, out PlayerItemState? newValue);
                return new ItemDelta
                {
                    ItemId = itemId,
                    Name = newValue?.Name ?? oldValue?.Name,
                    Quantity = (newValue?.Quantity ?? 0) - (oldValue?.Quantity ?? 0)
                };
            })
            .Where(item => item.Quantity != 0)
            .OrderBy(item => item.ItemId, StringComparer.Ordinal)
            .Take(32)
            .ToArray();
    }

    private static bool HasCaughtFish(IReadOnlyDictionary<string, PlayerItemState> before)
    {
        IReadOnlyDictionary<string, PlayerItemState> after = PlayerInventorySnapshot();
        return after.Any(pair =>
            pair.Value.Category == SObject.FishCategory &&
            pair.Value.Quantity > (before.TryGetValue(pair.Key, out PlayerItemState? oldValue) ? oldValue.Quantity : 0));
    }

    private static int MineFloor(string locationName)
    {
        const string prefix = "UndergroundMine";
        if (!locationName.StartsWith(prefix, StringComparison.Ordinal) ||
            !int.TryParse(locationName[prefix.Length..], out int floor))
            return 0;
        return floor;
    }

    private static string TargetKind(EventKind kind) => kind switch
    {
        EventKind.RefillCan => "water",
        EventKind.DepositItems => "chest",
        _ => "crop"
    };

    private static bool Unsupported(string errorCode, out string? error)
    {
        error = errorCode;
        return false;
    }

    private static ActionResult Result(HighLevelAction action, bool success, string? error) => new()
    {
        SaveId = action.SaveId,
        SessionId = action.SessionId,
        SnapshotVersion = action.SnapshotVersion,
        Action = action,
        Status = success ? ActionStatus.Succeeded : ActionStatus.Failed,
        ErrorCode = error
    };

    private sealed record ActiveExecution(
        HighLevelAction Action,
        Vector2 Target,
        Queue<BridgePosition> Path,
        TaskCompletionSource<ActionResult> Completion
    );

    private sealed record PlayerItemState(string Name, int Quantity, int Category);

    private sealed record SemanticProbe(
        SemanticActivityFamily Family,
        Vector2 TargetTile,
        string TargetId,
        string TargetKind,
        string Tool,
        IReadOnlyDictionary<string, PlayerItemState> InventoryBefore
    );
}

internal sealed record ObservationProbe(EventKind Kind, long Tick, Vector2 TargetTile, GameStateSample Before);
