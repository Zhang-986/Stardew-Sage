package intelligence

import (
	"context"
	"errors"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

type intentGeneratorStub struct {
	result IntentInference
	err    error
}

func (s *intentGeneratorStub) GenerateJSON(_ context.Context, _ string, input any, output any) error {
	if s.err != nil {
		return s.err
	}
	if _, ok := input.(IntentInput); !ok {
		return errors.New("unexpected intent input")
	}
	target, ok := output.(*IntentInference)
	if !ok {
		return errors.New("unexpected intent output")
	}
	*target = s.result
	return nil
}

func TestIntentGraphReturnsEvidenceBackedIntent(t *testing.T) {
	graph, err := NewIntentGraph(&intentGeneratorStub{result: IntentInference{
		Intent: domain.PlayerIntentWatering, EvidenceTargetIDs: []string{"crop-1"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := graph.InferIntent(context.Background(), intentInput())
	if err != nil || got != domain.PlayerIntentWatering {
		t.Fatalf("InferIntent() = %q, %v", got, err)
	}
}

func TestIntentGraphRejectsForgedTargetEvidence(t *testing.T) {
	graph, err := NewIntentGraph(&intentGeneratorStub{result: IntentInference{
		Intent: domain.PlayerIntentWatering, EvidenceTargetIDs: []string{"invented"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.InferIntent(context.Background(), intentInput()); !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("InferIntent() error = %v, want ErrInvalidModelOutput", err)
	}
}

func TestIntentGraphRejectsUnknownIntent(t *testing.T) {
	graph, err := NewIntentGraph(&intentGeneratorStub{result: IntentInference{Intent: "shopping"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.InferIntent(context.Background(), intentInput()); !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("InferIntent() error = %v, want ErrInvalidModelOutput", err)
	}
}

func intentInput() IntentInput {
	return IntentInput{
		SaveID: "farm-1", Day: 4, TimeOfDay: 700,
		Activities:  []domain.PlayerActivity{{Kind: domain.EventWater, TargetID: "crop-1", Tick: 100, Success: true}},
		PlayerModel: domain.PlayerModel{SaveID: "farm-1", Revision: 3, EnergyReserve: 40},
	}
}
