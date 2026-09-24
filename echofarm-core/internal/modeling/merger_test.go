package modeling

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestMergeAddsAndStrengthensTrait(t *testing.T) {
	firstDemo := demonstration("demo-1", 1, domain.WeatherSunny, "water-1")
	observation := domain.TraitObservation{
		Key: domain.PreferenceTaskOrder, Value: "watering,harvesting,depositing",
		Context: domain.TraitContextSunny, SupportingEventIDs: []string{"water-1"}, Strength: 0.7,
	}

	first, firstChange, err := Merge(nil, firstDemo, []domain.TraitObservation{observation})
	if err != nil {
		t.Fatal(err)
	}
	if first.Revision != 1 || first.LearnedThroughDay != 1 {
		t.Fatalf("first model revision/day = %d/%d", first.Revision, first.LearnedThroughDay)
	}
	if len(first.Traits) != 1 || first.Traits[0].Confidence != 0.7 || first.Traits[0].ObservationCount != 1 {
		t.Fatalf("first trait = %+v", first.Traits)
	}
	if !reflect.DeepEqual(first.Traits[0].EvidenceRefs, []string{"demo-1:water-1"}) {
		t.Fatalf("evidence refs = %v", first.Traits[0].EvidenceRefs)
	}
	if firstChange.Kind != domain.LearningChangeAdded {
		t.Fatalf("first change = %+v", firstChange)
	}

	secondDemo := demonstration("demo-2", 2, domain.WeatherSunny, "water-2")
	observation.SupportingEventIDs = []string{"water-2"}
	second, secondChange, err := Merge(&first, secondDemo, []domain.TraitObservation{observation})
	if err != nil {
		t.Fatal(err)
	}
	wantConfidence := 1 - (1-0.7)*(1-0.7*0.5)
	if math.Abs(second.Traits[0].Confidence-wantConfidence) > 1e-9 {
		t.Fatalf("confidence = %f, want %f", second.Traits[0].Confidence, wantConfidence)
	}
	if second.Traits[0].ObservationCount != 2 || secondChange.Kind != domain.LearningChangeStrengthened {
		t.Fatalf("second trait/change = %+v / %+v", second.Traits[0], secondChange)
	}
	if first.Traits[0].ObservationCount != 1 {
		t.Fatalf("Merge mutated input model: %+v", first.Traits[0])
	}
}

func TestMergeDecaysContradictionOnlyInSameContext(t *testing.T) {
	existing := domain.PlayerModel{
		SaveID: "farm-1", Revision: 2, LearnedThroughDay: 2, EnergyReserve: 40,
		Traits: []domain.TraitMemory{
			{Key: domain.PreferenceTaskOrder, Value: "watering,harvesting", Context: domain.TraitContextSunny, Confidence: 0.8, ObservationCount: 2, FirstSeenDay: 1, LastSeenDay: 2, EvidenceRefs: []string{"demo-1:water-1"}},
			{Key: domain.PreferenceTaskOrder, Value: "harvesting", Context: domain.TraitContextRainy, Confidence: 0.6, ObservationCount: 1, FirstSeenDay: 2, LastSeenDay: 2, EvidenceRefs: []string{"demo-2:harvest-1"}},
		},
	}
	demo := demonstration("demo-3", 3, domain.WeatherSunny, "harvest-2")
	observation := domain.TraitObservation{
		Key: domain.PreferenceTaskOrder, Value: "harvesting,watering",
		Context: domain.TraitContextSunny, SupportingEventIDs: []string{"harvest-2"}, Strength: 0.7,
	}

	got, _, err := Merge(&existing, demo, []domain.TraitObservation{observation})
	if err != nil {
		t.Fatal(err)
	}
	sunnyOld := findTrait(t, got.Traits, domain.PreferenceTaskOrder, "watering,harvesting", domain.TraitContextSunny)
	if sunnyOld.ContradictionCount != 1 || math.Abs(sunnyOld.Confidence-0.64) > 1e-9 {
		t.Fatalf("sunny old trait = %+v", sunnyOld)
	}
	rainy := findTrait(t, got.Traits, domain.PreferenceTaskOrder, "harvesting", domain.TraitContextRainy)
	if rainy.ContradictionCount != 0 || rainy.Confidence != 0.6 {
		t.Fatalf("rainy trait was incorrectly contradicted: %+v", rainy)
	}
}

func TestMergeCapsEvidenceAndProjectsLegacyFields(t *testing.T) {
	refs := make([]string, 12)
	for i := range refs {
		refs[i] = fmt.Sprintf("old:%02d", i)
	}
	existing := domain.PlayerModel{
		SaveID: "farm-1", Revision: 12, EnergyReserve: 40,
		Traits: []domain.TraitMemory{{
			Key: domain.PreferencePreferredChest, Value: "east-chest", Context: domain.TraitContextAny,
			Confidence: 0.9, ObservationCount: 12, FirstSeenDay: 1, LastSeenDay: 12, EvidenceRefs: refs,
		}},
	}
	demo := demonstration("demo-13", 13, domain.WeatherSunny, "deposit-13")
	observation := domain.TraitObservation{
		Key: domain.PreferencePreferredChest, Value: "east-chest", Context: domain.TraitContextAny,
		SupportingEventIDs: []string{"deposit-13"}, Strength: 0.8,
	}

	got, _, err := Merge(&existing, demo, []domain.TraitObservation{observation})
	if err != nil {
		t.Fatal(err)
	}
	trait := got.Traits[0]
	if len(trait.EvidenceRefs) != 12 || trait.EvidenceRefs[0] != "old:01" || trait.EvidenceRefs[11] != "demo-13:deposit-13" {
		t.Fatalf("bounded evidence refs = %v", trait.EvidenceRefs)
	}
	if got.PreferredChestID != "east-chest" {
		t.Fatalf("preferred chest = %q", got.PreferredChestID)
	}
}

func TestMergeRejectsDifferentSave(t *testing.T) {
	existing := domain.PlayerModel{SaveID: "farm-a", Revision: 1, EnergyReserve: 40}
	demo := demonstration("demo-2", 2, domain.WeatherSunny, "water-2")
	demo.SaveID = "farm-b"
	observation := domain.TraitObservation{
		Key: domain.PreferenceTaskOrder, Value: "watering", Context: domain.TraitContextSunny,
		SupportingEventIDs: []string{"water-2"}, Strength: 0.6,
	}
	if _, _, err := Merge(&existing, demo, []domain.TraitObservation{observation}); err == nil {
		t.Fatal("Merge() error = nil, want save mismatch")
	}
}

func TestMergeRejectsMultipleValuesForSameTraitContext(t *testing.T) {
	demo := demonstration("demo-1", 1, domain.WeatherSunny, "water-1")
	observations := []domain.TraitObservation{
		{
			Key: domain.PreferenceTaskOrder, Value: "watering,harvesting", Context: domain.TraitContextSunny,
			SupportingEventIDs: []string{"water-1"}, Strength: 0.7,
		},
		{
			Key: domain.PreferenceTaskOrder, Value: "harvesting,watering", Context: domain.TraitContextSunny,
			SupportingEventIDs: []string{"water-1"}, Strength: 0.7,
		},
	}

	if _, _, err := Merge(nil, demo, observations); err == nil {
		t.Fatal("Merge() error = nil, want ambiguous trait observation error")
	}
}

func demonstration(id string, day int, weather domain.Weather, eventID string) domain.Demonstration {
	return domain.Demonstration{
		ID: id, SaveID: "farm-1", SessionID: "teaching-" + id,
		Day: day, Weather: weather, StartedAt: 1, EndedAt: 2,
		Events: []domain.DemonstrationEvent{{
			ID: eventID, Kind: domain.EventWater, Tick: 1,
			Position: domain.Position{X: 1, Y: 1}, TargetID: "crop-1", Success: true,
		}},
	}
}

func findTrait(t *testing.T, traits []domain.TraitMemory, key domain.PreferenceKey, value string, context domain.TraitContext) domain.TraitMemory {
	t.Helper()
	for _, trait := range traits {
		if trait.Key == key && trait.Value == value && trait.Context == context {
			return trait
		}
	}
	t.Fatalf("trait %s/%s/%s not found in %+v", key, value, context, traits)
	return domain.TraitMemory{}
}
