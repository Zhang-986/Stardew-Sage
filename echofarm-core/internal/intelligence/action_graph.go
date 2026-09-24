package intelligence

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/cloudwego/eino/compose"
)

const actionSystemPrompt = `You are EchoFarm's runtime policy.
Choose exactly one high-level action from the current world snapshot, learned skill, and evidence-backed player model.
Reason about current targets rather than coordinates from the teaching day. Respect weather, tool capacity, energy reserve, and skill stop conditions.
Use the coordination context to complement the player's inferred intent. Never select a target claimed by the player or listed in playerClaimedTargets.
Use only: move_to, equip_tool, water_target, refill_can, harvest_target, deposit_items, stop_session.
Return only one HighLevelAction JSON object using the current saveId, sessionId, and snapshotVersion.`

const replanSystemPrompt = `You are EchoFarm's failure replanner.
The previous high-level action failed. Use its failure code and the latest current world snapshot to choose one safe recovery action.
Do not repeat an impossible action and never invent a target. Preserve the learned player's preferences when more than one recovery is valid.
Never select a target claimed by the player or listed in the coordination context's playerClaimedTargets.
Use only: move_to, equip_tool, water_target, refill_can, harvest_target, deposit_items, stop_session.
Return only one HighLevelAction JSON object using the current saveId, sessionId, and snapshotVersion.`

type ActionGraph struct {
	choose compose.Runnable[ActionInput, domain.HighLevelAction]
	replan compose.Runnable[ReplanInput, domain.HighLevelAction]
}

func NewActionGraph(generator StructuredGenerator) (*ActionGraph, error) {
	if generator == nil {
		return nil, errors.New("structured generator is required")
	}
	choose, err := compileActionChain[ActionInput](generator, actionSystemPrompt)
	if err != nil {
		return nil, fmt.Errorf("compile Eino action graph: %w", err)
	}
	replan, err := compileActionChain[ReplanInput](generator, replanSystemPrompt)
	if err != nil {
		return nil, fmt.Errorf("compile Eino replan graph: %w", err)
	}
	return &ActionGraph{choose: choose, replan: replan}, nil
}

func (g *ActionGraph) ChooseAction(ctx context.Context, input ActionInput) (domain.HighLevelAction, error) {
	return g.choose.Invoke(ctx, input)
}

func (g *ActionGraph) Replan(ctx context.Context, input ReplanInput) (domain.HighLevelAction, error) {
	return g.replan.Invoke(ctx, input)
}

func compileActionChain[I any](generator StructuredGenerator, prompt string) (compose.Runnable[I, domain.HighLevelAction], error) {
	chain := compose.NewChain[I, domain.HighLevelAction]()
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input I) (domain.HighLevelAction, error) {
		var action domain.HighLevelAction
		if err := generator.GenerateJSON(ctx, prompt, input, &action); err != nil {
			if errors.Is(err, ErrModelUnavailable) {
				return domain.HighLevelAction{}, err
			}
			return domain.HighLevelAction{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		if err := action.Validate(); err != nil {
			return domain.HighLevelAction{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		return action, nil
	}))
	return chain.Compile(context.Background())
}
