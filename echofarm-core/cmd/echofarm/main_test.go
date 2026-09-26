package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/relay"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc"
)

func TestLoadConfigUsesProductionModelBudgetDefaults(t *testing.T) {
	config, err := loadConfig(mapLookup(map[string]string{"ECHOFARM_MODEL_MODE": "fixture"}))
	if err != nil {
		t.Fatal(err)
	}
	if config.MaxModelCallsPerSession != 32 || config.MaxReportedTokensPerSession != 100000 {
		t.Fatalf("model budgets = %d/%d", config.MaxModelCallsPerSession, config.MaxReportedTokensPerSession)
	}
}

func TestLoadConfigUsesConfiguredModelTimeout(t *testing.T) {
	config, err := loadConfig(mapLookup(map[string]string{
		"ECHOFARM_MODEL_MODE":            "fixture",
		"ECHOFARM_MODEL_TIMEOUT_SECONDS": "90",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if config.ModelTimeout != 90*time.Second {
		t.Fatalf("model timeout = %s, want 1m30s", config.ModelTimeout)
	}
}

func TestLoadConfigRejectsInvalidModelBudgets(t *testing.T) {
	for _, test := range []struct {
		key   string
		value string
	}{
		{key: "ECHOFARM_MAX_MODEL_CALLS_PER_SESSION", value: "0"},
		{key: "ECHOFARM_MAX_MODEL_CALLS_PER_SESSION", value: "many"},
		{key: "ECHOFARM_MAX_REPORTED_TOKENS_PER_SESSION", value: "-1"},
	} {
		t.Run(test.key+"="+test.value, func(t *testing.T) {
			values := map[string]string{"ECHOFARM_MODEL_MODE": "fixture", test.key: test.value}
			_, err := loadConfig(mapLookup(values))
			if err == nil || !strings.Contains(err.Error(), test.key) {
				t.Fatalf("loadConfig() error = %v", err)
			}
		})
	}
}

func TestLoadConfigRejectsNonLoopbackAddressWithoutLANOptIn(t *testing.T) {
	_, err := loadConfig(mapLookup(map[string]string{
		"ECHOFARM_MODEL_MODE": "fixture",
		"ECHOFARM_ADDRESS":    "0.0.0.0:18472",
	}))
	if err == nil || !strings.Contains(err.Error(), "ECHOFARM_ALLOW_LAN") {
		t.Fatalf("loadConfig() error = %v", err)
	}
}

func TestLoadConfigRejectsIncompleteLANMode(t *testing.T) {
	complete := map[string]string{
		"ECHOFARM_MODEL_MODE":    "fixture",
		"ECHOFARM_ADDRESS":       "0.0.0.0:18472",
		"ECHOFARM_ALLOW_LAN":     "true",
		"ECHOFARM_LAN_TOKEN":     strings.Repeat("a", 64),
		"ECHOFARM_TLS_CERT_FILE": "/private/cert.pem",
		"ECHOFARM_TLS_KEY_FILE":  "/private/key.pem",
	}
	for _, key := range []string{
		"ECHOFARM_LAN_TOKEN",
		"ECHOFARM_TLS_CERT_FILE",
		"ECHOFARM_TLS_KEY_FILE",
	} {
		t.Run("missing_"+key, func(t *testing.T) {
			values := make(map[string]string, len(complete))
			for name, value := range complete {
				values[name] = value
			}
			delete(values, key)
			_, err := loadConfig(mapLookup(values))
			if err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("loadConfig() error = %v", err)
			}
		})
	}
}

func TestLoadConfigRejectsInvalidLANToken(t *testing.T) {
	for _, token := range []string{strings.Repeat("a", 63), strings.Repeat("z", 64)} {
		_, err := loadConfig(mapLookup(map[string]string{
			"ECHOFARM_MODEL_MODE":    "fixture",
			"ECHOFARM_ADDRESS":       "0.0.0.0:18472",
			"ECHOFARM_ALLOW_LAN":     "true",
			"ECHOFARM_LAN_TOKEN":     token,
			"ECHOFARM_TLS_CERT_FILE": "/private/cert.pem",
			"ECHOFARM_TLS_KEY_FILE":  "/private/key.pem",
		}))
		if err == nil || !strings.Contains(err.Error(), "ECHOFARM_LAN_TOKEN") {
			t.Fatalf("loadConfig() token length %d error = %v", len(token), err)
		}
	}
}

func TestLoadConfigAcceptsCompleteLANMode(t *testing.T) {
	config, err := loadConfig(mapLookup(map[string]string{
		"ECHOFARM_MODEL_MODE":    "fixture",
		"ECHOFARM_ADDRESS":       "0.0.0.0:18472",
		"ECHOFARM_ALLOW_LAN":     "true",
		"ECHOFARM_LAN_TOKEN":     strings.Repeat("a", 64),
		"ECHOFARM_TLS_CERT_FILE": "/private/cert.pem",
		"ECHOFARM_TLS_KEY_FILE":  "/private/key.pem",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !config.AllowLAN || config.LANToken == "" || config.TLSCertFile == "" || config.TLSKeyFile == "" {
		t.Fatalf("LAN config = %+v", config)
	}
}

func TestNewHTTPServerRejectsLANConfiguration(t *testing.T) {
	cfg := config{
		Address:      "0.0.0.0:18472",
		ModelTimeout: 90 * time.Second,
		AllowLAN:     true,
		LANToken:     strings.Repeat("a", 64),
	}
	if _, err := newHTTPServer(cfg, http.NotFoundHandler()); err == nil {
		t.Fatal("LAN configuration was accepted by the HTTP server")
	}
}

func TestRunRPCServerSurvivesAbortedTLSHandshakes(t *testing.T) {
	address := reserveTCPAddress(t)
	certificatePath, keyPath, certificateFingerprint := writeTestTLSIdentity(t)
	token := strings.Repeat("a", 64)
	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- runRPCServer(ctx, config{
			Address:      address,
			ModelMode:    "fixture",
			ModelTimeout: 500 * time.Millisecond,
			LANToken:     token,
			TLSCertFile:  certificatePath,
			TLSKeyFile:   keyPath,
		}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}), log.New(io.Discard, "", 0))
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-serverDone:
		case <-time.After(7 * time.Second):
			t.Error("RPC server did not stop")
		}
	})

	for attempt := 0; attempt < 40; attempt++ {
		connection, err := net.DialTimeout("tcp", address, 50*time.Millisecond)
		if err == nil {
			_ = connection.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	for attempt := 0; attempt < 12; attempt++ {
		connection, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
		if err != nil {
			t.Fatalf("open aborted TLS connection %d: %v", attempt, err)
		}
		_ = connection.Close()
	}

	client, err := relay.NewKitexClient(relay.Config{
		UpstreamAddress: address,
		LANToken:        token,
		CertSHA256:      certificateFingerprint,
		RequestTimeout:  500 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Call(context.Background(), relay.OperationHealth, &echofarmrpc.RelayRequest{
		RequestId: "0123456789abcdef0123456789abcdef",
		Token:     token,
	})
	if err != nil {
		t.Fatalf("health after aborted TLS handshakes: %v", err)
	}
	if response.StatusCode != http.StatusOK || string(response.Body) != `{"status":"ok"}` {
		t.Fatalf("health response = %+v", response)
	}
}

func reserveTCPAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

func writeTestTLSIdentity(t *testing.T) (string, string, [sha256.Size]byte) {
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
	temporaryDirectory := t.TempDir()
	certificatePath := filepath.Join(temporaryDirectory, "server.crt")
	keyPath := filepath.Join(temporaryDirectory, "server.key")
	if err := os.WriteFile(certificatePath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}), 0o600); err != nil {
		t.Fatal(err)
	}
	return certificatePath, keyPath, sha256.Sum256(der)
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
