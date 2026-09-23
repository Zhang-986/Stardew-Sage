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
