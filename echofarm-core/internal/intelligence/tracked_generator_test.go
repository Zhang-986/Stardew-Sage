package intelligence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

type modelUsageStoreStub struct {
	reserved   []domain.ModelCallRecord
	completed  []domain.ModelCallRecord
	reserveErr error
}

func (s *modelUsageStoreStub) ReserveModelCall(_ context.Context, record domain.ModelCallRecord) error {
	s.reserved = append(s.reserved, record)
	return s.reserveErr
}

func (s *modelUsageStoreStub) CompleteModelCall(_ context.Context, record domain.ModelCallRecord) error {
	s.completed = append(s.completed, record)
	return nil
}

type usageReportingGeneratorStub struct {
	usage GenerationUsage
	err   error
	calls int
}

func (s *usageReportingGeneratorStub) GenerateJSON(context.Context, string, any, any) error {
	s.calls++
	return s.err
}

func (s *usageReportingGeneratorStub) GenerateJSONWithUsage(context.Context, string, any, any) (GenerationUsage, error) {
	s.calls++
	return s.usage, s.err
}

func TestTrackingGeneratorInfersEveryModelPurpose(t *testing.T) {
	baseSnapshot := domain.WorldSnapshot{SaveID: "farm-a", SessionID: "echo-4", Day: 4}
	inputs := []struct {
		name    string
		input   any
		purpose domain.ModelCallPurpose
	}{
		{name: "learning", input: LearningInput{Demonstration: domain.Demonstration{SaveID: "farm-a", SessionID: "teach-4", Day: 4}}, purpose: domain.ModelCallLearning},
		{name: "intent", input: IntentInput{SaveID: "farm-a", SessionID: "echo-4", Day: 4}, purpose: domain.ModelCallIntent},
		{name: "action", input: ActionInput{Snapshot: baseSnapshot}, purpose: domain.ModelCallAction},
		{name: "recovery", input: ReplanInput{ActionInput: ActionInput{Snapshot: baseSnapshot}}, purpose: domain.ModelCallRecovery},
		{name: "reflection", input: ReflectionInput{Snapshot: baseSnapshot}, purpose: domain.ModelCallReflection},
	}
	for _, test := range inputs {
		t.Run(test.name, func(t *testing.T) {
			store := &modelUsageStoreStub{}
			delegate := &usageReportingGeneratorStub{}
			generator := newTrackingGeneratorForTest(t, delegate, store)

			if err := generator.GenerateJSON(context.Background(), "prompt", test.input, &struct{}{}); err != nil {
				t.Fatal(err)
			}
			if len(store.reserved) != 1 || len(store.completed) != 1 {
				t.Fatalf("reserved/completed = %d/%d", len(store.reserved), len(store.completed))
			}
			record := store.completed[0]
			if record.Purpose != test.purpose || record.SaveID != "farm-a" || record.Day != 4 {
				t.Fatalf("record identity = %+v", record)
			}
		})
	}
}

func TestTrackingGeneratorRecordsReportedUsageAndLatency(t *testing.T) {
	store := &modelUsageStoreStub{}
	delegate := &usageReportingGeneratorStub{usage: GenerationUsage{
		Reported: true, PromptTokens: 20, CompletionTokens: 8, TotalTokens: 28,
	}}
	generator := newTrackingGeneratorForTest(t, delegate, store)

	if err := generator.GenerateJSON(context.Background(), "prompt", ActionInput{
		Snapshot: domain.WorldSnapshot{SaveID: "farm-a", SessionID: "echo-4", Day: 4},
	}, &struct{}{}); err != nil {
		t.Fatal(err)
	}
	record := store.completed[0]
	if record.Status != domain.ModelCallSucceeded || record.LatencyMS != 25 {
		t.Fatalf("completion = %+v", record)
	}
	if record.PromptTokens == nil || *record.PromptTokens != 20 || record.TotalTokens == nil || *record.TotalTokens != 28 {
		t.Fatalf("token usage = %+v", record)
	}
}

func TestTrackingGeneratorStoresOnlyBoundedErrorClass(t *testing.T) {
	store := &modelUsageStoreStub{}
	delegate := &usageReportingGeneratorStub{err: errors.Join(ErrModelUnavailable, errors.New("Bearer sk-provider-secret"))}
	generator := newTrackingGeneratorForTest(t, delegate, store)

	err := generator.GenerateJSON(context.Background(), "prompt", ActionInput{
		Snapshot: domain.WorldSnapshot{SaveID: "farm-a", SessionID: "echo-4", Day: 4},
	}, &struct{}{})
	if !errors.Is(err, ErrModelUnavailable) {
		t.Fatalf("GenerateJSON() error = %v", err)
	}
	if got := store.completed[0].ErrorClass; got != domain.ModelErrorUnavailable {
		t.Fatalf("error class = %q", got)
	}
}

func TestTrackingGeneratorDoesNotInvokeDelegateAfterBudgetRejection(t *testing.T) {
	store := &modelUsageStoreStub{reserveErr: memory.ErrModelBudgetExceeded}
	delegate := &usageReportingGeneratorStub{}
	generator := newTrackingGeneratorForTest(t, delegate, store)

	err := generator.GenerateJSON(context.Background(), "prompt", ActionInput{
		Snapshot: domain.WorldSnapshot{SaveID: "farm-a", SessionID: "echo-4", Day: 4},
	}, &struct{}{})
	if !errors.Is(err, ErrModelBudgetExceeded) {
		t.Fatalf("GenerateJSON() error = %v, want ErrModelBudgetExceeded", err)
	}
	if delegate.calls != 0 || len(store.completed) != 0 {
		t.Fatalf("delegate calls/completions = %d/%d", delegate.calls, len(store.completed))
	}
}

func newTrackingGeneratorForTest(t *testing.T, delegate StructuredGenerator, store modelUsageStore) *TrackingGenerator {
	t.Helper()
	generator, err := NewTrackingGenerator(delegate, store, domain.ModelBudgetLimits{
		MaxCallsPerSession: 8, MaxReportedTokensPerSession: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	current := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	generator.now = func() time.Time {
		result := current
		current = current.Add(25 * time.Millisecond)
		return result
	}
	generator.newRequestID = func() (string, error) { return "request-1", nil }
	return generator
}
