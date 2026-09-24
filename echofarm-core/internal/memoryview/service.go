package memoryview

import (
	"context"
	"errors"
	"sort"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

type store interface {
	GetPlayerModel(context.Context, string) (domain.PlayerModel, error)
	GetLatestLearningOutcome(context.Context, string) (domain.LearningOutcome, error)
	GetLatestDecision(context.Context, string) (domain.DecisionRecord, error)
	GetActiveSession(context.Context, string) (domain.EchoSessionMemory, error)
}

type Service struct {
	store store
}

func NewService(store store) (*Service, error) {
	if store == nil {
		return nil, errors.New("memory view store is required")
	}
	return &Service{store: store}, nil
}

func (s *Service) Get(ctx context.Context, saveID string) (domain.EchoMemoryView, error) {
	if saveID == "" {
		return domain.EchoMemoryView{}, errors.New("save ID is required")
	}
	model, err := s.store.GetPlayerModel(ctx, saveID)
	if err != nil {
		return domain.EchoMemoryView{}, err
	}
	view := domain.EchoMemoryView{
		SaveID: saveID, ModelRevision: model.Revision, LearnedThroughDay: model.LearnedThroughDay,
	}
	for _, trait := range model.Traits {
		if trait.ObservationCount >= 2 {
			trait.EvidenceRefs = append([]string(nil), trait.EvidenceRefs...)
			view.StableTraits = append(view.StableTraits, trait)
		}
	}
	sort.Slice(view.StableTraits, func(i, j int) bool {
		left, right := view.StableTraits[i], view.StableTraits[j]
		if left.Confidence != right.Confidence {
			return left.Confidence > right.Confidence
		}
		return left.Key < right.Key
	})

	if outcome, err := s.store.GetLatestLearningOutcome(ctx, saveID); err == nil {
		change := outcome.Change
		view.RecentLearningChange = &change
	} else if !errors.Is(err, memory.ErrNotFound) {
		return domain.EchoMemoryView{}, err
	}
	if decision, err := s.store.GetLatestDecision(ctx, saveID); err == nil {
		view.LastDecision = &decision
	} else if !errors.Is(err, memory.ErrNotFound) {
		return domain.EchoMemoryView{}, err
	}
	if session, err := s.store.GetActiveSession(ctx, saveID); err == nil {
		view.ActiveSession = &session
	} else if !errors.Is(err, memory.ErrNotFound) {
		return domain.EchoMemoryView{}, err
	}
	return view, nil
}
