package memory

import (
	"context"
	"errors"
	"time"

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

type DecisionStore interface {
	GetPlayerModel(ctx context.Context, saveID string) (domain.PlayerModel, error)
	GetSkill(ctx context.Context, saveID, skillName string) (domain.SkillProgram, error)
	ListPolicyExperiences(ctx context.Context, saveID string) ([]domain.PolicyExperience, error)
	SaveDecision(ctx context.Context, record domain.DecisionRecord) error
	GetDecision(ctx context.Context, saveID, sessionID string, snapshotVersion int64) (domain.DecisionRecord, error)
	AttachDecisionResult(ctx context.Context, result domain.ActionResult, snapshot domain.WorldSnapshot) (bool, error)
}

type ReflectionJobStore interface {
	ClaimReflectionJob(ctx context.Context, saveID string, leaseDuration time.Duration) (ReflectionJobLease, bool, error)
	CompleteReflectionJob(ctx context.Context, lease ReflectionJobLease) error
	ReleaseReflectionJob(ctx context.Context, lease ReflectionJobLease, failureCode string) error
}

type ExperienceStore interface {
	GetPlayerModel(ctx context.Context, saveID string) (domain.PlayerModel, error)
	SaveExperienceOutcome(ctx context.Context, outcome domain.ExperienceOutcome) error
	GetExperienceOutcome(ctx context.Context, saveID, sourceID string) (domain.ExperienceOutcome, error)
	ListPolicyExperiences(ctx context.Context, saveID string) ([]domain.PolicyExperience, error)
	GetPlayerCorrection(ctx context.Context, saveID, correctionID string) (domain.PlayerCorrection, error)
}
