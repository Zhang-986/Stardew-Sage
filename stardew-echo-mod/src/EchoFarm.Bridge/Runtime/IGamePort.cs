using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Runtime;

public interface IGamePort
{
    void ShowEcho();
    void HideEcho();
    Task<WorldSnapshot> CaptureSnapshotAsync(string saveId, string sessionId, CancellationToken cancellationToken);
    Task<ActionResult> ExecuteAsync(HighLevelAction action, CancellationToken cancellationToken);
}
