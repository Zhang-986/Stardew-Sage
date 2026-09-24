package experience

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

type reflector interface {
	Reflect(context.Context, intelligence.ReflectionInput) (domain.ExperienceObservation, error)
}

type experienceStore interface {
	GetPlayerModel(context.Context, string) (domain.PlayerModel, error)
	GetExperienceOutcome(context.Context, string, string) (domain.ExperienceOutcome, error)
	SaveExperienceOutcome(context.Context, domain.ExperienceOutcome) error
	ListPolicyExperiences(context.Context, string) ([]domain.PolicyExperience, error)
}

type Service struct {
	store     experienceStore
	reflector reflector
}

func NewService(store experienceStore, reflector reflector) (*Service, error) {
	if store == nil {
		return nil, errors.New("experience store is required")
	}
	if reflector == nil {
		return nil, errors.New("reflection graph is required")
	}
	return &Service{store: store, reflector: reflector}, nil
}

func FailureSourceID(result domain.ActionResult) string {
	return "decision:" + result.SessionID + ":" + strconv.FormatInt(result.SnapshotVersion, 10)
}

func CorrectionSourceID(correction domain.PlayerCorrection) string {
	return correction.ID
}

func (s *Service) LearnFromResult(ctx context.Context, snapshot domain.WorldSnapshot, result domain.ActionResult) (domain.ExperienceOutcome, error) {
	if err := result.Validate(); err != nil {
		return domain.ExperienceOutcome{}, fmt.Errorf("validate reflection result: %w", err)
	}
	if result.Status != domain.ActionFailed {
		return domain.ExperienceOutcome{}, errors.New("only failed actions produce reflection experience")
	}
	if err := snapshot.Validate(); err != nil {
		return domain.ExperienceOutcome{}, fmt.Errorf("validate reflection snapshot: %w", err)
	}
	if result.SaveID != snapshot.SaveID || result.SessionID != snapshot.SessionID || result.SnapshotVersion >= snapshot.SnapshotVersion {
		return domain.ExperienceOutcome{}, errors.New("reflection result does not precede current snapshot")
	}
	return s.learn(ctx, snapshot, FailureSourceID(result), domain.ExperienceSourceFailure, &result, nil)
}

func (s *Service) LearnFromCorrection(ctx context.Context, correction domain.PlayerCorrection) (domain.ExperienceOutcome, error) {
	if err := correction.Validate(); err != nil {
		return domain.ExperienceOutcome{}, fmt.Errorf("validate player correction: %w", err)
	}
	return s.learn(ctx, correction.Snapshot, CorrectionSourceID(correction), domain.ExperienceSourceCorrection, nil, &correction)
}

func (s *Service) learn(ctx context.Context, snapshot domain.WorldSnapshot, sourceID string, source domain.ExperienceSource, result *domain.ActionResult, correction *domain.PlayerCorrection) (domain.ExperienceOutcome, error) {
	stored, err := s.store.GetExperienceOutcome(ctx, snapshot.SaveID, sourceID)
	if err == nil {
		return stored, nil
	}
	if !errors.Is(err, memory.ErrNotFound) {
		return domain.ExperienceOutcome{}, fmt.Errorf("load experience outcome: %w", err)
	}
	model, err := s.store.GetPlayerModel(ctx, snapshot.SaveID)
	if err != nil {
		return domain.ExperienceOutcome{}, fmt.Errorf("load player model for reflection: %w", err)
	}
	existing, err := s.store.ListPolicyExperiences(ctx, snapshot.SaveID)
	if err != nil {
		return domain.ExperienceOutcome{}, fmt.Errorf("load policy experiences: %w", err)
	}
	observation, err := s.reflector.Reflect(ctx, intelligence.ReflectionInput{
		Snapshot: snapshot, PlayerModel: model, ExistingExperiences: existing,
		Result: result, Correction: correction, EvidenceRef: sourceID,
	})
	if err != nil {
		return domain.ExperienceOutcome{}, err
	}
	updated, learned, err := Merge(existing, snapshot.SaveID, snapshot.Day, source, observation)
	if err != nil {
		return domain.ExperienceOutcome{}, fmt.Errorf("merge policy experience: %w", err)
	}
	outcome := domain.ExperienceOutcome{
		SourceID: sourceID, Source: source, Observation: observation,
		Experience: learned, UpdatedExperiences: updated, Correction: correction,
	}
	if err := s.store.SaveExperienceOutcome(ctx, outcome); err != nil {
		return domain.ExperienceOutcome{}, fmt.Errorf("persist experience outcome: %w", err)
	}
	stored, err = s.store.GetExperienceOutcome(ctx, snapshot.SaveID, sourceID)
	if err != nil {
		return domain.ExperienceOutcome{}, fmt.Errorf("reload persisted experience outcome: %w", err)
	}
	return stored, nil
}
