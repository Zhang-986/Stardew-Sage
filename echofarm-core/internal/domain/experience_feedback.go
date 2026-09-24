package domain

import (
	"errors"
	"math"
)

type ExperienceFeedbackOutcome string

const (
	ExperienceFeedbackSucceeded    ExperienceFeedbackOutcome = "succeeded"
	ExperienceFeedbackContradicted ExperienceFeedbackOutcome = "contradicted"
	ExperienceFeedbackNeutral      ExperienceFeedbackOutcome = "neutral"
)

const experienceFeedbackPriorWeight = 4.0

func ClassifyExperienceFeedback(result ActionResult) ExperienceFeedbackOutcome {
	if result.Status == ActionSucceeded {
		return ExperienceFeedbackSucceeded
	}
	switch {
	case result.Action.Kind == ActionWaterTarget && result.ErrorCode == "out_of_water":
		return ExperienceFeedbackContradicted
	case result.Action.Kind == ActionHarvestTarget && result.ErrorCode == "inventory_full":
		return ExperienceFeedbackContradicted
	case result.Action.Kind == ActionDepositItems && (result.ErrorCode == "chest_full" || result.ErrorCode == "inventory_empty"):
		return ExperienceFeedbackContradicted
	default:
		return ExperienceFeedbackNeutral
	}
}

func NormalizeExperienceFeedbackError(result ActionResult) string {
	if result.Status == ActionSucceeded {
		return ""
	}
	switch result.ErrorCode {
	case "inventory_full", "out_of_water", "chest_full", "inventory_empty", "path_blocked", "target_changed", "unsupported_crop":
		return result.ErrorCode
	default:
		return "other"
	}
}

func EffectiveExperienceConfidence(base float64, successCount, failureCount int) (float64, error) {
	if math.IsNaN(base) || math.IsInf(base, 0) || base < 0 || base > 1 {
		return 0, errors.New("experience confidence must be between zero and one")
	}
	if successCount < 0 || failureCount < 0 {
		return 0, errors.New("experience feedback counts cannot be negative")
	}
	return (base*experienceFeedbackPriorWeight + float64(successCount)) /
		(experienceFeedbackPriorWeight + float64(successCount+failureCount)), nil
}

func ExperienceRankingConfidence(experience PolicyExperience) float64 {
	if experience.EffectiveConfidence > 0 || experience.SuccessCount+experience.FailureCount+experience.NeutralCount > 0 {
		return experience.EffectiveConfidence
	}
	return experience.Confidence
}

func ExperienceIsCooled(experience PolicyExperience) bool {
	return experience.FailureCount >= 3 && ExperienceRankingConfidence(experience) < 0.40
}
