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
	case *LearningResult:
		learningInput, ok := input.(LearningInput)
		if !ok {
			return errors.New("fixture learning generator received unexpected input")
		}
		*target = fixtureLearningResult(learningInput)
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
	default:
		return fmt.Errorf("fixture generator cannot populate %T", output)
	}
}

func fixtureLearningResult(input LearningInput) LearningResult {
	revision := 1
	if input.ExistingModel != nil {
		revision = input.ExistingModel.Revision + 1
	}
	evidence := make([]string, 0, len(input.Demonstration.Events))
	for _, event := range input.Demonstration.Events {
		evidence = append(evidence, event.ID)
	}

	order := make([]domain.BehaviorKind, 0, len(input.Segments))
	steps := make([]domain.SkillStep, 0, len(input.Segments))
	seen := make(map[domain.BehaviorKind]struct{})
	preferredChest := ""
	for _, segment := range input.Segments {
		if _, exists := seen[segment.Kind]; !exists {
			seen[segment.Kind] = struct{}{}
			order = append(order, segment.Kind)
			if step, ok := fixtureStep(segment.Kind); ok {
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

	preferences := []domain.ObservedPreference{{
		Key: domain.PreferenceTaskOrder, Value: behaviorOrderValue(order),
		EvidenceEventIDs: evidence, ObservationCount: 1, Confidence: 0.65,
	}}
	if preferredChest != "" {
		preferences = append(preferences, domain.ObservedPreference{
			Key: domain.PreferencePreferredChest, Value: preferredChest,
			EvidenceEventIDs: depositEvidence(input.Demonstration.Events), ObservationCount: 1, Confidence: 0.8,
		})
	}

	return LearningResult{
		PlayerModel: domain.PlayerModel{
			SaveID: input.Demonstration.SaveID, Revision: revision, CommonTaskOrder: order,
			PreferredChestID: preferredChest, EnergyReserve: 40, RouteStyle: "demonstrated_order",
			Preferences: preferences,
		},
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
			},
			EvidenceEventIDs: evidence,
		},
	}
}

func fixtureAction(input ActionInput) domain.HighLevelAction {
	snapshot := input.Snapshot
	newAction := func(kind domain.ActionKind, targetID, reason string) domain.HighLevelAction {
		return domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
			Kind: kind, TargetID: targetID, Reason: reason,
		}
	}
	for _, crop := range snapshot.Crops {
		if crop.Mature {
			return newAction(domain.ActionHarvestTarget, crop.ID, "harvest a currently mature crop")
		}
	}
	if snapshot.Weather != domain.WeatherRainy && snapshot.Weather != domain.WeatherStorm {
		for _, crop := range snapshot.Crops {
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

func fixtureReplan(input ReplanInput) domain.HighLevelAction {
	if input.LastResult.ErrorCode == "path_blocked" && input.LastResult.Action.TargetID != "" {
		return domain.HighLevelAction{
			SaveID: input.ActionInput.Snapshot.SaveID, SessionID: input.ActionInput.Snapshot.SessionID,
			SnapshotVersion: input.ActionInput.Snapshot.SnapshotVersion,
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

func depositEvidence(events []domain.DemonstrationEvent) []string {
	result := make([]string, 0)
	for _, event := range events {
		if event.Kind == domain.EventDeposit {
			result = append(result, event.ID)
		}
	}
	return result
}
