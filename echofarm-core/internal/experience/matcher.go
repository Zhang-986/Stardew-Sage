package experience

import (
	"sort"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func Match(snapshot domain.WorldSnapshot, experiences []domain.PolicyExperience, limit int) []domain.PolicyExperience {
	if limit <= 0 {
		return nil
	}
	availableSignals := snapshotSignals(snapshot)
	availableTargets := snapshotTargets(snapshot)
	result := make([]domain.PolicyExperience, 0, min(limit, len(experiences)))
	for _, experience := range experiences {
		if experience.SaveID != snapshot.SaveID || !contextMatches(experience.Context, snapshot.Weather) ||
			!signalsMatch(experience.WhenSignals, availableSignals) || domain.ExperienceIsCooled(experience) {
			continue
		}
		if experience.PreferredTargetID != "" {
			if _, ok := availableTargets[experience.PreferredTargetID]; !ok {
				continue
			}
		}
		copy := experience
		copy.WhenSignals = append([]domain.SituationSignal(nil), experience.WhenSignals...)
		copy.EvidenceRefs = append([]string(nil), experience.EvidenceRefs...)
		result = append(result, copy)
	}
	sort.Slice(result, func(i, j int) bool {
		leftConfidence := domain.ExperienceRankingConfidence(result[i])
		rightConfidence := domain.ExperienceRankingConfidence(result[j])
		if leftConfidence != rightConfidence {
			return leftConfidence > rightConfidence
		}
		if result[i].ObservationCount != result[j].ObservationCount {
			return result[i].ObservationCount > result[j].ObservationCount
		}
		return result[i].ID < result[j].ID
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

func snapshotSignals(snapshot domain.WorldSnapshot) map[domain.SituationSignal]struct{} {
	signals := make(map[domain.SituationSignal]struct{})
	if snapshot.Inventory.FreeSlots == 0 {
		signals[domain.SignalInventoryFull] = struct{}{}
	}
	if len(snapshot.Inventory.Items) > 0 {
		signals[domain.SignalInventoryHasItems] = struct{}{}
	}
	if snapshot.WateringCan.Water == 0 {
		signals[domain.SignalCanEmpty] = struct{}{}
	}
	if snapshot.Weather == domain.WeatherRainy || snapshot.Weather == domain.WeatherStorm {
		signals[domain.SignalRaining] = struct{}{}
	}
	return signals
}

func snapshotTargets(snapshot domain.WorldSnapshot) map[string]struct{} {
	targets := make(map[string]struct{}, len(snapshot.Crops)+len(snapshot.WaterSources)+len(snapshot.Chests))
	for _, crop := range snapshot.Crops {
		targets[crop.ID] = struct{}{}
	}
	for _, source := range snapshot.WaterSources {
		targets[source.ID] = struct{}{}
	}
	for _, chest := range snapshot.Chests {
		targets[chest.ID] = struct{}{}
	}
	return targets
}

func contextMatches(context domain.TraitContext, weather domain.Weather) bool {
	return context == domain.TraitContextAny || string(context) == string(weather)
}

func signalsMatch(required []domain.SituationSignal, available map[domain.SituationSignal]struct{}) bool {
	for _, signal := range required {
		if _, ok := available[signal]; !ok {
			return false
		}
	}
	return true
}
