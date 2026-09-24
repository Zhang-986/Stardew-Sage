package intelligence

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

var ErrModelBudgetExceeded = memory.ErrModelBudgetExceeded

type modelUsageStore interface {
	ReserveModelCall(context.Context, domain.ModelCallRecord) error
	CompleteModelCall(context.Context, domain.ModelCallRecord) error
}

type TrackingGenerator struct {
	delegate     StructuredGenerator
	store        modelUsageStore
	limits       domain.ModelBudgetLimits
	now          func() time.Time
	newRequestID func() (string, error)
}

func NewTrackingGenerator(delegate StructuredGenerator, store modelUsageStore, limits domain.ModelBudgetLimits) (*TrackingGenerator, error) {
	if delegate == nil {
		return nil, errors.New("structured generator is required")
	}
	if store == nil {
		return nil, errors.New("model usage store is required")
	}
	if limits.MaxCallsPerSession <= 0 || limits.MaxReportedTokensPerSession <= 0 {
		return nil, errors.New("model budget limits must be positive")
	}
	return &TrackingGenerator{
		delegate: delegate, store: store, limits: limits,
		now: time.Now, newRequestID: newModelRequestID,
	}, nil
}

func (g *TrackingGenerator) GenerateJSON(ctx context.Context, systemPrompt string, input any, output any) error {
	saveID, sessionID, day, purpose, err := modelCallIdentity(input)
	if err != nil {
		return err
	}
	requestID, err := g.newRequestID()
	if err != nil {
		return fmt.Errorf("generate model request ID: %w", err)
	}
	started := g.now().UTC()
	record := domain.ModelCallRecord{
		RequestID: requestID, SaveID: saveID, SessionID: sessionID, Day: day,
		Purpose: purpose, StartedAt: started, Status: domain.ModelCallStarted,
		CallBudget: g.limits.MaxCallsPerSession, TokenBudget: g.limits.MaxReportedTokensPerSession,
	}
	if err := g.store.ReserveModelCall(ctx, record); err != nil {
		if errors.Is(err, memory.ErrModelBudgetExceeded) {
			return fmt.Errorf("%w: %v", ErrModelBudgetExceeded, err)
		}
		return fmt.Errorf("reserve model call usage: %w", err)
	}

	var usage GenerationUsage
	if reporting, ok := g.delegate.(UsageReportingGenerator); ok {
		usage, err = reporting.GenerateJSONWithUsage(ctx, systemPrompt, input, output)
	} else {
		err = g.delegate.GenerateJSON(ctx, systemPrompt, input, output)
	}
	finished := g.now().UTC()
	record.FinishedAt = finished
	record.LatencyMS = finished.Sub(started).Milliseconds()
	if record.LatencyMS < 0 {
		record.LatencyMS = 0
	}
	if usage.Reported {
		record.PromptTokens = intPointer(usage.PromptTokens)
		record.CompletionTokens = intPointer(usage.CompletionTokens)
		record.TotalTokens = intPointer(usage.TotalTokens)
	}
	if err == nil {
		record.Status = domain.ModelCallSucceeded
	} else {
		record.Status = domain.ModelCallFailed
		record.ErrorClass = modelErrorClass(err)
	}
	if completionErr := g.store.CompleteModelCall(context.WithoutCancel(ctx), record); completionErr != nil {
		return errors.Join(err, fmt.Errorf("complete model call usage: %w", completionErr))
	}
	return err
}

func modelCallIdentity(input any) (string, string, int, domain.ModelCallPurpose, error) {
	switch value := input.(type) {
	case LearningInput:
		return value.Demonstration.SaveID, value.Demonstration.SessionID, value.Demonstration.Day, domain.ModelCallLearning, nil
	case IntentInput:
		return value.SaveID, value.SessionID, value.Day, domain.ModelCallIntent, nil
	case ActionInput:
		return value.Snapshot.SaveID, value.Snapshot.SessionID, value.Snapshot.Day, domain.ModelCallAction, nil
	case ReplanInput:
		snapshot := value.ActionInput.Snapshot
		return snapshot.SaveID, snapshot.SessionID, snapshot.Day, domain.ModelCallRecovery, nil
	case ReflectionInput:
		return value.Snapshot.SaveID, value.Snapshot.SessionID, value.Snapshot.Day, domain.ModelCallReflection, nil
	default:
		return "", "", 0, "", fmt.Errorf("unsupported tracked model input %T", input)
	}
}

func modelErrorClass(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return domain.ModelErrorCanceled
	case errors.Is(err, context.DeadlineExceeded):
		return domain.ModelErrorDeadline
	case errors.Is(err, ErrModelUnavailable):
		return domain.ModelErrorUnavailable
	case errors.Is(err, ErrInvalidModelOutput):
		return domain.ModelErrorInvalidOutput
	default:
		return domain.ModelErrorInternal
	}
}

func newModelRequestID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func intPointer(value int) *int {
	return &value
}
