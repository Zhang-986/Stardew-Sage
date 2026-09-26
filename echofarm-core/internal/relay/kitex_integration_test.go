package relay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/rpcgateway"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc/echofarmgateway"
	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	kitexserver "github.com/cloudwego/kitex/server"
)

func TestKitexThriftClientCallsTLSGateway(t *testing.T) {
	certificate, parsedCertificate := newTestCertificate(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	tlsListener := tls.NewListener(listener, &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	})
	gateway, err := rpcgateway.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Errorf("core path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}), testRelayToken, 2<<20, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	server := echofarmgateway.NewServer(
		gateway,
		kitexserver.WithListener(tlsListener),
		kitexserver.WithTransServerFactory(gonet.NewTransServerFactory()),
		kitexserver.WithTransHandlerFactory(gonet.NewSvrTransHandlerFactory()),
		kitexserver.WithReadWriteTimeout(2*time.Second),
		kitexserver.WithExitWaitTime(10*time.Millisecond),
	)
	serverDone := make(chan error, 1)
	go func() { serverDone <- server.Run() }()
	t.Cleanup(func() {
		_ = server.Stop()
		select {
		case <-serverDone:
		case <-time.After(time.Second):
			t.Error("Kitex server did not stop")
		}
	})

	config := relayTestConfig()
	config.UpstreamAddress = listener.Addr().String()
	config.CertSHA256 = sha256.Sum256(parsedCertificate.Raw)
	client, err := NewKitexClient(config)
	if err != nil {
		t.Fatal(err)
	}
	request := &echofarmrpc.RelayRequest{
		RequestId: "0123456789abcdef0123456789abcdef",
		Token:     testRelayToken,
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		response, callErr := client.Call(context.Background(), OperationHealth, request)
		if callErr == nil {
			if response.StatusCode != http.StatusOK || string(response.Body) != `{"status":"ok"}` {
				t.Fatalf("response = %+v", response)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Kitex call did not succeed: %v", callErr)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func newTestCertificate(t *testing.T) (tls.Certificate, *x509.Certificate) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "EchoFarm Test"},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: privateKey, Leaf: parsed}, parsed
}
