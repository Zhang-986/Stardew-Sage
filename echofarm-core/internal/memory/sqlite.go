package memory

import (
	"context"
	"database/sql"
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
);`

type SQLite struct {
	db *sql.DB
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
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
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
	status := "active"
	if record.FinalAction.Kind == domain.ActionStopSession {
		status = "completed"
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO echo_sessions(save_id, session_id, day, status, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(save_id, session_id) DO UPDATE SET
  day=excluded.day, status=excluded.status, updated_at=excluded.updated_at`,
		record.SaveID, record.SessionID, record.Day, status, now); err != nil {
		return fmt.Errorf("save echo session: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO decision_records(save_id, session_id, snapshot_version, payload_json, created_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(save_id, session_id, snapshot_version) DO NOTHING`,
		record.SaveID, record.SessionID, record.SnapshotVersion, payload, now); err != nil {
		return fmt.Errorf("save decision: %w", err)
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

func (s *SQLite) AttachDecisionResult(ctx context.Context, result domain.ActionResult) error {
	record, err := s.GetDecision(ctx, result.SaveID, result.SessionID, result.SnapshotVersion)
	if err != nil {
		return err
	}
	finalActionJSON, _ := json.Marshal(record.FinalAction)
	reportedActionJSON, _ := json.Marshal(result.Action)
	if string(finalActionJSON) != string(reportedActionJSON) {
		return errors.New("action result does not match the recorded final action")
	}
	if record.Result != nil {
		existingJSON, _ := json.Marshal(record.Result)
		resultJSON, _ := json.Marshal(result)
		if string(existingJSON) == string(resultJSON) {
			return nil
		}
		return errors.New("decision result already recorded with different content")
	}
	record.Result = &result
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal decision result: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `
UPDATE decision_records SET payload_json=?
WHERE save_id=? AND session_id=? AND snapshot_version=?`,
		payload, result.SaveID, result.SessionID, result.SnapshotVersion); err != nil {
		return fmt.Errorf("attach decision result: %w", err)
	}
	return nil
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
