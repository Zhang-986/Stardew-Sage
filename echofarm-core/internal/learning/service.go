package learning

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/modeling"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/trace"
)

type learner interface {
	Learn(ctx context.Context, input intelligence.LearningInput) (intelligence.LearningInference, error)
}

type Service struct {
	store   memory.LearningStore
	learner learner
}

func NewService(store memory.LearningStore, learner learner) (*Service, error) {
	if store == nil {
		return nil, errors.New("memory store is required")
	}
	if learner == nil {
		return nil, errors.New("learner is required")
	}
	return &Service{store: store, learner: learner}, nil
}

func (s *Service) Teach(ctx context.Context, demonstration domain.Demonstration) (domain.PlayerModel, domain.SkillProgram, error) {
	outcome, err := s.TeachOutcome(ctx, demonstration)
	return outcome.PlayerModel, outcome.Skill, err
}

func (s *Service) TeachOutcome(ctx context.Context, demonstration domain.Demonstration) (domain.LearningOutcome, error) {
	if err := demonstration.Validate(); err != nil {
		return domain.LearningOutcome{}, fmt.Errorf("validate demonstration: %w", err)
	}
	stored, err := s.store.GetLearningOutcome(ctx, demonstration.SaveID, demonstration.ID)
	if err == nil {
		return stored, nil
	}
	if !errors.Is(err, memory.ErrNotFound) {
		return domain.LearningOutcome{}, fmt.Errorf("load learning outcome: %w", err)
	}
	segments := trace.Segment(demonstration.Events)
	if len(segments) == 0 {
		return domain.LearningOutcome{}, errors.New("demonstration contains no learnable behavior")
	}

	var existing *domain.PlayerModel
	current, err := s.store.GetPlayerModel(ctx, demonstration.SaveID)
	if err == nil {
		existing = &current
	} else if !errors.Is(err, memory.ErrNotFound) {
		return domain.LearningOutcome{}, fmt.Errorf("load existing player model: %w", err)
	}

	result, err := s.learner.Learn(ctx, intelligence.LearningInput{
		Demonstration: demonstration,
		Segments:      segments,
		ExistingModel: existing,
	})
	if err != nil {
		return domain.LearningOutcome{}, err
	}
	model, change, err := modeling.Merge(existing, demonstration, result.Observations)
	if err != nil {
		return domain.LearningOutcome{}, fmt.Errorf("merge player model: %w", err)
	}
	result.Skill.Revision = model.Revision
	if err := result.Skill.Validate(); err != nil {
		return domain.LearningOutcome{}, fmt.Errorf("validate learned skill: %w", err)
	}
	outcome := domain.LearningOutcome{Demonstration: demonstration, PlayerModel: model, Skill: result.Skill, Change: change}
	if err := s.store.SaveLearningOutcome(ctx, outcome); err != nil {
		return domain.LearningOutcome{}, fmt.Errorf("persist learning result: %w", err)
	}
	stored, err = s.store.GetLearningOutcome(ctx, demonstration.SaveID, demonstration.ID)
	if err != nil {
		return domain.LearningOutcome{}, fmt.Errorf("reload persisted learning result: %w", err)
	}
	return stored, nil
}
