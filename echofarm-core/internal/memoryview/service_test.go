package memoryview

import (
	"context"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

type storeStub struct {
	model    domain.PlayerModel
	outcome  domain.LearningOutcome
	decision domain.DecisionRecord
	session  domain.EchoSessionMemory
}

func (s *storeStub) GetPlayerModel(context.Context, string) (domain.PlayerModel, error) {
	return s.model, nil
}
func (s *storeStub) GetLatestLearningOutcome(context.Context, string) (domain.LearningOutcome, error) {
	return s.outcome, nil
}
func (s *storeStub) GetLatestDecision(context.Context, string) (domain.DecisionRecord, error) {
	return s.decision, nil
}
func (s *storeStub) GetActiveSession(context.Context, string) (domain.EchoSessionMemory, error) {
	return s.session, nil
}

func TestGetBuildsStableSortedMemoryView(t *testing.T) {
	store := &storeStub{
		model: domain.PlayerModel{
			SaveID: "farm-1", Revision: 3, LearnedThroughDay: 3, EnergyReserve: 40,
			Traits: []domain.TraitMemory{
				{Key: domain.PreferenceRouteStyle, Value: "ordered", Context: domain.TraitContextAny, Confidence: 0.95, ObservationCount: 1, EvidenceRefs: []string{"d:e"}},
				{Key: domain.PreferencePreferredChest, Value: "east", Context: domain.TraitContextAny, Confidence: 0.82, ObservationCount: 2, EvidenceRefs: []string{"d1:e", "d2:e"}},
				{Key: domain.PreferenceTaskOrder, Value: "watering,harvesting", Context: domain.TraitContextSunny, Confidence: 0.91, ObservationCount: 3, EvidenceRefs: []string{"d1:e"}},
			},
		},
		outcome:  domain.LearningOutcome{Change: domain.LearningChange{ModelRevision: 3, Kind: domain.LearningChangeStrengthened, Summary: "task order strengthened"}},
		decision: domain.DecisionRecord{SaveID: "farm-1", SessionID: "echo-4", SnapshotVersion: 8},
		session:  domain.EchoSessionMemory{SessionID: "echo-4", Day: 4, Status: "active"},
	}
	service, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}

	view, err := service.Get(context.Background(), "farm-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(view.StableTraits) != 2 || view.StableTraits[0].Key != domain.PreferenceTaskOrder {
		t.Fatalf("stable traits = %+v", view.StableTraits)
	}
	if view.RecentLearningChange == nil || view.LastDecision == nil || view.ActiveSession == nil {
		t.Fatalf("incomplete memory view = %+v", view)
	}
}

func TestGetRequiresSaveID(t *testing.T) {
	service, err := NewService(&storeStub{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(context.Background(), ""); err == nil {
		t.Fatal("Get() error = nil")
	}
}
