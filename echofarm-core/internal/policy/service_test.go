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
	nextAction     domain.HighLevelAction
	replanAction   domain.HighLevelAction
	nextProposal   domain.ActionProposal
	replanProposal domain.ActionProposal
	chooseCalls    int
	replanCalls    int
	lastAction     intelligence.ActionInput
	lastReplan     intelligence.ReplanInput
}

func (s *actorStub) ProposeAction(_ context.Context, input intelligence.ActionInput) (domain.ActionProposal, error) {
	s.chooseCalls++
	s.lastAction = input
	if s.nextProposal.Primary.Kind != "" {
		return s.nextProposal, nil
	}
	return domain.ActionProposal{Primary: s.nextAction, ModelConfidence: 0.8}, nil
}

func (s *actorStub) ProposeRecovery(_ context.Context, input intelligence.ReplanInput) (domain.ActionProposal, error) {
	s.replanCalls++
	s.lastReplan = input
	if s.replanProposal.Primary.Kind != "" {
		return s.replanProposal, nil
	}
	return domain.ActionProposal{Primary: s.replanAction, ModelConfidence: 0.8}, nil
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
	model          domain.PlayerModel
	skill          domain.SkillProgram
	priorDecision  domain.DecisionRecord
	decisionErr    error
	savedDecisions []domain.DecisionRecord
	attachedResult *domain.ActionResult
	saveWinner     *domain.DecisionRecord
	experiences    []domain.PolicyExperience
}

func (s *policyStoreStub) ListPolicyExperiences(context.Context, string) ([]domain.PolicyExperience, error) {
	return append([]domain.PolicyExperience(nil), s.experiences...), nil
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

func (s *policyStoreStub) SaveDecision(_ context.Context, record domain.DecisionRecord) error {
	s.savedDecisions = append(s.savedDecisions, record)
	s.priorDecision = record
	if s.saveWinner != nil {
		s.priorDecision = *s.saveWinner
	}
	s.decisionErr = nil
	return nil
}

func (s *policyStoreStub) GetDecision(context.Context, string, string, int64) (domain.DecisionRecord, error) {
	return s.priorDecision, s.decisionErr
}

func (s *policyStoreStub) AttachDecisionResult(_ context.Context, result domain.ActionResult) error {
	s.attachedResult = &result
	return nil
}

type coordinatorStub struct {
	context domain.CoordinationContext
	calls   int
}

type experienceLearnerStub struct {
	calls  int
	result domain.ActionResult
}

func (s *experienceLearnerStub) LearnFromResult(_ context.Context, _ domain.WorldSnapshot, result domain.ActionResult) (domain.ExperienceOutcome, error) {
	s.calls++
	s.result = result
	return domain.ExperienceOutcome{}, nil
}

func (s *coordinatorStub) Prepare(context.Context, domain.WorldSnapshot, domain.PlayerModel) (domain.CoordinationContext, error) {
	s.calls++
	return s.context, nil
}

func TestNextActionRecordsCoordinatedDecisionAndStopsOnClaimConflict(t *testing.T) {
	snapshot := validSnapshot()
	actor := &actorStub{nextAction: actionFor(snapshot, domain.ActionWaterTarget, "crop-new")}
	store := validPolicyStore()
	coordinator := &coordinatorStub{context: domain.CoordinationContext{
		InferredIntent: domain.PlayerIntentWatering, PlayerClaimedTargets: []string{"crop-new"}, ModelRevision: 1,
	}}
	service, err := NewService(store, actor, coordinator)
	if err != nil {
		t.Fatal(err)
	}

	action, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != domain.ActionStopSession || !strings.Contains(action.Reason, "player") {
		t.Fatalf("final action = %+v", action)
	}
	if len(store.savedDecisions) != 1 || store.savedDecisions[0].CandidateAction.TargetID != "crop-new" || store.savedDecisions[0].FinalAction.Kind != domain.ActionStopSession {
		t.Fatalf("saved decisions = %+v", store.savedDecisions)
	}
	if actor.lastAction.Coordination.InferredIntent != domain.PlayerIntentWatering {
		t.Fatalf("actor coordination = %+v", actor.lastAction.Coordination)
	}
}

func TestNextActionReturnsPersistedDecisionWithoutCallingActor(t *testing.T) {
	snapshot := validSnapshot()
	want := actionFor(snapshot, domain.ActionHarvestTarget, "crop-mature")
	store := validPolicyStore()
	store.priorDecision = domain.DecisionRecord{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		CandidateAction: want, FinalAction: want,
	}
	store.decisionErr = nil
	actor := &actorStub{}
	service, err := NewService(store, actor)
	if err != nil {
		t.Fatal(err)
	}

	got, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if got != want || actor.chooseCalls != 0 {
		t.Fatalf("action/calls = %+v/%d", got, actor.chooseCalls)
	}
}

func TestNextActionReturnsCanonicalPersistedWinner(t *testing.T) {
	snapshot := validSnapshot()
	winnerAction := actionFor(snapshot, domain.ActionHarvestTarget, "crop-mature")
	winner := domain.DecisionRecord{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Day: snapshot.Day, ModelRevision: 1, CandidateAction: winnerAction, FinalAction: winnerAction,
	}
	store := validPolicyStore()
	store.saveWinner = &winner
	actor := &actorStub{nextAction: actionFor(snapshot, domain.ActionWaterTarget, "crop-new")}
	service, err := NewService(store, actor)
	if err != nil {
		t.Fatal(err)
	}

	got, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if got != winnerAction {
		t.Fatalf("NextAction() = %+v, want persisted winner %+v", got, winnerAction)
	}
}

func TestNextDecisionUsesExperienceAndFallsBackToSafeAlternative(t *testing.T) {
	snapshot := validSnapshot()
	experience := validPolicyExperience(snapshot.SaveID, "exp-inventory", 0.8, 2)
	store := validPolicyStore()
	store.experiences = []domain.PolicyExperience{experience}
	store.model.Traits = []domain.TraitMemory{{
		Key: domain.PreferenceTaskOrder, Value: "watering,harvesting", Context: domain.TraitContextSunny,
		Confidence: 0.8, ObservationCount: 2, FirstSeenDay: 1, LastSeenDay: 2,
		EvidenceRefs: []string{"demo-1:water-1"},
	}}
	unsafe := actionFor(snapshot, domain.ActionWaterTarget, "missing-crop")
	fallback := actionFor(snapshot, domain.ActionHarvestTarget, "crop-mature")
	actor := &actorStub{nextProposal: domain.ActionProposal{
		Primary: unsafe, Alternatives: []domain.HighLevelAction{fallback}, ModelConfidence: 0.7,
		AppliedExperienceIDs: []string{experience.ID},
	}}
	service, err := NewService(store, actor)
	if err != nil {
		t.Fatal(err)
	}

	decision, err := service.NextDecision(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Action != fallback || decision.Confidence != 0.95 {
		t.Fatalf("NextDecision() = %+v", decision)
	}
	if len(actor.lastAction.ApplicableExperiences) != 1 || actor.lastAction.ApplicableExperiences[0].ID != experience.ID {
		t.Fatalf("actor experiences = %+v", actor.lastAction.ApplicableExperiences)
	}

	replayed, err := service.NextDecision(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Action != fallback || len(replayed.Alternatives) != 0 {
		t.Fatalf("replayed decision exposed rejected candidates: %+v", replayed)
	}
}

func TestNextDecisionRejectsFabricatedExperienceReference(t *testing.T) {
	snapshot := validSnapshot()
	action := actionFor(snapshot, domain.ActionHarvestTarget, "crop-mature")
	actor := &actorStub{nextProposal: domain.ActionProposal{
		Primary: action, ModelConfidence: 0.8, AppliedExperienceIDs: []string{"made-up"},
	}}
	service := newPolicyService(t, actor)

	if _, err := service.NextDecision(context.Background(), snapshot.SaveID, snapshot); err == nil || !strings.Contains(err.Error(), "experience") {
		t.Fatalf("NextDecision() error = %v, want experience reference error", err)
	}
}

func TestNextDecisionStopsWhenPolicyConfidenceIsLow(t *testing.T) {
	snapshot := validSnapshot()
	primary := actionFor(snapshot, domain.ActionHarvestTarget, "crop-mature")
	actor := &actorStub{nextProposal: domain.ActionProposal{
		Primary: primary, ModelConfidence: 0.4,
		UncertaintyCodes: []domain.UncertaintyCode{domain.UncertaintyNovelContext},
	}}
	service := newPolicyService(t, actor)

	decision, err := service.NextDecision(context.Background(), snapshot.SaveID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Action.Kind != domain.ActionStopSession || decision.Confidence != 0.2 {
		t.Fatalf("NextDecision() = %+v", decision)
	}
}

func TestHandleResultReflectsFailureBeforeReplanning(t *testing.T) {
	executed := validSnapshot()
	current := executed
	current.SnapshotVersion++
	failed := actionFor(executed, domain.ActionHarvestTarget, "crop-mature")
	replanned := actionFor(current, domain.ActionMoveTo, "crop-mature")
	store := validPolicyStore()
	reflection := &experienceLearnerStub{}
	actor := &actorStub{replanAction: replanned}
	service, err := NewReflectiveService(store, actor, noOpCollaborator{}, reflection)
	if err != nil {
		t.Fatal(err)
	}
	result := domain.ActionResult{
		SaveID: failed.SaveID, SessionID: failed.SessionID, SnapshotVersion: failed.SnapshotVersion,
		Action: failed, Status: domain.ActionFailed, ErrorCode: "path_blocked",
	}

	if _, err := service.HandleResult(context.Background(), current.SaveID, current, result); err != nil {
		t.Fatal(err)
	}
	if reflection.calls != 1 || reflection.result.ErrorCode != "path_blocked" {
		t.Fatalf("reflection calls/result = %d / %+v", reflection.calls, reflection.result)
	}
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
	executedSnapshot := validSnapshot()
	snapshot := executedSnapshot
	snapshot.SnapshotVersion++
	failed := actionFor(executedSnapshot, domain.ActionWaterTarget, "crop-new")
	replanned := actionFor(snapshot, domain.ActionMoveTo, "crop-new")
	actor := &actorStub{replanAction: replanned}
	service := newPolicyService(t, actor)

	action, err := service.HandleResult(context.Background(), snapshot.SaveID, snapshot, domain.ActionResult{
		SaveID: executedSnapshot.SaveID, SessionID: executedSnapshot.SessionID, SnapshotVersion: executedSnapshot.SnapshotVersion,
		Action: failed, Status: domain.ActionFailed, ErrorCode: "path_blocked",
	})
	if err != nil {
		t.Fatalf("HandleResult() error = %v", err)
	}
	if action.Kind != domain.ActionMoveTo || actor.replanCalls != 1 {
		t.Fatalf("action = %+v, replan calls = %d", action, actor.replanCalls)
	}
}

func TestHandleResultRejectsUnknownStatusBeforeWritingLedger(t *testing.T) {
	snapshot := validSnapshot()
	store := validPolicyStore()
	service, err := NewService(store, &actorStub{})
	if err != nil {
		t.Fatal(err)
	}
	result := domain.ActionResult{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Action: actionFor(snapshot, domain.ActionWaterTarget, "crop-new"), Status: "maybe",
	}

	if _, err := service.HandleResult(context.Background(), snapshot.SaveID, snapshot, result); err == nil {
		t.Fatal("HandleResult() error = nil, want invalid status")
	}
	if store.attachedResult != nil {
		t.Fatalf("invalid result was persisted: %+v", store.attachedResult)
	}
}

func TestHandleResultRequiresNewerSnapshotThanExecutedAction(t *testing.T) {
	snapshot := validSnapshot()
	store := validPolicyStore()
	service, err := NewService(store, &actorStub{})
	if err != nil {
		t.Fatal(err)
	}
	result := domain.ActionResult{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Action: actionFor(snapshot, domain.ActionWaterTarget, "crop-new"), Status: domain.ActionSucceeded,
	}

	_, err = service.HandleResult(context.Background(), snapshot.SaveID, snapshot, result)
	if err == nil || !strings.Contains(err.Error(), "newer") {
		t.Fatalf("HandleResult() error = %v, want newer snapshot error", err)
	}
	if store.attachedResult != nil {
		t.Fatalf("same-version result was persisted: %+v", store.attachedResult)
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

func TestNextActionRejectsStructurallyInvalidProposal(t *testing.T) {
	snapshot := validSnapshot()
	actor := &actorStub{nextAction: actionFor(snapshot, "teleport", "crop-new")}
	service := newPolicyService(t, actor)

	_, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
	if err == nil || !strings.Contains(err.Error(), "unsupported action") {
		t.Fatalf("NextAction() error = %v, want unsupported action", err)
	}
}

func TestNextActionStopsWhenEveryCandidateIsUnsafe(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*domain.WorldSnapshot)
		action func(domain.WorldSnapshot) domain.HighLevelAction
	}{
		{
			name: "stale target",
			action: func(snapshot domain.WorldSnapshot) domain.HighLevelAction {
				return actionFor(snapshot, domain.ActionWaterTarget, "crop-yesterday")
			},
		},
		{
			name:   "watering in rain",
			mutate: func(snapshot *domain.WorldSnapshot) { snapshot.Weather = domain.WeatherRainy },
			action: func(snapshot domain.WorldSnapshot) domain.HighLevelAction {
				return actionFor(snapshot, domain.ActionWaterTarget, "crop-new")
			},
		},
		{
			name: "depositing an empty inventory",
			action: func(snapshot domain.WorldSnapshot) domain.HighLevelAction {
				return actionFor(snapshot, domain.ActionDepositItems, "chest-1")
			},
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
			action, err := service.NextAction(context.Background(), snapshot.SaveID, snapshot)
			if err != nil {
				t.Fatal(err)
			}
			if action.Kind != domain.ActionStopSession || !strings.Contains(action.Reason, "safety") {
				t.Fatalf("NextAction() = %+v, want safety stop", action)
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
	store := validPolicyStore()
	service, err := NewService(store, actor)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func validPolicyStore() *policyStoreStub {
	return &policyStoreStub{
		model: domain.PlayerModel{SaveID: "farm-1", Revision: 1, EnergyReserve: 40},
		skill: domain.SkillProgram{
			Name: "morning-farm-routine", Revision: 1, Goal: "care for crops", TargetSelector: "actionable_crops",
			Steps:             []domain.SkillStep{{Action: domain.ActionWaterTarget, TargetSelector: "dry_crops"}},
			SuccessConditions: []string{"all crops cared for"}, EvidenceEventIDs: []string{"water-1"},
		},
		decisionErr: memory.ErrNotFound,
	}
}

func validPolicyExperience(saveID, id string, confidence float64, observations int) domain.PolicyExperience {
	return domain.PolicyExperience{
		ID: id, SaveID: saveID, Trigger: domain.ExperienceInventoryFull, Context: domain.TraitContextSunny,
		WhenSignals: []domain.SituationSignal{domain.SignalInventoryFull},
		AvoidAction: domain.ActionHarvestTarget, PreferAction: domain.ActionDepositItems,
		PreferredTargetID: "chest-1", Summary: "deposit first", Confidence: confidence,
		ObservationCount: observations, FirstSeenDay: 1, LastSeenDay: 2,
		EvidenceRefs: []string{"decision:echo-1:1"}, Source: domain.ExperienceSourceFailure,
	}
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
