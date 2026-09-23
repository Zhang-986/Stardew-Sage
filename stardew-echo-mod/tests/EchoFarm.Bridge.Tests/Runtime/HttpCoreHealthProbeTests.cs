using System.Net;
using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class HttpCoreHealthProbeTests
{
    [Fact]
    public async Task HealthyEndpointReturnsTrue()
    {
        var handler = new HandlerStub(HttpStatusCode.OK);
        var probe = new HttpCoreHealthProbe(new HttpClient(handler), TimeSpan.FromSeconds(1));

        bool healthy = await probe.IsHealthyAsync(new Uri("http://127.0.0.1:18471"), CancellationToken.None);

        Assert.True(healthy);
        Assert.Equal("http://127.0.0.1:18471/healthz", handler.RequestUri?.AbsoluteUri);
    }

    [Fact]
    public async Task UnavailableEndpointReturnsFalse()
    {
        var probe = new HttpCoreHealthProbe(
            new HttpClient(new HandlerStub(HttpStatusCode.ServiceUnavailable)),
            TimeSpan.FromSeconds(1)
        );

        Assert.False(await probe.IsHealthyAsync(new Uri("http://127.0.0.1:18471"), CancellationToken.None));
    }

    [Fact]
    public async Task NetworkFailureReturnsFalse()
    {
        var probe = new HttpCoreHealthProbe(
            new HttpClient(new HandlerStub(new HttpRequestException("offline"))),
            TimeSpan.FromSeconds(1)
        );

        Assert.False(await probe.IsHealthyAsync(new Uri("http://127.0.0.1:18471"), CancellationToken.None));
    }

    private sealed class HandlerStub : HttpMessageHandler
    {
        private readonly HttpStatusCode? statusCode;
        private readonly Exception? error;

        public HandlerStub(HttpStatusCode statusCode)
        {
            this.statusCode = statusCode;
        }

        public HandlerStub(Exception error)
        {
            this.error = error;
        }

        public Uri? RequestUri { get; private set; }

        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            RequestUri = request.RequestUri;
            return error is null
                ? Task.FromResult(new HttpResponseMessage(statusCode!.Value))
                : Task.FromException<HttpResponseMessage>(error);
        }
    }
}
