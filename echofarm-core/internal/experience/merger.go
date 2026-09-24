package experience

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

const (
	maxFailureInitialConfidence    = 0.65
	maxCorrectionInitialConfidence = 0.85
	maxExperienceEvidence          = 12
)

func Merge(existing []domain.PolicyExperience, saveID string, day int, source domain.ExperienceSource, observation domain.ExperienceObservation) ([]domain.PolicyExperience, domain.PolicyExperience, error) {
	if saveID == "" || day <= 0 {
		return nil, domain.PolicyExperience{}, errors.New("save ID and positive day are required")
	}
	if source != domain.ExperienceSourceFailure && source != domain.ExperienceSourceCorrection {
		return nil, domain.PolicyExperience{}, errors.New("unsupported experience source")
	}
	evidence := map[string]struct{}{observation.EvidenceRef: {}}
	targets := make(map[string]struct{})
	if observation.PreferredTargetID != "" {
		targets[observation.PreferredTargetID] = struct{}{}
	}
	if err := observation.Validate(evidence, targets); err != nil {
		return nil, domain.PolicyExperience{}, err
	}

	result := cloneExperiences(existing)
	for _, item := range result {
		if err := item.Validate(); err != nil {
			return nil, domain.PolicyExperience{}, err
		}
		if item.SaveID != saveID {
			return nil, domain.PolicyExperience{}, errors.New("existing experience belongs to another save")
		}
	}
	normalizedSignals := normalizeSignals(observation.WhenSignals)
	id := experienceID(observation, normalizedSignals)
	for i := range result {
		if result[i].ID != id {
			continue
		}
		experience := &result[i]
		experience.Confidence = 1 - (1-experience.Confidence)*(1-observation.Strength*0.5)
		experience.ObservationCount++
		experience.LastSeenDay = max(experience.LastSeenDay, day)
		experience.EvidenceRefs = boundedEvidence(experience.EvidenceRefs, observation.EvidenceRef)
		if source == domain.ExperienceSourceCorrection {
			experience.Source = source
		}
		if err := refreshEffectiveConfidence(experience); err != nil {
			return nil, domain.PolicyExperience{}, err
		}
		learned := *experience
		learned.WhenSignals = append([]domain.SituationSignal(nil), experience.WhenSignals...)
		learned.EvidenceRefs = append([]string(nil), experience.EvidenceRefs...)
		sortExperiences(result)
		return result, learned, nil
	}

	for i := range result {
		if sameScope(result[i], observation, normalizedSignals) {
			result[i].Confidence *= 0.8
			result[i].ContradictionCount++
			if err := refreshEffectiveConfidence(&result[i]); err != nil {
				return nil, domain.PolicyExperience{}, err
			}
		}
	}
	confidence := observation.Strength
	cap := maxFailureInitialConfidence
	if source == domain.ExperienceSourceCorrection {
		cap = maxCorrectionInitialConfidence
	}
	if confidence > cap {
		confidence = cap
	}
	learned := domain.PolicyExperience{
		ID: id, SaveID: saveID, Trigger: observation.Trigger, Context: observation.Context,
		WhenSignals: normalizedSignals, AvoidAction: observation.AvoidAction, PreferAction: observation.PreferAction,
		PreferredTargetID: observation.PreferredTargetID, Summary: observation.Summary,
		Confidence: confidence, EffectiveConfidence: confidence,
		ObservationCount: 1, FirstSeenDay: day, LastSeenDay: day,
		EvidenceRefs: []string{observation.EvidenceRef}, Source: source,
	}
	result = append(result, learned)
	sortExperiences(result)
	return result, learned, nil
}

func refreshEffectiveConfidence(experience *domain.PolicyExperience) error {
	effective, err := domain.EffectiveExperienceConfidence(experience.Confidence, experience.SuccessCount, experience.FailureCount)
	if err != nil {
		return err
	}
	experience.EffectiveConfidence = effective
	return nil
}

func experienceID(observation domain.ExperienceObservation, signals []domain.SituationSignal) string {
	parts := []string{
		string(observation.Trigger), string(observation.Context),
		string(observation.AvoidAction), string(observation.PreferAction), observation.PreferredTargetID,
	}
	for _, signal := range signals {
		parts = append(parts, string(signal))
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "exp-" + hex.EncodeToString(sum[:8])
}

func sameScope(existing domain.PolicyExperience, observation domain.ExperienceObservation, signals []domain.SituationSignal) bool {
	return existing.Trigger == observation.Trigger && existing.Context == observation.Context &&
		existing.AvoidAction == observation.AvoidAction && equalSignals(existing.WhenSignals, signals) &&
		(existing.PreferAction != observation.PreferAction || existing.PreferredTargetID != observation.PreferredTargetID)
}

func normalizeSignals(signals []domain.SituationSignal) []domain.SituationSignal {
	result := append([]domain.SituationSignal(nil), signals...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func equalSignals(left, right []domain.SituationSignal) bool {
	left = normalizeSignals(left)
	right = normalizeSignals(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func boundedEvidence(existing []string, added string) []string {
	refs := append(append([]string(nil), existing...), added)
	if len(refs) > maxExperienceEvidence {
		refs = refs[len(refs)-maxExperienceEvidence:]
	}
	return refs
}

func cloneExperiences(existing []domain.PolicyExperience) []domain.PolicyExperience {
	result := append([]domain.PolicyExperience(nil), existing...)
	for i := range result {
		result[i].WhenSignals = append([]domain.SituationSignal(nil), existing[i].WhenSignals...)
		result[i].EvidenceRefs = append([]string(nil), existing[i].EvidenceRefs...)
	}
	return result
}

func sortExperiences(experiences []domain.PolicyExperience) {
	sort.Slice(experiences, func(i, j int) bool { return experiences[i].ID < experiences[j].ID })
}
