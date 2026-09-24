package memory

import (
	"context"
	"encoding/json"
	"math"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestSQLiteActionResultAppendsExperienceFeedbackAtomically(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	first := seedFeedbackExperience(t, store, "source-1", "chest-east")
	second := seedFeedbackExperience(t, store, "source-2", "chest-west")
	record := feedbackDecision(7, []string{first.ID, second.ID})
	if err := store.SaveDecision(ctx, record); err != nil {
		t.Fatal(err)
	}
	result := domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: record.FinalAction, Status: domain.ActionSucceeded,
	}

	attached, err := store.AttachDecisionResult(ctx, result, snapshotAfter(record))
	if err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v", attached, err)
	}
	want := []feedbackRow{
		{experienceID: first.ID, outcome: domain.ExperienceFeedbackSucceeded, errorCode: ""},
		{experienceID: second.ID, outcome: domain.ExperienceFeedbackSucceeded, errorCode: ""},
	}
	if got := readFeedbackRows(t, store, record.SaveID); !reflect.DeepEqual(got, want) {
		t.Fatalf("feedback rows = %+v, want %+v", got, want)
	}
	attached, err = store.AttachDecisionResult(ctx, result, snapshotAfter(record))
	if err != nil || attached {
		t.Fatalf("duplicate AttachDecisionResult() = %v, %v", attached, err)
	}
	conflicting := result
	conflicting.Status = domain.ActionFailed
	conflicting.ErrorCode = "inventory_full"
	if _, err := store.AttachDecisionResult(ctx, conflicting, snapshotAfter(record)); err == nil {
		t.Fatal("conflicting result error = nil")
	}
	if got := readFeedbackRows(t, store, record.SaveID); !reflect.DeepEqual(got, want) {
		t.Fatalf("feedback rows after retries = %+v, want %+v", got, want)
	}
}

func TestSQLiteActionResultRollsBackWhenAppliedExperienceIsMissing(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	record := feedbackDecision(7, []string{"exp-missing"})
	if err := store.SaveDecision(ctx, record); err != nil {
		t.Fatal(err)
	}
	result := domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: record.FinalAction, Status: domain.ActionSucceeded,
	}

	if _, err := store.AttachDecisionResult(ctx, result, snapshotAfter(record)); err == nil {
		t.Fatal("AttachDecisionResult() error = nil, want missing experience")
	}
	stored, err := store.GetDecision(ctx, record.SaveID, record.SessionID, record.SnapshotVersion)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Result != nil {
		t.Fatalf("result persisted after feedback failure: %+v", stored.Result)
	}
}

func TestSQLiteActionResultSkipsFeedbackWhenAlternativeWasSelected(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	experience := seedFeedbackExperience(t, store, "source-1", "chest-east")
	record := feedbackDecision(7, []string{experience.ID})
	record.SelectedCandidate = 1
	record.Proposal.Alternatives = []domain.HighLevelAction{record.FinalAction}
	record.Proposal.Primary = record.FinalAction
	record.Proposal.Primary.TargetID = "crop-primary"
	if err := store.SaveDecision(ctx, record); err != nil {
		t.Fatal(err)
	}
	result := domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: record.FinalAction, Status: domain.ActionSucceeded,
	}
	if attached, err := store.AttachDecisionResult(ctx, result, snapshotAfter(record)); err != nil || !attached {
		t.Fatalf("AttachDecisionResult() = %v, %v", attached, err)
	}
	if got := readFeedbackRows(t, store, record.SaveID); len(got) != 0 {
		t.Fatalf("alternative feedback rows = %+v, want none", got)
	}
}

func TestSQLiteConcurrentActionResultsAppendOneCanonicalFeedback(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	experience := seedFeedbackExperience(t, store, "source-1", "chest-east")
	record := feedbackDecision(7, []string{experience.ID})
	if err := store.SaveDecision(ctx, record); err != nil {
		t.Fatal(err)
	}
	succeeded := domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: record.FinalAction, Status: domain.ActionSucceeded,
	}
	contradicted := succeeded
	contradicted.Status = domain.ActionFailed
	contradicted.ErrorCode = "inventory_full"
	type outcome struct {
		attached bool
		err      error
	}
	start := make(chan struct{})
	results := make(chan outcome, 16)
	for index := range 16 {
		result := succeeded
		if index%2 == 1 {
			result = contradicted
		}
		go func() {
			<-start
			attached, err := store.AttachDecisionResult(ctx, result, snapshotAfter(record))
			results <- outcome{attached: attached, err: err}
		}()
	}
	close(start)
	winners := 0
	for range 16 {
		if result := <-results; result.attached {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("canonical result winners = %d, want 1", winners)
	}
	stored, err := store.GetDecision(ctx, record.SaveID, record.SessionID, record.SnapshotVersion)
	if err != nil || stored.Result == nil {
		t.Fatalf("GetDecision() result/error = %+v/%v", stored.Result, err)
	}
	rows := readFeedbackRows(t, store, record.SaveID)
	if len(rows) != 1 {
		t.Fatalf("feedback rows = %+v, want one", rows)
	}
	wantOutcome := domain.ExperienceFeedbackSucceeded
	if stored.Result.Status == domain.ActionFailed {
		wantOutcome = domain.ExperienceFeedbackContradicted
	}
	if rows[0].outcome != wantOutcome {
		t.Fatalf("feedback outcome = %q, want %q", rows[0].outcome, wantOutcome)
	}
}

func TestSQLiteExperienceFeedbackProjectionSurvivesReopenAndBaseRewrite(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "echo.db")
	store, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	experience := seedFeedbackExperience(t, store, "source-1", "chest-east")
	results := []struct {
		version   int64
		status    domain.ActionStatus
		errorCode string
	}{
		{version: 7, status: domain.ActionSucceeded},
		{version: 8, status: domain.ActionFailed, errorCode: "inventory_full"},
		{version: 9, status: domain.ActionFailed, errorCode: "path_blocked"},
	}
	for _, item := range results {
		record := feedbackDecision(item.version, []string{experience.ID})
		if err := store.SaveDecision(ctx, record); err != nil {
			t.Fatal(err)
		}
		result := domain.ActionResult{
			SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
			Action: record.FinalAction, Status: item.status, ErrorCode: item.errorCode,
		}
		if attached, err := store.AttachDecisionResult(ctx, result, snapshotAfter(record)); err != nil || !attached {
			t.Fatalf("AttachDecisionResult(%d) = %v, %v", item.version, attached, err)
		}
	}
	assertFeedbackProjection(t, store, experience.ID, 1, 1, 1, 0.6)

	var raw []byte
	if err := store.db.QueryRow(`SELECT payload_json FROM policy_experiences WHERE save_id=? AND experience_id=?`, experience.SaveID, experience.ID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var base domain.PolicyExperience
	if err := json.Unmarshal(raw, &base); err != nil {
		t.Fatal(err)
	}
	if base.SuccessCount != 0 || base.FailureCount != 0 || base.NeutralCount != 0 || base.EffectiveConfidence != 0 {
		t.Fatalf("base experience contains projection: %+v", base)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	assertFeedbackProjection(t, store, experience.ID, 1, 1, 1, 0.6)
	rewrite := experienceOutcome("farm-a", "source-rewrite", domain.ExperienceSourceFailure, "chest-east")
	rewrite.Experience.SuccessCount = 99
	rewrite.Experience.EffectiveConfidence = 0.99
	if err := store.SaveExperienceOutcome(ctx, rewrite); err != nil {
		t.Fatal(err)
	}
	assertFeedbackProjection(t, store, experience.ID, 1, 1, 1, 0.6)
}

func assertFeedbackProjection(t *testing.T, store *SQLite, experienceID string, success, failure, neutral int, effective float64) {
	t.Helper()
	experiences, err := store.ListPolicyExperiences(context.Background(), "farm-a")
	if err != nil {
		t.Fatal(err)
	}
	for _, experience := range experiences {
		if experience.ID != experienceID {
			continue
		}
		if experience.SuccessCount != success || experience.FailureCount != failure || experience.NeutralCount != neutral ||
			math.Abs(experience.EffectiveConfidence-effective) > 1e-9 {
			t.Fatalf("feedback projection = %+v", experience)
		}
		return
	}
	t.Fatalf("experience %q not found in %+v", experienceID, experiences)
}

type feedbackRow struct {
	experienceID string
	outcome      domain.ExperienceFeedbackOutcome
	errorCode    string
}

func readFeedbackRows(t *testing.T, store *SQLite, saveID string) []feedbackRow {
	t.Helper()
	rows, err := store.db.Query(`
SELECT experience_id, outcome, error_code FROM experience_feedback
WHERE save_id=? ORDER BY experience_id`, saveID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	result := make([]feedbackRow, 0)
	for rows.Next() {
		var row feedbackRow
		if err := rows.Scan(&row.experienceID, &row.outcome, &row.errorCode); err != nil {
			t.Fatal(err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func seedFeedbackExperience(t *testing.T, store *SQLite, sourceID, targetID string) domain.PolicyExperience {
	t.Helper()
	outcome := experienceOutcome("farm-a", sourceID, domain.ExperienceSourceFailure, targetID)
	if err := store.SaveExperienceOutcome(context.Background(), outcome); err != nil {
		t.Fatal(err)
	}
	return outcome.Experience
}

func feedbackDecision(version int64, experienceIDs []string) domain.DecisionRecord {
	record := decisionRecord(version, "crop-free")
	record.Proposal = &domain.ActionProposal{
		Primary: record.FinalAction, ModelConfidence: 0.8,
		AppliedExperienceIDs: append([]string(nil), experienceIDs...),
	}
	record.SelectedCandidate = 0
	return record
}
