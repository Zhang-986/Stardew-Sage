package domain

import (
	"math"
	"strings"
	"testing"
)

func TestDemonstrationValidateLearningContext(t *testing.T) {
	valid := Demonstration{
		ID: "demo-day-1", SaveID: "farm-1", SessionID: "teaching-1",
		Day: 1, Weather: WeatherSunny, StartedAt: 10, EndedAt: 20,
		Events: []DemonstrationEvent{{
			ID: "water-1", Kind: EventWater, Tick: 12,
			Position: Position{X: 4, Y: 5}, TargetID: "crop-1", Success: true,
		}},
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	invalidDay := valid
	invalidDay.Day = -1
	if err := invalidDay.Validate(); err == nil || !strings.Contains(err.Error(), "day") {
		t.Fatalf("Validate() error = %v, want day error", err)
	}

	invalidWeather := valid
	invalidWeather.Weather = "fog"
	if err := invalidWeather.Validate(); err == nil || !strings.Contains(err.Error(), "weather") {
		t.Fatalf("Validate() error = %v, want weather error", err)
	}
}

func TestTraitObservationValidate(t *testing.T) {
	evidence := map[string]struct{}{"water-1": {}}
	valid := TraitObservation{
		Key: PreferenceTaskOrder, Value: "watering,harvesting",
		Context: TraitContextSunny, SupportingEventIDs: []string{"water-1"}, Strength: 0.7,
	}
	if err := valid.Validate(evidence); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*TraitObservation)
		wantErr string
	}{
		{name: "unknown key", mutate: func(o *TraitObservation) { o.Key = "favorite_hat" }, wantErr: "key"},
		{name: "unknown context", mutate: func(o *TraitObservation) { o.Context = "festival" }, wantErr: "context"},
		{name: "strength above one", mutate: func(o *TraitObservation) { o.Strength = 1.1 }, wantErr: "strength"},
		{name: "nan strength", mutate: func(o *TraitObservation) { o.Strength = math.NaN() }, wantErr: "strength"},
		{name: "forged evidence", mutate: func(o *TraitObservation) { o.SupportingEventIDs = []string{"made-up"} }, wantErr: "evidence"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observation := valid
			tt.mutate(&observation)
			if err := observation.Validate(evidence); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestPlayerActivityValidate(t *testing.T) {
	valid := PlayerActivity{Kind: EventWater, TargetID: "crop-1", Tick: 120, Success: true}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	invalidTarget := valid
	invalidTarget.TargetID = ""
	if err := invalidTarget.Validate(); err == nil || !strings.Contains(err.Error(), "target") {
		t.Fatalf("Validate() error = %v, want target error", err)
	}

	invalidTick := valid
	invalidTick.Tick = -1
	if err := invalidTick.Validate(); err == nil || !strings.Contains(err.Error(), "tick") {
		t.Fatalf("Validate() error = %v, want tick error", err)
	}

	unsupported := valid
	unsupported.Kind = EventMove
	if err := unsupported.Validate(); err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("Validate() error = %v, want kind error", err)
	}
}

func TestWorldSnapshotValidate(t *testing.T) {
	valid := WorldSnapshot{
		SaveID:          "farm-1",
		SessionID:       "day-2",
		SnapshotVersion: 3,
		Day:             2,
		TimeOfDay:       620,
		Weather:         WeatherSunny,
		Location:        "Farm",
		Energy:          250,
		MaxEnergy:       270,
		WateringCan:     ToolState{Name: "Watering Can", Water: 20, Capacity: 40},
		Crops:           []Crop{{ID: "crop-1", NeedsWater: true}},
		WaterSources:    []WaterSource{{ID: "pond-1"}},
		Chests:          []Chest{{ID: "chest-1"}},
	}

	tests := []struct {
		name    string
		mutate  func(*WorldSnapshot)
		wantErr string
	}{
		{name: "valid"},
		{name: "missing save", mutate: func(s *WorldSnapshot) { s.SaveID = "" }, wantErr: "save_id"},
		{name: "duplicate target", mutate: func(s *WorldSnapshot) { s.Chests[0].ID = "crop-1" }, wantErr: "duplicate target id"},
		{name: "energy exceeds maximum", mutate: func(s *WorldSnapshot) { s.Energy = 300 }, wantErr: "energy"},
		{name: "water exceeds capacity", mutate: func(s *WorldSnapshot) { s.WateringCan.Water = 50 }, wantErr: "watering_can"},
		{name: "negative game tick", mutate: func(s *WorldSnapshot) { s.Tick = -1 }, wantErr: "tick"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valid
			got.Crops = append([]Crop(nil), valid.Crops...)
			got.Chests = append([]Chest(nil), valid.Chests...)
			if tt.mutate != nil {
				tt.mutate(&got)
			}
			err := got.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestPlayerModelRejectsOutOfRangeConfidence(t *testing.T) {
	model := PlayerModel{
		SaveID:   "farm-1",
		Revision: 1,
		Preferences: []ObservedPreference{{
			Key:              PreferenceTaskOrder,
			Value:            "water_before_harvest",
			EvidenceEventIDs: []string{"event-1"},
			ObservationCount: 1,
			Confidence:       1.2,
		}},
	}

	if err := model.Validate(); err == nil || !strings.Contains(err.Error(), "confidence") {
		t.Fatalf("Validate() error = %v, want confidence error", err)
	}
}

func TestSkillProgramRequiresSuccessCondition(t *testing.T) {
	skill := SkillProgram{
		Name:  "morning-farm-routine",
		Goal:  "care for crops",
		Steps: []SkillStep{{Action: ActionWaterTarget, TargetSelector: "dry_crops"}},
	}

	if err := skill.Validate(); err == nil || !strings.Contains(err.Error(), "success_conditions") {
		t.Fatalf("Validate() error = %v, want success_conditions error", err)
	}
}

func TestHighLevelActionAllowsOnlyCatalog(t *testing.T) {
	tests := []struct {
		name    string
		action  HighLevelAction
		wantErr string
	}{
		{
			name: "known target action",
			action: HighLevelAction{
				SaveID: "farm-1", SessionID: "day-2", SnapshotVersion: 3,
				Kind: ActionWaterTarget, TargetID: "crop-1",
			},
		},
		{
			name: "stop needs no target",
			action: HighLevelAction{
				SaveID: "farm-1", SessionID: "day-2", SnapshotVersion: 3,
				Kind: ActionStopSession, Reason: "routine complete",
			},
		},
		{
			name: "unknown action",
			action: HighLevelAction{
				SaveID: "farm-1", SessionID: "day-2", SnapshotVersion: 3,
				Kind: "teleport", TargetID: "crop-1",
			},
			wantErr: "unsupported action kind",
		},
		{
			name: "missing target",
			action: HighLevelAction{
				SaveID: "farm-1", SessionID: "day-2", SnapshotVersion: 3,
				Kind: ActionWaterTarget,
			},
			wantErr: "target or destination",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}
