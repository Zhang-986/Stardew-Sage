package intelligence

import (
	"context"
	"errors"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

var (
	ErrInvalidModelOutput = errors.New("invalid model output")
	ErrModelUnavailable   = errors.New("model unavailable")
)

type LearningInput struct {
	Demonstration domain.Demonstration     `json:"demonstration"`
	Segments      []domain.BehaviorSegment `json:"segments"`
	ExistingModel *domain.PlayerModel      `json:"existingPlayerModel,omitempty"`
}

type LearningInference struct {
	Observations []domain.TraitObservation `json:"observations"`
	Skill        domain.SkillProgram       `json:"skill"`
}

type ActionInput struct {
	Snapshot    domain.WorldSnapshot `json:"snapshot"`
	PlayerModel domain.PlayerModel   `json:"playerModel"`
	Skill       domain.SkillProgram  `json:"skill"`
}

type ReplanInput struct {
	ActionInput ActionInput         `json:"actionInput"`
	LastResult  domain.ActionResult `json:"lastResult"`
}

type Intelligence interface {
	Learn(ctx context.Context, input LearningInput) (LearningInference, error)
	ChooseAction(ctx context.Context, input ActionInput) (domain.HighLevelAction, error)
	Replan(ctx context.Context, input ReplanInput) (domain.HighLevelAction, error)
}

type StructuredGenerator interface {
	GenerateJSON(ctx context.Context, systemPrompt string, input any, output any) error
}
