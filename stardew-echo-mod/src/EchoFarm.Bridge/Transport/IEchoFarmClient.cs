using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Transport;

public interface IEchoFarmClient
{
    Task<LearnResponse> LearnAsync(Demonstration demonstration, CancellationToken cancellationToken);
    Task<HighLevelAction> NextActionAsync(WorldSnapshot snapshot, CancellationToken cancellationToken);
    Task<HighLevelAction> ReportActionResultAsync(ActionResultRequest request, CancellationToken cancellationToken);
}
