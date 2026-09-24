package intelligence

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/compose"
)

type LearningGraph struct {
	runnable compose.Runnable[LearningInput, LearningInference]
}

func NewLearningGraph(generator StructuredGenerator) (*LearningGraph, error) {
	if generator == nil {
		return nil, errors.New("structured generator is required")
	}

	chain := compose.NewChain[LearningInput, LearningInference]()
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input LearningInput) (LearningInference, error) {
		var result LearningInference
		if err := generator.GenerateJSON(ctx, learningSystemPrompt, input, &result); err != nil {
			if errors.Is(err, ErrModelUnavailable) {
				return LearningInference{}, err
			}
			return LearningInference{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		if err := validateLearningInference(input, result); err != nil {
			return LearningInference{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		return result, nil
	}))

	runnable, err := chain.Compile(context.Background())
	if err != nil {
		return nil, fmt.Errorf("compile Eino learning graph: %w", err)
	}
	return &LearningGraph{runnable: runnable}, nil
}

func (g *LearningGraph) Learn(ctx context.Context, input LearningInput) (LearningInference, error) {
	return g.runnable.Invoke(ctx, input)
}

func validateLearningInference(input LearningInput, result LearningInference) error {
	if len(result.Observations) == 0 {
		return errors.New("trait observations are required")
	}
	if err := result.Skill.Validate(); err != nil {
		return fmt.Errorf("skill: %w", err)
	}

	evidence := make(map[string]struct{}, len(input.Demonstration.Events))
	for _, event := range input.Demonstration.Events {
		evidence[event.ID] = struct{}{}
	}
	for i, observation := range result.Observations {
		if err := observation.Validate(evidence); err != nil {
			return fmt.Errorf("trait observation %d: %w", i, err)
		}
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
	return checkEvidence("skill", result.Skill.EvidenceEventIDs)
}
