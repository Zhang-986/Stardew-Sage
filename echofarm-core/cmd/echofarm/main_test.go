package main

import (
	"testing"
)

func TestLoadConfigUsesSafeLocalDefaultsInFixtureMode(t *testing.T) {
	env := map[string]string{"ECHOFARM_MODEL_MODE": "fixture"}
	config, err := loadConfig(func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	})
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if config.Address != "127.0.0.1:18471" {
		t.Fatalf("address = %q", config.Address)
	}
	if config.DatabasePath != "echofarm.db" {
		t.Fatalf("database path = %q", config.DatabasePath)
	}
}

func TestLoadConfigRequiresRealModelSettings(t *testing.T) {
	_, err := loadConfig(func(key string) (string, bool) {
		if key == "ECHOFARM_MODEL_MODE" {
			return "openai", true
		}
		return "", false
	})
	if err == nil {
		t.Fatal("loadConfig() error = nil, want missing model configuration error")
	}
}

func TestLoadConfigRejectsNonLoopbackDefaultOverride(t *testing.T) {
	env := map[string]string{
		"ECHOFARM_MODEL_MODE": "fixture",
		"ECHOFARM_ADDRESS":    "0.0.0.0:18471",
	}
	_, err := loadConfig(func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	})
	if err == nil {
		t.Fatal("loadConfig() error = nil, want non-loopback rejection")
	}
}
