package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestNewHTTPServerAppliesLANAuthentication(t *testing.T) {
	cfg := config{
		Address:      "0.0.0.0:18472",
		ModelTimeout: 90 * time.Second,
		AllowLAN:     true,
		LANToken:     strings.Repeat("a", 64),
	}
	server, err := newHTTPServer(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
