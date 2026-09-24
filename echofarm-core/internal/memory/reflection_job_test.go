package memory

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestSQLiteFailedResultEnqueuesOneReflectionJob(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	record := decisionRecord(7, "crop-free")
	if err := store.SaveDecision(ctx, record); err != nil {
		t.Fatal(err)
	}
	failed := domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: record.FinalAction, Status: domain.ActionFailed, ErrorCode: "path_blocked",
	}
	snapshot := correctionSnapshot(record.SaveID)
	snapshot.SessionID = record.SessionID
	snapshot.SnapshotVersion = record.SnapshotVersion + 1

	attached, err := store.AttachDecisionResult(ctx, failed, snapshot)
	if err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v; want true, nil", attached, err)
	}
	attached, err = store.AttachDecisionResult(ctx, failed, snapshot)
	if err != nil || attached {
		t.Fatalf("duplicate AttachDecisionResult() = %v, %v; want false, nil", attached, err)
	}

	lease, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute)
	if err != nil || !ok {
		t.Fatalf("ClaimReflectionJob() = %+v, %v, %v; want a job", lease, ok, err)
	}
	if lease.Token == "" || lease.Job.SourceID != "decision:echo-day-4:7" || lease.Job.Result != failed {
		t.Fatalf("reflection lease = %+v", lease)
	}
	if lease.Job.Snapshot.SnapshotVersion != snapshot.SnapshotVersion {
		t.Fatalf("job snapshot version = %d, want %d", lease.Job.Snapshot.SnapshotVersion, snapshot.SnapshotVersion)
	}
	if _, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute); err != nil || ok {
		t.Fatalf("second ClaimReflectionJob() ok/error = %v/%v, want false/nil", ok, err)
	}
}

func TestSQLiteReflectionJobLeaseLifecycle(t *testing.T) {
	ctx := context.Background()
	store, record, failed, snapshot := seededFailedDecision(t)
	fixedNow := time.Date(2026, 9, 24, 8, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return fixedNow }
	if attached, err := store.AttachDecisionResult(ctx, failed, snapshot); err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v", attached, err)
	}

	first, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute)
	if err != nil || !ok {
		t.Fatalf("first claim = %+v, %v, %v", first, ok, err)
	}
	if _, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute); err != nil || ok {
		t.Fatalf("live lease claim ok/error = %v/%v, want false/nil", ok, err)
	}
	wrong := first
	wrong.Token = "not-the-owner"
	if err := store.CompleteReflectionJob(ctx, wrong); !errors.Is(err, ErrReflectionLeaseLost) {
		t.Fatalf("wrong-owner complete error = %v, want ErrReflectionLeaseLost", err)
	}
	if err := store.ReleaseReflectionJob(ctx, first, ReflectionFailureModelUnavailable); err != nil {
		t.Fatalf("ReleaseReflectionJob() error = %v", err)
	}

	second, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute)
	if err != nil || !ok || second.Job.AttemptCount != 1 || second.Token == first.Token {
		t.Fatalf("retry claim = %+v, %v, %v", second, ok, err)
	}
	if err := store.CompleteReflectionJob(ctx, second); err != nil {
		t.Fatalf("CompleteReflectionJob() error = %v", err)
	}
	if _, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute); err != nil || ok {
		t.Fatalf("completed job claim ok/error = %v/%v, want false/nil", ok, err)
	}
}

func TestSQLiteReflectionJobExpiredLeaseCanBeReclaimed(t *testing.T) {
	ctx := context.Background()
	store, record, failed, snapshot := seededFailedDecision(t)
	now := time.Date(2026, 9, 24, 8, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }
	if attached, err := store.AttachDecisionResult(ctx, failed, snapshot); err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v", attached, err)
	}
	stale, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute)
	if err != nil || !ok {
		t.Fatalf("first claim = %+v, %v, %v", stale, ok, err)
	}
	now = now.Add(2 * time.Minute)
	reclaimed, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute)
	if err != nil || !ok || reclaimed.Token == stale.Token {
		t.Fatalf("reclaimed lease = %+v, %v, %v", reclaimed, ok, err)
	}
	if err := store.CompleteReflectionJob(ctx, stale); !errors.Is(err, ErrReflectionLeaseLost) {
		t.Fatalf("stale complete error = %v, want ErrReflectionLeaseLost", err)
	}
	if err := store.CompleteReflectionJob(ctx, reclaimed); err != nil {
		t.Fatalf("reclaimed complete error = %v", err)
	}
}

func TestSQLiteReflectionJobComparesFractionalLeaseTimesChronologically(t *testing.T) {
	ctx := context.Background()
	store, record, failed, snapshot := seededFailedDecision(t)
	now := time.Date(2026, 9, 24, 8, 0, 0, 500_000_000, time.UTC)
	store.now = func() time.Time { return now }
	if attached, err := store.AttachDecisionResult(ctx, failed, snapshot); err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v", attached, err)
	}
	first, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, 500*time.Millisecond)
	if err != nil || !ok {
		t.Fatalf("first claim = %+v, %v, %v", first, ok, err)
	}
	now = now.Add(600 * time.Millisecond)
	second, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute)
	if err != nil || !ok || second.Token == first.Token {
		t.Fatalf("fractional-time reclaim = %+v, %v, %v", second, ok, err)
	}
}

func TestSQLiteReflectionJobConcurrentClaimsHaveOneWinner(t *testing.T) {
	ctx := context.Background()
	store, record, failed, snapshot := seededFailedDecision(t)
	if attached, err := store.AttachDecisionResult(ctx, failed, snapshot); err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v", attached, err)
	}
	type outcome struct {
		ok  bool
		err error
	}
	start := make(chan struct{})
	results := make(chan outcome, 16)
	for range 16 {
		go func() {
			<-start
			_, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute)
			results <- outcome{ok: ok, err: err}
		}()
	}
	close(start)
	winners := 0
	for range 16 {
		result := <-results
		if result.err != nil {
			t.Fatalf("ClaimReflectionJob() error = %v", result.err)
		}
		if result.ok {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("claim winners = %d, want 1", winners)
	}
}

func TestSQLiteReflectionJobSurvivesReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "echo.db")
	store, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	record := decisionRecord(7, "crop-free")
	if err := store.SaveDecision(ctx, record); err != nil {
		t.Fatal(err)
	}
	failed := failedResult(record)
	if attached, err := store.AttachDecisionResult(ctx, failed, snapshotAfter(record)); err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v", attached, err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	lease, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute)
	if err != nil || !ok || lease.Job.Result != failed {
		t.Fatalf("reopened claim = %+v, %v, %v", lease, ok, err)
	}
}

func TestSQLiteSuccessfulResultDoesNotEnqueueReflectionJob(t *testing.T) {
	ctx := context.Background()
	store, record, _, snapshot := seededFailedDecision(t)
	succeeded := domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: record.FinalAction, Status: domain.ActionSucceeded,
	}
	if attached, err := store.AttachDecisionResult(ctx, succeeded, snapshot); err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v", attached, err)
	}
	if _, ok, err := store.ClaimReflectionJob(ctx, record.SaveID, time.Minute); err != nil || ok {
		t.Fatalf("successful-result claim ok/error = %v/%v, want false/nil", ok, err)
	}
}

func TestSQLiteRejectsInvalidReflectionClaim(t *testing.T) {
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if _, _, err := store.ClaimReflectionJob(context.Background(), "", time.Minute); err == nil {
		t.Fatal("empty save ID error = nil")
	}
	if _, _, err := store.ClaimReflectionJob(context.Background(), "farm-a", 0); err == nil {
		t.Fatal("zero lease duration error = nil")
	}
}

func seededFailedDecision(t *testing.T) (*SQLite, domain.DecisionRecord, domain.ActionResult, domain.WorldSnapshot) {
	t.Helper()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	record := decisionRecord(7, "crop-free")
	if err := store.SaveDecision(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	return store, record, failedResult(record), snapshotAfter(record)
}

func failedResult(record domain.DecisionRecord) domain.ActionResult {
	return domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: record.FinalAction, Status: domain.ActionFailed, ErrorCode: "path_blocked",
	}
}
