package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

type apiStub struct {
	model  domain.PlayerModel
	skill  domain.SkillProgram
	action domain.HighLevelAction
	err    error
}

func (s *apiStub) Teach(context.Context, domain.Demonstration) (domain.PlayerModel, domain.SkillProgram, error) {
	return s.model, s.skill, s.err
}

func (s *apiStub) NextAction(context.Context, string, domain.WorldSnapshot) (domain.HighLevelAction, error) {
	return s.action, s.err
}

func (s *apiStub) HandleResult(context.Context, string, domain.WorldSnapshot, domain.ActionResult) (domain.HighLevelAction, error) {
	return s.action, s.err
}

func (s *apiStub) GetPlayerModel(context.Context, string) (domain.PlayerModel, error) {
	return s.model, s.err
}

func (s *apiStub) GetSkill(context.Context, string, string) (domain.SkillProgram, error) {
	return s.skill, s.err
}

func TestHealthEndpoint(t *testing.T) {
	handler := newTestHandler(t, &apiStub{})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("health response = %d %q", response.Code, response.Header().Get("Content-Type"))
	}
}

func TestLearnEndpointReturnsPlayerMemory(t *testing.T) {
	demo := validDemonstration()
	stub := &apiStub{model: validModel(), skill: validSkill()}
	handler := newTestHandler(t, stub)
	body, _ := json.Marshal(demo)
	request := httptest.NewRequest(http.MethodPost, "/v1/demonstrations/learn", bytes.NewReader(body))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var got LearnResponse
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.PlayerModel.SaveID != "farm-1" || got.Skill.Name != "morning-farm-routine" {
		t.Fatalf("learn response = %+v", got)
	}
}

func TestLearnEndpointRejectsUnknownFields(t *testing.T) {
	handler := newTestHandler(t, &apiStub{})
	request := httptest.NewRequest(http.MethodPost, "/v1/demonstrations/learn", bytes.NewBufferString(`{
      "id":"demo-1","saveId":"farm-1","sessionId":"teach-1","startedAt":1,"endedAt":2,
      "events":[],"secretOverride":true
    }`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestNextActionEndpointUsesSnapshotSaveID(t *testing.T) {
	snapshot := validWorldSnapshot()
	want := domain.HighLevelAction{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Kind: domain.ActionWaterTarget, TargetID: "crop-new", Reason: "learned routine",
	}
	handler := newTestHandler(t, &apiStub{action: want})
	body, _ := json.Marshal(snapshot)
	request := httptest.NewRequest(http.MethodPost, "/v1/echo/next-action", bytes.NewReader(body))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var got ActionResponse
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Action.TargetID != "crop-new" {
		t.Fatalf("action = %+v", got.Action)
	}
}

func TestActionResultEndpointRejectsSaveMismatch(t *testing.T) {
	snapshot := validWorldSnapshot()
	requestBody := ActionResultRequest{
		SaveID: "another-save", Snapshot: snapshot,
		Result: domain.ActionResult{SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, Status: domain.ActionFailed},
	}
	handler := newTestHandler(t, &apiStub{})
	body, _ := json.Marshal(requestBody)
	request := httptest.NewRequest(http.MethodPost, "/v1/echo/action-result", bytes.NewReader(body))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestModelUnavailableMapsToServiceUnavailableWithoutDetails(t *testing.T) {
	handler := newTestHandler(t, &apiStub{err: errors.Join(intelligence.ErrModelUnavailable, errors.New("api-key-secret"))})
	body, _ := json.Marshal(validWorldSnapshot())
	request := httptest.NewRequest(http.MethodPost, "/v1/echo/next-action", bytes.NewReader(body))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	if bytes.Contains(response.Body.Bytes(), []byte("api-key-secret")) {
		t.Fatalf("response leaked internal error: %s", response.Body.String())
	}
}

func TestGetPlayerModelMapsMissingMemoryToNotFound(t *testing.T) {
	handler := newTestHandler(t, &apiStub{err: memory.ErrNotFound})
	request := httptest.NewRequest(http.MethodGet, "/v1/player-model?saveId=farm-1", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func newTestHandler(t *testing.T, stub *apiStub) http.Handler {
	t.Helper()
	handler, err := NewHandler(stub, stub, stub)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return handler
}

func validDemonstration() domain.Demonstration {
	return domain.Demonstration{
		ID: "demo-1", SaveID: "farm-1", SessionID: "teach-1", StartedAt: 1, EndedAt: 2,
		Events: []domain.DemonstrationEvent{{ID: "water-1", Kind: domain.EventWater, Tick: 1, TargetID: "crop-old", Success: true}},
	}
}

func validModel() domain.PlayerModel {
	return domain.PlayerModel{SaveID: "farm-1", Revision: 1, EnergyReserve: 40}
}

func validSkill() domain.SkillProgram {
	return domain.SkillProgram{
		Name: "morning-farm-routine", Revision: 1, Goal: "care for crops", TargetSelector: "actionable_crops",
		Steps:             []domain.SkillStep{{Action: domain.ActionWaterTarget, TargetSelector: "dry_crops"}},
		SuccessConditions: []string{"all crops cared for"}, EvidenceEventIDs: []string{"water-1"},
	}
}

func validWorldSnapshot() domain.WorldSnapshot {
	return domain.WorldSnapshot{
		SaveID: "farm-1", SessionID: "day-2", SnapshotVersion: 7, Day: 2, TimeOfDay: 620,
		Weather: domain.WeatherSunny, Location: "Farm", Energy: 200, MaxEnergy: 270,
		WateringCan: domain.ToolState{Name: "Watering Can", Water: 5, Capacity: 40},
		Crops:       []domain.Crop{{ID: "crop-new", NeedsWater: true}},
	}
}
