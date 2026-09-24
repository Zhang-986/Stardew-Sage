using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Recording;

namespace EchoFarm.Bridge.Runtime;

public sealed class PlayerActivityWindow
{
    public const long DefaultWindowTicks = 20 * 60;

    private readonly long windowTicks;
    private readonly List<PlayerActivity> activities = new();

    public PlayerActivityWindow(long windowTicks = DefaultWindowTicks)
    {
        if (windowTicks <= 0)
            throw new ArgumentOutOfRangeException(nameof(windowTicks));
        this.windowTicks = windowTicks;
    }

    public bool Add(ObservedGameEvent observed)
    {
        if (!observed.Success || string.IsNullOrWhiteSpace(observed.TargetId) || !IsSemanticActivity(observed.Kind))
            return false;

        activities.RemoveAll(activity => activity.Kind == observed.Kind && activity.TargetId == observed.TargetId);
        activities.Add(new PlayerActivity
        {
            Kind = observed.Kind,
            TargetId = observed.TargetId,
            Tick = observed.Tick,
            Success = true
        });
        return true;
    }

    public IReadOnlyList<PlayerActivity> Snapshot(long currentTick)
    {
        activities.RemoveAll(activity => activity.Tick > currentTick || currentTick - activity.Tick > windowTicks);
        return Array.AsReadOnly(activities.OrderBy(activity => activity.Tick).ToArray());
    }

    public void Reset() => activities.Clear();

    private static bool IsSemanticActivity(EventKind kind) => kind is
        EventKind.WaterTarget or EventKind.RefillCan or EventKind.HarvestTarget or EventKind.DepositItems;
}
