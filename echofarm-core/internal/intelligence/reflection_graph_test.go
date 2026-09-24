package intelligence

import (
	"context"
	"errors"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

type reflectionGeneratorStub struct {
	observation domain.ExperienceObservation
	prompt      string
	input       any
}

func (s *reflectionGeneratorStub) GenerateJSON(_ context.Context, prompt string, input any, output any) error {
	s.prompt = prompt
	s.input = input
	target, ok := output.(*domain.ExperienceObservation)
	if !ok {
		return errors.New("unexpected output type")
	}
	*target = s.observation
	return nil
}

func TestReflectionGraphProducesBoundedFailureExperience(t *testing.T) {
	input := failureReflectionInput()
	want := domain.ExperienceObservation{
		Trigger: domain.ExperienceInventoryFull, Context: domain.TraitContextSunny,
		WhenSignals: []domain.SituationSignal{domain.SignalInventoryFull},
		AvoidAction: domain.ActionHarvestTarget, PreferAction: domain.ActionDepositItems,
		PreferredTargetID: "chest-1", Summary: "deposit before harvesting",
		EvidenceRef: input.EvidenceRef, Strength: 0.7,
	}
	generator := &reflectionGeneratorStub{observation: want}
	graph, err := NewReflectionGraph(generator)
	if err != nil {
		t.Fatal(err)
	}

	got, err := graph.Reflect(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if got.EvidenceRef != input.EvidenceRef || got.PreferAction != domain.ActionDepositItems {
		t.Fatalf("Reflect() = %+v", got)
	}
}

func TestReflectionGraphRejectsFabricatedEvidence(t *testing.T) {
	input := failureReflectionInput()
	observation := domain.ExperienceObservation{
		Trigger: domain.ExperienceInventoryFull, Context: domain.TraitContextSunny,
		WhenSignals: []domain.SituationSignal{domain.SignalInventoryFull},
		AvoidAction: domain.ActionHarvestTarget, PreferAction: domain.ActionDepositItems,
		PreferredTargetID: "chest-1", Summary: "deposit before harvesting",
		EvidenceRef: "made-up", Strength: 0.7,
	}
	graph, err := NewReflectionGraph(&reflectionGeneratorStub{observation: observation})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := graph.Reflect(context.Background(), input); !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("Reflect() error = %v, want ErrInvalidModelOutput", err)
	}
}

func failureReflectionInput() ReflectionInput {
	snapshot := validActionInput().Snapshot
	snapshot.Inventory = domain.InventorySummary{
		FreeSlots: 0, Items: []domain.InventoryItem{{ItemID: "parsnip", Name: "Parsnip", Quantity: 1}},
	}
	snapshot.Chests = []domain.Chest{{ID: "chest-1"}}
	action := domain.HighLevelAction{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion - 1,
		Kind: domain.ActionHarvestTarget, TargetID: "crop-new", Reason: "harvest",
	}
	return ReflectionInput{
		Snapshot:    snapshot,
		PlayerModel: domain.PlayerModel{SaveID: snapshot.SaveID, Revision: 1, EnergyReserve: 40, PreferredChestID: "chest-1"},
		Result: &domain.ActionResult{
			SaveID: action.SaveID, SessionID: action.SessionID, SnapshotVersion: action.SnapshotVersion,
			Action: action, Status: domain.ActionFailed, ErrorCode: "inventory_full",
		},
		EvidenceRef: "decision:day-2:6",
	}
}
