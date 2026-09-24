package policy

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/coordination"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/experience"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

const MorningFarmRoutine = "morning-farm-routine"

type actor interface {
	ProposeAction(ctx context.Context, input intelligence.ActionInput) (domain.ActionProposal, error)
	ProposeRecovery(ctx context.Context, input intelligence.ReplanInput) (domain.ActionProposal, error)
}

type experienceLearner interface {
	LearnFromResult(context.Context, domain.WorldSnapshot, domain.ActionResult) (domain.ExperienceOutcome, error)
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
	reflection   experienceLearner
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

func NewReflectiveService(store memory.DecisionStore, actor actor, collaborator collaborator, reflection experienceLearner) (*Service, error) {
	if collaborator == nil {
		return nil, errors.New("collaborator is required")
	}
	if reflection == nil {
		return nil, errors.New("experience learner is required")
	}
	service, err := NewService(store, actor, collaborator)
	if err != nil {
		return nil, err
	}
	service.reflection = reflection
	return service, nil
}

func (s *Service) NextAction(ctx context.Context, saveID string, snapshot domain.WorldSnapshot) (domain.HighLevelAction, error) {
	decision, err := s.NextDecision(ctx, saveID, snapshot)
	return decision.Action, err
}

func (s *Service) NextDecision(ctx context.Context, saveID string, snapshot domain.WorldSnapshot) (domain.ActionDecision, error) {
	if existing, err := s.store.GetDecision(ctx, saveID, snapshot.SessionID, snapshot.SnapshotVersion); err == nil {
		return decisionFromRecord(existing), nil
	} else if !errors.Is(err, memory.ErrNotFound) {
		return domain.ActionDecision{}, fmt.Errorf("load existing decision: %w", err)
	}
	input, stop, err := s.prepare(ctx, saveID, snapshot)
	if err != nil {
		return domain.ActionDecision{}, err
	}
	if stop != nil {
		proposal := domain.ActionProposal{Primary: *stop, ModelConfidence: 1}
		decision := domain.ActionDecision{Action: *stop, Confidence: 1}
		return s.saveDecision(ctx, snapshot, input.Coordination, proposal, decision, 0)
	}
	proposal, err := s.actor.ProposeAction(ctx, input)
	if err != nil {
		return domain.ActionDecision{}, err
	}
	decision, selected, err := s.selectDecision(proposal, input)
	if err != nil {
		return domain.ActionDecision{}, err
	}
	return s.saveDecision(ctx, snapshot, input.Coordination, proposal, decision, selected)
}

func (s *Service) HandleResult(ctx context.Context, saveID string, snapshot domain.WorldSnapshot, result domain.ActionResult) (domain.HighLevelAction, error) {
	decision, err := s.HandleResultDecision(ctx, saveID, snapshot, result)
	return decision.Action, err
}

func (s *Service) HandleResultDecision(ctx context.Context, saveID string, snapshot domain.WorldSnapshot, result domain.ActionResult) (domain.ActionDecision, error) {
	if err := result.Validate(); err != nil {
		return domain.ActionDecision{}, fmt.Errorf("validate action result: %w", err)
	}
	if result.SaveID != saveID || result.SaveID != snapshot.SaveID || result.SessionID != snapshot.SessionID {
		return domain.ActionDecision{}, errors.New("action result identity does not match current snapshot")
	}
	if snapshot.SnapshotVersion <= result.SnapshotVersion {
		return domain.ActionDecision{}, errors.New("current snapshot must be newer than action result")
	}
	if err := s.store.AttachDecisionResult(ctx, result); err != nil {
		return domain.ActionDecision{}, fmt.Errorf("record action result: %w", err)
	}
	if result.Status == domain.ActionFailed && s.reflection != nil {
		_, _ = s.reflection.LearnFromResult(ctx, snapshot, result)
	}
	if existing, err := s.store.GetDecision(ctx, saveID, snapshot.SessionID, snapshot.SnapshotVersion); err == nil && snapshot.SnapshotVersion != result.SnapshotVersion {
		return decisionFromRecord(existing), nil
	} else if err != nil && !errors.Is(err, memory.ErrNotFound) {
		return domain.ActionDecision{}, fmt.Errorf("load existing decision: %w", err)
	}
	input, stop, err := s.prepare(ctx, saveID, snapshot)
	if err != nil {
		return domain.ActionDecision{}, err
	}
	if stop != nil {
		proposal := domain.ActionProposal{Primary: *stop, ModelConfidence: 1}
		decision := domain.ActionDecision{Action: *stop, Confidence: 1}
		return s.saveDecision(ctx, snapshot, input.Coordination, proposal, decision, 0)
	}

	var proposal domain.ActionProposal
	if result.Status == domain.ActionFailed {
		proposal, err = s.actor.ProposeRecovery(ctx, intelligence.ReplanInput{ActionInput: input, LastResult: result})
	} else {
		proposal, err = s.actor.ProposeAction(ctx, input)
	}
	if err != nil {
		return domain.ActionDecision{}, err
	}
	decision, selected, err := s.selectDecision(proposal, input)
	if err != nil {
		return domain.ActionDecision{}, err
	}
	return s.saveDecision(ctx, snapshot, input.Coordination, proposal, decision, selected)
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
	storedExperiences, err := s.store.ListPolicyExperiences(ctx, saveID)
	if err != nil {
		return intelligence.ActionInput{}, nil, fmt.Errorf("load policy experiences: %w", err)
	}
	applicableExperiences := experience.Match(snapshot, storedExperiences, 3)
	collaborationContext, err := s.collaborator.Prepare(ctx, snapshot, model)
	if err != nil {
		return intelligence.ActionInput{}, nil, err
	}
	input := intelligence.ActionInput{
		Snapshot: snapshot, PlayerModel: model, Skill: skill, Coordination: collaborationContext,
		ApplicableExperiences: applicableExperiences,
	}
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

func (s *Service) selectDecision(proposal domain.ActionProposal, input intelligence.ActionInput) (domain.ActionDecision, int, error) {
	availableExperiences := make(map[string]struct{}, len(input.ApplicableExperiences))
	for _, item := range input.ApplicableExperiences {
		availableExperiences[item.ID] = struct{}{}
	}
	if err := proposal.Validate(input.Snapshot, availableExperiences); err != nil {
		return domain.ActionDecision{}, 0, fmt.Errorf("reject AI proposal: %w", err)
	}
	confidence := policyConfidence(proposal, input)
	if confidence < 0.35 {
		stop := domain.HighLevelAction{
			SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID, SnapshotVersion: input.Snapshot.SnapshotVersion,
			Kind: domain.ActionStopSession, Reason: "policy confidence is below the safe execution threshold",
		}
		return domain.ActionDecision{
			Action: stop, Confidence: confidence,
			Alternatives:         append([]domain.HighLevelAction(nil), proposal.Alternatives...),
			AppliedExperienceIDs: append([]string(nil), proposal.AppliedExperienceIDs...),
		}, -1, nil
	}
	candidates := append([]domain.HighLevelAction{proposal.Primary}, proposal.Alternatives...)
	claimConflict := false
	for index, candidate := range candidates {
		if validateActionForSnapshot(candidate, input.Snapshot) != nil {
			continue
		}
		if coordination.ValidateChoice(input.Coordination, candidate) != nil {
			claimConflict = true
			continue
		}
		alternatives := make([]domain.HighLevelAction, 0, len(candidates)-1)
		for otherIndex, alternative := range candidates {
			if otherIndex == index || validateActionForSnapshot(alternative, input.Snapshot) != nil || coordination.ValidateChoice(input.Coordination, alternative) != nil {
				continue
			}
			alternatives = append(alternatives, alternative)
		}
		return domain.ActionDecision{
			Action: candidate, Confidence: confidence, Alternatives: alternatives,
			AppliedExperienceIDs: append([]string(nil), proposal.AppliedExperienceIDs...),
		}, index, nil
	}
	reason := "no proposed action passed deterministic safety checks"
	if claimConflict {
		reason = "player is already handling every safe proposed target"
	}
	stop := domain.HighLevelAction{
		SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID, SnapshotVersion: input.Snapshot.SnapshotVersion,
		Kind: domain.ActionStopSession, Reason: reason,
	}
	return domain.ActionDecision{
		Action: stop, Confidence: confidence,
		AppliedExperienceIDs: append([]string(nil), proposal.AppliedExperienceIDs...),
	}, -1, nil
}

func policyConfidence(proposal domain.ActionProposal, input intelligence.ActionInput) float64 {
	confidence := proposal.ModelConfidence
	for _, trait := range input.PlayerModel.Traits {
		if trait.ObservationCount >= 2 {
			confidence += 0.10
			break
		}
	}
	applied := make(map[string]struct{}, len(proposal.AppliedExperienceIDs))
	for _, id := range proposal.AppliedExperienceIDs {
		applied[id] = struct{}{}
	}
	for _, item := range input.ApplicableExperiences {
		if _, ok := applied[item.ID]; ok && (item.ObservationCount >= 2 || item.Source == domain.ExperienceSourceCorrection) {
			confidence += 0.15
			break
		}
	}
	confidence -= 0.20 * float64(len(proposal.UncertaintyCodes))
	confidence = math.Max(0, math.Min(1, confidence))
	return math.Round(confidence*100) / 100
}

func decisionFromRecord(record domain.DecisionRecord) domain.ActionDecision {
	decision := domain.ActionDecision{Action: record.FinalAction, Confidence: record.PolicyConfidence}
	if record.Proposal != nil {
		decision.AppliedExperienceIDs = append([]string(nil), record.Proposal.AppliedExperienceIDs...)
		if record.SafeAlternatives != nil {
			decision.Alternatives = append([]domain.HighLevelAction(nil), record.SafeAlternatives...)
		} else {
			candidates := append([]domain.HighLevelAction{record.Proposal.Primary}, record.Proposal.Alternatives...)
			for index, candidate := range candidates {
				if index != record.SelectedCandidate {
					decision.Alternatives = append(decision.Alternatives, candidate)
				}
			}
		}
	}
	return decision
}

func (s *Service) saveDecision(ctx context.Context, snapshot domain.WorldSnapshot, collaborationContext domain.CoordinationContext, proposal domain.ActionProposal, decision domain.ActionDecision, selected int) (domain.ActionDecision, error) {
	proposalCopy := proposal
	proposalCopy.Alternatives = append([]domain.HighLevelAction(nil), proposal.Alternatives...)
	proposalCopy.UncertaintyCodes = append([]domain.UncertaintyCode(nil), proposal.UncertaintyCodes...)
	proposalCopy.AppliedExperienceIDs = append([]string(nil), proposal.AppliedExperienceIDs...)
	record := domain.DecisionRecord{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Day: snapshot.Day, ModelRevision: collaborationContext.ModelRevision,
		InferredIntent:       collaborationContext.InferredIntent,
		PlayerClaimedTargets: append([]string(nil), collaborationContext.PlayerClaimedTargets...),
		CandidateAction:      proposal.Primary, FinalAction: decision.Action,
		Proposal: &proposalCopy, PolicyConfidence: decision.Confidence, SelectedCandidate: selected,
		SafeAlternatives: append([]domain.HighLevelAction{}, decision.Alternatives...),
	}
	if err := s.store.SaveDecision(ctx, record); err != nil {
		return domain.ActionDecision{}, fmt.Errorf("persist decision: %w", err)
	}
	stored, err := s.store.GetDecision(ctx, record.SaveID, record.SessionID, record.SnapshotVersion)
	if err != nil {
		return domain.ActionDecision{}, fmt.Errorf("reload persisted decision: %w", err)
	}
	return decisionFromRecord(stored), nil
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
