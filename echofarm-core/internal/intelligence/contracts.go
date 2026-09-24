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
	Snapshot              domain.WorldSnapshot       `json:"snapshot"`
	PlayerModel           domain.PlayerModel         `json:"playerModel"`
	Skill                 domain.SkillProgram        `json:"skill"`
	Coordination          domain.CoordinationContext `json:"coordination"`
	ApplicableExperiences []domain.PolicyExperience  `json:"applicableExperiences,omitempty"`
}

type IntentInput struct {
	SaveID      string                  `json:"saveId"`
	Day         int                     `json:"day"`
	TimeOfDay   int                     `json:"timeOfDay"`
	Activities  []domain.PlayerActivity `json:"activities"`
	PlayerModel domain.PlayerModel      `json:"playerModel"`
}

type IntentInference struct {
	Intent            domain.PlayerIntent `json:"intent"`
	EvidenceTargetIDs []string            `json:"evidenceTargetIds,omitempty"`
}

type IntentInferer interface {
	InferIntent(ctx context.Context, input IntentInput) (domain.PlayerIntent, error)
}

type ReplanInput struct {
	ActionInput ActionInput         `json:"actionInput"`
	LastResult  domain.ActionResult `json:"lastResult"`
}

type ReflectionInput struct {
	Snapshot            domain.WorldSnapshot      `json:"snapshot"`
	PlayerModel         domain.PlayerModel        `json:"playerModel"`
	ExistingExperiences []domain.PolicyExperience `json:"existingExperiences,omitempty"`
	Result              *domain.ActionResult      `json:"result,omitempty"`
	Correction          *domain.PlayerCorrection  `json:"correction,omitempty"`
	EvidenceRef         string                    `json:"evidenceRef"`
}

type Intelligence interface {
	Learn(ctx context.Context, input LearningInput) (LearningInference, error)
	ChooseAction(ctx context.Context, input ActionInput) (domain.HighLevelAction, error)
	Replan(ctx context.Context, input ReplanInput) (domain.HighLevelAction, error)
}

type StructuredGenerator interface {
	GenerateJSON(ctx context.Context, systemPrompt string, input any, output any) error
}

type GenerationUsage struct {
	Reported         bool
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type UsageReportingGenerator interface {
	GenerateJSONWithUsage(ctx context.Context, systemPrompt string, input any, output any) (GenerationUsage, error)
}
