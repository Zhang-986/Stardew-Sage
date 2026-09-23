package policy

import (
	"context"
	"strings"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

type actorStub struct {
	nextAction   domain.HighLevelAction
	replanAction domain.HighLevelAction
	chooseCalls  int
	replanCalls  int
	lastAction   intelligence.ActionInput
	lastReplan   intelligence.ReplanInput
}

func (s *actorStub) ChooseAction(_ context.Context, input intelligence.ActionInput) (domain.HighLevelAction, error) {
	s.chooseCalls++
	s.lastAction = input
	return s.nextAction, nil
}

func (s *actorStub) Replan(_ context.Context, input intelligence.ReplanInput) (domain.HighLevelAction, error) {
	s.replanCalls++
	s.lastReplan = input
	return s.replanAction, nil
}

type policyStoreStub struct {
	model domain.PlayerModel
	skill domain.SkillProgram
}

func (s *policyStoreStub) SaveLearning(context.Context, domain.Demonstration, domain.PlayerModel, domain.SkillProgram) error {
	return nil
}

func (s *policyStoreStub) GetDemonstration(context.Context, string, string) (domain.Demonstration, error) {
	return domain.Demonstration{}, memory.ErrNotFound
}

func (s *policyStoreStub) GetPlayerModel(_ context.Context, _ string) (domain.PlayerModel, error) {
	return s.model, nil
}

func (s *policyStoreStub) GetSkill(_ context.Context, _, _ string) (domain.SkillProgram, error) {
	return s.skill, nil
}

func TestNextActionUsesAIToHarvestInsteadOfWateringOnRainyDay(t *testing.T) {
	snapshot := validSnapshot()
	snapshot.Weather = domain.WeatherRainy
	actor := &actorStub{nextAction: actionFor(snapshot, domain.ActionHarvestTarget, "crop-mature")}
	service := newPolicyService(t, actor)

	action, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatalf("NextAction() error = %v", err)
	}
	if action.Kind != domain.ActionHarvestTarget || actor.chooseCalls != 1 {
		t.Fatalf("action = %+v, choose calls = %d", action, actor.chooseCalls)
	}
}

func TestNextActionCanTargetCropNotPresentInTeachingTrace(t *testing.T) {
	snapshot := validSnapshot()
	actor := &actorStub{nextAction: actionFor(snapshot, domain.ActionWaterTarget, "crop-new")}
	service := newPolicyService(t, actor)

	action, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatalf("NextAction() error = %v", err)
	}
	if action.TargetID != "crop-new" {
		t.Fatalf("target = %q, want crop-new", action.TargetID)
	}
}

func TestNextActionAcceptsRefillWhenCanIsEmpty(t *testing.T) {
	snapshot := validSnapshot()
	snapshot.WateringCan.Water = 0
	actor := &actorStub{nextAction: actionFor(snapshot, domain.ActionRefillCan, "pond-1")}
	service := newPolicyService(t, actor)

	action, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatalf("NextAction() error = %v", err)
	}
	if action.Kind != domain.ActionRefillCan {
		t.Fatalf("action = %+v, want refill_can", action)
	}
}

func TestHandleResultReplansRecoverableFailure(t *testing.T) {
	snapshot := validSnapshot()
	failed := actionFor(snapshot, domain.ActionWaterTarget, "crop-new")
	replanned := actionFor(snapshot, domain.ActionMoveTo, "crop-new")
	actor := &actorStub{replanAction: replanned}
	service := newPolicyService(t, actor)

	action, err := service.HandleResult(context.Background(), snapshot.SaveID, snapshot, domain.ActionResult{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Action: failed, Status: domain.ActionFailed, ErrorCode: "path_blocked",
	})
	if err != nil {
		t.Fatalf("HandleResult() error = %v", err)
	}
	if action.Kind != domain.ActionMoveTo || actor.replanCalls != 1 {
		t.Fatalf("action = %+v, replan calls = %d", action, actor.replanCalls)
	}
}

func TestNextActionStopsBeforeViolatingEnergyReserve(t *testing.T) {
	snapshot := validSnapshot()
	snapshot.Energy = 40
	actor := &actorStub{nextAction: actionFor(snapshot, domain.ActionWaterTarget, "crop-new")}
	service := newPolicyService(t, actor)

	action, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatalf("NextAction() error = %v", err)
	}
	if action.Kind != domain.ActionStopSession || actor.chooseCalls != 0 {
		t.Fatalf("action = %+v, choose calls = %d", action, actor.chooseCalls)
	}
}

func TestNextActionRejectsUnsafeOrInventedTargets(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*domain.WorldSnapshot)
		action func(domain.WorldSnapshot) domain.HighLevelAction
		want   string
	}{
		{
			name: "invented action",
			action: func(snapshot domain.WorldSnapshot) domain.HighLevelAction {
				return actionFor(snapshot, "teleport", "crop-new")
			},
			want: "unsupported action",
		},
		{
			name: "stale target",
			action: func(snapshot domain.WorldSnapshot) domain.HighLevelAction {
				return actionFor(snapshot, domain.ActionWaterTarget, "crop-yesterday")
			},
			want: "not present",
		},
		{
			name:   "watering in rain",
			mutate: func(snapshot *domain.WorldSnapshot) { snapshot.Weather = domain.WeatherRainy },
			action: func(snapshot domain.WorldSnapshot) domain.HighLevelAction {
				return actionFor(snapshot, domain.ActionWaterTarget, "crop-new")
			},
			want: "rain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := validSnapshot()
			if tt.mutate != nil {
				tt.mutate(&snapshot)
			}
			actor := &actorStub{nextAction: tt.action(snapshot)}
			service := newPolicyService(t, actor)
			_, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("NextAction() error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestNextActionRejectsMismatchedSave(t *testing.T) {
	snapshot := validSnapshot()
	service := newPolicyService(t, &actorStub{})
	if _, err := service.NextAction(context.Background(), "another-save", snapshot); err == nil {
		t.Fatal("NextAction() error = nil, want save mismatch")
	}
}

func TestNewServiceRequiresDependencies(t *testing.T) {
	if _, err := NewService(nil, &actorStub{}); err == nil {
		t.Fatal("NewService(nil store) error = nil")
	}
	if _, err := NewService(&policyStoreStub{}, nil); err == nil {
		t.Fatal("NewService(nil actor) error = nil")
	}
}

func newPolicyService(t *testing.T, actor *actorStub) *Service {
	t.Helper()
	store := &policyStoreStub{
		model: domain.PlayerModel{SaveID: "farm-1", Revision: 1, EnergyReserve: 40},
		skill: domain.SkillProgram{
			Name: "morning-farm-routine", Revision: 1, Goal: "care for crops", TargetSelector: "actionable_crops",
			Steps:             []domain.SkillStep{{Action: domain.ActionWaterTarget, TargetSelector: "dry_crops"}},
			SuccessConditions: []string{"all crops cared for"}, EvidenceEventIDs: []string{"water-1"},
		},
	}
	service, err := NewService(store, actor)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func validSnapshot() domain.WorldSnapshot {
	return domain.WorldSnapshot{
		SaveID: "farm-1", SessionID: "day-2", SnapshotVersion: 7, Day: 2, TimeOfDay: 620,
		Weather: domain.WeatherSunny, Location: "Farm", Energy: 200, MaxEnergy: 270,
		WateringCan: domain.ToolState{Name: "Watering Can", Water: 5, Capacity: 40},
		Crops: []domain.Crop{
			{ID: "crop-new", Position: domain.Position{X: 4, Y: 2}, NeedsWater: true},
			{ID: "crop-mature", Position: domain.Position{X: 5, Y: 2}, Mature: true},
		},
		WaterSources: []domain.WaterSource{{ID: "pond-1", Position: domain.Position{X: 8, Y: 8}}},
		Chests:       []domain.Chest{{ID: "chest-1", Position: domain.Position{X: 2, Y: 2}}},
	}
}

func actionFor(snapshot domain.WorldSnapshot, kind domain.ActionKind, targetID string) domain.HighLevelAction {
	return domain.HighLevelAction{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Kind: kind, TargetID: targetID, Reason: "player-model decision",
	}
}
