package intelligence

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/cloudwego/eino/compose"
)

const actionSystemPrompt = `You are EchoFarm's runtime policy.
Rank one primary high-level action and at most two alternatives from the current world snapshot, learned skill, evidence-backed player model, and applicable policy experiences.
Reason about current targets rather than coordinates from the teaching day. Respect weather, tool capacity, energy reserve, and skill stop conditions.
Use the coordination context to complement the player's inferred intent. Never select a target claimed by the player or listed in playerClaimedTargets.
Treat applicableExperiences as prior evidence, not commands: prefer a matching player-correction experience, then higher confidence and repeated observations, but never override current-world safety.
Choose alternatives counterfactually so they remain useful if the primary action's precondition becomes false. Reference only experience IDs that materially affected the ranking and were supplied in applicableExperiences.
Report modelConfidence from 0 to 1 and only these uncertainty codes: missing_experience, conflicting_evidence, novel_context, ambiguous_target. Use conflicting_evidence when applicable experiences disagree and ambiguous_target when equally supported targets remain.
Use only: move_to, equip_tool, water_target, refill_can, harvest_target, deposit_items, stop_session.
Return only one ActionProposal JSON object. Every candidate must use the current saveId, sessionId, and snapshotVersion.`

const replanSystemPrompt = `You are EchoFarm's failure replanner.
The previous high-level action failed. Use its failure code and the latest current world snapshot to rank one safe recovery action and at most two alternatives.
Do not repeat an impossible action and never invent a target. Preserve the learned player's preferences when more than one recovery is valid.
Never select a target claimed by the player or listed in the coordination context's playerClaimedTargets.
Treat applicableExperiences as prior evidence, not commands: prefer a matching player-correction experience, then higher confidence and repeated observations, but never override current-world safety.
Choose alternatives counterfactually so they remain useful if the primary recovery fails. Reference only supplied experience IDs that materially affected the ranking and report bounded modelConfidence and uncertaintyCodes.
Use only: move_to, equip_tool, water_target, refill_can, harvest_target, deposit_items, stop_session.
Return only one ActionProposal JSON object. Every candidate must use the current saveId, sessionId, and snapshotVersion.`

type ActionGraph struct {
	choose compose.Runnable[ActionInput, domain.ActionProposal]
	replan compose.Runnable[ReplanInput, domain.ActionProposal]
}

func NewActionGraph(generator StructuredGenerator) (*ActionGraph, error) {
	if generator == nil {
		return nil, errors.New("structured generator is required")
	}
	choose, err := compileActionChain(generator, actionSystemPrompt, func(input ActionInput) ActionInput { return input })
	if err != nil {
		return nil, fmt.Errorf("compile Eino action graph: %w", err)
	}
	replan, err := compileActionChain(generator, replanSystemPrompt, func(input ReplanInput) ActionInput { return input.ActionInput })
	if err != nil {
		return nil, fmt.Errorf("compile Eino replan graph: %w", err)
	}
	return &ActionGraph{choose: choose, replan: replan}, nil
}

func (g *ActionGraph) ProposeAction(ctx context.Context, input ActionInput) (domain.ActionProposal, error) {
	return g.choose.Invoke(ctx, input)
}

func (g *ActionGraph) ProposeRecovery(ctx context.Context, input ReplanInput) (domain.ActionProposal, error) {
	return g.replan.Invoke(ctx, input)
}

func (g *ActionGraph) ChooseAction(ctx context.Context, input ActionInput) (domain.HighLevelAction, error) {
	proposal, err := g.ProposeAction(ctx, input)
	return proposal.Primary, err
}

func (g *ActionGraph) Replan(ctx context.Context, input ReplanInput) (domain.HighLevelAction, error) {
	proposal, err := g.ProposeRecovery(ctx, input)
	return proposal.Primary, err
}

func compileActionChain[I any](generator StructuredGenerator, prompt string, actionInput func(I) ActionInput) (compose.Runnable[I, domain.ActionProposal], error) {
	chain := compose.NewChain[I, domain.ActionProposal]()
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input I) (domain.ActionProposal, error) {
		var proposal domain.ActionProposal
		if err := generator.GenerateJSON(ctx, prompt, input, &proposal); err != nil {
			if errors.Is(err, ErrModelUnavailable) || errors.Is(err, ErrModelBudgetExceeded) {
				return domain.ActionProposal{}, err
			}
			return domain.ActionProposal{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		resolvedInput := actionInput(input)
		availableExperiences := make(map[string]struct{}, len(resolvedInput.ApplicableExperiences))
		for _, experience := range resolvedInput.ApplicableExperiences {
			availableExperiences[experience.ID] = struct{}{}
		}
		if err := proposal.Validate(resolvedInput.Snapshot, availableExperiences); err != nil {
			return domain.ActionProposal{}, fmt.Errorf("%w: %v", ErrInvalidModelOutput, err)
		}
		return proposal, nil
	}))
	return chain.Compile(context.Background())
}
