using System.Net;
using System.Text;
using EchoFarm.Bridge.Contracts;

namespace EchoFarm.Bridge.Transport;

public sealed class EchoFarmClient : IEchoFarmClient
{
    private readonly HttpClient httpClient;
    private readonly TimeSpan requestTimeout;

    public EchoFarmClient(HttpClient httpClient, TimeSpan requestTimeout)
    {
        this.httpClient = httpClient ?? throw new ArgumentNullException(nameof(httpClient));
        if (httpClient.BaseAddress is null || !httpClient.BaseAddress.IsLoopback)
            throw new ArgumentException("EchoFarm client requires a loopback base address.", nameof(httpClient));
        if (requestTimeout <= TimeSpan.Zero)
            throw new ArgumentOutOfRangeException(nameof(requestTimeout));
        this.requestTimeout = requestTimeout;
    }

    public async Task<LearnResponse> LearnAsync(Demonstration demonstration, CancellationToken cancellationToken)
    {
        LearnResponse response = await PostAsync<Demonstration, LearnResponse>(
            "/v1/demonstrations/learn", demonstration, cancellationToken).ConfigureAwait(false);
        if (response.PlayerModel.SaveId != demonstration.SaveId)
            throw new EchoFarmProtocolException("EchoFarm returned player memory for another save.");
        return response;
    }

    public async Task<HighLevelAction> NextActionAsync(WorldSnapshot snapshot, CancellationToken cancellationToken)
    {
        ActionResponse response = await PostAsync<WorldSnapshot, ActionResponse>(
            "/v1/echo/next-action", snapshot, cancellationToken).ConfigureAwait(false);
        ValidateCorrelation(response.Action, snapshot);
        return response.Action;
    }

    public async Task<HighLevelAction> ReportActionResultAsync(ActionResultRequest request, CancellationToken cancellationToken)
    {
        if (request.SaveId != request.Snapshot.SaveId || request.SaveId != request.Result.SaveId)
            throw new EchoFarmProtocolException("Action result save IDs do not match.");
        ActionResponse response = await PostAsync<ActionResultRequest, ActionResponse>(
            "/v1/echo/action-result", request, cancellationToken).ConfigureAwait(false);
        ValidateCorrelation(response.Action, request.Snapshot);
        return response.Action;
    }

    public async Task<PlayerModel> GetPlayerModelAsync(string saveId, CancellationToken cancellationToken)
    {
        if (string.IsNullOrWhiteSpace(saveId))
            throw new ArgumentException("Save ID is required.", nameof(saveId));

        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
        timeout.CancelAfter(requestTimeout);
        HttpResponseMessage response;
        try
        {
            response = await httpClient.GetAsync(
                $"/v1/player-model?saveId={Uri.EscapeDataString(saveId)}",
                timeout.Token
            ).ConfigureAwait(false);
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            throw new ModelUnavailableException();
        }
        catch (HttpRequestException)
        {
            throw new ModelUnavailableException();
        }

        using (response)
        {
            if (response.StatusCode == HttpStatusCode.NotFound)
                throw new EchoMemoryNotFoundException();
            EnsureAvailable(response);
            PlayerModel model = await DeserializeResponse<PlayerModel>(response, timeout.Token).ConfigureAwait(false);
            if (model.SaveId != saveId)
                throw new EchoFarmProtocolException("EchoFarm returned player memory for another save.");
            return model;
        }
    }

    public async Task<EchoMemoryView> GetMemoryAsync(string saveId, CancellationToken cancellationToken)
    {
        if (string.IsNullOrWhiteSpace(saveId))
            throw new ArgumentException("Save ID is required.", nameof(saveId));

        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
        timeout.CancelAfter(requestTimeout);
        HttpResponseMessage response;
        try
        {
            response = await httpClient.GetAsync(
                $"/v1/echo/memory?saveId={Uri.EscapeDataString(saveId)}",
                timeout.Token
            ).ConfigureAwait(false);
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            throw new ModelUnavailableException();
        }
        catch (HttpRequestException)
        {
            throw new ModelUnavailableException();
        }

        using (response)
        {
            if (response.StatusCode == HttpStatusCode.NotFound)
                throw new EchoMemoryNotFoundException();
            EnsureAvailable(response);
            EchoMemoryView view = await DeserializeResponse<EchoMemoryView>(response, timeout.Token).ConfigureAwait(false);
            if (view.SaveId != saveId)
                throw new EchoFarmProtocolException("EchoFarm returned memory for another save.");
            return view;
        }
    }

    private async Task<TResponse> PostAsync<TRequest, TResponse>(string path, TRequest request, CancellationToken cancellationToken)
    {
        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
        timeout.CancelAfter(requestTimeout);
        using var content = new StringContent(EchoJson.Serialize(request), Encoding.UTF8, "application/json");
        HttpResponseMessage response;
        try
        {
            response = await httpClient.PostAsync(path, content, timeout.Token).ConfigureAwait(false);
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            throw new ModelUnavailableException();
        }
        catch (HttpRequestException)
        {
            throw new ModelUnavailableException();
        }

        using (response)
        {
            EnsureAvailable(response);
            return await DeserializeResponse<TResponse>(response, timeout.Token).ConfigureAwait(false);
        }
    }

    private static void EnsureAvailable(HttpResponseMessage response)
    {
        if (response.StatusCode == HttpStatusCode.ServiceUnavailable)
            throw new ModelUnavailableException();
        if (!response.IsSuccessStatusCode)
            throw new EchoFarmException($"EchoFarm rejected the request with HTTP {(int)response.StatusCode}.");
    }

    private static async Task<TResponse> DeserializeResponse<TResponse>(HttpResponseMessage response, CancellationToken cancellationToken)
    {
        string json = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
        try
        {
            return EchoJson.Deserialize<TResponse>(json);
        }
        catch (Exception error) when (error is System.Text.Json.JsonException or NotSupportedException)
        {
            throw new EchoFarmProtocolException("EchoFarm returned an invalid JSON contract.", error);
        }
    }

    private static void ValidateCorrelation(HighLevelAction action, WorldSnapshot snapshot)
    {
        if (action.SaveId != snapshot.SaveId ||
            action.SessionId != snapshot.SessionId ||
            action.SnapshotVersion != snapshot.SnapshotVersion)
        {
            throw new StaleActionException();
        }
    }
}
