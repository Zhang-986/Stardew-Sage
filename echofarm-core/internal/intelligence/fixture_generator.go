package intelligence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

// FixtureGenerator provides deterministic, offline responses for contract and
// smoke demonstrations. It is not a fallback when the configured model fails.
type FixtureGenerator struct{}

func NewFixtureGenerator() *FixtureGenerator {
	return &FixtureGenerator{}
}

func (g *FixtureGenerator) GenerateJSON(_ context.Context, _ string, input any, output any) error {
	switch target := output.(type) {
	case *LearningInference:
		learningInput, ok := input.(LearningInput)
		if !ok {
			return errors.New("fixture learning generator received unexpected input")
		}
		*target = fixtureLearningInference(learningInput)
		return nil
	case *domain.HighLevelAction:
		switch actionInput := input.(type) {
		case ActionInput:
			*target = fixtureAction(actionInput)
			return nil
		case ReplanInput:
			*target = fixtureReplan(actionInput)
			return nil
		default:
			return errors.New("fixture action generator received unexpected input")
		}
	case *domain.ActionProposal:
		switch actionInput := input.(type) {
		case ActionInput:
			*target = fixtureProposal(actionInput, fixtureAction(actionInput))
			return nil
		case ReplanInput:
			*target = fixtureProposal(actionInput.ActionInput, fixtureReplan(actionInput))
			return nil
		default:
			return errors.New("fixture proposal generator received unexpected input")
		}
	case *domain.ExperienceObservation:
		reflectionInput, ok := input.(ReflectionInput)
		if !ok {
			return errors.New("fixture reflection generator received unexpected input")
		}
		observation, err := fixtureReflection(reflectionInput)
		if err != nil {
			return err
		}
		*target = observation
		return nil
	case *IntentInference:
		intentInput, ok := input.(IntentInput)
		if !ok {
			return errors.New("fixture intent generator received unexpected input")
		}
		*target = fixtureIntent(intentInput)
		return nil
	default:
		return fmt.Errorf("fixture generator cannot populate %T", output)
	}
}

func fixtureProposal(input ActionInput, primary domain.HighLevelAction) domain.ActionProposal {
	proposal := domain.ActionProposal{Primary: primary, ModelConfidence: 0.78}
	for _, item := range input.ApplicableExperiences {
		if item.PreferAction == primary.Kind && (item.PreferredTargetID == "" || item.PreferredTargetID == primary.TargetID) {
			proposal.AppliedExperienceIDs = []string{item.ID}
			break
		}
	}
	if len(proposal.AppliedExperienceIDs) == 0 {
		proposal.UncertaintyCodes = []domain.UncertaintyCode{domain.UncertaintyMissingExperience}
	}
	if primary.Kind != domain.ActionStopSession {
		if primary.Kind != domain.ActionHarvestTarget && domain.ActionEnabled(input.Snapshot, domain.ActionHarvestTarget) {
			for _, crop := range input.Snapshot.Crops {
				if crop.Mature && !containsString(input.Coordination.PlayerClaimedTargets, crop.ID) {
					proposal.Alternatives = append(proposal.Alternatives, domain.HighLevelAction{
						SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID,
						SnapshotVersion: input.Snapshot.SnapshotVersion,
						Kind:            domain.ActionHarvestTarget, TargetID: crop.ID,
						Reason: "fallback to the currently mature crop",
					})
					break
				}
			}
		}
		proposal.Alternatives = append(proposal.Alternatives, domain.HighLevelAction{
			SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID,
			SnapshotVersion: input.Snapshot.SnapshotVersion,
			Kind:            domain.ActionStopSession, Reason: "safe fallback if the primary action becomes invalid",
		})
	}
	return proposal
}

func fixtureReflection(input ReflectionInput) (domain.ExperienceObservation, error) {
	context := traitContext(input.Snapshot.Weather)
	if input.Correction != nil {
		return domain.ExperienceObservation{
			Trigger: domain.ExperiencePlayerCorrection, Context: context,
			WhenSignals: situationSignals(input.Snapshot),
			AvoidAction: input.Correction.RejectedAction.Kind, PreferAction: input.Correction.PreferredAction.Kind,
			PreferredTargetID: input.Correction.PreferredAction.TargetID,
			Summary:           "follow the player's explicit correction in this situation",
			EvidenceRef:       input.EvidenceRef, Strength: 0.9,
		}, nil
	}
	if input.Result == nil {
		return domain.ExperienceObservation{}, errors.New("fixture reflection requires result or correction")
	}
	observation := domain.ExperienceObservation{
		Trigger: domain.ExperienceTrigger(input.Result.ErrorCode), Context: context,
		WhenSignals: situationSignals(input.Snapshot), AvoidAction: input.Result.Action.Kind,
		EvidenceRef: input.EvidenceRef, Strength: 0.7,
	}
	switch input.Result.ErrorCode {
	case string(domain.ExperienceInventoryFull):
		observation.WhenSignals = []domain.SituationSignal{domain.SignalInventoryFull, domain.SignalInventoryHasItems}
		observation.PreferAction = domain.ActionDepositItems
		observation.PreferredTargetID = preferredChest(input)
		observation.Summary = "deposit carried items before attempting another harvest"
	case string(domain.ExperienceOutOfWater):
		observation.WhenSignals = []domain.SituationSignal{domain.SignalCanEmpty}
		observation.PreferAction = domain.ActionRefillCan
		if len(input.Snapshot.WaterSources) > 0 {
			observation.PreferredTargetID = input.Snapshot.WaterSources[0].ID
		}
		observation.Summary = "refill the watering can before watering another crop"
	case string(domain.ExperiencePathBlocked):
		observation.WhenSignals = []domain.SituationSignal{domain.SignalTargetBlocked}
		observation.PreferAction = domain.ActionMoveTo
		observation.PreferredTargetID = input.Result.Action.TargetID
		observation.Summary = "move around the obstruction before retrying the task"
	case string(domain.ExperienceChestFull):
		observation.WhenSignals = []domain.SituationSignal{domain.SignalInventoryHasItems}
		observation.PreferAction = domain.ActionDepositItems
		observation.PreferredTargetID = anotherChest(input.Snapshot.Chests, input.Result.Action.TargetID)
		observation.Summary = "use another available chest when the selected chest is full"
	default:
		return domain.ExperienceObservation{}, fmt.Errorf("fixture cannot reflect failure %q", input.Result.ErrorCode)
	}
	return observation, nil
}

func situationSignals(snapshot domain.WorldSnapshot) []domain.SituationSignal {
	result := make([]domain.SituationSignal, 0, 4)
	if snapshot.Inventory.FreeSlots == 0 {
		result = append(result, domain.SignalInventoryFull)
	}
	if len(snapshot.Inventory.Items) > 0 {
		result = append(result, domain.SignalInventoryHasItems)
	}
	if snapshot.WateringCan.Water == 0 {
		result = append(result, domain.SignalCanEmpty)
	}
	if snapshot.Weather == domain.WeatherRainy || snapshot.Weather == domain.WeatherStorm {
		result = append(result, domain.SignalRaining)
	}
	return result
}

func preferredChest(input ReflectionInput) string {
	for _, chest := range input.Snapshot.Chests {
		if chest.ID == input.PlayerModel.PreferredChestID {
			return chest.ID
		}
	}
	if len(input.Snapshot.Chests) > 0 {
		return input.Snapshot.Chests[0].ID
	}
	return ""
}

func anotherChest(chests []domain.Chest, rejected string) string {
	for _, chest := range chests {
		if chest.ID != rejected {
			return chest.ID
		}
	}
	return ""
}

func fixtureLearningInference(input LearningInput) LearningInference {
	revision := 1
	if input.ExistingModel != nil {
		revision = input.ExistingModel.Revision + 1
	}
	evidence := make([]string, 0, len(input.Demonstration.Events))
	for _, event := range input.Demonstration.Events {
		evidence = append(evidence, event.ID)
	}

	order := make([]domain.BehaviorKind, 0, len(input.Segments))
	activityOrder := make([]domain.BehaviorKind, 0, len(input.Segments))
	steps := make([]domain.SkillStep, 0, len(input.Segments))
	seen := make(map[domain.BehaviorKind]struct{})
	seenActivity := make(map[domain.BehaviorKind]struct{})
	preferredChest := ""
	for _, segment := range input.Segments {
		if _, exists := seenActivity[segment.Kind]; !exists {
			seenActivity[segment.Kind] = struct{}{}
			activityOrder = append(activityOrder, segment.Kind)
		}
		if step, ok := fixtureStep(segment.Kind); ok {
			if _, exists := seen[segment.Kind]; !exists {
				seen[segment.Kind] = struct{}{}
				order = append(order, segment.Kind)
				steps = append(steps, step)
			}
		}
		if segment.Kind == domain.BehaviorDepositing && len(segment.TargetIDs) > 0 {
			preferredChest = segment.TargetIDs[len(segment.TargetIDs)-1]
		}
	}
	if len(steps) == 0 {
		steps = []domain.SkillStep{{Action: domain.ActionStopSession}}
	}

	context := traitContext(input.Demonstration.Weather)
	observations := []domain.TraitObservation{{
		Key: domain.PreferenceTaskOrder, Value: behaviorOrderValue(order), Context: context,
		SupportingEventIDs: evidence, Strength: 0.7,
	}}
	if hasExtendedActivity(activityOrder) {
		observations = append(observations, domain.TraitObservation{
			Key: domain.PreferenceActivityOrder, Value: behaviorOrderValue(activityOrder), Context: context,
			SupportingEventIDs: evidence, Strength: 0.7,
		})
	}
	if itemID, itemEvidence := leadingResource(input.Segments); itemID != "" {
		observations = append(observations, domain.TraitObservation{
			Key: domain.PreferenceResourcePriority, Value: itemID, Context: domain.TraitContextAny,
			SupportingEventIDs: itemEvidence, Strength: 0.65,
		})
	}
	if segment, ok := firstSegment(input.Segments, domain.BehaviorMineTraversal); ok {
		observations = append(observations, domain.TraitObservation{
			Key: domain.PreferenceMineExitPolicy, Value: "observed_mine_progression", Context: domain.TraitContextAny,
			SupportingEventIDs: append([]string(nil), segment.EventIDs...), Strength: 0.55,
		})
	}
	if segment, ok := firstSegment(input.Segments, domain.BehaviorFishing); ok {
		observations = append(observations, domain.TraitObservation{
			Key: domain.PreferenceFishingContext, Value: string(context), Context: context,
			SupportingEventIDs: append([]string(nil), segment.EventIDs...), Strength: 0.6,
		})
	}
	if preferredChest != "" {
		observations = append(observations, domain.TraitObservation{
			Key: domain.PreferencePreferredChest, Value: preferredChest, Context: domain.TraitContextAny,
			SupportingEventIDs: depositEvidence(input.Demonstration.Events), Strength: 0.8,
		})
	}
	observations = append(observations, domain.TraitObservation{
		Key: domain.PreferenceRouteStyle, Value: "demonstrated_order", Context: domain.TraitContextAny,
		SupportingEventIDs: evidence, Strength: 0.6,
	})

	return LearningInference{
		Observations: observations,
		Skill: domain.SkillProgram{
			Name: "morning-farm-routine", Revision: revision,
			Goal:          "care for all currently actionable farm crops",
			Preconditions: []string{"player is on the farm"}, TargetSelector: "current_actionable_crops",
			PreferredOrder: order, Steps: steps,
			SuccessConditions: []string{"no mature or unwatered crops remain"},
			StopConditions:    []string{"energy reserve reached", "time limit reached"},
			RecoveryStrategies: []domain.RecoveryStrategy{
				{FailureCode: "out_of_water", Actions: []domain.ActionKind{domain.ActionRefillCan}},
				{FailureCode: "path_blocked", Actions: []domain.ActionKind{domain.ActionMoveTo, domain.ActionStopSession}},
				{FailureCode: "inventory_full", Actions: []domain.ActionKind{domain.ActionDepositItems, domain.ActionStopSession}},
				{FailureCode: "chest_full", Actions: []domain.ActionKind{domain.ActionStopSession}},
			},
			EvidenceEventIDs: evidence,
		},
	}
}

func traitContext(weather domain.Weather) domain.TraitContext {
	switch weather {
	case domain.WeatherSunny:
		return domain.TraitContextSunny
	case domain.WeatherRainy:
		return domain.TraitContextRainy
	case domain.WeatherStorm:
		return domain.TraitContextStorm
	case domain.WeatherSnow:
		return domain.TraitContextSnow
	default:
		return domain.TraitContextAny
	}
}

func fixtureAction(input ActionInput) domain.HighLevelAction {
	snapshot := input.Snapshot
	claimed := make(map[string]struct{}, len(input.Coordination.PlayerClaimedTargets))
	for _, targetID := range input.Coordination.PlayerClaimedTargets {
		claimed[targetID] = struct{}{}
	}
	newAction := func(kind domain.ActionKind, targetID, reason string) domain.HighLevelAction {
		return domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
			Kind: kind, TargetID: targetID, Reason: reason,
		}
	}
	for _, item := range input.ApplicableExperiences {
		if !domain.ActionEnabled(snapshot, item.PreferAction) {
			continue
		}
		if action, ok := actionFromExperience(item, input, newAction); ok {
			return action
		}
	}
	for _, crop := range snapshot.Crops {
		if _, occupied := claimed[crop.ID]; occupied {
			continue
		}
		if crop.Mature && domain.ActionEnabled(snapshot, domain.ActionHarvestTarget) {
			return newAction(domain.ActionHarvestTarget, crop.ID, "harvest a currently mature crop")
		}
	}
	if snapshot.Weather != domain.WeatherRainy && snapshot.Weather != domain.WeatherStorm {
		for _, crop := range snapshot.Crops {
			if _, occupied := claimed[crop.ID]; occupied {
				continue
			}
			if crop.NeedsWater {
				if snapshot.WateringCan.Water <= 0 && len(snapshot.WaterSources) > 0 {
					return newAction(domain.ActionRefillCan, snapshot.WaterSources[0].ID, "refill before continuing the learned watering goal")
				}
				return newAction(domain.ActionWaterTarget, crop.ID, "water a currently dry crop")
			}
		}
	}
	if len(snapshot.Inventory.Items) > 0 {
		for _, chest := range snapshot.Chests {
			if chest.ID == input.PlayerModel.PreferredChestID {
				return newAction(domain.ActionDepositItems, chest.ID, "use the player's preferred chest")
			}
		}
	}
	return newAction(domain.ActionStopSession, "", "morning routine is complete")
}

func actionFromExperience(item domain.PolicyExperience, input ActionInput, newAction func(domain.ActionKind, string, string) domain.HighLevelAction) (domain.HighLevelAction, bool) {
	if !domain.ActionEnabled(input.Snapshot, item.PreferAction) {
		return domain.HighLevelAction{}, false
	}
	targetID := item.PreferredTargetID
	switch item.PreferAction {
	case domain.ActionDepositItems:
		if targetID == "" && len(input.Snapshot.Chests) > 0 {
			targetID = input.Snapshot.Chests[0].ID
		}
	case domain.ActionRefillCan:
		if targetID == "" && len(input.Snapshot.WaterSources) > 0 {
			targetID = input.Snapshot.WaterSources[0].ID
		}
	case domain.ActionHarvestTarget:
		if targetID == "" {
			for _, crop := range input.Snapshot.Crops {
				if crop.Mature && !containsString(input.Coordination.PlayerClaimedTargets, crop.ID) {
					targetID = crop.ID
					break
				}
			}
		}
	case domain.ActionWaterTarget:
		if targetID == "" {
			for _, crop := range input.Snapshot.Crops {
				if crop.NeedsWater && !containsString(input.Coordination.PlayerClaimedTargets, crop.ID) {
					targetID = crop.ID
					break
				}
			}
		}
	default:
		return domain.HighLevelAction{}, false
	}
	if targetID == "" {
		return domain.HighLevelAction{}, false
	}
	return newAction(item.PreferAction, targetID, "apply learned experience: "+item.Summary), true
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func fixtureIntent(input IntentInput) IntentInference {
	counts := map[domain.PlayerIntent]int{}
	evidence := map[domain.PlayerIntent][]string{}
	for _, activity := range input.Activities {
		var intent domain.PlayerIntent
		switch activity.Kind {
		case domain.EventWater, domain.EventRefill:
			intent = domain.PlayerIntentWatering
		case domain.EventHarvest:
			intent = domain.PlayerIntentHarvesting
		case domain.EventDeposit:
			intent = domain.PlayerIntentDepositing
		default:
			continue
		}
		counts[intent]++
		evidence[intent] = append(evidence[intent], activity.TargetID)
	}
	selected := domain.PlayerIntentUnknown
	for _, intent := range []domain.PlayerIntent{
		domain.PlayerIntentWatering,
		domain.PlayerIntentHarvesting,
		domain.PlayerIntentDepositing,
	} {
		if counts[intent] > counts[selected] {
			selected = intent
		}
	}
	return IntentInference{Intent: selected, EvidenceTargetIDs: evidence[selected]}
}

func fixtureReplan(input ReplanInput) domain.HighLevelAction {
	snapshot := input.ActionInput.Snapshot
	if input.LastResult.ErrorCode == "inventory_full" {
		for _, chest := range snapshot.Chests {
			if chest.ID == input.ActionInput.PlayerModel.PreferredChestID {
				return domain.HighLevelAction{
					SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
					Kind: domain.ActionDepositItems, TargetID: chest.ID,
					Reason: "deposit the full Echo inventory into the player's preferred chest",
				}
			}
		}
		return domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
			Kind: domain.ActionStopSession, Reason: "Echo inventory is full and the preferred chest is unavailable",
		}
	}
	if input.LastResult.ErrorCode == "chest_full" {
		return domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
			Kind: domain.ActionStopSession, Reason: "the preferred chest cannot accept the remaining harvest",
		}
	}
	if input.LastResult.ErrorCode == "path_blocked" && input.LastResult.Action.TargetID != "" {
		return domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID,
			SnapshotVersion: snapshot.SnapshotVersion,
			Kind:            domain.ActionMoveTo, TargetID: input.LastResult.Action.TargetID,
			Reason: "replan a route around the blocked tile",
		}
	}
	return fixtureAction(input.ActionInput)
}

func fixtureStep(kind domain.BehaviorKind) (domain.SkillStep, bool) {
	switch kind {
	case domain.BehaviorWatering:
		return domain.SkillStep{Action: domain.ActionWaterTarget, TargetSelector: "dry_crops"}, true
	case domain.BehaviorRefilling:
		return domain.SkillStep{Action: domain.ActionRefillCan, TargetSelector: "reachable_water_source"}, true
	case domain.BehaviorHarvesting:
		return domain.SkillStep{Action: domain.ActionHarvestTarget, TargetSelector: "mature_crops"}, true
	case domain.BehaviorDepositing:
		return domain.SkillStep{Action: domain.ActionDepositItems, TargetSelector: "preferred_chest"}, true
	default:
		return domain.SkillStep{}, false
	}
}

func behaviorOrderValue(order []domain.BehaviorKind) string {
	values := make([]string, 0, len(order))
	for _, kind := range order {
		values = append(values, string(kind))
	}
	return strings.Join(values, ",")
}

func hasExtendedActivity(order []domain.BehaviorKind) bool {
	for _, kind := range order {
		switch kind {
		case domain.BehaviorWoodcutting, domain.BehaviorMining, domain.BehaviorMineTraversal, domain.BehaviorFishing:
			return true
		}
	}
	return false
}

func leadingResource(segments []domain.BehaviorSegment) (string, []string) {
	quantities := make(map[string]int)
	evidence := make(map[string][]string)
	for _, segment := range segments {
		for _, item := range segment.ItemDeltas {
			if item.Quantity <= 0 {
				continue
			}
			quantities[item.ItemID] += item.Quantity
			evidence[item.ItemID] = append(evidence[item.ItemID], segment.EventIDs...)
		}
	}
	selected := ""
	for itemID, quantity := range quantities {
		if selected == "" || quantity > quantities[selected] || (quantity == quantities[selected] && itemID < selected) {
			selected = itemID
		}
	}
	return selected, append([]string(nil), evidence[selected]...)
}

func firstSegment(segments []domain.BehaviorSegment, kind domain.BehaviorKind) (domain.BehaviorSegment, bool) {
	for _, segment := range segments {
		if segment.Kind == kind {
			return segment, true
		}
	}
	return domain.BehaviorSegment{}, false
}

func depositEvidence(events []domain.DemonstrationEvent) []string {
	result := make([]string, 0)
	for _, event := range events {
		if event.Kind == domain.EventDeposit {
			result = append(result, event.ID)
		}
	}
	return result
}
