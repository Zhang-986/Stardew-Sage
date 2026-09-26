package main

import (
	"io"
	"log"
	"strings"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/relay"
)

func TestNewRelayServerUsesLoopbackConfiguration(t *testing.T) {
	config, err := relay.LoadConfig(mapLookup(map[string]string{
		"ECHOFARM_RELAY_UPSTREAM_ADDRESS": "192.168.1.10:18472",
		"ECHOFARM_RELAY_TOKEN":            strings.Repeat("a", 64),
		"ECHOFARM_RELAY_CERT_SHA256":      strings.Repeat("b", 64),
	}))
	if err != nil {
		t.Fatal(err)
	}
	server, err := newRelayServer(config, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	if server.Addr != "127.0.0.1:18471" {
		t.Fatalf("server address = %q", server.Addr)
	}
	if server.Handler == nil {
		t.Fatal("server handler is nil")
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
