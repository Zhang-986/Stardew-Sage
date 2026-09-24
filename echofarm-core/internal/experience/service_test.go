package experience

import (
	"context"
	"errors"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

type reflectorStub struct {
	observation domain.ExperienceObservation
	err         error
	calls       int
	input       intelligence.ReflectionInput
}

func (s *reflectorStub) Reflect(_ context.Context, input intelligence.ReflectionInput) (domain.ExperienceObservation, error) {
	s.calls++
	s.input = input
	return s.observation, s.err
}

type experienceStoreStub struct {
	model           domain.PlayerModel
	experiences     []domain.PolicyExperience
	outcomes        map[string]domain.ExperienceOutcome
	saveCalls       int
	canonicalWinner *domain.ExperienceOutcome
	latestDecision  domain.DecisionRecord
	latestErr       error
}

func (s *experienceStoreStub) GetPlayerModel(context.Context, string) (domain.PlayerModel, error) {
	return s.model, nil
}

func (s *experienceStoreStub) GetExperienceOutcome(_ context.Context, _ string, sourceID string) (domain.ExperienceOutcome, error) {
	if outcome, ok := s.outcomes[sourceID]; ok {
		return outcome, nil
	}
	return domain.ExperienceOutcome{}, memory.ErrNotFound
}

func (s *experienceStoreStub) SaveExperienceOutcome(_ context.Context, outcome domain.ExperienceOutcome) error {
	s.saveCalls++
	if s.outcomes == nil {
		s.outcomes = make(map[string]domain.ExperienceOutcome)
	}
	if s.canonicalWinner != nil {
		s.outcomes[outcome.SourceID] = *s.canonicalWinner
	} else {
		s.outcomes[outcome.SourceID] = outcome
	}
	if len(outcome.UpdatedExperiences) > 0 {
		s.experiences = append([]domain.PolicyExperience(nil), outcome.UpdatedExperiences...)
	}
	return nil
}

func (s *experienceStoreStub) ListPolicyExperiences(context.Context, string) ([]domain.PolicyExperience, error) {
	return append([]domain.PolicyExperience(nil), s.experiences...), nil
}

func (s *experienceStoreStub) GetLatestDecision(context.Context, string) (domain.DecisionRecord, error) {
	if s.latestErr != nil {
		return domain.DecisionRecord{}, s.latestErr
	}
	if s.latestDecision.SaveID == "" {
		return domain.DecisionRecord{}, memory.ErrNotFound
	}
	return s.latestDecision, nil
}

func TestLearnFromResultPersistsGeneralizedExperience(t *testing.T) {
	snapshot, result := reflectiveFailure()
	sourceID := FailureSourceID(result)
	observation := experienceObservation("chest-east", sourceID)
	store := &experienceStoreStub{
		model: domain.PlayerModel{SaveID: snapshot.SaveID, Revision: 2, EnergyReserve: 40},
	}
	reflector := &reflectorStub{observation: observation}
	service, err := NewService(store, reflector)
	if err != nil {
		t.Fatal(err)
	}

	outcome, err := service.LearnFromResult(context.Background(), snapshot, result)
	if err != nil {
		t.Fatal(err)
	}
	if reflector.calls != 1 || store.saveCalls != 1 || outcome.Source != domain.ExperienceSourceFailure {
		t.Fatalf("outcome/calls = %+v / %d / %d", outcome, reflector.calls, store.saveCalls)
	}
	if outcome.Experience.PreferAction != domain.ActionDepositItems || len(outcome.UpdatedExperiences) != 1 {
		t.Fatalf("learned outcome = %+v", outcome)
	}
}

func TestLearnFromResultReturnsExistingOutcomeWithoutReflecting(t *testing.T) {
	snapshot, result := reflectiveFailure()
	want := domain.ExperienceOutcome{
		SourceID: FailureSourceID(result), Source: domain.ExperienceSourceFailure,
		Experience: policyExperience("exp-existing", domain.TraitContextSunny, "chest-east", 0.7, 1),
	}
	store := &experienceStoreStub{
		model:    domain.PlayerModel{SaveID: snapshot.SaveID, Revision: 2, EnergyReserve: 40},
		outcomes: map[string]domain.ExperienceOutcome{want.SourceID: want},
	}
	reflector := &reflectorStub{}
	service, err := NewService(store, reflector)
	if err != nil {
		t.Fatal(err)
	}

	got, err := service.LearnFromResult(context.Background(), snapshot, result)
	if err != nil {
		t.Fatal(err)
	}
	if reflector.calls != 0 || got.SourceID != want.SourceID {
		t.Fatalf("outcome/calls = %+v / %d", got, reflector.calls)
	}
}

func TestLearnFromCorrectionUsesHighSignalCapAndCanonicalWinner(t *testing.T) {
	correction := reflectiveCorrection()
	sourceID := CorrectionSourceID(correction)
	observation := experienceObservation("chest-west", sourceID)
	observation.Trigger = domain.ExperiencePlayerCorrection
	winner := domain.ExperienceOutcome{
		SourceID: sourceID, Source: domain.ExperienceSourceCorrection,
		Experience: policyExperience("exp-winner", domain.TraitContextSunny, "chest-west", 0.82, 2),
	}
	store := &experienceStoreStub{
		model:           domain.PlayerModel{SaveID: correction.SaveID, Revision: 2, EnergyReserve: 40},
		canonicalWinner: &winner,
		latestDecision: domain.DecisionRecord{
			SaveID: correction.SaveID, SessionID: correction.SessionID,
			SnapshotVersion: correction.RejectedDecisionSnapshotVersion, FinalAction: correction.RejectedAction,
		},
	}
	service, err := NewService(store, &reflectorStub{observation: observation})
	if err != nil {
		t.Fatal(err)
	}

	got, err := service.LearnFromCorrection(context.Background(), correction)
	if err != nil {
		t.Fatal(err)
	}
	if got.Experience.ID != winner.Experience.ID || store.saveCalls != 1 {
		t.Fatalf("outcome = %+v, want canonical winner %+v", got, winner)
	}
}

func TestLearnFromCorrectionRejectsAnythingButLatestPersistedDecision(t *testing.T) {
	correction := reflectiveCorrection()
	observation := experienceObservation("chest-west", correction.ID)
	observation.Trigger = domain.ExperiencePlayerCorrection
	reflector := &reflectorStub{observation: observation}
	store := &experienceStoreStub{
		model: domain.PlayerModel{SaveID: correction.SaveID, Revision: 2, EnergyReserve: 40},
		latestDecision: domain.DecisionRecord{
			SaveID: correction.SaveID, SessionID: correction.SessionID,
			SnapshotVersion: correction.RejectedDecisionSnapshotVersion,
			FinalAction: domain.HighLevelAction{
				SaveID: correction.SaveID, SessionID: correction.SessionID,
				SnapshotVersion: correction.RejectedDecisionSnapshotVersion,
				Kind:            domain.ActionDepositItems, TargetID: "chest-east", Reason: "different recorded action",
			},
		},
	}
	service, err := NewService(store, reflector)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.LearnFromCorrection(context.Background(), correction); err == nil {
		t.Fatal("LearnFromCorrection() error = nil, want latest-decision mismatch")
	}
	if reflector.calls != 0 || store.saveCalls != 0 {
		t.Fatalf("reflect/save calls = %d/%d, want 0/0", reflector.calls, store.saveCalls)
	}
}

func TestLearnFromResultDoesNotWriteWhenModelUnavailable(t *testing.T) {
	snapshot, result := reflectiveFailure()
	store := &experienceStoreStub{model: domain.PlayerModel{SaveID: snapshot.SaveID, Revision: 2, EnergyReserve: 40}}
	service, err := NewService(store, &reflectorStub{err: intelligence.ErrModelUnavailable})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.LearnFromResult(context.Background(), snapshot, result); !errors.Is(err, intelligence.ErrModelUnavailable) {
		t.Fatalf("LearnFromResult() error = %v", err)
	}
	if store.saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", store.saveCalls)
	}
}

func reflectiveFailure() (domain.WorldSnapshot, domain.ActionResult) {
	snapshot := domain.WorldSnapshot{
		SaveID: "farm-1", SessionID: "echo-day-4", SnapshotVersion: 2,
		Day: 4, TimeOfDay: 710, Weather: domain.WeatherSunny, Location: "Farm",
		Energy: 200, MaxEnergy: 270,
		Inventory:   domain.InventorySummary{FreeSlots: 0, Items: []domain.InventoryItem{{ItemID: "parsnip", Name: "Parsnip", Quantity: 1}}},
		WateringCan: domain.ToolState{Name: "Watering Can", Water: 10, Capacity: 40},
		Crops:       []domain.Crop{{ID: "crop-1", Mature: true}}, Chests: []domain.Chest{{ID: "chest-east"}, {ID: "chest-west"}},
	}
	action := domain.HighLevelAction{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: 1,
		Kind: domain.ActionHarvestTarget, TargetID: "crop-1", Reason: "harvest",
	}
	return snapshot, domain.ActionResult{
		SaveID: action.SaveID, SessionID: action.SessionID, SnapshotVersion: action.SnapshotVersion,
		Action: action, Status: domain.ActionFailed, ErrorCode: "inventory_full",
	}
}

func reflectiveCorrection() domain.PlayerCorrection {
	snapshot, result := reflectiveFailure()
	return domain.PlayerCorrection{
		ID: "correction-1", SaveID: snapshot.SaveID, SessionID: snapshot.SessionID,
		RejectedDecisionSnapshotVersion: result.SnapshotVersion, RejectedAction: result.Action,
		Snapshot: snapshot,
		PreferredAction: domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
			Kind: domain.ActionDepositItems, TargetID: "chest-west", Reason: "player correction",
		},
		ObservedAtTick: 200,
	}
}
