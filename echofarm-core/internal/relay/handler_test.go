package relay

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc"
)

const testRelayToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type gatewayClientFunc func(context.Context, Operation, *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error)

func (f gatewayClientFunc) Call(ctx context.Context, operation Operation, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return f(ctx, operation, request)
}

func TestHandlerMapsHTTPToTypedRPC(t *testing.T) {
	var receivedOperation Operation
	var received *echofarmrpc.RelayRequest
	client := gatewayClientFunc(func(_ context.Context, operation Operation, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
		receivedOperation, received = operation, request
		return &echofarmrpc.RelayResponse{
			RequestId: request.RequestId, StatusCode: http.StatusCreated,
			ContentType: "application/json", Body: []byte(`{"ok":true}`),
		}, nil
	})
	var logs bytes.Buffer
	handler, err := NewHandler(relayTestConfig(), client, log.New(&logs, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/demonstrations/learn?source=game", strings.NewReader(`{"private":"game-data"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "must-not-cross")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if receivedOperation != OperationLearn || received == nil || received.GetQuery() != "source=game" || string(received.Body) != `{"private":"game-data"}` {
		t.Fatalf("RPC request = operation %q request %+v", receivedOperation, received)
	}
	if received.Token != testRelayToken || received.GetContentType() != "application/json" {
		t.Fatalf("RPC auth/content type = %q/%q", received.Token, received.GetContentType())
	}
	if response.Code != http.StatusCreated || response.Body.String() != `{"ok":true}` {
		t.Fatalf("HTTP response = %d %q", response.Code, response.Body.String())
	}
	for _, forbidden := range []string{testRelayToken, "game-data", "source=game", "must-not-cross"} {
		if strings.Contains(logs.String(), forbidden) {
			t.Fatalf("log contains %q: %s", forbidden, logs.String())
		}
	}
}

func TestHandlerRejectsUnsupportedRouteAndOversizedBody(t *testing.T) {
	var calls atomic.Int32
	client := gatewayClientFunc(func(context.Context, Operation, *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
		calls.Add(1)
		return nil, errors.New("must not be called")
	})
	config := relayTestConfig()
	config.MaxRequestBytes = 8
	handler, err := NewHandler(config, client, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		request  *http.Request
		wantCode int
	}{
		{request: httptest.NewRequest(http.MethodPost, "/admin/files", nil), wantCode: http.StatusNotFound},
		{request: httptest.NewRequest(http.MethodPost, "/v1/demonstrations/learn", strings.NewReader("123456789")), wantCode: http.StatusRequestEntityTooLarge},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, test.request)
		if response.Code != test.wantCode {
			t.Fatalf("status=%d want=%d", response.Code, test.wantCode)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("RPC calls = %d", calls.Load())
	}
}

func TestHandlerRetriesOnlyIdempotentLearningTransportFailure(t *testing.T) {
	for _, test := range []struct {
		path      string
		wantCalls int32
		wantCode  int
	}{
		{path: "/v1/demonstrations/learn", wantCalls: 2, wantCode: http.StatusOK},
		{path: "/v1/echo/corrections", wantCalls: 2, wantCode: http.StatusOK},
		{path: "/v1/echo/next-action", wantCalls: 1, wantCode: http.StatusServiceUnavailable},
	} {
		t.Run(test.path, func(t *testing.T) {
			var calls atomic.Int32
			client := gatewayClientFunc(func(_ context.Context, _ Operation, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
				if calls.Add(1) == 1 {
					return nil, errors.New("connection reset")
				}
				return &echofarmrpc.RelayResponse{
					RequestId: request.RequestId, StatusCode: http.StatusOK,
					ContentType: "application/json", Body: []byte(`{"ok":true}`),
				}, nil
			})
			handler, err := NewHandler(relayTestConfig(), client, log.New(io.Discard, "", 0))
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(`{"id":"stable"}`)))
			if response.Code != test.wantCode || calls.Load() != test.wantCalls {
				t.Fatalf("status=%d calls=%d", response.Code, calls.Load())
			}
		})
	}
}

func TestHandlerRejectsMismatchedRPCResponse(t *testing.T) {
	client := gatewayClientFunc(func(context.Context, Operation, *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
		return &echofarmrpc.RelayResponse{
			RequestId: "ffffffffffffffffffffffffffffffff", StatusCode: http.StatusOK,
			ContentType: "application/json", Body: []byte(`{}`),
		}, nil
	})
	handler, err := NewHandler(relayTestConfig(), client, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestHandlerRejectsConcurrentRequestAboveLimit(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	client := gatewayClientFunc(func(_ context.Context, _ Operation, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
		close(started)
		<-release
		return &echofarmrpc.RelayResponse{
			RequestId: request.RequestId, StatusCode: http.StatusOK,
			ContentType: "application/json", Body: []byte(`{}`),
		}, nil
	})
	config := relayTestConfig()
	config.MaxInFlight = 1
	handler, err := NewHandler(config, client, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first request did not reach RPC client")
	}
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d", second.Code)
	}
	close(release)
	select {
	case <-firstDone:
	case <-time.After(time.Second):
		t.Fatal("first request did not finish")
	}
}

func relayTestConfig() Config {
	return Config{
		ListenAddress: "127.0.0.1:18471", UpstreamAddress: "192.168.1.10:18472",
		LANToken: testRelayToken, RequestTimeout: time.Second, MaxRequestBytes: 2 << 20, MaxInFlight: 2,
	}
}

func TestPrivateHostAcceptsLocalNames(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "192.168.1.2", "echo-mac.local"} {
		if !privateHost(host) {
			t.Fatalf("privateHost(%q) = false", host)
		}
	}
	if privateHost("8.8.8.8") {
		t.Fatal("public host accepted")
	}
}
