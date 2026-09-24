package memory

import (
	"context"
	"errors"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

var ErrNotFound = errors.New("memory not found")

type Store interface {
	SaveLearning(ctx context.Context, demonstration domain.Demonstration, model domain.PlayerModel, skill domain.SkillProgram) error
	GetDemonstration(ctx context.Context, saveID, demonstrationID string) (domain.Demonstration, error)
	GetPlayerModel(ctx context.Context, saveID string) (domain.PlayerModel, error)
	GetSkill(ctx context.Context, saveID, skillName string) (domain.SkillProgram, error)
}

type LearningStore interface {
	Store
	SaveLearningOutcome(ctx context.Context, outcome domain.LearningOutcome) error
	GetLearningOutcome(ctx context.Context, saveID, demonstrationID string) (domain.LearningOutcome, error)
}
