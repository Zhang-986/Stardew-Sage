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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin learning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
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
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit learning transaction: %w", err)
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
