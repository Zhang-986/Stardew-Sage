package memory

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestSQLiteModelUsagePersistsKnownAndUnknownTokensAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "echo.db")
	store, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	limits := domain.ModelBudgetLimits{MaxCallsPerSession: 4, MaxReportedTokensPerSession: 100}
	first := modelCall("request-1", "session-a", domain.ModelCallAction, started, limits)
	if err := store.ReserveModelCall(ctx, first); err != nil {
		t.Fatal(err)
	}
	prompt, completion, total := 10, 5, 15
	first.Status = domain.ModelCallSucceeded
	first.FinishedAt = started.Add(120 * time.Millisecond)
	first.LatencyMS = 120
	first.PromptTokens, first.CompletionTokens, first.TotalTokens = &prompt, &completion, &total
	if err := store.CompleteModelCall(ctx, first); err != nil {
		t.Fatal(err)
	}
	second := modelCall("request-2", "session-a", domain.ModelCallReflection, started.Add(time.Second), limits)
	if err := store.ReserveModelCall(ctx, second); err != nil {
		t.Fatal(err)
	}
	second.Status = domain.ModelCallFailed
	second.FinishedAt = second.StartedAt.Add(80 * time.Millisecond)
	second.LatencyMS = 80
	second.ErrorClass = domain.ModelErrorUnavailable
	if err := store.CompleteModelCall(ctx, second); err != nil {
		t.Fatal(err)
	}
	third := modelCall("request-3", "session-b", domain.ModelCallIntent, started.Add(2*time.Second), limits)
	if err := store.ReserveModelCall(ctx, third); err != nil {
		t.Fatal(err)
	}
	third.Status = domain.ModelCallSucceeded
	third.FinishedAt = third.StartedAt.Add(50 * time.Millisecond)
	third.LatencyMS = 50
	if err := store.CompleteModelCall(ctx, third); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	summary, err := store.GetModelUsageSummary(ctx, "farm-a", "session-a", 4)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Session.Calls != 2 || summary.Session.Succeeded != 1 || summary.Session.Failed != 1 {
		t.Fatalf("session totals = %+v", summary.Session)
	}
	if summary.Session.ReportedTokenCalls != 1 || summary.Session.TotalTokens != 15 || summary.Session.TokensKnown {
		t.Fatalf("session token totals = %+v, want partial/unknown", summary.Session)
	}
	if summary.Session.LastLatencyMS != 80 || summary.Session.AverageLatencyMS != 100 {
		t.Fatalf("session latency = %+v", summary.Session)
	}
	if summary.DayTotals.Calls != 3 || summary.DayTotals.Succeeded != 2 || summary.DayTotals.Failed != 1 {
		t.Fatalf("day totals = %+v", summary.DayTotals)
	}
	if summary.CallBudget != 4 || summary.TokenBudget != 100 || summary.BudgetExhausted {
		t.Fatalf("budget summary = %+v", summary)
	}
}

func TestSQLiteModelUsageTokenBudgetBlocksTheNextCall(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	limits := domain.ModelBudgetLimits{MaxCallsPerSession: 10, MaxReportedTokensPerSession: 50}
	record := modelCall("request-1", "session-a", domain.ModelCallLearning, time.Now().UTC(), limits)
	if err := store.ReserveModelCall(ctx, record); err != nil {
		t.Fatal(err)
	}
	prompt, completion, total := 40, 10, 50
	record.Status = domain.ModelCallSucceeded
	record.FinishedAt = record.StartedAt.Add(time.Millisecond)
	record.LatencyMS = 1
	record.PromptTokens, record.CompletionTokens, record.TotalTokens = &prompt, &completion, &total
	if err := store.CompleteModelCall(ctx, record); err != nil {
		t.Fatal(err)
	}

	err = store.ReserveModelCall(ctx, modelCall("request-2", "session-a", domain.ModelCallAction, time.Now().UTC(), limits))
	if !errors.Is(err, ErrModelBudgetExceeded) {
		t.Fatalf("ReserveModelCall() error = %v, want ErrModelBudgetExceeded", err)
	}
}

func TestSQLiteConcurrentModelBudgetReservationsDoNotOvershoot(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	limits := domain.ModelBudgetLimits{MaxCallsPerSession: 3, MaxReportedTokensPerSession: 1000}
	start := make(chan struct{})
	results := make(chan error, 16)
	var group sync.WaitGroup
	for index := range 16 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			results <- store.ReserveModelCall(ctx, modelCall(
				"request-"+string(rune('a'+index)), "session-a", domain.ModelCallAction, time.Now().UTC(), limits,
			))
		}()
	}
	close(start)
	group.Wait()
	close(results)
	reserved, rejected := 0, 0
	for err := range results {
		switch {
		case err == nil:
			reserved++
		case errors.Is(err, ErrModelBudgetExceeded):
			rejected++
		default:
			t.Fatalf("unexpected reservation error: %v", err)
		}
	}
	if reserved != 3 || rejected != 13 {
		t.Fatalf("reserved/rejected = %d/%d, want 3/13", reserved, rejected)
	}
}

func TestSQLiteModelUsageSchemaStoresNoPromptResponseOrAPIKey(t *testing.T) {
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	rows, err := store.db.Query(`PRAGMA table_info(model_usage)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns := map[string]bool{}
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, kind string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		columns[name] = true
	}
	for _, forbidden := range []string{"prompt", "response", "api_key", "provider_error"} {
		if columns[forbidden] {
			t.Fatalf("model_usage contains forbidden column %q", forbidden)
		}
	}
}

func modelCall(requestID, sessionID string, purpose domain.ModelCallPurpose, started time.Time, limits domain.ModelBudgetLimits) domain.ModelCallRecord {
	return domain.ModelCallRecord{
		RequestID: requestID, SaveID: "farm-a", SessionID: sessionID, Day: 4,
		Purpose: purpose, StartedAt: started, Status: domain.ModelCallStarted,
		CallBudget: limits.MaxCallsPerSession, TokenBudget: limits.MaxReportedTokensPerSession,
	}
}
