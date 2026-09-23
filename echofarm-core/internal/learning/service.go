package learning

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/trace"
)

type learner interface {
	Learn(ctx context.Context, input intelligence.LearningInput) (intelligence.LearningResult, error)
}

type Service struct {
	store   memory.Store
	learner learner
}

func NewService(store memory.Store, learner learner) (*Service, error) {
	if store == nil {
		return nil, errors.New("memory store is required")
	}
	if learner == nil {
		return nil, errors.New("learner is required")
	}
	return &Service{store: store, learner: learner}, nil
}

func (s *Service) Teach(ctx context.Context, demonstration domain.Demonstration) (domain.PlayerModel, domain.SkillProgram, error) {
	if err := demonstration.Validate(); err != nil {
		return domain.PlayerModel{}, domain.SkillProgram{}, fmt.Errorf("validate demonstration: %w", err)
	}
	segments := trace.Segment(demonstration.Events)
	if len(segments) == 0 {
		return domain.PlayerModel{}, domain.SkillProgram{}, errors.New("demonstration contains no learnable behavior")
	}

	var existing *domain.PlayerModel
	current, err := s.store.GetPlayerModel(ctx, demonstration.SaveID)
	if err == nil {
		existing = &current
	} else if !errors.Is(err, memory.ErrNotFound) {
		return domain.PlayerModel{}, domain.SkillProgram{}, fmt.Errorf("load existing player model: %w", err)
	}

	result, err := s.learner.Learn(ctx, intelligence.LearningInput{
		Demonstration: demonstration,
		Segments:      segments,
		ExistingModel: existing,
	})
	if err != nil {
		return domain.PlayerModel{}, domain.SkillProgram{}, err
	}
	if err := result.PlayerModel.Validate(); err != nil {
		return domain.PlayerModel{}, domain.SkillProgram{}, fmt.Errorf("validate learned player model: %w", err)
	}
	if result.PlayerModel.SaveID != demonstration.SaveID {
		return domain.PlayerModel{}, domain.SkillProgram{}, errors.New("learned player model belongs to another save")
	}
	if existing != nil && result.PlayerModel.Revision <= existing.Revision {
		return domain.PlayerModel{}, domain.SkillProgram{}, errors.New("learned player model revision did not advance")
	}
	if err := result.Skill.Validate(); err != nil {
		return domain.PlayerModel{}, domain.SkillProgram{}, fmt.Errorf("validate learned skill: %w", err)
	}
	if err := s.store.SaveLearning(ctx, demonstration, result.PlayerModel, result.Skill); err != nil {
		return domain.PlayerModel{}, domain.SkillProgram{}, fmt.Errorf("persist learning result: %w", err)
	}
	return result.PlayerModel, result.Skill, nil
}
