package coordination

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
)

const ActivityWindowTicks int64 = 20 * 60

var ErrTargetClaimed = errors.New("target is claimed by the player")

type Service struct {
	inferer intelligence.IntentInferer
}

func NewService(inferer intelligence.IntentInferer) (*Service, error) {
	if inferer == nil {
		return nil, errors.New("intent inferer is required")
	}
	return &Service{inferer: inferer}, nil
}

func (s *Service) Prepare(ctx context.Context, snapshot domain.WorldSnapshot, model domain.PlayerModel) (domain.CoordinationContext, error) {
	activities := recentSuccessfulActivities(snapshot)
	coordination := domain.CoordinationContext{
		InferredIntent: domain.PlayerIntentUnknown,
		ModelRevision:  model.Revision,
	}
	claimed := make(map[string]struct{}, len(activities))
	for _, activity := range activities {
		if _, exists := claimed[activity.TargetID]; exists {
			continue
		}
		claimed[activity.TargetID] = struct{}{}
		coordination.PlayerClaimedTargets = append(coordination.PlayerClaimedTargets, activity.TargetID)
	}
	coordination.AvailableGoals = availableGoals(snapshot, claimed)
	if len(activities) == 0 {
		return coordination, nil
	}
	intent, err := s.inferer.InferIntent(ctx, intelligence.IntentInput{
		SaveID: snapshot.SaveID, Day: snapshot.Day, TimeOfDay: snapshot.TimeOfDay,
		Activities: activities, PlayerModel: model,
	})
	if err != nil {
		return domain.CoordinationContext{}, fmt.Errorf("infer player intent: %w", err)
	}
	coordination.InferredIntent = intent
	return coordination, nil
}

func ValidateChoice(coordination domain.CoordinationContext, action domain.HighLevelAction) error {
	for _, targetID := range coordination.PlayerClaimedTargets {
		if action.TargetID != "" && action.TargetID == targetID {
			return fmt.Errorf("%w: %s", ErrTargetClaimed, targetID)
		}
	}
	return nil
}

func recentSuccessfulActivities(snapshot domain.WorldSnapshot) []domain.PlayerActivity {
	result := make([]domain.PlayerActivity, 0, len(snapshot.RecentPlayerActions))
	for _, activity := range snapshot.RecentPlayerActions {
		if !activity.Success || activity.Tick > snapshot.Tick || snapshot.Tick-activity.Tick > ActivityWindowTicks {
			continue
		}
		result = append(result, activity)
	}
	return result
}

func availableGoals(snapshot domain.WorldSnapshot, claimed map[string]struct{}) []string {
	goals := make([]string, 0, len(snapshot.Crops)*2+len(snapshot.Chests))
	for _, crop := range snapshot.Crops {
		if _, occupied := claimed[crop.ID]; !occupied && crop.Mature {
			goals = append(goals, "harvest:"+crop.ID)
		}
	}
	if snapshot.Weather != domain.WeatherRainy && snapshot.Weather != domain.WeatherStorm {
		for _, crop := range snapshot.Crops {
			if _, occupied := claimed[crop.ID]; !occupied && crop.NeedsWater {
				goals = append(goals, "water:"+crop.ID)
			}
		}
	}
	if len(snapshot.Inventory.Items) > 0 {
		for _, chest := range snapshot.Chests {
			if _, occupied := claimed[chest.ID]; !occupied {
				goals = append(goals, "deposit:"+chest.ID)
			}
		}
	}
	return goals
}
