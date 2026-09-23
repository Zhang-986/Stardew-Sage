package policy

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

const MorningFarmRoutine = "morning-farm-routine"

type actor interface {
	ChooseAction(ctx context.Context, input intelligence.ActionInput) (domain.HighLevelAction, error)
	Replan(ctx context.Context, input intelligence.ReplanInput) (domain.HighLevelAction, error)
}

type Service struct {
	store memory.Store
	actor actor
}

func NewService(store memory.Store, actor actor) (*Service, error) {
	if store == nil {
		return nil, errors.New("memory store is required")
	}
	if actor == nil {
		return nil, errors.New("actor is required")
	}
	return &Service{store: store, actor: actor}, nil
}

func (s *Service) NextAction(ctx context.Context, saveID string, snapshot domain.WorldSnapshot) (domain.HighLevelAction, error) {
	input, stop, err := s.prepare(ctx, saveID, snapshot)
	if err != nil {
		return domain.HighLevelAction{}, err
	}
	if stop != nil {
		return *stop, nil
	}
	action, err := s.actor.ChooseAction(ctx, input)
	if err != nil {
		return domain.HighLevelAction{}, err
	}
	if err := validateActionForSnapshot(action, snapshot); err != nil {
		return domain.HighLevelAction{}, fmt.Errorf("reject AI action: %w", err)
	}
	return action, nil
}

func (s *Service) HandleResult(ctx context.Context, saveID string, snapshot domain.WorldSnapshot, result domain.ActionResult) (domain.HighLevelAction, error) {
	if result.SaveID != saveID || result.SaveID != snapshot.SaveID || result.SessionID != snapshot.SessionID {
		return domain.HighLevelAction{}, errors.New("action result identity does not match current snapshot")
	}
	if result.SnapshotVersion > snapshot.SnapshotVersion {
		return domain.HighLevelAction{}, errors.New("action result is newer than current snapshot")
	}
	input, stop, err := s.prepare(ctx, saveID, snapshot)
	if err != nil {
		return domain.HighLevelAction{}, err
	}
	if stop != nil {
		return *stop, nil
	}

	var action domain.HighLevelAction
	if result.Status == domain.ActionFailed {
		action, err = s.actor.Replan(ctx, intelligence.ReplanInput{ActionInput: input, LastResult: result})
	} else {
		action, err = s.actor.ChooseAction(ctx, input)
	}
	if err != nil {
		return domain.HighLevelAction{}, err
	}
	if err := validateActionForSnapshot(action, snapshot); err != nil {
		return domain.HighLevelAction{}, fmt.Errorf("reject AI action: %w", err)
	}
	return action, nil
}

func (s *Service) prepare(ctx context.Context, saveID string, snapshot domain.WorldSnapshot) (intelligence.ActionInput, *domain.HighLevelAction, error) {
	if saveID == "" || saveID != snapshot.SaveID {
		return intelligence.ActionInput{}, nil, errors.New("save ID does not match snapshot")
	}
	if err := snapshot.Validate(); err != nil {
		return intelligence.ActionInput{}, nil, fmt.Errorf("validate world snapshot: %w", err)
	}
	model, err := s.store.GetPlayerModel(ctx, saveID)
	if err != nil {
		return intelligence.ActionInput{}, nil, fmt.Errorf("load player model: %w", err)
	}
	skill, err := s.store.GetSkill(ctx, saveID, MorningFarmRoutine)
	if err != nil {
		return intelligence.ActionInput{}, nil, fmt.Errorf("load morning skill: %w", err)
	}
	input := intelligence.ActionInput{Snapshot: snapshot, PlayerModel: model, Skill: skill}
	if snapshot.Energy <= model.EnergyReserve {
		stop := domain.HighLevelAction{
			SaveID: saveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
			Kind: domain.ActionStopSession, Reason: "player energy reserve reached",
		}
		return input, &stop, nil
	}
	return input, nil, nil
}

func validateActionForSnapshot(action domain.HighLevelAction, snapshot domain.WorldSnapshot) error {
	if err := action.Validate(); err != nil {
		return err
	}
	if action.SaveID != snapshot.SaveID || action.SessionID != snapshot.SessionID || action.SnapshotVersion != snapshot.SnapshotVersion {
		return errors.New("action identity or snapshot version is stale")
	}

	switch action.Kind {
	case domain.ActionWaterTarget:
		crop, ok := findCrop(snapshot.Crops, action.TargetID)
		if !ok {
			return fmt.Errorf("crop target %q is not present", action.TargetID)
		}
		if snapshot.Weather == domain.WeatherRainy || snapshot.Weather == domain.WeatherStorm {
			return errors.New("cannot water crops in rain or storm")
		}
		if !crop.NeedsWater {
			return fmt.Errorf("crop target %q does not need water", action.TargetID)
		}
		if snapshot.WateringCan.Water <= 0 {
			return errors.New("watering can is empty")
		}
	case domain.ActionHarvestTarget:
		crop, ok := findCrop(snapshot.Crops, action.TargetID)
		if !ok {
			return fmt.Errorf("crop target %q is not present", action.TargetID)
		}
		if !crop.Mature {
			return fmt.Errorf("crop target %q is not mature", action.TargetID)
		}
	case domain.ActionRefillCan:
		if !hasWaterSource(snapshot.WaterSources, action.TargetID) {
			return fmt.Errorf("water source target %q is not present", action.TargetID)
		}
		if snapshot.WateringCan.Water >= snapshot.WateringCan.Capacity {
			return errors.New("watering can is already full")
		}
	case domain.ActionDepositItems:
		if !hasChest(snapshot.Chests, action.TargetID) {
			return fmt.Errorf("chest target %q is not present", action.TargetID)
		}
	case domain.ActionMoveTo:
		if action.Destination == nil && !hasTarget(snapshot, action.TargetID) {
			return fmt.Errorf("move target %q is not present", action.TargetID)
		}
	}
	return nil
}

func findCrop(crops []domain.Crop, id string) (domain.Crop, bool) {
	for _, crop := range crops {
		if crop.ID == id {
			return crop, true
		}
	}
	return domain.Crop{}, false
}

func hasWaterSource(sources []domain.WaterSource, id string) bool {
	for _, source := range sources {
		if source.ID == id {
			return true
		}
	}
	return false
}

func hasChest(chests []domain.Chest, id string) bool {
	for _, chest := range chests {
		if chest.ID == id {
			return true
		}
	}
	return false
}

func hasTarget(snapshot domain.WorldSnapshot, id string) bool {
	if _, ok := findCrop(snapshot.Crops, id); ok {
		return true
	}
	return hasWaterSource(snapshot.WaterSources, id) || hasChest(snapshot.Chests, id)
}
