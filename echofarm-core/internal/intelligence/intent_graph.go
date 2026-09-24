package intelligence

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/cloudwego/eino/compose"
)

type IntentGraph struct {
	runnable compose.Runnable[IntentInput, IntentInference]
}

func NewIntentGraph(generator StructuredGenerator) (*IntentGraph, error) {
	if generator == nil {
		return nil, errors.New("structured generator is required")
	}
	chain := compose.NewChain[IntentInput, IntentInference]()
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input IntentInput) (IntentInference, error) {
		var inference IntentInference
		if err := generator.GenerateJSON(ctx, intentSystemPrompt, input, &inference); err != nil {
			if errors.Is(err, ErrModelUnavailable) || errors.Is(err, ErrModelBudgetExceeded) {
				return IntentInference{}, err
			}
			return IntentInference{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		if err := validateIntentInference(input, inference); err != nil {
			return IntentInference{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		return inference, nil
	}))
	runnable, err := chain.Compile(context.Background())
	if err != nil {
		return nil, fmt.Errorf("compile Eino intent graph: %w", err)
	}
	return &IntentGraph{runnable: runnable}, nil
}

func (g *IntentGraph) InferIntent(ctx context.Context, input IntentInput) (domain.PlayerIntent, error) {
	inference, err := g.runnable.Invoke(ctx, input)
	return inference.Intent, err
}

func validateIntentInference(input IntentInput, inference IntentInference) error {
	if !validIntent(inference.Intent) {
		return fmt.Errorf("unsupported player intent %q", inference.Intent)
	}
	available := make(map[string]struct{}, len(input.Activities))
	for _, activity := range input.Activities {
		available[activity.TargetID] = struct{}{}
	}
	if inference.Intent != domain.PlayerIntentUnknown && len(inference.EvidenceTargetIDs) == 0 {
		return errors.New("non-unknown intent requires target evidence")
	}
	for _, targetID := range inference.EvidenceTargetIDs {
		if _, ok := available[targetID]; !ok {
			return fmt.Errorf("intent cites unknown target %q", targetID)
		}
	}
	return nil
}

func validIntent(intent domain.PlayerIntent) bool {
	switch intent {
	case domain.PlayerIntentUnknown, domain.PlayerIntentWatering, domain.PlayerIntentHarvesting, domain.PlayerIntentDepositing:
		return true
	default:
		return false
	}
}
