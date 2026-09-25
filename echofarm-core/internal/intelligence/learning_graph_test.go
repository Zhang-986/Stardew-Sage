package intelligence

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestLearningPromptDefinesExactJSONFieldNames(t *testing.T) {
	for _, field := range []string{
		`"observations"`,
		`"skill"`,
		`"supportingEventIds"`,
		`"targetSelector"`,
		`"successConditions"`,
		`"evidenceEventIds"`,
	} {
		if !strings.Contains(learningSystemPrompt, field) {
			t.Fatalf("learning prompt does not define JSON field %s", field)
		}
	}
}

type stubGenerator struct {
	result LearningInference
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
	target, ok := output.(*LearningInference)
	if !ok {
		return errors.New("unexpected output type")
	}
	*target = s.result
	return nil
}

func TestLearningGraphReturnsValidatedEvidenceBackedInference(t *testing.T) {
	generator := &stubGenerator{result: validLearningInference()}
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
	if got.Observations[0].SupportingEventIDs[0] != "water-1" {
		t.Fatalf("evidence was not preserved: %+v", got.Observations[0])
	}
	if got.Skill.Name != "morning-farm-routine" {
		t.Fatalf("skill name = %q", got.Skill.Name)
	}
}

func TestLearningGraphRejectsUnknownEvidence(t *testing.T) {
	result := validLearningInference()
	result.Observations[0].SupportingEventIDs = []string{"invented-event"}
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
	result := validLearningInference()
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

func TestLearningGraphRejectsUnsupportedTraitKey(t *testing.T) {
	result := validLearningInference()
	result.Observations[0].Key = "favorite_hat"
	graph, err := NewLearningGraph(&stubGenerator{result: result})
	if err != nil {
		t.Fatalf("NewLearningGraph() error = %v", err)
	}

	_, err = graph.Learn(context.Background(), learningInput())
	if !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("Learn() error = %v, want ErrInvalidModelOutput", err)
	}
}

func TestLearningGraphAcceptsWeatherScopedTrait(t *testing.T) {
	input := learningInput()
	input.Demonstration.Day = 3
	input.Demonstration.Weather = domain.WeatherRainy
	result := validLearningInference()
	result.Observations[0].Context = domain.TraitContextRainy
	graph, err := NewLearningGraph(&stubGenerator{result: result})
	if err != nil {
		t.Fatalf("NewLearningGraph() error = %v", err)
	}

	got, err := graph.Learn(context.Background(), input)
	if err != nil {
		t.Fatalf("Learn() error = %v", err)
	}
	if got.Observations[0].Context != domain.TraitContextRainy {
		t.Fatalf("context = %q", got.Observations[0].Context)
	}
}

func TestLearningGraphAcceptsEvidenceBackedLifestyleTraits(t *testing.T) {
	input := learningInput()
	input.Demonstration.SchemaVersion = 2
	input.Demonstration.Events = append(input.Demonstration.Events, domain.DemonstrationEvent{
		ID: "tree-1", Kind: domain.EventChopTree, Tick: 3, TargetID: "tree-1", TargetKind: "tree", Success: true,
	})
	input.Segments = append(input.Segments, domain.BehaviorSegment{
		Kind: domain.BehaviorWoodcutting, EventIDs: []string{"tree-1"}, TargetIDs: []string{"tree-1"},
	})
	result := validLearningInference()
	result.Observations = append(result.Observations, domain.TraitObservation{
		Key: domain.PreferenceActivityOrder, Value: "watering,woodcutting", Context: domain.TraitContextSunny,
		SupportingEventIDs: []string{"water-1", "tree-1"}, Strength: 0.7,
	})
	graph, err := NewLearningGraph(&stubGenerator{result: result})
	if err != nil {
		t.Fatal(err)
	}

	got, err := graph.Learn(context.Background(), input)
	if err != nil {
		t.Fatalf("Learn() error = %v", err)
	}
	if got.Observations[1].Key != domain.PreferenceActivityOrder {
		t.Fatalf("lifestyle observation = %+v", got.Observations[1])
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

func validLearningInference() LearningInference {
	return LearningInference{
		Observations: []domain.TraitObservation{{
			Key: domain.PreferenceTaskOrder, Value: "watering,depositing", Context: domain.TraitContextAny,
			SupportingEventIDs: []string{"water-1", "deposit-1"}, Strength: 0.7,
		}},
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
