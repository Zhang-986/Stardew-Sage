package intelligence

import (
	"context"
	"errors"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

type stubGenerator struct {
	result LearningResult
	err    error
	calls  int
}

func (s *stubGenerator) GenerateJSON(_ context.Context, _ string, input any, output any) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	request, ok := input.(LearningInput)
	if !ok || request.Demonstration.ID == "" {
		return errors.New("missing learning input")
	}
	target, ok := output.(*LearningResult)
	if !ok {
		return errors.New("unexpected output type")
	}
	*target = s.result
	return nil
}

func TestLearningGraphReturnsValidatedEvidenceBackedResult(t *testing.T) {
	generator := &stubGenerator{result: validLearningResult()}
	graph, err := NewLearningGraph(generator)
	if err != nil {
		t.Fatalf("NewLearningGraph() error = %v", err)
	}

	got, err := graph.Learn(context.Background(), learningInput())
	if err != nil {
		t.Fatalf("Learn() error = %v", err)
	}
	if generator.calls != 1 {
		t.Fatalf("generator calls = %d, want 1", generator.calls)
	}
	if got.PlayerModel.Preferences[0].EvidenceEventIDs[0] != "water-1" {
		t.Fatalf("evidence was not preserved: %+v", got.PlayerModel.Preferences[0])
	}
	if got.Skill.Name != "morning-farm-routine" {
		t.Fatalf("skill name = %q", got.Skill.Name)
	}
}

func TestLearningGraphRejectsUnknownEvidence(t *testing.T) {
	result := validLearningResult()
	result.PlayerModel.Preferences[0].EvidenceEventIDs = []string{"invented-event"}
	graph, err := NewLearningGraph(&stubGenerator{result: result})
	if err != nil {
		t.Fatalf("NewLearningGraph() error = %v", err)
	}

	_, err = graph.Learn(context.Background(), learningInput())
	if !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("Learn() error = %v, want ErrInvalidModelOutput", err)
	}
}

func TestLearningGraphRejectsUnsupportedSkillAction(t *testing.T) {
	result := validLearningResult()
	result.Skill.Steps[0].Action = "teleport"
	graph, err := NewLearningGraph(&stubGenerator{result: result})
	if err != nil {
		t.Fatalf("NewLearningGraph() error = %v", err)
	}

	_, err = graph.Learn(context.Background(), learningInput())
	if !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("Learn() error = %v, want ErrInvalidModelOutput", err)
	}
}

func TestLearningGraphWrapsGeneratorFailure(t *testing.T) {
	graph, err := NewLearningGraph(&stubGenerator{err: errors.New("bad json")})
	if err != nil {
		t.Fatalf("NewLearningGraph() error = %v", err)
	}

	_, err = graph.Learn(context.Background(), learningInput())
	if !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("Learn() error = %v, want ErrInvalidModelOutput", err)
	}
}

func learningInput() LearningInput {
	demo := domain.Demonstration{
		ID: "demo-1", SaveID: "farm-1", SessionID: "teaching-1", StartedAt: 1, EndedAt: 2,
		Events: []domain.DemonstrationEvent{
			{ID: "water-1", Kind: domain.EventWater, Tick: 1, TargetID: "crop-old", Success: true},
			{ID: "deposit-1", Kind: domain.EventDeposit, Tick: 2, TargetID: "chest-1", Success: true},
		},
	}
	return LearningInput{
		Demonstration: demo,
		Segments: []domain.BehaviorSegment{
			{Kind: domain.BehaviorWatering, EventIDs: []string{"water-1"}, TargetIDs: []string{"crop-old"}},
			{Kind: domain.BehaviorDepositing, EventIDs: []string{"deposit-1"}, TargetIDs: []string{"chest-1"}},
		},
	}
}

func validLearningResult() LearningResult {
	return LearningResult{
		PlayerModel: domain.PlayerModel{
			SaveID: "farm-1", Revision: 1, EnergyReserve: 40,
			CommonTaskOrder:  []domain.BehaviorKind{domain.BehaviorWatering, domain.BehaviorDepositing},
			PreferredChestID: "chest-1",
			Preferences: []domain.ObservedPreference{{
				Key: domain.PreferenceTaskOrder, Value: "water_before_deposit",
				EvidenceEventIDs: []string{"water-1", "deposit-1"}, ObservationCount: 1, Confidence: 0.7,
			}},
		},
		Skill: domain.SkillProgram{
			Name: "morning-farm-routine", Revision: 1, Goal: "care for all crops",
			TargetSelector: "current_actionable_crops",
			Steps: []domain.SkillStep{
				{Action: domain.ActionWaterTarget, TargetSelector: "dry_crops"},
				{Action: domain.ActionDepositItems, TargetSelector: "preferred_chest"},
			},
			SuccessConditions: []string{"no actionable crops remain"},
			EvidenceEventIDs:  []string{"water-1", "deposit-1"},
		},
	}
}
