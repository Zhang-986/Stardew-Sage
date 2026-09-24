package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
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

func TestBuildHandlerWiresContinuumServices(t *testing.T) {
	store, err := memory.OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if _, err := buildHandler(context.Background(), store, intelligence.NewFixtureGenerator()); err != nil {
		t.Fatalf("buildHandler() error = %v", err)
	}
}

func TestBuildHandlerWiresReflectiveCorrectionLoop(t *testing.T) {
	ctx := context.Background()
	store, err := memory.OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	demo := domain.Demonstration{
		ID: "demo-1", SaveID: "farm-1", SessionID: "teach-1", Day: 1, StartedAt: 1, EndedAt: 2,
		Events: []domain.DemonstrationEvent{{ID: "event-1", Kind: domain.EventWater, Tick: 1, TargetID: "crop-1", Success: true}},
	}
	model := domain.PlayerModel{SaveID: "farm-1", Revision: 1, LearnedThroughDay: 1, EnergyReserve: 30}
	skill := domain.SkillProgram{
		Name: "morning-farm-routine", Revision: 1, Goal: "care for crops",
		Steps:             []domain.SkillStep{{Action: domain.ActionHarvestTarget, TargetSelector: "mature_crops"}},
		SuccessConditions: []string{"crops cared for"}, EvidenceEventIDs: []string{"event-1"},
	}
	if err := store.SaveLearning(ctx, demo, model, skill); err != nil {
		t.Fatal(err)
	}
	snapshot := domain.WorldSnapshot{
		SaveID: "farm-1", SessionID: "echo-day-2", SnapshotVersion: 2, Tick: 240,
		Day: 2, TimeOfDay: 700, Weather: domain.WeatherSunny, Location: "Farm",
		Energy: 200, MaxEnergy: 270, WateringCan: domain.ToolState{Name: "Watering Can", Water: 10, Capacity: 40},
		Inventory: domain.InventorySummary{FreeSlots: 0, Items: []domain.InventoryItem{{ItemID: "parsnip", Name: "Parsnip", Quantity: 1}}},
		Chests:    []domain.Chest{{ID: "chest-east"}, {ID: "chest-west"}},
	}
	rejected := domain.HighLevelAction{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: 1,
		Kind: domain.ActionDepositItems, TargetID: "chest-east", Reason: "initial choice",
	}
	if err := store.SaveDecision(ctx, domain.DecisionRecord{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: 1, Day: 2,
		ModelRevision: 1, CandidateAction: rejected, FinalAction: rejected,
	}); err != nil {
		t.Fatal(err)
	}
	correction := domain.PlayerCorrection{
		ID: "correction-1", SaveID: snapshot.SaveID, SessionID: snapshot.SessionID,
		RejectedDecisionSnapshotVersion: 1, RejectedAction: rejected, Snapshot: snapshot,
		PreferredAction: domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: 2,
			Kind: domain.ActionDepositItems, TargetID: "chest-west", Reason: "player demonstration",
		},
		ObservedAtTick: snapshot.Tick,
	}
	handler, err := buildHandler(ctx, store, intelligence.NewFixtureGenerator())
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(correction)
	request := httptest.NewRequest(http.MethodPost, "/v1/echo/corrections", bytes.NewReader(body))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	experiences, err := store.ListPolicyExperiences(ctx, snapshot.SaveID)
	if err != nil || len(experiences) != 1 || experiences[0].PreferredTargetID != "chest-west" {
		t.Fatalf("experiences = %+v, err = %v", experiences, err)
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
