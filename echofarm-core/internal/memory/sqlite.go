package memory

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS demonstrations (
  save_id TEXT NOT NULL,
  demonstration_id TEXT NOT NULL,
  payload_json BLOB NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (save_id, demonstration_id)
);
CREATE TABLE IF NOT EXISTS player_models (
  save_id TEXT PRIMARY KEY,
  revision INTEGER NOT NULL,
  payload_json BLOB NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS skills (
  save_id TEXT NOT NULL,
  skill_name TEXT NOT NULL,
  revision INTEGER NOT NULL,
  payload_json BLOB NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (save_id, skill_name)
);
CREATE TABLE IF NOT EXISTS learning_revisions (
  save_id TEXT NOT NULL,
  demonstration_id TEXT NOT NULL,
  model_revision INTEGER NOT NULL,
  outcome_json BLOB NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (save_id, demonstration_id)
);
CREATE TABLE IF NOT EXISTS echo_sessions (
  save_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  day INTEGER NOT NULL,
  status TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (save_id, session_id)
);
CREATE TABLE IF NOT EXISTS decision_records (
  save_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  snapshot_version INTEGER NOT NULL,
  payload_json BLOB NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (save_id, session_id, snapshot_version)
);
CREATE TABLE IF NOT EXISTS policy_experiences (
  save_id TEXT NOT NULL,
  experience_id TEXT NOT NULL,
  payload_json BLOB NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (save_id, experience_id)
);
CREATE TABLE IF NOT EXISTS experience_revisions (
  save_id TEXT NOT NULL,
  source_id TEXT NOT NULL,
  experience_id TEXT NOT NULL,
  outcome_json BLOB NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (save_id, source_id)
);
CREATE TABLE IF NOT EXISTS player_corrections (
  save_id TEXT NOT NULL,
  correction_id TEXT NOT NULL,
  payload_json BLOB NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (save_id, correction_id)
);
CREATE TABLE IF NOT EXISTS reflection_jobs (
  save_id TEXT NOT NULL,
  source_id TEXT NOT NULL,
  payload_json BLOB NOT NULL,
  status TEXT NOT NULL,
  lease_token TEXT,
  lease_until TEXT,
  attempt_count INTEGER NOT NULL,
  last_error TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (save_id, source_id)
);
CREATE TABLE IF NOT EXISTS experience_feedback (
  save_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  snapshot_version INTEGER NOT NULL,
  experience_id TEXT NOT NULL,
  outcome TEXT NOT NULL,
  error_code TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (save_id, session_id, snapshot_version, experience_id)
);
CREATE TABLE IF NOT EXISTS model_usage (
  request_id TEXT PRIMARY KEY,
  save_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  day INTEGER NOT NULL,
  purpose TEXT NOT NULL,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  status TEXT NOT NULL,
  prompt_tokens INTEGER,
  completion_tokens INTEGER,
  total_tokens INTEGER,
  latency_ms INTEGER NOT NULL,
  error_class TEXT NOT NULL,
  call_budget INTEGER NOT NULL,
  token_budget INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS model_usage_session_idx
  ON model_usage(save_id, session_id, started_at);
CREATE INDEX IF NOT EXISTS model_usage_day_idx
  ON model_usage(save_id, day, started_at);
`

type SQLite struct {
	db  *sql.DB
	now func() time.Time
}

func OpenSQLite(path string) (*SQLite, error) {
	if path == "" {
		return nil, errors.New("SQLite path is required")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate SQLite: %w", err)
	}
	return &SQLite{db: db, now: time.Now}, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) ReserveModelCall(ctx context.Context, record domain.ModelCallRecord) error {
	if err := record.Validate(); err != nil {
		return fmt.Errorf("model call reservation: %w", err)
	}
	if record.Status != domain.ModelCallStarted {
		return errors.New("model call reservation must have started status")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin model call reservation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var calls, reportedTokens int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*), COALESCE(SUM(total_tokens), 0)
FROM model_usage WHERE save_id=? AND session_id=?`, record.SaveID, record.SessionID).Scan(&calls, &reportedTokens); err != nil {
		return fmt.Errorf("read model budget: %w", err)
	}
	if calls >= record.CallBudget {
		return fmt.Errorf("%w: session call limit %d reached", ErrModelBudgetExceeded, record.CallBudget)
	}
	if reportedTokens >= record.TokenBudget {
		return fmt.Errorf("%w: session reported-token limit %d reached", ErrModelBudgetExceeded, record.TokenBudget)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO model_usage(
  request_id, save_id, session_id, day, purpose, started_at, finished_at, status,
  prompt_tokens, completion_tokens, total_tokens, latency_ms, error_class, call_budget, token_budget
) VALUES (?, ?, ?, ?, ?, ?, NULL, ?, NULL, NULL, NULL, 0, '', ?, ?)`,
		record.RequestID, record.SaveID, record.SessionID, record.Day, record.Purpose,
		record.StartedAt.UTC().Format(time.RFC3339Nano), record.Status, record.CallBudget, record.TokenBudget); err != nil {
		return fmt.Errorf("reserve model call: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit model call reservation: %w", err)
	}
	return nil
}

func (s *SQLite) CompleteModelCall(ctx context.Context, record domain.ModelCallRecord) error {
	if err := record.Validate(); err != nil {
		return fmt.Errorf("model call completion: %w", err)
	}
	if record.Status != domain.ModelCallSucceeded && record.Status != domain.ModelCallFailed {
		return errors.New("model call completion must have a terminal status")
	}
	write, err := s.db.ExecContext(ctx, `
UPDATE model_usage SET
  finished_at=?, status=?, prompt_tokens=?, completion_tokens=?, total_tokens=?, latency_ms=?, error_class=?
WHERE request_id=? AND save_id=? AND session_id=? AND status=?`,
		record.FinishedAt.UTC().Format(time.RFC3339Nano), record.Status,
		nullableModelTokens(record.PromptTokens), nullableModelTokens(record.CompletionTokens), nullableModelTokens(record.TotalTokens),
		record.LatencyMS, record.ErrorClass, record.RequestID, record.SaveID, record.SessionID, domain.ModelCallStarted)
	if err != nil {
		return fmt.Errorf("complete model call: %w", err)
	}
	affected, err := write.RowsAffected()
	if err != nil {
		return fmt.Errorf("read completed model call count: %w", err)
	}
	if affected != 1 {
		return errors.New("model call reservation was not found or already completed")
	}
	return nil
}

func (s *SQLite) GetModelUsageSummary(ctx context.Context, saveID, sessionID string, day int) (domain.ModelUsageSummary, error) {
	if saveID == "" || sessionID == "" || day < 0 {
		return domain.ModelUsageSummary{}, errors.New("model usage save, session, and day are required")
	}
	var callBudget, tokenBudget int
	err := s.db.QueryRowContext(ctx, `
SELECT call_budget, token_budget FROM model_usage
WHERE save_id=? AND session_id=? ORDER BY started_at DESC, request_id DESC LIMIT 1`, saveID, sessionID).Scan(&callBudget, &tokenBudget)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ModelUsageSummary{}, ErrNotFound
	}
	if err != nil {
		return domain.ModelUsageSummary{}, fmt.Errorf("read model usage limits: %w", err)
	}
	session, err := s.modelUsageTotals(ctx, `save_id=? AND session_id=?`, saveID, sessionID)
	if err != nil {
		return domain.ModelUsageSummary{}, err
	}
	dayTotals, err := s.modelUsageTotals(ctx, `save_id=? AND day=?`, saveID, day)
	if err != nil {
		return domain.ModelUsageSummary{}, err
	}
	return domain.ModelUsageSummary{
		SaveID: saveID, SessionID: sessionID, Day: day,
		Session: session, DayTotals: dayTotals,
		CallBudget: callBudget, TokenBudget: tokenBudget,
		BudgetExhausted: session.Calls >= callBudget || session.TotalTokens >= tokenBudget,
	}, nil
}

func (s *SQLite) modelUsageTotals(ctx context.Context, where string, arguments ...any) (domain.ModelUsageTotals, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT status, prompt_tokens, completion_tokens, total_tokens, latency_ms, finished_at
FROM model_usage WHERE `+where+` ORDER BY started_at, request_id`, arguments...)
	if err != nil {
		return domain.ModelUsageTotals{}, fmt.Errorf("read model usage totals: %w", err)
	}
	defer rows.Close()
	var totals domain.ModelUsageTotals
	var completed int64
	var latencyTotal int64
	for rows.Next() {
		var status domain.ModelCallStatus
		var prompt, completion, total sql.NullInt64
		var latency int64
		var finished sql.NullString
		if err := rows.Scan(&status, &prompt, &completion, &total, &latency, &finished); err != nil {
			return domain.ModelUsageTotals{}, fmt.Errorf("scan model usage totals: %w", err)
		}
		totals.Calls++
		switch status {
		case domain.ModelCallSucceeded:
			totals.Succeeded++
		case domain.ModelCallFailed:
			totals.Failed++
		}
		if total.Valid {
			totals.ReportedTokenCalls++
			totals.PromptTokens += int(prompt.Int64)
			totals.CompletionTokens += int(completion.Int64)
			totals.TotalTokens += int(total.Int64)
		}
		if finished.Valid {
			completed++
			latencyTotal += latency
			totals.LastLatencyMS = latency
		}
	}
	if err := rows.Err(); err != nil {
		return domain.ModelUsageTotals{}, fmt.Errorf("iterate model usage totals: %w", err)
	}
	totals.TokensKnown = totals.Calls > 0 && totals.ReportedTokenCalls == totals.Calls
	if completed > 0 {
		totals.AverageLatencyMS = latencyTotal / completed
	}
	return totals, nil
}

func nullableModelTokens(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func (s *SQLite) SaveLearning(ctx context.Context, demonstration domain.Demonstration, model domain.PlayerModel, skill domain.SkillProgram) error {
	return s.SaveLearningOutcome(ctx, domain.LearningOutcome{
		Demonstration: demonstration,
		PlayerModel:   model,
		Skill:         skill,
		Change: domain.LearningChange{
			ModelRevision: model.Revision,
			Kind:          domain.LearningChangeUnchanged,
			Summary:       "legacy learning commit",
		},
	})
}

func (s *SQLite) SaveLearningOutcome(ctx context.Context, outcome domain.LearningOutcome) error {
	demonstration, model, skill := outcome.Demonstration, outcome.PlayerModel, outcome.Skill
	if err := demonstration.Validate(); err != nil {
		return fmt.Errorf("demonstration: %w", err)
	}
	if err := model.Validate(); err != nil {
		return fmt.Errorf("player model: %w", err)
	}
	if err := skill.Validate(); err != nil {
		return fmt.Errorf("skill: %w", err)
	}
	if demonstration.SaveID != model.SaveID {
		return errors.New("demonstration and player model save IDs do not match")
	}
	if outcome.Change.ModelRevision != model.Revision {
		return errors.New("learning change revision does not match player model")
	}

	demoJSON, err := json.Marshal(demonstration)
	if err != nil {
		return fmt.Errorf("marshal demonstration: %w", err)
	}
	modelJSON, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("marshal player model: %w", err)
	}
	skillJSON, err := json.Marshal(skill)
	if err != nil {
		return fmt.Errorf("marshal skill: %w", err)
	}
	outcomeJSON, err := json.Marshal(outcome)
	if err != nil {
		return fmt.Errorf("marshal learning outcome: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin learning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var alreadyStored int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM learning_revisions WHERE save_id=? AND demonstration_id=?`, demonstration.SaveID, demonstration.ID).Scan(&alreadyStored)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check learning revision: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `
INSERT INTO demonstrations(save_id, demonstration_id, payload_json, created_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(save_id, demonstration_id) DO UPDATE SET payload_json=excluded.payload_json`,
		demonstration.SaveID, demonstration.ID, demoJSON, now); err != nil {
		return fmt.Errorf("save demonstration: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO player_models(save_id, revision, payload_json, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(save_id) DO UPDATE SET
  revision=excluded.revision, payload_json=excluded.payload_json, updated_at=excluded.updated_at
WHERE excluded.revision > player_models.revision`,
		model.SaveID, model.Revision, modelJSON, now); err != nil {
		return fmt.Errorf("save player model: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO skills(save_id, skill_name, revision, payload_json, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(save_id, skill_name) DO UPDATE SET
  revision=excluded.revision, payload_json=excluded.payload_json, updated_at=excluded.updated_at
WHERE excluded.revision > skills.revision`,
		demonstration.SaveID, skill.Name, skill.Revision, skillJSON, now); err != nil {
		return fmt.Errorf("save skill: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO learning_revisions(save_id, demonstration_id, model_revision, outcome_json, created_at)
VALUES (?, ?, ?, ?, ?)`, demonstration.SaveID, demonstration.ID, model.Revision, outcomeJSON, now); err != nil {
		return fmt.Errorf("save learning revision: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit learning transaction: %w", err)
	}
	return nil
}

func (s *SQLite) GetLearningOutcome(ctx context.Context, saveID, demonstrationID string) (domain.LearningOutcome, error) {
	var value domain.LearningOutcome
	err := scanJSON(s.db.QueryRowContext(ctx,
		`SELECT outcome_json FROM learning_revisions WHERE save_id=? AND demonstration_id=?`, saveID, demonstrationID), &value)
	return value, err
}

func (s *SQLite) GetLatestLearningOutcome(ctx context.Context, saveID string) (domain.LearningOutcome, error) {
	var value domain.LearningOutcome
	err := scanJSON(s.db.QueryRowContext(ctx, `
SELECT outcome_json FROM learning_revisions
WHERE save_id=? ORDER BY model_revision DESC LIMIT 1`, saveID), &value)
	return value, err
}

func (s *SQLite) SaveDecision(ctx context.Context, record domain.DecisionRecord) error {
	if record.SaveID == "" || record.SessionID == "" || record.Day <= 0 || record.SnapshotVersion < 0 {
		return errors.New("decision identity, day, and snapshot version are required")
	}
	if err := record.CandidateAction.Validate(); err != nil {
		return fmt.Errorf("candidate action: %w", err)
	}
	if err := record.FinalAction.Validate(); err != nil {
		return fmt.Errorf("final action: %w", err)
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal decision: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin decision transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `
INSERT INTO decision_records(save_id, session_id, snapshot_version, payload_json, created_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(save_id, session_id, snapshot_version) DO NOTHING`,
		record.SaveID, record.SessionID, record.SnapshotVersion, payload, now); err != nil {
		return fmt.Errorf("save decision: %w", err)
	}
	var canonical domain.DecisionRecord
	if err := scanJSON(tx.QueryRowContext(ctx, `
SELECT payload_json FROM decision_records
WHERE save_id=? AND session_id=? AND snapshot_version=?`,
		record.SaveID, record.SessionID, record.SnapshotVersion), &canonical); err != nil {
		return fmt.Errorf("reload canonical decision: %w", err)
	}
	status := "active"
	if canonical.FinalAction.Kind == domain.ActionStopSession {
		status = "completed"
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO echo_sessions(save_id, session_id, day, status, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(save_id, session_id) DO UPDATE SET
  day=excluded.day, status=excluded.status, updated_at=excluded.updated_at`,
		canonical.SaveID, canonical.SessionID, canonical.Day, status, now); err != nil {
		return fmt.Errorf("save echo session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit decision: %w", err)
	}
	return nil
}

func (s *SQLite) GetDecision(ctx context.Context, saveID, sessionID string, snapshotVersion int64) (domain.DecisionRecord, error) {
	var value domain.DecisionRecord
	err := scanJSON(s.db.QueryRowContext(ctx, `
SELECT payload_json FROM decision_records
WHERE save_id=? AND session_id=? AND snapshot_version=?`, saveID, sessionID, snapshotVersion), &value)
	return value, err
}

func (s *SQLite) AttachDecisionResult(ctx context.Context, result domain.ActionResult, snapshot domain.WorldSnapshot) (bool, error) {
	if err := result.Validate(); err != nil {
		return false, fmt.Errorf("validate action result: %w", err)
	}
	if err := snapshot.Validate(); err != nil {
		return false, fmt.Errorf("validate reflection snapshot: %w", err)
	}
	if result.SaveID != snapshot.SaveID || result.SessionID != snapshot.SessionID || result.SnapshotVersion >= snapshot.SnapshotVersion {
		return false, errors.New("action result does not precede reflection snapshot")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin action result transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var originalPayload []byte
	err = tx.QueryRowContext(ctx, `
SELECT payload_json FROM decision_records
WHERE save_id=? AND session_id=? AND snapshot_version=?`,
		result.SaveID, result.SessionID, result.SnapshotVersion).Scan(&originalPayload)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrNotFound
	}
	if err != nil {
		return false, fmt.Errorf("read decision for result: %w", err)
	}
	var record domain.DecisionRecord
	if err := json.Unmarshal(originalPayload, &record); err != nil {
		return false, fmt.Errorf("decode decision for result: %w", err)
	}
	finalActionJSON, _ := json.Marshal(record.FinalAction)
	reportedActionJSON, _ := json.Marshal(result.Action)
	if string(finalActionJSON) != string(reportedActionJSON) {
		return false, errors.New("action result does not match the recorded final action")
	}
	if record.Result != nil {
		existingJSON, _ := json.Marshal(record.Result)
		resultJSON, _ := json.Marshal(result)
		if string(existingJSON) == string(resultJSON) {
			return false, nil
		}
		return false, errors.New("decision result already recorded with different content")
	}
	record.Result = &result
	payload, err := json.Marshal(record)
	if err != nil {
		return false, fmt.Errorf("marshal decision result: %w", err)
	}
	write, err := tx.ExecContext(ctx, `
UPDATE decision_records SET payload_json=?
WHERE save_id=? AND session_id=? AND snapshot_version=? AND payload_json=?`,
		payload, result.SaveID, result.SessionID, result.SnapshotVersion, originalPayload)
	if err != nil {
		return false, fmt.Errorf("attach decision result: %w", err)
	}
	affected, err := write.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read attached decision result count: %w", err)
	}
	if affected == 1 {
		now := s.now().UTC().Format(time.RFC3339Nano)
		if result.Status == domain.ActionFailed {
			jobPayload, marshalErr := json.Marshal(reflectionJobPayload{Snapshot: snapshot, Result: result})
			if marshalErr != nil {
				return false, fmt.Errorf("marshal reflection job: %w", marshalErr)
			}
			sourceID := fmt.Sprintf("decision:%s:%d", result.SessionID, result.SnapshotVersion)
			if _, err := tx.ExecContext(ctx, `
INSERT INTO reflection_jobs(
  save_id, source_id, payload_json, status, lease_token, lease_until,
  attempt_count, last_error, created_at, updated_at
) VALUES (?, ?, ?, ?, NULL, NULL, 0, '', ?, ?)
ON CONFLICT(save_id, source_id) DO NOTHING`,
				result.SaveID, sourceID, jobPayload, reflectionJobPending, now, now); err != nil {
				return false, fmt.Errorf("enqueue reflection job: %w", err)
			}
		}
		if err := appendExperienceFeedback(ctx, tx, record, result, now); err != nil {
			return false, err
		}
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit action result: %w", err)
		}
		return true, nil
	}
	var canonical domain.DecisionRecord
	err = scanJSON(tx.QueryRowContext(ctx, `
SELECT payload_json FROM decision_records
WHERE save_id=? AND session_id=? AND snapshot_version=?`,
		result.SaveID, result.SessionID, result.SnapshotVersion), &canonical)
	if err != nil {
		return false, fmt.Errorf("reload decision after result conflict: %w", err)
	}
	existingJSON, _ := json.Marshal(canonical.Result)
	resultJSON, _ := json.Marshal(result)
	if canonical.Result != nil && string(existingJSON) == string(resultJSON) {
		return false, nil
	}
	return false, errors.New("decision result already recorded with different content")
}

func appendExperienceFeedback(ctx context.Context, tx *sql.Tx, record domain.DecisionRecord, result domain.ActionResult, now string) error {
	if record.Proposal == nil || record.SelectedCandidate != 0 || record.FinalAction.Kind == domain.ActionStopSession || len(record.Proposal.AppliedExperienceIDs) == 0 {
		return nil
	}
	outcome := domain.ClassifyExperienceFeedback(result)
	errorCode := domain.NormalizeExperienceFeedbackError(result)
	for _, experienceID := range record.Proposal.AppliedExperienceIDs {
		var exists int
		err := tx.QueryRowContext(ctx, `
SELECT 1 FROM policy_experiences WHERE save_id=? AND experience_id=?`,
			record.SaveID, experienceID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("applied experience %q is not persisted for this save", experienceID)
		}
		if err != nil {
			return fmt.Errorf("verify applied experience %q: %w", experienceID, err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO experience_feedback(
  save_id, session_id, snapshot_version, experience_id, outcome, error_code, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(save_id, session_id, snapshot_version, experience_id) DO NOTHING`,
			record.SaveID, record.SessionID, record.SnapshotVersion, experienceID, outcome, errorCode, now); err != nil {
			return fmt.Errorf("save experience feedback %q: %w", experienceID, err)
		}
	}
	return nil
}

func (s *SQLite) ClaimReflectionJob(ctx context.Context, saveID string, leaseDuration time.Duration) (ReflectionJobLease, bool, error) {
	if saveID == "" {
		return ReflectionJobLease{}, false, errors.New("reflection job save ID is required")
	}
	if leaseDuration <= 0 {
		return ReflectionJobLease{}, false, errors.New("reflection job lease duration must be positive")
	}
	token, err := newLeaseToken()
	if err != nil {
		return ReflectionJobLease{}, false, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ReflectionJobLease{}, false, fmt.Errorf("begin reflection job claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	now := s.now().UTC()
	var sourceID string
	var payload []byte
	var attemptCount int
	err = tx.QueryRowContext(ctx, `
SELECT source_id, payload_json, attempt_count FROM reflection_jobs
WHERE save_id=? AND (
  status=? OR (status=? AND lease_until IS NOT NULL AND lease_until<=?)
) ORDER BY created_at, source_id LIMIT 1`,
		saveID, reflectionJobPending, reflectionJobProcessing, formatReflectionLeaseTime(now)).Scan(&sourceID, &payload, &attemptCount)
	if errors.Is(err, sql.ErrNoRows) {
		return ReflectionJobLease{}, false, nil
	}
	if err != nil {
		return ReflectionJobLease{}, false, fmt.Errorf("select reflection job: %w", err)
	}
	write, err := tx.ExecContext(ctx, `
UPDATE reflection_jobs SET status=?, lease_token=?, lease_until=?, updated_at=?
WHERE save_id=? AND source_id=? AND (
  status=? OR (status=? AND lease_until IS NOT NULL AND lease_until<=?)
)`,
		reflectionJobProcessing, token, formatReflectionLeaseTime(now.Add(leaseDuration)), now.Format(time.RFC3339Nano),
		saveID, sourceID, reflectionJobPending, reflectionJobProcessing, formatReflectionLeaseTime(now))
	if err != nil {
		return ReflectionJobLease{}, false, fmt.Errorf("claim reflection job: %w", err)
	}
	affected, err := write.RowsAffected()
	if err != nil {
		return ReflectionJobLease{}, false, fmt.Errorf("read reflection claim count: %w", err)
	}
	if affected != 1 {
		return ReflectionJobLease{}, false, nil
	}
	var decoded reflectionJobPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return ReflectionJobLease{}, false, fmt.Errorf("decode reflection job: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ReflectionJobLease{}, false, fmt.Errorf("commit reflection job claim: %w", err)
	}
	return ReflectionJobLease{
		Job: ReflectionJob{
			SaveID: saveID, SourceID: sourceID, Snapshot: decoded.Snapshot,
			Result: decoded.Result, AttemptCount: attemptCount,
		},
		Token: token,
	}, true, nil
}

func (s *SQLite) CompleteReflectionJob(ctx context.Context, lease ReflectionJobLease) error {
	if err := validateReflectionLease(lease); err != nil {
		return err
	}
	now := s.now().UTC().Format(time.RFC3339Nano)
	write, err := s.db.ExecContext(ctx, `
UPDATE reflection_jobs
SET status=?, lease_token=NULL, lease_until=NULL, updated_at=?
WHERE save_id=? AND source_id=? AND status=? AND lease_token=?`,
		reflectionJobCompleted, now, lease.Job.SaveID, lease.Job.SourceID, reflectionJobProcessing, lease.Token)
	if err != nil {
		return fmt.Errorf("complete reflection job: %w", err)
	}
	affected, err := write.RowsAffected()
	if err != nil {
		return fmt.Errorf("read completed reflection job count: %w", err)
	}
	if affected != 1 {
		return ErrReflectionLeaseLost
	}
	return nil
}

func (s *SQLite) ReleaseReflectionJob(ctx context.Context, lease ReflectionJobLease, failureCode string) error {
	if err := validateReflectionLease(lease); err != nil {
		return err
	}
	if !validReflectionFailureCode(failureCode) {
		return errors.New("unsupported reflection failure code")
	}
	now := s.now().UTC().Format(time.RFC3339Nano)
	write, err := s.db.ExecContext(ctx, `
UPDATE reflection_jobs
SET status=?, lease_token=NULL, lease_until=NULL, attempt_count=attempt_count+1,
    last_error=?, updated_at=?
WHERE save_id=? AND source_id=? AND status=? AND lease_token=?`,
		reflectionJobPending, failureCode, now, lease.Job.SaveID, lease.Job.SourceID, reflectionJobProcessing, lease.Token)
	if err != nil {
		return fmt.Errorf("release reflection job: %w", err)
	}
	affected, err := write.RowsAffected()
	if err != nil {
		return fmt.Errorf("read released reflection job count: %w", err)
	}
	if affected != 1 {
		return ErrReflectionLeaseLost
	}
	return nil
}

func newLeaseToken() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate reflection lease token: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

func (s *SQLite) GetLatestDecision(ctx context.Context, saveID string) (domain.DecisionRecord, error) {
	var value domain.DecisionRecord
	err := scanJSON(s.db.QueryRowContext(ctx, `
SELECT payload_json FROM decision_records
WHERE save_id=? ORDER BY created_at DESC, snapshot_version DESC LIMIT 1`, saveID), &value)
	return value, err
}

func (s *SQLite) GetActiveSession(ctx context.Context, saveID string) (domain.EchoSessionMemory, error) {
	var value domain.EchoSessionMemory
	err := s.db.QueryRowContext(ctx, `
SELECT session_id, day, status FROM echo_sessions
WHERE save_id=? AND status='active' ORDER BY updated_at DESC LIMIT 1`, saveID).Scan(&value.SessionID, &value.Day, &value.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.EchoSessionMemory{}, ErrNotFound
	}
	if err != nil {
		return domain.EchoSessionMemory{}, fmt.Errorf("read active echo session: %w", err)
	}
	return value, nil
}

func (s *SQLite) SaveExperienceOutcome(ctx context.Context, outcome domain.ExperienceOutcome) error {
	if err := validateExperienceOutcome(outcome); err != nil {
		return err
	}
	persistedOutcome := outcome
	persistedOutcome.Experience = basePolicyExperience(outcome.Experience)
	persistedOutcome.UpdatedExperiences = make([]domain.PolicyExperience, len(outcome.UpdatedExperiences))
	for index, experience := range outcome.UpdatedExperiences {
		persistedOutcome.UpdatedExperiences[index] = basePolicyExperience(experience)
	}
	outcomeJSON, err := json.Marshal(persistedOutcome)
	if err != nil {
		return fmt.Errorf("marshal experience outcome: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin experience transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var alreadyStored int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM experience_revisions WHERE save_id=? AND source_id=?`,
		outcome.Experience.SaveID, outcome.SourceID).Scan(&alreadyStored)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check experience revision: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if outcome.Correction != nil {
		correctionJSON, err := json.Marshal(outcome.Correction)
		if err != nil {
			return fmt.Errorf("marshal player correction: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO player_corrections(save_id, correction_id, payload_json, created_at)
VALUES (?, ?, ?, ?)`, outcome.Experience.SaveID, outcome.Correction.ID, correctionJSON, now); err != nil {
			return fmt.Errorf("save player correction: %w", err)
		}
	}
	experiences := outcome.UpdatedExperiences
	if len(experiences) == 0 {
		experiences = []domain.PolicyExperience{outcome.Experience}
	}
	for _, experience := range experiences {
		experience = basePolicyExperience(experience)
		if experience.SaveID != outcome.Experience.SaveID {
			return errors.New("updated experience belongs to another save")
		}
		if err := experience.Validate(); err != nil {
			return fmt.Errorf("updated policy experience: %w", err)
		}
		experienceJSON, err := json.Marshal(experience)
		if err != nil {
			return fmt.Errorf("marshal policy experience: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO policy_experiences(save_id, experience_id, payload_json, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(save_id, experience_id) DO UPDATE SET
  payload_json=excluded.payload_json, updated_at=excluded.updated_at`,
			experience.SaveID, experience.ID, experienceJSON, now); err != nil {
			return fmt.Errorf("save policy experience: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO experience_revisions(save_id, source_id, experience_id, outcome_json, created_at)
VALUES (?, ?, ?, ?, ?)`, outcome.Experience.SaveID, outcome.SourceID, outcome.Experience.ID, outcomeJSON, now); err != nil {
		return fmt.Errorf("save experience revision: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit experience transaction: %w", err)
	}
	return nil
}

func (s *SQLite) GetExperienceOutcome(ctx context.Context, saveID, sourceID string) (domain.ExperienceOutcome, error) {
	var value domain.ExperienceOutcome
	err := scanJSON(s.db.QueryRowContext(ctx, `
SELECT outcome_json FROM experience_revisions WHERE save_id=? AND source_id=?`, saveID, sourceID), &value)
	return value, err
}

func (s *SQLite) ListPolicyExperiences(ctx context.Context, saveID string) ([]domain.PolicyExperience, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT payload_json FROM policy_experiences WHERE save_id=? ORDER BY experience_id`, saveID)
	if err != nil {
		return nil, fmt.Errorf("list policy experiences: %w", err)
	}
	result := make([]domain.PolicyExperience, 0)
	for rows.Next() {
		var experience domain.PolicyExperience
		if err := scanJSON(rows, &experience); err != nil {
			_ = rows.Close()
			return nil, err
		}
		experience = basePolicyExperience(experience)
		result = append(result, experience)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate policy experiences: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close policy experiences: %w", err)
	}
	indexes := make(map[string]int, len(result))
	for index := range result {
		indexes[result[index].ID] = index
	}
	feedbackRows, err := s.db.QueryContext(ctx, `
SELECT experience_id, outcome, COUNT(*) FROM experience_feedback
WHERE save_id=? GROUP BY experience_id, outcome ORDER BY experience_id, outcome`, saveID)
	if err != nil {
		return nil, fmt.Errorf("list experience feedback: %w", err)
	}
	defer feedbackRows.Close()
	for feedbackRows.Next() {
		var experienceID string
		var outcome domain.ExperienceFeedbackOutcome
		var count int
		if err := feedbackRows.Scan(&experienceID, &outcome, &count); err != nil {
			return nil, fmt.Errorf("scan experience feedback: %w", err)
		}
		index, ok := indexes[experienceID]
		if !ok {
			return nil, fmt.Errorf("experience feedback references missing experience %q", experienceID)
		}
		switch outcome {
		case domain.ExperienceFeedbackSucceeded:
			result[index].SuccessCount += count
		case domain.ExperienceFeedbackContradicted:
			result[index].FailureCount += count
		case domain.ExperienceFeedbackNeutral:
			result[index].NeutralCount += count
		default:
			return nil, fmt.Errorf("unsupported experience feedback outcome %q", outcome)
		}
	}
	if err := feedbackRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate experience feedback: %w", err)
	}
	for index := range result {
		effective, err := domain.EffectiveExperienceConfidence(result[index].Confidence, result[index].SuccessCount, result[index].FailureCount)
		if err != nil {
			return nil, fmt.Errorf("project experience %q: %w", result[index].ID, err)
		}
		result[index].EffectiveConfidence = effective
	}
	return result, nil
}

func basePolicyExperience(experience domain.PolicyExperience) domain.PolicyExperience {
	experience.EffectiveConfidence = 0
	experience.SuccessCount = 0
	experience.FailureCount = 0
	experience.NeutralCount = 0
	return experience
}

func (s *SQLite) GetPlayerCorrection(ctx context.Context, saveID, correctionID string) (domain.PlayerCorrection, error) {
	var value domain.PlayerCorrection
	err := scanJSON(s.db.QueryRowContext(ctx, `
SELECT payload_json FROM player_corrections WHERE save_id=? AND correction_id=?`, saveID, correctionID), &value)
	return value, err
}

func validateExperienceOutcome(outcome domain.ExperienceOutcome) error {
	if outcome.SourceID == "" || outcome.SourceID != outcome.Observation.EvidenceRef {
		return errors.New("experience source ID must match observation evidence")
	}
	if outcome.Source != domain.ExperienceSourceFailure && outcome.Source != domain.ExperienceSourceCorrection {
		return errors.New("unsupported experience source")
	}
	if err := outcome.Experience.Validate(); err != nil {
		return fmt.Errorf("policy experience: %w", err)
	}
	if outcome.Experience.Source != outcome.Source {
		return errors.New("experience source does not match outcome")
	}
	targets := make(map[string]struct{})
	if outcome.Observation.PreferredTargetID != "" {
		targets[outcome.Observation.PreferredTargetID] = struct{}{}
	}
	if err := outcome.Observation.Validate(map[string]struct{}{outcome.SourceID: {}}, targets); err != nil {
		return fmt.Errorf("experience observation: %w", err)
	}
	if outcome.Source == domain.ExperienceSourceCorrection {
		if outcome.Correction == nil {
			return errors.New("correction experience requires player correction")
		}
		if err := outcome.Correction.Validate(); err != nil {
			return err
		}
		if outcome.Correction.ID != outcome.SourceID || outcome.Correction.SaveID != outcome.Experience.SaveID {
			return errors.New("player correction does not match experience outcome")
		}
	} else if outcome.Correction != nil {
		return errors.New("failure experience cannot contain player correction")
	}
	return nil
}

func (s *SQLite) GetDemonstration(ctx context.Context, saveID, demonstrationID string) (domain.Demonstration, error) {
	var value domain.Demonstration
	err := scanJSON(s.db.QueryRowContext(ctx,
		`SELECT payload_json FROM demonstrations WHERE save_id=? AND demonstration_id=?`, saveID, demonstrationID), &value)
	return value, err
}

func (s *SQLite) GetPlayerModel(ctx context.Context, saveID string) (domain.PlayerModel, error) {
	var value domain.PlayerModel
	err := scanJSON(s.db.QueryRowContext(ctx,
		`SELECT payload_json FROM player_models WHERE save_id=?`, saveID), &value)
	return value, err
}

func (s *SQLite) GetSkill(ctx context.Context, saveID, skillName string) (domain.SkillProgram, error) {
	var value domain.SkillProgram
	err := scanJSON(s.db.QueryRowContext(ctx,
		`SELECT payload_json FROM skills WHERE save_id=? AND skill_name=?`, saveID, skillName), &value)
	return value, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJSON(row rowScanner, output any) error {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("read SQLite memory: %w", err)
	}
	if err := json.Unmarshal(payload, output); err != nil {
		return fmt.Errorf("decode SQLite memory: %w", err)
	}
	return nil
}
