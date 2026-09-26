package relay

import (
	"strings"
	"testing"
	"time"
)

func TestLoadConfigUsesSafeDefaults(t *testing.T) {
	config, err := LoadConfig(mapLookup(validConfigValues()))
	if err != nil {
		t.Fatal(err)
	}
	if config.ListenAddress != "127.0.0.1:18471" {
		t.Fatalf("listen address = %q", config.ListenAddress)
	}
	if config.UpstreamAddress != "192.168.1.10:18472" {
		t.Fatalf("upstream address = %q", config.UpstreamAddress)
	}
	if config.RequestTimeout != 100*time.Second || config.MaxRequestBytes != 2<<20 || config.MaxInFlight != 2 {
		t.Fatalf("limits = %s/%d/%d", config.RequestTimeout, config.MaxRequestBytes, config.MaxInFlight)
	}
}

func TestLoadConfigRejectsUnsafeValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "non-loopback listen", key: "ECHOFARM_RELAY_ADDRESS", value: "0.0.0.0:18471"},
		{name: "URL instead of host port", key: "ECHOFARM_RELAY_UPSTREAM_ADDRESS", value: "https://192.168.1.10:18472"},
		{name: "public upstream", key: "ECHOFARM_RELAY_UPSTREAM_ADDRESS", value: "8.8.8.8:18472"},
		{name: "short token", key: "ECHOFARM_RELAY_TOKEN", value: strings.Repeat("a", 63)},
		{name: "non-hex token", key: "ECHOFARM_RELAY_TOKEN", value: strings.Repeat("z", 64)},
		{name: "short fingerprint", key: "ECHOFARM_RELAY_CERT_SHA256", value: strings.Repeat("b", 63)},
		{name: "zero timeout", key: "ECHOFARM_RELAY_TIMEOUT_SECONDS", value: "0"},
		{name: "zero bytes", key: "ECHOFARM_RELAY_MAX_REQUEST_BYTES", value: "0"},
		{name: "too much concurrency", key: "ECHOFARM_RELAY_MAX_IN_FLIGHT", value: "9"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := validConfigValues()
			values[test.key] = test.value
			if _, err := LoadConfig(mapLookup(values)); err == nil || !strings.Contains(err.Error(), test.key) {
				t.Fatalf("LoadConfig() error = %v", err)
			}
		})
	}
}

func TestLoadConfigRequiresUpstreamTokenAndFingerprint(t *testing.T) {
	for _, key := range []string{
		"ECHOFARM_RELAY_UPSTREAM_ADDRESS",
		"ECHOFARM_RELAY_TOKEN",
		"ECHOFARM_RELAY_CERT_SHA256",
	} {
		t.Run(key, func(t *testing.T) {
			values := validConfigValues()
			delete(values, key)
			if _, err := LoadConfig(mapLookup(values)); err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("LoadConfig() error = %v", err)
			}
		})
	}
}

func validConfigValues() map[string]string {
	return map[string]string{
		"ECHOFARM_RELAY_UPSTREAM_ADDRESS": "192.168.1.10:18472",
		"ECHOFARM_RELAY_TOKEN":            strings.Repeat("a", 64),
		"ECHOFARM_RELAY_CERT_SHA256":      strings.Repeat("b", 64),
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
