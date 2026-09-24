package policy

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/coordination"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

const MorningFarmRoutine = "morning-farm-routine"

type actor interface {
	ChooseAction(ctx context.Context, input intelligence.ActionInput) (domain.HighLevelAction, error)
	Replan(ctx context.Context, input intelligence.ReplanInput) (domain.HighLevelAction, error)
}

type collaborator interface {
	Prepare(context.Context, domain.WorldSnapshot, domain.PlayerModel) (domain.CoordinationContext, error)
}

type noOpCollaborator struct{}

func (noOpCollaborator) Prepare(_ context.Context, _ domain.WorldSnapshot, model domain.PlayerModel) (domain.CoordinationContext, error) {
	return domain.CoordinationContext{InferredIntent: domain.PlayerIntentUnknown, ModelRevision: model.Revision}, nil
}

type Service struct {
	store        memory.DecisionStore
	actor        actor
	collaborator collaborator
}

func NewService(store memory.DecisionStore, actor actor, collaborators ...collaborator) (*Service, error) {
	if store == nil {
		return nil, errors.New("memory store is required")
	}
	if actor == nil {
		return nil, errors.New("actor is required")
	}
	if len(collaborators) > 1 || (len(collaborators) == 1 && collaborators[0] == nil) {
		return nil, errors.New("at most one non-nil collaborator is allowed")
	}
	var collaborationService collaborator = noOpCollaborator{}
	if len(collaborators) == 1 {
		collaborationService = collaborators[0]
	}
	return &Service{store: store, actor: actor, collaborator: collaborationService}, nil
}

func (s *Service) NextAction(ctx context.Context, saveID string, snapshot domain.WorldSnapshot) (domain.HighLevelAction, error) {
	if existing, err := s.store.GetDecision(ctx, saveID, snapshot.SessionID, snapshot.SnapshotVersion); err == nil {
		return existing.FinalAction, nil
	} else if !errors.Is(err, memory.ErrNotFound) {
		return domain.HighLevelAction{}, fmt.Errorf("load existing decision: %w", err)
	}
	input, stop, err := s.prepare(ctx, saveID, snapshot)
	if err != nil {
		return domain.HighLevelAction{}, err
	}
	if stop != nil {
		return *stop, s.saveDecision(ctx, snapshot, input.Coordination, *stop, *stop)
	}
	candidate, err := s.actor.ChooseAction(ctx, input)
	if err != nil {
		return domain.HighLevelAction{}, err
	}
	if err := validateActionForSnapshot(candidate, snapshot); err != nil {
		return domain.HighLevelAction{}, fmt.Errorf("reject AI action: %w", err)
	}
	final := s.resolveClaimConflict(candidate, input.Coordination, snapshot)
	if err := s.saveDecision(ctx, snapshot, input.Coordination, candidate, final); err != nil {
		return domain.HighLevelAction{}, err
	}
	return final, nil
}

func (s *Service) HandleResult(ctx context.Context, saveID string, snapshot domain.WorldSnapshot, result domain.ActionResult) (domain.HighLevelAction, error) {
	if err := result.Validate(); err != nil {
		return domain.HighLevelAction{}, fmt.Errorf("validate action result: %w", err)
	}
	if result.SaveID != saveID || result.SaveID != snapshot.SaveID || result.SessionID != snapshot.SessionID {
		return domain.HighLevelAction{}, errors.New("action result identity does not match current snapshot")
	}
	if result.SnapshotVersion > snapshot.SnapshotVersion {
		return domain.HighLevelAction{}, errors.New("action result is newer than current snapshot")
	}
	if err := s.store.AttachDecisionResult(ctx, result); err != nil {
		return domain.HighLevelAction{}, fmt.Errorf("record action result: %w", err)
	}
	if existing, err := s.store.GetDecision(ctx, saveID, snapshot.SessionID, snapshot.SnapshotVersion); err == nil && snapshot.SnapshotVersion != result.SnapshotVersion {
		return existing.FinalAction, nil
	} else if err != nil && !errors.Is(err, memory.ErrNotFound) {
		return domain.HighLevelAction{}, fmt.Errorf("load existing decision: %w", err)
	}
	input, stop, err := s.prepare(ctx, saveID, snapshot)
	if err != nil {
		return domain.HighLevelAction{}, err
	}
	if stop != nil {
		return *stop, s.saveDecision(ctx, snapshot, input.Coordination, *stop, *stop)
	}

	var candidate domain.HighLevelAction
	if result.Status == domain.ActionFailed {
		candidate, err = s.actor.Replan(ctx, intelligence.ReplanInput{ActionInput: input, LastResult: result})
	} else {
		candidate, err = s.actor.ChooseAction(ctx, input)
	}
	if err != nil {
		return domain.HighLevelAction{}, err
	}
	if err := validateActionForSnapshot(candidate, snapshot); err != nil {
		return domain.HighLevelAction{}, fmt.Errorf("reject AI action: %w", err)
	}
	final := s.resolveClaimConflict(candidate, input.Coordination, snapshot)
	if err := s.saveDecision(ctx, snapshot, input.Coordination, candidate, final); err != nil {
		return domain.HighLevelAction{}, err
	}
	return final, nil
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
	collaborationContext, err := s.collaborator.Prepare(ctx, snapshot, model)
	if err != nil {
		return intelligence.ActionInput{}, nil, err
	}
	input := intelligence.ActionInput{Snapshot: snapshot, PlayerModel: model, Skill: skill, Coordination: collaborationContext}
	if snapshot.Energy <= model.EnergyReserve {
		stop := domain.HighLevelAction{
			SaveID: saveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
			Kind: domain.ActionStopSession, Reason: "player energy reserve reached",
		}
		return input, &stop, nil
	}
	return input, nil, nil
}

func (s *Service) resolveClaimConflict(candidate domain.HighLevelAction, collaborationContext domain.CoordinationContext, snapshot domain.WorldSnapshot) domain.HighLevelAction {
	if err := coordination.ValidateChoice(collaborationContext, candidate); err == nil {
		return candidate
	}
	return domain.HighLevelAction{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Kind: domain.ActionStopSession, Reason: "player is already handling the selected target",
	}
}

func (s *Service) saveDecision(ctx context.Context, snapshot domain.WorldSnapshot, collaborationContext domain.CoordinationContext, candidate, final domain.HighLevelAction) error {
	record := domain.DecisionRecord{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Day: snapshot.Day, ModelRevision: collaborationContext.ModelRevision,
		InferredIntent:       collaborationContext.InferredIntent,
		PlayerClaimedTargets: append([]string(nil), collaborationContext.PlayerClaimedTargets...),
		CandidateAction:      candidate, FinalAction: final,
	}
	if err := s.store.SaveDecision(ctx, record); err != nil {
		return fmt.Errorf("persist decision: %w", err)
	}
	return nil
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
		if len(snapshot.Inventory.Items) == 0 {
			return errors.New("Echo inventory is empty")
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
