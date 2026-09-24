package memory

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestSQLitePersistsLearningArtifactsAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "echo.db")
	store, err := OpenSQLite(path)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	demo, model, skill := learningArtifacts("farm-a", 1)
	if err := store.SaveLearning(ctx, demo, model, skill); err != nil {
		t.Fatalf("SaveLearning() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	store, err = OpenSQLite(path)
	if err != nil {
		t.Fatalf("reopen SQLite error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	gotDemo, err := store.GetDemonstration(ctx, "farm-a", demo.ID)
	if err != nil || !reflect.DeepEqual(gotDemo, demo) {
		t.Fatalf("GetDemonstration() = (%+v, %v), want %+v", gotDemo, err, demo)
	}
	gotModel, err := store.GetPlayerModel(ctx, "farm-a")
	if err != nil || !reflect.DeepEqual(gotModel, model) {
		t.Fatalf("GetPlayerModel() = (%+v, %v), want %+v", gotModel, err, model)
	}
	gotSkill, err := store.GetSkill(ctx, "farm-a", skill.Name)
	if err != nil || !reflect.DeepEqual(gotSkill, skill) {
		t.Fatalf("GetSkill() = (%+v, %v), want %+v", gotSkill, err, skill)
	}
}

func TestSQLiteIsolatesSavesAndKeepsNewestRevision(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	demoA, modelA, skillA := learningArtifacts("farm-a", 2)
	demoB, modelB, skillB := learningArtifacts("farm-b", 1)
	if err := store.SaveLearning(ctx, demoA, modelA, skillA); err != nil {
		t.Fatalf("save farm-a error = %v", err)
	}
	if err := store.SaveLearning(ctx, demoB, modelB, skillB); err != nil {
		t.Fatalf("save farm-b error = %v", err)
	}

	olderDemo, olderModel, olderSkill := learningArtifacts("farm-a", 1)
	olderDemo.ID = "demo-older"
	if err := store.SaveLearning(ctx, olderDemo, olderModel, olderSkill); err != nil {
		t.Fatalf("save older revision error = %v", err)
	}
	gotA, err := store.GetPlayerModel(ctx, "farm-a")
	if err != nil {
		t.Fatalf("GetPlayerModel(farm-a) error = %v", err)
	}
	if gotA.Revision != 2 {
		t.Fatalf("farm-a revision = %d, want 2", gotA.Revision)
	}
	gotB, err := store.GetPlayerModel(ctx, "farm-b")
	if err != nil {
		t.Fatalf("GetPlayerModel(farm-b) error = %v", err)
	}
	if gotB.SaveID != "farm-b" || gotB.PreferredChestID != "farm-b-chest" {
		t.Fatalf("farm-b model leaked across saves: %+v", gotB)
	}
	if _, err := store.GetSkill(ctx, "farm-c", skillA.Name); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetSkill(farm-c) error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteRejectsMismatchedSaveIDsAtomically(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	demo, model, skill := learningArtifacts("farm-a", 1)
	model.SaveID = "farm-b"
	if err := store.SaveLearning(ctx, demo, model, skill); err == nil {
		t.Fatal("SaveLearning() error = nil, want save mismatch error")
	}
	if _, err := store.GetDemonstration(ctx, "farm-a", demo.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("partial demonstration persisted, error = %v", err)
	}
}

func TestSQLitePersistsLearningOutcomeAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "echo.db")
	store, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	demo, model, skill := learningArtifacts("farm-a", 2)
	outcome := domain.LearningOutcome{
		Demonstration: demo, PlayerModel: model, Skill: skill,
		Change: domain.LearningChange{
			ModelRevision: 2, Kind: domain.LearningChangeStrengthened,
			Key: domain.PreferenceTaskOrder, Value: "watering", Confidence: 0.8, Summary: "strengthened task order",
		},
	}
	if err := store.SaveLearningOutcome(ctx, outcome); err != nil {
		t.Fatalf("SaveLearningOutcome() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	got, err := store.GetLearningOutcome(ctx, demo.SaveID, demo.ID)
	if err != nil {
		t.Fatalf("GetLearningOutcome() error = %v", err)
	}
	if !reflect.DeepEqual(got, outcome) {
		t.Fatalf("GetLearningOutcome() = %+v, want %+v", got, outcome)
	}
}

func TestSQLitePersistsDecisionAndAttachesResultIdempotently(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	record := decisionRecord(7, "crop-free")
	if err := store.SaveDecision(ctx, record); err != nil {
		t.Fatalf("SaveDecision() error = %v", err)
	}
	duplicate := record
	duplicate.FinalAction.TargetID = "should-not-replace"
	if err := store.SaveDecision(ctx, duplicate); err != nil {
		t.Fatalf("duplicate SaveDecision() error = %v", err)
	}

	result := domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: record.FinalAction, Status: domain.ActionSucceeded,
	}
	if err := store.AttachDecisionResult(ctx, result); err != nil {
		t.Fatalf("AttachDecisionResult() error = %v", err)
	}
	if err := store.AttachDecisionResult(ctx, result); err != nil {
		t.Fatalf("idempotent AttachDecisionResult() error = %v", err)
	}

	got, err := store.GetDecision(ctx, record.SaveID, record.SessionID, record.SnapshotVersion)
	if err != nil {
		t.Fatal(err)
	}
	if got.FinalAction.TargetID != "crop-free" || got.Result == nil || got.Result.Status != domain.ActionSucceeded {
		t.Fatalf("decision = %+v", got)
	}
	latest, err := store.GetLatestDecision(ctx, record.SaveID)
	if err != nil || !reflect.DeepEqual(latest, got) {
		t.Fatalf("GetLatestDecision() = %+v, %v", latest, err)
	}
	session, err := store.GetActiveSession(ctx, record.SaveID)
	if err != nil || session.SessionID != record.SessionID || session.Status != "active" {
		t.Fatalf("GetActiveSession() = %+v, %v", session, err)
	}
}

func TestSQLiteDecisionConflictKeepsSessionStatusFromCanonicalWinner(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	winner := decisionRecord(7, "crop-free")
	if err := store.SaveDecision(ctx, winner); err != nil {
		t.Fatal(err)
	}
	loser := winner
	stop := domain.HighLevelAction{
		SaveID: winner.SaveID, SessionID: winner.SessionID, SnapshotVersion: winner.SnapshotVersion,
		Kind: domain.ActionStopSession, Reason: "conflicting local stop",
	}
	loser.CandidateAction = stop
	loser.FinalAction = stop
	if err := store.SaveDecision(ctx, loser); err != nil {
		t.Fatal(err)
	}

	session, err := store.GetActiveSession(ctx, winner.SaveID)
	if err != nil || session.SessionID != winner.SessionID || session.Status != "active" {
		t.Fatalf("GetActiveSession() = %+v, %v; want canonical winner to remain active", session, err)
	}
}

func TestSQLiteReturnsLatestLearningOutcome(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	for revision := 1; revision <= 2; revision++ {
		demo, model, skill := learningArtifacts("farm-a", revision)
		demo.ID = fmt.Sprintf("demo-%d", revision)
		outcome := domain.LearningOutcome{
			Demonstration: demo, PlayerModel: model, Skill: skill,
			Change: domain.LearningChange{ModelRevision: revision, Kind: domain.LearningChangeAdded, Summary: fmt.Sprintf("revision %d", revision)},
		}
		if err := store.SaveLearningOutcome(ctx, outcome); err != nil {
			t.Fatal(err)
		}
	}
	got, err := store.GetLatestLearningOutcome(ctx, "farm-a")
	if err != nil || got.PlayerModel.Revision != 2 {
		t.Fatalf("GetLatestLearningOutcome() = %+v, %v", got, err)
	}
}

func TestSQLiteRejectsResultForDifferentAction(t *testing.T) {
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
	mismatched := domain.ActionResult{
		SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
		Action: domain.HighLevelAction{
			SaveID: record.SaveID, SessionID: record.SessionID, SnapshotVersion: record.SnapshotVersion,
			Kind: domain.ActionHarvestTarget, TargetID: "crop-other", Reason: "different action",
		},
		Status: domain.ActionSucceeded,
	}
	if err := store.AttachDecisionResult(ctx, mismatched); err == nil {
		t.Fatal("AttachDecisionResult() error = nil, want action mismatch")
	}
}

func TestSQLitePersistsExperienceOutcomeAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "echo.db")
	store, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	want := experienceOutcome("farm-a", "decision:echo-day-4:1", domain.ExperienceSourceFailure, "chest-east")
	if err := store.SaveExperienceOutcome(ctx, want); err != nil {
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
	got, err := store.GetExperienceOutcome(ctx, want.Experience.SaveID, want.SourceID)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("GetExperienceOutcome() = %+v, %v; want %+v", got, err, want)
	}
	experiences, err := store.ListPolicyExperiences(ctx, want.Experience.SaveID)
	if err != nil || len(experiences) != 1 || !reflect.DeepEqual(experiences[0], want.Experience) {
		t.Fatalf("ListPolicyExperiences() = %+v, %v", experiences, err)
	}
}

func TestSQLiteExperienceOutcomeIsIdempotentBySource(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	winner := experienceOutcome("farm-a", "decision:echo-day-4:1", domain.ExperienceSourceFailure, "chest-east")
	if err := store.SaveExperienceOutcome(ctx, winner); err != nil {
		t.Fatal(err)
	}
	loser := experienceOutcome("farm-a", winner.SourceID, domain.ExperienceSourceFailure, "chest-west")
	if err := store.SaveExperienceOutcome(ctx, loser); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetExperienceOutcome(ctx, "farm-a", winner.SourceID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Experience.PreferredTargetID != "chest-east" {
		t.Fatalf("canonical outcome = %+v, want first writer", got)
	}
}

func TestSQLitePersistsCorrectionWithExperienceAtomically(t *testing.T) {
	ctx := context.Background()
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "echo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	outcome := experienceOutcome("farm-a", "correction-1", domain.ExperienceSourceCorrection, "chest-west")
	outcome.Observation.Trigger = domain.ExperiencePlayerCorrection
	outcome.Experience.Trigger = domain.ExperiencePlayerCorrection
	snapshot := correctionSnapshot("farm-a")
	outcome.Correction = &domain.PlayerCorrection{
		ID: outcome.SourceID, SaveID: snapshot.SaveID, SessionID: snapshot.SessionID,
		RejectedDecisionSnapshotVersion: 1,
		RejectedAction: domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: 1,
			Kind: domain.ActionDepositItems, TargetID: "chest-east", Reason: "old preference",
		},
		Snapshot: snapshot,
		PreferredAction: domain.HighLevelAction{
			SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
			Kind: domain.ActionDepositItems, TargetID: "chest-west", Reason: "player correction",
		},
		ObservedAtTick: 200,
	}

	if err := store.SaveExperienceOutcome(ctx, outcome); err != nil {
		t.Fatal(err)
	}
	correction, err := store.GetPlayerCorrection(ctx, "farm-a", "correction-1")
	if err != nil || !reflect.DeepEqual(correction, *outcome.Correction) {
		t.Fatalf("GetPlayerCorrection() = %+v, %v", correction, err)
	}
	if experiences, err := store.ListPolicyExperiences(ctx, "farm-other"); err != nil || len(experiences) != 0 {
		t.Fatalf("other save experiences = %+v, %v", experiences, err)
	}
}

func decisionRecord(version int64, targetID string) domain.DecisionRecord {
	action := domain.HighLevelAction{
		SaveID: "farm-a", SessionID: "echo-day-4", SnapshotVersion: version,
		Kind: domain.ActionHarvestTarget, TargetID: targetID, Reason: "complement player work",
	}
	return domain.DecisionRecord{
		SaveID: "farm-a", SessionID: "echo-day-4", SnapshotVersion: version,
		Day: 4, ModelRevision: 3, InferredIntent: domain.PlayerIntentWatering,
		PlayerClaimedTargets: []string{"crop-claimed"}, CandidateAction: action, FinalAction: action,
	}
}

func learningArtifacts(saveID string, revision int) (domain.Demonstration, domain.PlayerModel, domain.SkillProgram) {
	eventID := saveID + "-water"
	demo := domain.Demonstration{
		ID: "demo-" + saveID, SaveID: saveID, SessionID: "teaching", StartedAt: 1, EndedAt: 2,
		Events: []domain.DemonstrationEvent{{ID: eventID, Kind: domain.EventWater, Tick: 1, TargetID: "crop-1", Success: true}},
	}
	model := domain.PlayerModel{
		SaveID: saveID, Revision: revision, PreferredChestID: saveID + "-chest", EnergyReserve: 40,
		Preferences: []domain.ObservedPreference{{
			Key: domain.PreferenceTaskOrder, Value: "water_first", EvidenceEventIDs: []string{eventID},
			ObservationCount: 1, Confidence: 0.8,
		}},
	}
	skill := domain.SkillProgram{
		Name: "morning-farm-routine", Revision: revision, Goal: "care for crops", TargetSelector: "actionable_crops",
		Steps:             []domain.SkillStep{{Action: domain.ActionWaterTarget, TargetSelector: "dry_crops"}},
		SuccessConditions: []string{"all crops cared for"}, EvidenceEventIDs: []string{eventID},
	}
	return demo, model, skill
}

func experienceOutcome(saveID, sourceID string, source domain.ExperienceSource, targetID string) domain.ExperienceOutcome {
	observation := domain.ExperienceObservation{
		Trigger: domain.ExperienceInventoryFull, Context: domain.TraitContextSunny,
		WhenSignals: []domain.SituationSignal{domain.SignalInventoryFull, domain.SignalInventoryHasItems},
		AvoidAction: domain.ActionHarvestTarget, PreferAction: domain.ActionDepositItems,
		PreferredTargetID: targetID, Summary: "deposit before harvesting", EvidenceRef: sourceID, Strength: 0.8,
	}
	return domain.ExperienceOutcome{
		SourceID: sourceID, Source: source, Observation: observation,
		Experience: domain.PolicyExperience{
			ID: "exp-" + targetID, SaveID: saveID, Trigger: observation.Trigger, Context: observation.Context,
			WhenSignals: append([]domain.SituationSignal(nil), observation.WhenSignals...),
			AvoidAction: observation.AvoidAction, PreferAction: observation.PreferAction,
			PreferredTargetID: targetID, Summary: observation.Summary,
			Confidence: 0.65, ObservationCount: 1, FirstSeenDay: 4, LastSeenDay: 4,
			EvidenceRefs: []string{sourceID}, Source: source,
		},
	}
}

func correctionSnapshot(saveID string) domain.WorldSnapshot {
	return domain.WorldSnapshot{
		SaveID: saveID, SessionID: "echo-day-4", SnapshotVersion: 2,
		Day: 4, TimeOfDay: 710, Weather: domain.WeatherSunny, Location: "Farm",
		Energy: 200, MaxEnergy: 270,
		Inventory:   domain.InventorySummary{FreeSlots: 0, Items: []domain.InventoryItem{{ItemID: "parsnip", Name: "Parsnip", Quantity: 1}}},
		WateringCan: domain.ToolState{Name: "Watering Can", Water: 10, Capacity: 40},
		Chests:      []domain.Chest{{ID: "chest-east"}, {ID: "chest-west"}},
	}
}
