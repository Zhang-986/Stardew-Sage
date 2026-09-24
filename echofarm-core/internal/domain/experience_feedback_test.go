package domain

import (
	"math"
	"testing"
)

func TestClassifyExperienceFeedbackUsesNarrowCausalCatalog(t *testing.T) {
	tests := []struct {
		name      string
		action    ActionKind
		status    ActionStatus
		errorCode string
		want      ExperienceFeedbackOutcome
	}{
		{name: "success", action: ActionHarvestTarget, status: ActionSucceeded, want: ExperienceFeedbackSucceeded},
		{name: "water without water", action: ActionWaterTarget, status: ActionFailed, errorCode: "out_of_water", want: ExperienceFeedbackContradicted},
		{name: "harvest with full inventory", action: ActionHarvestTarget, status: ActionFailed, errorCode: "inventory_full", want: ExperienceFeedbackContradicted},
		{name: "full chest", action: ActionDepositItems, status: ActionFailed, errorCode: "chest_full", want: ExperienceFeedbackContradicted},
		{name: "empty deposit", action: ActionDepositItems, status: ActionFailed, errorCode: "inventory_empty", want: ExperienceFeedbackContradicted},
		{name: "blocked route", action: ActionHarvestTarget, status: ActionFailed, errorCode: "path_blocked", want: ExperienceFeedbackNeutral},
		{name: "changed target", action: ActionHarvestTarget, status: ActionFailed, errorCode: "target_changed", want: ExperienceFeedbackNeutral},
		{name: "unsupported crop", action: ActionHarvestTarget, status: ActionFailed, errorCode: "unsupported_crop", want: ExperienceFeedbackNeutral},
		{name: "unrelated precondition", action: ActionDepositItems, status: ActionFailed, errorCode: "out_of_water", want: ExperienceFeedbackNeutral},
		{name: "unknown failure", action: ActionHarvestTarget, status: ActionFailed, errorCode: "provider-secret-text", want: ExperienceFeedbackNeutral},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := feedbackResult(test.action, test.status, test.errorCode)
			if got := ClassifyExperienceFeedback(result); got != test.want {
				t.Fatalf("ClassifyExperienceFeedback() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeExperienceFeedbackErrorRejectsArbitraryText(t *testing.T) {
	known := feedbackResult(ActionHarvestTarget, ActionFailed, "inventory_full")
	if got := NormalizeExperienceFeedbackError(known); got != "inventory_full" {
		t.Fatalf("known error = %q", got)
	}
	unknown := feedbackResult(ActionHarvestTarget, ActionFailed, "api-key-secret")
	if got := NormalizeExperienceFeedbackError(unknown); got != "other" {
		t.Fatalf("unknown error = %q, want other", got)
	}
	succeeded := feedbackResult(ActionHarvestTarget, ActionSucceeded, "")
	if got := NormalizeExperienceFeedbackError(succeeded); got != "" {
		t.Fatalf("successful error = %q, want empty", got)
	}
}

func TestEffectiveExperienceConfidenceUsesOrderIndependentPrior(t *testing.T) {
	tests := []struct {
		base             float64
		success, failure int
		want             float64
	}{
		{base: 0.65, want: 0.65},
		{base: 0.65, success: 1, want: 0.72},
		{base: 0.65, failure: 1, want: 0.52},
		{base: 0.65, success: 3, failure: 2, want: 5.6 / 9},
	}
	for _, test := range tests {
		got, err := EffectiveExperienceConfidence(test.base, test.success, test.failure)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got-test.want) > 1e-9 {
			t.Fatalf("EffectiveExperienceConfidence(%v,%d,%d) = %v, want %v", test.base, test.success, test.failure, got, test.want)
		}
	}
	if _, err := EffectiveExperienceConfidence(math.NaN(), 0, 0); err == nil {
		t.Fatal("NaN base error = nil")
	}
	if _, err := EffectiveExperienceConfidence(0.5, -1, 0); err == nil {
		t.Fatal("negative success count error = nil")
	}
}

func TestExperienceCoolingRequiresRepeatedContradiction(t *testing.T) {
	experience := PolicyExperience{Confidence: 0.65, EffectiveConfidence: 0.39, FailureCount: 3}
	if !ExperienceIsCooled(experience) {
		t.Fatal("ExperienceIsCooled() = false, want true")
	}
	experience.FailureCount = 2
	if ExperienceIsCooled(experience) {
		t.Fatal("two failures cooled experience")
	}
	experience.FailureCount = 3
	experience.EffectiveConfidence = 0.40
	if ExperienceIsCooled(experience) {
		t.Fatal("threshold confidence cooled experience")
	}
	legacy := PolicyExperience{Confidence: 0.30, FailureCount: 3}
	if !ExperienceIsCooled(legacy) {
		t.Fatal("legacy experience should fall back to semantic confidence")
	}
}

func feedbackResult(kind ActionKind, status ActionStatus, errorCode string) ActionResult {
	action := HighLevelAction{
		SaveID: "farm-1", SessionID: "day-2", SnapshotVersion: 7,
		Kind: kind, TargetID: "target-1", Reason: "feedback test",
	}
	return ActionResult{
		SaveID: action.SaveID, SessionID: action.SessionID, SnapshotVersion: action.SnapshotVersion,
		Action: action, Status: status, ErrorCode: errorCode,
	}
}
