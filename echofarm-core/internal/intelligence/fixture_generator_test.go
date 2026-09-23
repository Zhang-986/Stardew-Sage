package intelligence

import (
	"context"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestFixtureGeneratorBuildsEvidenceBackedLearningResult(t *testing.T) {
	generator := NewFixtureGenerator()
	input := learningInput()
	var output LearningResult

	if err := generator.GenerateJSON(context.Background(), learningSystemPrompt, input, &output); err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}
	if output.PlayerModel.SaveID != input.Demonstration.SaveID || output.PlayerModel.Revision != 1 {
		t.Fatalf("player model = %+v", output.PlayerModel)
	}
	if len(output.Skill.EvidenceEventIDs) != len(input.Demonstration.Events) {
		t.Fatalf("skill evidence = %v", output.Skill.EvidenceEventIDs)
	}
}

func TestFixtureLearningResultIncludesInventoryRecoveryStrategies(t *testing.T) {
	generator := NewFixtureGenerator()
	var output LearningResult

	if err := generator.GenerateJSON(context.Background(), learningSystemPrompt, learningInput(), &output); err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}
	want := map[string]domain.ActionKind{
		"inventory_full": domain.ActionDepositItems,
		"chest_full":     domain.ActionStopSession,
	}
	for _, recovery := range output.Skill.RecoveryStrategies {
		if action, ok := want[recovery.FailureCode]; ok && len(recovery.Actions) > 0 && recovery.Actions[0] == action {
			delete(want, recovery.FailureCode)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing inventory recovery strategies: %v", want)
	}
}

func TestFixtureGeneratorDemonstratesChangedWorldDecisions(t *testing.T) {
	generator := NewFixtureGenerator()
	tests := []struct {
		name   string
		mutate func(*ActionInput)
		want   domain.ActionKind
		target string
	}{
		{
			name: "rain skips watering and harvests",
			mutate: func(input *ActionInput) {
				input.Snapshot.Weather = domain.WeatherRainy
				input.Snapshot.Crops = append(input.Snapshot.Crops, domain.Crop{ID: "crop-ripe", Mature: true})
			},
			want: domain.ActionHarvestTarget, target: "crop-ripe",
		},
		{
			name: "empty can refills before new crop",
			mutate: func(input *ActionInput) {
				input.Snapshot.WateringCan.Water = 0
				input.Snapshot.WaterSources = []domain.WaterSource{{ID: "pond-1"}}
			},
			want: domain.ActionRefillCan, target: "pond-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validActionInput()
			tt.mutate(&input)
			var output domain.HighLevelAction
			if err := generator.GenerateJSON(context.Background(), actionSystemPrompt, input, &output); err != nil {
				t.Fatalf("GenerateJSON() error = %v", err)
			}
			if output.Kind != tt.want || output.TargetID != tt.target {
				t.Fatalf("action = %+v, want %s %s", output, tt.want, tt.target)
			}
		})
	}
}

func TestFixtureGeneratorReplansFullInventoryToPreferredChest(t *testing.T) {
	generator := NewFixtureGenerator()
	input := validActionInput()
	input.PlayerModel.PreferredChestID = "chest-1"
	input.Snapshot.Chests = []domain.Chest{{ID: "chest-1"}}
	input.Snapshot.Inventory.Items = []domain.InventoryItem{{ItemID: "(O)24", Name: "Parsnip", Quantity: 1}}
	input.Snapshot.Crops = []domain.Crop{{ID: "crop-ripe", Mature: true}}
	failed := domain.ActionResult{
		SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID,
		SnapshotVersion: input.Snapshot.SnapshotVersion, Status: domain.ActionFailed,
		ErrorCode: "inventory_full",
		Action:    domain.HighLevelAction{Kind: domain.ActionHarvestTarget, TargetID: "crop-ripe"},
	}
	var output domain.HighLevelAction

	if err := generator.GenerateJSON(context.Background(), replanSystemPrompt, ReplanInput{ActionInput: input, LastResult: failed}, &output); err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}
	if output.Kind != domain.ActionDepositItems || output.TargetID != "chest-1" {
		t.Fatalf("replanned action = %+v, want deposit_items chest-1", output)
	}
}

func TestFixtureGeneratorStopsWhenPreferredChestIsFull(t *testing.T) {
	generator := NewFixtureGenerator()
	input := validActionInput()
	input.PlayerModel.PreferredChestID = "chest-1"
	input.Snapshot.Chests = []domain.Chest{{ID: "chest-1"}}
	input.Snapshot.Inventory.Items = []domain.InventoryItem{{ItemID: "(O)24", Name: "Parsnip", Quantity: 1}}
	input.Snapshot.Crops = []domain.Crop{{ID: "crop-ripe", Mature: true}}
	failed := domain.ActionResult{
		SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID,
		SnapshotVersion: input.Snapshot.SnapshotVersion, Status: domain.ActionFailed,
		ErrorCode: "chest_full",
		Action:    domain.HighLevelAction{Kind: domain.ActionDepositItems, TargetID: "chest-1"},
	}
	var output domain.HighLevelAction

	if err := generator.GenerateJSON(context.Background(), replanSystemPrompt, ReplanInput{ActionInput: input, LastResult: failed}, &output); err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}
	if output.Kind != domain.ActionStopSession {
		t.Fatalf("replanned action = %+v, want stop_session", output)
	}
}
