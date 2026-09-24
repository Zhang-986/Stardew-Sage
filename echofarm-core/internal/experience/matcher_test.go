package experience

import (
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestMatchSelectsContextualTopThreeExperiences(t *testing.T) {
	snapshot := domain.WorldSnapshot{
		SaveID: "farm-1", SessionID: "echo-5", SnapshotVersion: 1, Day: 5, TimeOfDay: 700,
		Weather: domain.WeatherSunny, Location: "Farm", Energy: 200, MaxEnergy: 270,
		Inventory:   domain.InventorySummary{FreeSlots: 0, Items: []domain.InventoryItem{{ItemID: "parsnip", Name: "Parsnip", Quantity: 1}}},
		WateringCan: domain.ToolState{Name: "Watering Can", Water: 10, Capacity: 40},
		Crops:       []domain.Crop{{ID: "crop-1", Mature: true}},
		Chests:      []domain.Chest{{ID: "chest-east"}, {ID: "chest-west"}},
	}
	experiences := []domain.PolicyExperience{
		policyExperience("exp-low", domain.TraitContextSunny, "chest-east", 0.5, 4),
		policyExperience("exp-high", domain.TraitContextSunny, "chest-west", 0.9, 2),
		policyExperience("exp-any", domain.TraitContextAny, "chest-east", 0.8, 3),
		policyExperience("exp-fourth", domain.TraitContextSunny, "chest-east", 0.4, 5),
		policyExperience("exp-rain", domain.TraitContextRainy, "chest-east", 1, 9),
	}

	got := Match(snapshot, experiences, 3)
	if len(got) != 3 || got[0].ID != "exp-high" || got[1].ID != "exp-any" || got[2].ID != "exp-low" {
		t.Fatalf("Match() = %+v", got)
	}
}

func TestMatchRequiresAllSituationSignalsAndExistingPreferredTarget(t *testing.T) {
	snapshot := domain.WorldSnapshot{
		SaveID: "farm-1", SessionID: "echo-5", SnapshotVersion: 1, Day: 5, TimeOfDay: 700,
		Weather: domain.WeatherSunny, Location: "Farm", Energy: 200, MaxEnergy: 270,
		Inventory:   domain.InventorySummary{FreeSlots: 1},
		WateringCan: domain.ToolState{Name: "Watering Can", Water: 10, Capacity: 40},
		Chests:      []domain.Chest{{ID: "chest-east"}},
	}
	experiences := []domain.PolicyExperience{
		policyExperience("exp-full", domain.TraitContextSunny, "chest-east", 0.9, 2),
		policyExperience("exp-missing-target", domain.TraitContextSunny, "chest-west", 0.8, 2),
	}

	if got := Match(snapshot, experiences, 3); len(got) != 0 {
		t.Fatalf("Match() = %+v, want no applicable experience", got)
	}
}

func policyExperience(id string, context domain.TraitContext, targetID string, confidence float64, observations int) domain.PolicyExperience {
	return domain.PolicyExperience{
		ID: id, SaveID: "farm-1", Trigger: domain.ExperienceInventoryFull, Context: context,
		WhenSignals: []domain.SituationSignal{domain.SignalInventoryFull, domain.SignalInventoryHasItems},
		AvoidAction: domain.ActionHarvestTarget, PreferAction: domain.ActionDepositItems,
		PreferredTargetID: targetID, Summary: "deposit first", Confidence: confidence,
		ObservationCount: observations, FirstSeenDay: 1, LastSeenDay: 4,
		EvidenceRefs: []string{"decision:echo-1:1"}, Source: domain.ExperienceSourceFailure,
	}
}
