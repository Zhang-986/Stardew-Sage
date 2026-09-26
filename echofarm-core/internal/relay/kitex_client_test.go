package relay

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc"
	"github.com/cloudwego/kitex/client/callopt"
)

type kitexClientStub struct {
	method   string
	request  *echofarmrpc.RelayRequest
	response *echofarmrpc.RelayResponse
}

func (s *kitexClientStub) record(method string, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	s.method, s.request = method, request
	return s.response, nil
}

func (s *kitexClientStub) Health(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("Health", request)
}
func (s *kitexClientStub) Learn(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("Learn", request)
}
func (s *kitexClientStub) NextAction(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("NextAction", request)
}
func (s *kitexClientStub) ActionResult_(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("ActionResult", request)
}
func (s *kitexClientStub) Correction(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("Correction", request)
}
func (s *kitexClientStub) PlayerModel(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("PlayerModel", request)
}
func (s *kitexClientStub) Skill(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("Skill", request)
}
func (s *kitexClientStub) Memory(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("Memory", request)
}
func (s *kitexClientStub) ModelUsage(_ context.Context, request *echofarmrpc.RelayRequest, _ ...callopt.Option) (*echofarmrpc.RelayResponse, error) {
	return s.record("ModelUsage", request)
}

func TestKitexClientDispatchesEveryOperation(t *testing.T) {
	request := &echofarmrpc.RelayRequest{RequestId: "0123456789abcdef0123456789abcdef"}
	response := &echofarmrpc.RelayResponse{RequestId: request.RequestId, StatusCode: 200}
	for _, operation := range []Operation{
		OperationHealth, OperationLearn, OperationNextAction, OperationActionResult,
		OperationCorrection, OperationPlayerModel, OperationSkill, OperationMemory, OperationModelUsage,
	} {
		t.Run(string(operation), func(t *testing.T) {
			stub := &kitexClientStub{response: response}
			client := &kitexGatewayClient{client: stub}
			got, err := client.Call(context.Background(), operation, request)
			if err != nil || got != response || stub.method != string(operation) || stub.request != request {
				t.Fatalf("Call() response=%+v method=%q request=%p err=%v", got, stub.method, stub.request, err)
			}
		})
	}
}

func TestKitexClientRejectsUnknownOperation(t *testing.T) {
	client := &kitexGatewayClient{client: &kitexClientStub{}}
	if _, err := client.Call(context.Background(), Operation("Unknown"), &echofarmrpc.RelayRequest{}); err == nil {
		t.Fatal("unknown operation was accepted")
	}
}

func TestVerifyPinnedCertificateChecksFingerprintAndValidity(t *testing.T) {
	now := time.Now()
	certificate := &x509.Certificate{Raw: []byte("certificate"), NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour)}
	pin := sha256.Sum256(certificate.Raw)
	state := tls.ConnectionState{PeerCertificates: []*x509.Certificate{certificate}}
	if err := verifyPinnedCertificate(state, pin, now); err != nil {
		t.Fatalf("valid certificate rejected: %v", err)
	}
	badPin := sha256.Sum256([]byte("other"))
	if err := verifyPinnedCertificate(state, badPin, now); err == nil {
		t.Fatal("mismatched fingerprint accepted")
	}
	if err := verifyPinnedCertificate(state, pin, now.Add(2*time.Hour)); err == nil {
		t.Fatal("expired certificate accepted")
	}
}
