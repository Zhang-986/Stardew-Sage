package intelligence

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/cloudwego/eino/compose"
)

type ReflectionGraph struct {
	runnable compose.Runnable[ReflectionInput, domain.ExperienceObservation]
}

func NewReflectionGraph(generator StructuredGenerator) (*ReflectionGraph, error) {
	if generator == nil {
		return nil, errors.New("structured generator is required")
	}
	chain := compose.NewChain[ReflectionInput, domain.ExperienceObservation]()
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input ReflectionInput) (domain.ExperienceObservation, error) {
		if err := validateReflectionInput(input); err != nil {
			return domain.ExperienceObservation{}, err
		}
		var observation domain.ExperienceObservation
		if err := generator.GenerateJSON(ctx, reflectionSystemPrompt, input, &observation); err != nil {
			if errors.Is(err, ErrModelUnavailable) || errors.Is(err, ErrModelBudgetExceeded) {
				return domain.ExperienceObservation{}, err
			}
			return domain.ExperienceObservation{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		evidence := map[string]struct{}{input.EvidenceRef: {}}
		if err := observation.Validate(evidence, snapshotTargetIDs(input.Snapshot)); err != nil {
			return domain.ExperienceObservation{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		if input.Result != nil && observation.Trigger != domain.ExperienceTrigger(input.Result.ErrorCode) {
			return domain.ExperienceObservation{}, fmt.Errorf("%w: reflection trigger does not match action result", ErrInvalidModelOutput)
		}
		if input.Correction != nil && observation.Trigger != domain.ExperiencePlayerCorrection {
			return domain.ExperienceObservation{}, fmt.Errorf("%w: correction reflection must use player_correction trigger", ErrInvalidModelOutput)
		}
		return observation, nil
	}))
	runnable, err := chain.Compile(context.Background())
	if err != nil {
		return nil, fmt.Errorf("compile Eino reflection graph: %w", err)
	}
	return &ReflectionGraph{runnable: runnable}, nil
}

func (g *ReflectionGraph) Reflect(ctx context.Context, input ReflectionInput) (domain.ExperienceObservation, error) {
	return g.runnable.Invoke(ctx, input)
}

func validateReflectionInput(input ReflectionInput) error {
	if input.EvidenceRef == "" {
		return errors.New("reflection evidence reference is required")
	}
	if err := input.Snapshot.Validate(); err != nil {
		return fmt.Errorf("reflection snapshot: %w", err)
	}
	if err := input.PlayerModel.Validate(); err != nil {
		return fmt.Errorf("reflection player model: %w", err)
	}
	if input.PlayerModel.SaveID != input.Snapshot.SaveID {
		return errors.New("reflection player model belongs to another save")
	}
	if (input.Result == nil) == (input.Correction == nil) {
		return errors.New("reflection requires exactly one result or correction")
	}
	if input.Result != nil {
		if err := input.Result.Validate(); err != nil {
			return fmt.Errorf("reflection result: %w", err)
		}
		if input.Result.Status != domain.ActionFailed {
			return errors.New("reflection result must be a failure")
		}
		if input.Result.SaveID != input.Snapshot.SaveID || input.Result.SessionID != input.Snapshot.SessionID || input.Result.SnapshotVersion >= input.Snapshot.SnapshotVersion {
			return errors.New("reflection result does not precede the current snapshot")
		}
	}
	if input.Correction != nil {
		if err := input.Correction.Validate(); err != nil {
			return fmt.Errorf("reflection correction: %w", err)
		}
		if input.Correction.SaveID != input.Snapshot.SaveID || input.Correction.SessionID != input.Snapshot.SessionID || input.Correction.Snapshot.SnapshotVersion != input.Snapshot.SnapshotVersion {
			return errors.New("reflection correction does not match current snapshot")
		}
	}
	for i, experience := range input.ExistingExperiences {
		if err := experience.Validate(); err != nil {
			return fmt.Errorf("existing experience %d: %w", i, err)
		}
		if experience.SaveID != input.Snapshot.SaveID {
			return fmt.Errorf("existing experience %d belongs to another save", i)
		}
	}
	return nil
}

func snapshotTargetIDs(snapshot domain.WorldSnapshot) map[string]struct{} {
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
