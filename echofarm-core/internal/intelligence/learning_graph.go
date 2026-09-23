package intelligence

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/compose"
)

type LearningGraph struct {
	runnable compose.Runnable[LearningInput, LearningResult]
}

func NewLearningGraph(generator StructuredGenerator) (*LearningGraph, error) {
	if generator == nil {
		return nil, errors.New("structured generator is required")
	}

	chain := compose.NewChain[LearningInput, LearningResult]()
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input LearningInput) (LearningResult, error) {
		var result LearningResult
		if err := generator.GenerateJSON(ctx, learningSystemPrompt, input, &result); err != nil {
			if errors.Is(err, ErrModelUnavailable) {
				return LearningResult{}, err
			}
			return LearningResult{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		if err := validateLearningResult(input, result); err != nil {
			return LearningResult{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		return result, nil
	}))

	runnable, err := chain.Compile(context.Background())
	if err != nil {
		return nil, fmt.Errorf("compile Eino learning graph: %w", err)
	}
	return &LearningGraph{runnable: runnable}, nil
}

func (g *LearningGraph) Learn(ctx context.Context, input LearningInput) (LearningResult, error) {
	return g.runnable.Invoke(ctx, input)
}

func validateLearningResult(input LearningInput, result LearningResult) error {
	if err := result.PlayerModel.Validate(); err != nil {
		return fmt.Errorf("player model: %w", err)
	}
	if result.PlayerModel.SaveID != input.Demonstration.SaveID {
		return errors.New("player model save_id does not match demonstration")
	}
	if err := result.Skill.Validate(); err != nil {
		return fmt.Errorf("skill: %w", err)
	}

	evidence := make(map[string]struct{}, len(input.Demonstration.Events))
	for _, event := range input.Demonstration.Events {
		evidence[event.ID] = struct{}{}
	}
	checkEvidence := func(owner string, ids []string) error {
		if len(ids) == 0 {
			return fmt.Errorf("%s has no evidence", owner)
		}
		for _, id := range ids {
			if _, ok := evidence[id]; !ok {
				return fmt.Errorf("%s cites unknown event %q", owner, id)
			}
		}
		return nil
	}
	for i, preference := range result.PlayerModel.Preferences {
		if err := checkEvidence(fmt.Sprintf("preference %d", i), preference.EvidenceEventIDs); err != nil {
			return err
		}
	}
	return checkEvidence("skill", result.Skill.EvidenceEventIDs)
}
