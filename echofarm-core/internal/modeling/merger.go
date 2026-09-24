package modeling

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

const (
	maxInitialConfidence = 0.75
	maxEvidenceRefs      = 12
	defaultEnergyReserve = 40
)

func Merge(existing *domain.PlayerModel, demonstration domain.Demonstration, observations []domain.TraitObservation) (domain.PlayerModel, domain.LearningChange, error) {
	if err := demonstration.Validate(); err != nil {
		return domain.PlayerModel{}, domain.LearningChange{}, fmt.Errorf("validate demonstration: %w", err)
	}
	if len(observations) == 0 {
		return domain.PlayerModel{}, domain.LearningChange{}, errors.New("trait observations are required")
	}
	if existing != nil && existing.SaveID != demonstration.SaveID {
		return domain.PlayerModel{}, domain.LearningChange{}, errors.New("existing player model belongs to another save")
	}

	evidence := make(map[string]struct{}, len(demonstration.Events))
	for _, event := range demonstration.Events {
		evidence[event.ID] = struct{}{}
	}
	seen := make(map[string]string, len(observations))
	for _, observation := range observations {
		if err := observation.Validate(evidence); err != nil {
			return domain.PlayerModel{}, domain.LearningChange{}, err
		}
		slot := traitSlotIdentity(observation.Key, observation.Context)
		if previous, duplicate := seen[slot]; duplicate {
			return domain.PlayerModel{}, domain.LearningChange{}, fmt.Errorf(
				"ambiguous trait observations for key %q in context %q: %q and %q",
				observation.Key, observation.Context, previous, observation.Value,
			)
		}
		seen[slot] = observation.Value
	}

	model := cloneModel(existing, demonstration.SaveID)
	model.Revision++
	if demonstration.Day > model.LearnedThroughDay {
		model.LearnedThroughDay = demonstration.Day
	}
	change := domain.LearningChange{
		ModelRevision: model.Revision,
		Kind:          domain.LearningChangeUnchanged,
		Summary:       "no stable player trait changed",
	}
	for _, observation := range observations {
		for i := range model.Traits {
			trait := &model.Traits[i]
			if trait.Key == observation.Key && trait.Context == observation.Context && trait.Value != observation.Value {
				trait.Confidence *= 0.8
				trait.ContradictionCount++
			}
		}

		index := findTraitIndex(model.Traits, observation.Key, observation.Value, observation.Context)
		refs := scopedEvidence(demonstration.ID, observation.SupportingEventIDs)
		if index < 0 {
			confidence := observation.Strength
			if confidence > maxInitialConfidence {
				confidence = maxInitialConfidence
			}
			model.Traits = append(model.Traits, domain.TraitMemory{
				Key: observation.Key, Value: observation.Value, Context: observation.Context,
				Confidence: confidence, ObservationCount: 1,
				FirstSeenDay: demonstration.Day, LastSeenDay: demonstration.Day,
				EvidenceRefs: boundedEvidence(nil, refs),
			})
			if change.Kind == domain.LearningChangeUnchanged {
				change = domain.LearningChange{
					ModelRevision: model.Revision, Kind: domain.LearningChangeAdded,
					Key: observation.Key, Value: observation.Value, Confidence: confidence,
					Summary: fmt.Sprintf("learned %s=%s", observation.Key, observation.Value),
				}
			}
			continue
		}

		trait := &model.Traits[index]
		previous := trait.Confidence
		trait.Confidence = 1 - (1-trait.Confidence)*(1-observation.Strength*0.5)
		trait.ObservationCount++
		if demonstration.Day > trait.LastSeenDay {
			trait.LastSeenDay = demonstration.Day
		}
		trait.EvidenceRefs = boundedEvidence(trait.EvidenceRefs, refs)
		if change.Kind == domain.LearningChangeUnchanged {
			change = domain.LearningChange{
				ModelRevision: model.Revision, Kind: domain.LearningChangeStrengthened,
				Key: observation.Key, Value: observation.Value,
				PreviousConfidence: previous, Confidence: trait.Confidence,
				Summary: fmt.Sprintf("strengthened %s=%s", observation.Key, observation.Value),
			}
		}
	}

	sort.Slice(model.Traits, func(i, j int) bool {
		left, right := model.Traits[i], model.Traits[j]
		if left.Key != right.Key {
			return left.Key < right.Key
		}
		if left.Context != right.Context {
			return left.Context < right.Context
		}
		return left.Value < right.Value
	})
	projectLegacyFields(&model)
	if err := model.Validate(); err != nil {
		return domain.PlayerModel{}, domain.LearningChange{}, fmt.Errorf("validate merged player model: %w", err)
	}
	return model, change, nil
}

func cloneModel(existing *domain.PlayerModel, saveID string) domain.PlayerModel {
	if existing == nil {
		return domain.PlayerModel{SaveID: saveID, EnergyReserve: defaultEnergyReserve}
	}
	model := *existing
	model.CommonTaskOrder = append([]domain.BehaviorKind(nil), existing.CommonTaskOrder...)
	model.Preferences = append([]domain.ObservedPreference(nil), existing.Preferences...)
	for i := range model.Preferences {
		model.Preferences[i].EvidenceEventIDs = append([]string(nil), existing.Preferences[i].EvidenceEventIDs...)
	}
	model.Traits = append([]domain.TraitMemory(nil), existing.Traits...)
	for i := range model.Traits {
		model.Traits[i].EvidenceRefs = append([]string(nil), existing.Traits[i].EvidenceRefs...)
	}
	return model
}

func findTraitIndex(traits []domain.TraitMemory, key domain.PreferenceKey, value string, context domain.TraitContext) int {
	for i, trait := range traits {
		if trait.Key == key && trait.Value == value && trait.Context == context {
			return i
		}
	}
	return -1
}

func traitSlotIdentity(key domain.PreferenceKey, context domain.TraitContext) string {
	return string(key) + "\x00" + string(context)
}

func scopedEvidence(demonstrationID string, eventIDs []string) []string {
	result := make([]string, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		result = append(result, demonstrationID+":"+eventID)
	}
	return result
}

func boundedEvidence(existing, added []string) []string {
	refs := append(append([]string(nil), existing...), added...)
	if len(refs) > maxEvidenceRefs {
		refs = refs[len(refs)-maxEvidenceRefs:]
	}
	return refs
}

func projectLegacyFields(model *domain.PlayerModel) {
	model.Preferences = nil
	for _, key := range []domain.PreferenceKey{
		domain.PreferenceTaskOrder,
		domain.PreferencePreferredChest,
		domain.PreferenceEnergyReserve,
		domain.PreferenceRouteStyle,
	} {
		trait, ok := strongestTrait(model.Traits, key)
		if !ok {
			continue
		}
		model.Preferences = append(model.Preferences, domain.ObservedPreference{
			Key: trait.Key, Value: trait.Value,
			EvidenceEventIDs: append([]string(nil), trait.EvidenceRefs...),
			ObservationCount: trait.ObservationCount, Confidence: trait.Confidence,
		})
		switch key {
		case domain.PreferenceTaskOrder:
			model.CommonTaskOrder = parseBehaviorOrder(trait.Value)
		case domain.PreferencePreferredChest:
			model.PreferredChestID = trait.Value
		case domain.PreferenceEnergyReserve:
			if value, err := strconv.Atoi(trait.Value); err == nil && value >= 0 {
				model.EnergyReserve = value
			}
		case domain.PreferenceRouteStyle:
			model.RouteStyle = trait.Value
		}
	}
}

func strongestTrait(traits []domain.TraitMemory, key domain.PreferenceKey) (domain.TraitMemory, bool) {
	var selected domain.TraitMemory
	found := false
	for _, trait := range traits {
		if trait.Key != key {
			continue
		}
		if !found || trait.Confidence > selected.Confidence ||
			(trait.Confidence == selected.Confidence && trait.ObservationCount > selected.ObservationCount) {
			selected, found = trait, true
		}
	}
	return selected, found
}

func parseBehaviorOrder(value string) []domain.BehaviorKind {
	parts := strings.Split(value, ",")
	result := make([]domain.BehaviorKind, 0, len(parts))
	for _, part := range parts {
		kind := domain.BehaviorKind(strings.TrimSpace(part))
		switch kind {
		case domain.BehaviorWatering, domain.BehaviorRefilling, domain.BehaviorHarvesting, domain.BehaviorDepositing:
			result = append(result, kind)
		}
	}
	return result
}
