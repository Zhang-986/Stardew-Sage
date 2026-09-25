package main

import (
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

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
