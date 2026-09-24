package memoryview

import (
	"context"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

type storeStub struct {
	model       domain.PlayerModel
	outcome     domain.LearningOutcome
	decision    domain.DecisionRecord
	session     domain.EchoSessionMemory
	experiences []domain.PolicyExperience
	usage       domain.ModelUsageSummary
	usageErr    error
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
func (s *storeStub) ListPolicyExperiences(context.Context, string) ([]domain.PolicyExperience, error) {
	return append([]domain.PolicyExperience(nil), s.experiences...), nil
}
func (s *storeStub) GetLatestModelUsageSummary(context.Context, string) (domain.ModelUsageSummary, error) {
	return s.usage, s.usageErr
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
		outcome: domain.LearningOutcome{Change: domain.LearningChange{ModelRevision: 3, Kind: domain.LearningChangeStrengthened, Summary: "task order strengthened"}},
		decision: domain.DecisionRecord{
			SaveID: "farm-1", SessionID: "echo-4", SnapshotVersion: 8, PolicyConfidence: 0.91,
			SafeAlternatives: []domain.HighLevelAction{{SaveID: "farm-1", SessionID: "echo-4", SnapshotVersion: 8, Kind: domain.ActionStopSession}},
		},
		session: domain.EchoSessionMemory{SessionID: "echo-4", Day: 4, Status: "active"},
		experiences: []domain.PolicyExperience{
			{ID: "exp-low", SaveID: "farm-1", Trigger: domain.ExperienceInventoryFull, Context: domain.TraitContextSunny, WhenSignals: []domain.SituationSignal{domain.SignalInventoryFull}, PreferAction: domain.ActionDepositItems, Summary: "older", Confidence: 0.6, ObservationCount: 1, FirstSeenDay: 2, LastSeenDay: 2, EvidenceRefs: []string{"decision:echo-2:1"}, Source: domain.ExperienceSourceFailure},
			{ID: "exp-correction", SaveID: "farm-1", Trigger: domain.ExperiencePlayerCorrection, Context: domain.TraitContextSunny, WhenSignals: []domain.SituationSignal{domain.SignalInventoryHasItems}, PreferAction: domain.ActionDepositItems, PreferredTargetID: "chest-west", Summary: "newer", Confidence: 0.85, ObservationCount: 1, FirstSeenDay: 4, LastSeenDay: 4, EvidenceRefs: []string{"correction-1"}, Source: domain.ExperienceSourceCorrection},
		},
		usage: domain.ModelUsageSummary{
			SaveID: "farm-1", SessionID: "echo-4", Day: 4,
			Session:    domain.ModelUsageTotals{Calls: 3, Succeeded: 2, Failed: 1, TotalTokens: 120, ReportedTokenCalls: 3, TokensKnown: true},
			DayTotals:  domain.ModelUsageTotals{Calls: 5, Succeeded: 4, Failed: 1, TotalTokens: 200, ReportedTokenCalls: 5, TokensKnown: true},
			CallBudget: 32, TokenBudget: 100000,
		},
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
	if view.LastDecision.PolicyConfidence != 0.91 || len(view.LastDecision.SafeAlternatives) != 1 {
		t.Fatalf("last decision metadata = %+v", view.LastDecision)
	}
	if len(view.Experiences) != 2 || view.Experiences[0].ID != "exp-correction" {
		t.Fatalf("experiences = %+v", view.Experiences)
	}
	if view.ModelUsage == nil || view.ModelUsage.Session.Calls != 3 || view.ModelUsage.DayTotals.TotalTokens != 200 {
		t.Fatalf("model usage = %+v", view.ModelUsage)
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
