using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Transport;

public interface IEchoFarmClient
{
    Task<LearnResponse> LearnAsync(Demonstration demonstration, CancellationToken cancellationToken);
    Task<ActionResponse> NextActionAsync(WorldSnapshot snapshot, CancellationToken cancellationToken);
    Task<ActionResponse> ReportActionResultAsync(ActionResultRequest request, CancellationToken cancellationToken);
    Task<CorrectionResponse> CorrectAsync(PlayerCorrection correction, CancellationToken cancellationToken);
    Task<PlayerModel> GetPlayerModelAsync(string saveId, CancellationToken cancellationToken);
    Task<EchoMemoryView> GetMemoryAsync(string saveId, CancellationToken cancellationToken);
}
