package experience

import (
	"math"
	"reflect"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestMergeCapsInitialConfidenceByEvidenceSource(t *testing.T) {
	observation := experienceObservation("chest-east", "decision:echo-1:1")

	failureList, failure, err := Merge(nil, "farm-1", 2, domain.ExperienceSourceFailure, observation)
	if err != nil {
		t.Fatal(err)
	}
	correctionObservation := observation
	correctionObservation.Trigger = domain.ExperiencePlayerCorrection
	correctionObservation.EvidenceRef = "correction-1"
	correctionList, correction, err := Merge(nil, "farm-1", 2, domain.ExperienceSourceCorrection, correctionObservation)
	if err != nil {
		t.Fatal(err)
	}
	if len(failureList) != 1 || failure.Confidence != 0.65 {
		t.Fatalf("failure experience = %+v", failure)
	}
	if len(correctionList) != 1 || correction.Confidence != 0.85 {
		t.Fatalf("correction experience = %+v", correction)
	}
}

func TestMergeStrengthensMatchingExperienceWithoutMutatingInput(t *testing.T) {
	observation := experienceObservation("chest-east", "decision:echo-1:1")
	existing, _, err := Merge(nil, "farm-1", 2, domain.ExperienceSourceFailure, observation)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]domain.PolicyExperience(nil), existing...)
	before[0].WhenSignals = append([]domain.SituationSignal(nil), existing[0].WhenSignals...)
	before[0].EvidenceRefs = append([]string(nil), existing[0].EvidenceRefs...)
	second := observation
	second.EvidenceRef = "decision:echo-2:1"

	updated, learned, err := Merge(existing, "farm-1", 3, domain.ExperienceSourceFailure, second)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 1 || learned.ObservationCount != 2 || learned.Confidence <= existing[0].Confidence {
		t.Fatalf("learned experience = %+v", learned)
	}
	if !reflect.DeepEqual(existing, before) {
		t.Fatalf("Merge mutated input: before=%+v after=%+v", before, existing)
	}
}

func TestMergeReprojectsEffectivenessAfterSemanticEvidenceChanges(t *testing.T) {
	observation := experienceObservation("chest-east", "decision:echo-1:1")
	existing, _, err := Merge(nil, "farm-1", 2, domain.ExperienceSourceFailure, observation)
	if err != nil {
		t.Fatal(err)
	}
	existing[0].FailureCount = 1
	existing[0].EffectiveConfidence = 0.52
	second := observation
	second.EvidenceRef = "decision:echo-2:1"

	_, learned, err := Merge(existing, "farm-1", 3, domain.ExperienceSourceFailure, second)
	if err != nil {
		t.Fatal(err)
	}
	want, err := domain.EffectiveExperienceConfidence(learned.Confidence, learned.SuccessCount, learned.FailureCount)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(learned.EffectiveConfidence-want) > 1e-9 {
		t.Fatalf("effective confidence = %v, want %v after semantic update", learned.EffectiveConfidence, want)
	}
}

func TestMergeDecaysConflictingExperienceInSameScope(t *testing.T) {
	firstObservation := experienceObservation("chest-east", "correction-1")
	firstObservation.Trigger = domain.ExperiencePlayerCorrection
	existing, first, err := Merge(nil, "farm-1", 2, domain.ExperienceSourceCorrection, firstObservation)
	if err != nil {
		t.Fatal(err)
	}
	conflict := experienceObservation("chest-west", "correction-2")
	conflict.Trigger = domain.ExperiencePlayerCorrection

	updated, learned, err := Merge(existing, "farm-1", 3, domain.ExperienceSourceCorrection, conflict)
	if err != nil {
		t.Fatal(err)
	}
	if learned.ID == first.ID || len(updated) != 2 {
		t.Fatalf("updated experiences = %+v", updated)
	}
	for _, candidate := range updated {
		if candidate.ID == first.ID && (candidate.ContradictionCount != 1 || candidate.Confidence >= first.Confidence) {
			t.Fatalf("conflicting experience was not decayed: %+v", candidate)
		}
	}
}

func TestMergeReturnsTheStrengthenedExperienceAfterSorting(t *testing.T) {
	observation := experienceObservation("chest-east", "decision:echo-1:1")
	existing, first, err := Merge(nil, "farm-1", 2, domain.ExperienceSourceFailure, observation)
	if err != nil {
		t.Fatal(err)
	}
	existing = append(existing, policyExperience("aaa-before-learned", domain.TraitContextRainy, "chest-west", 0.9, 3))
	second := observation
	second.EvidenceRef = "decision:echo-2:1"

	_, learned, err := Merge(existing, "farm-1", 3, domain.ExperienceSourceFailure, second)
	if err != nil {
		t.Fatal(err)
	}
	if learned.ID != first.ID || learned.ObservationCount != 2 {
		t.Fatalf("Merge() learned = %+v, want strengthened %q", learned, first.ID)
	}
}

func experienceObservation(targetID, evidence string) domain.ExperienceObservation {
	return domain.ExperienceObservation{
		Trigger: domain.ExperienceInventoryFull, Context: domain.TraitContextSunny,
		WhenSignals: []domain.SituationSignal{domain.SignalInventoryHasItems, domain.SignalInventoryFull},
		AvoidAction: domain.ActionHarvestTarget, PreferAction: domain.ActionDepositItems,
		PreferredTargetID: targetID, Summary: "deposit before harvesting", EvidenceRef: evidence, Strength: 0.9,
	}
}
