package learning

import (
	"context"
	"errors"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

type learnerStub struct {
	result intelligence.LearningResult
	err    error
	input  intelligence.LearningInput
	calls  int
}

func (s *learnerStub) Learn(_ context.Context, input intelligence.LearningInput) (intelligence.LearningResult, error) {
	s.calls++
	s.input = input
	return s.result, s.err
}

type storeStub struct {
	currentModel domain.PlayerModel
	loadErr      error
	saveCalls    int
	savedDemo    domain.Demonstration
	savedModel   domain.PlayerModel
	savedSkill   domain.SkillProgram
}

func (s *storeStub) SaveLearning(_ context.Context, demo domain.Demonstration, model domain.PlayerModel, skill domain.SkillProgram) error {
	s.saveCalls++
	s.savedDemo, s.savedModel, s.savedSkill = demo, model, skill
	return nil
}

func (s *storeStub) GetDemonstration(context.Context, string, string) (domain.Demonstration, error) {
	return domain.Demonstration{}, memory.ErrNotFound
}

func (s *storeStub) GetPlayerModel(context.Context, string) (domain.PlayerModel, error) {
	return s.currentModel, s.loadErr
}

func (s *storeStub) GetSkill(context.Context, string, string) (domain.SkillProgram, error) {
	return domain.SkillProgram{}, memory.ErrNotFound
}

func TestTeachSegmentsLearnsAndPersistsAtomically(t *testing.T) {
	demo := teachingDemo()
	existing := domain.PlayerModel{SaveID: demo.SaveID, Revision: 1, EnergyReserve: 30}
	result := validResult(demo.SaveID, 2)
	store := &storeStub{currentModel: existing}
	learner := &learnerStub{result: result}
	service, err := NewService(store, learner)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	model, skill, err := service.Teach(context.Background(), demo)
	if err != nil {
		t.Fatalf("Teach() error = %v", err)
	}
	if learner.calls != 1 || learner.input.ExistingModel == nil || learner.input.ExistingModel.Revision != 1 {
		t.Fatalf("learner input = %+v, calls = %d", learner.input, learner.calls)
	}
	if len(learner.input.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(learner.input.Segments))
	}
	if store.saveCalls != 1 || store.savedDemo.ID != demo.ID {
		t.Fatalf("save calls = %d, saved demo = %+v", store.saveCalls, store.savedDemo)
	}
	if model.Revision != 2 || skill.Name != "morning-farm-routine" {
		t.Fatalf("Teach() = (%+v, %+v)", model, skill)
	}
}

func TestTeachAllowsFirstObservationWithoutExistingModel(t *testing.T) {
	demo := teachingDemo()
	store := &storeStub{loadErr: memory.ErrNotFound}
	learner := &learnerStub{result: validResult(demo.SaveID, 1)}
	service, err := NewService(store, learner)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, _, err := service.Teach(context.Background(), demo); err != nil {
		t.Fatalf("Teach() error = %v", err)
	}
	if learner.input.ExistingModel != nil {
		t.Fatalf("existing model = %+v, want nil", learner.input.ExistingModel)
	}
}

func TestTeachRejectsInvalidAIOutputBeforeWriting(t *testing.T) {
	demo := teachingDemo()
	result := validResult(demo.SaveID, 1)
	result.Skill.Steps[0].Action = "teleport"
	store := &storeStub{loadErr: memory.ErrNotFound}
	service, err := NewService(store, &learnerStub{result: result})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, _, err := service.Teach(context.Background(), demo); err == nil {
		t.Fatal("Teach() error = nil, want invalid model output error")
	}
	if store.saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", store.saveCalls)
	}
}

func TestTeachDoesNotReplaceMemoryWhenAIUnavailable(t *testing.T) {
	demo := teachingDemo()
	store := &storeStub{loadErr: memory.ErrNotFound}
	service, err := NewService(store, &learnerStub{err: intelligence.ErrModelUnavailable})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, _, err := service.Teach(context.Background(), demo); !errors.Is(err, intelligence.ErrModelUnavailable) {
		t.Fatalf("Teach() error = %v, want ErrModelUnavailable", err)
	}
	if store.saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", store.saveCalls)
	}
}

func teachingDemo() domain.Demonstration {
	return domain.Demonstration{
		ID: "demo-1", SaveID: "farm-1", SessionID: "teach-1", StartedAt: 1, EndedAt: 4,
		Events: []domain.DemonstrationEvent{
			{ID: "move-1", Kind: domain.EventMove, Tick: 1, Success: true},
			{ID: "water-1", Kind: domain.EventWater, Tick: 2, TargetID: "crop-1", Success: true},
			{ID: "harvest-1", Kind: domain.EventHarvest, Tick: 3, TargetID: "crop-2", Success: true},
		},
	}
}

func validResult(saveID string, revision int) intelligence.LearningResult {
	return intelligence.LearningResult{
		PlayerModel: domain.PlayerModel{SaveID: saveID, Revision: revision, EnergyReserve: 40},
		Skill: domain.SkillProgram{
			Name: "morning-farm-routine", Revision: revision, Goal: "care for crops", TargetSelector: "actionable_crops",
			Steps:             []domain.SkillStep{{Action: domain.ActionWaterTarget, TargetSelector: "dry_crops"}},
			SuccessConditions: []string{"all crops cared for"}, EvidenceEventIDs: []string{"water-1"},
		},
	}
}
