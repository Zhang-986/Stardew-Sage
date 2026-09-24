package domain

import (
	"fmt"
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

func TestDemonstrationValidateV2SemanticActivity(t *testing.T) {
	valid := Demonstration{
		SchemaVersion: 2,
		ID:            "demo-activity", SaveID: "farm-1", SessionID: "teaching-2",
		Day: 8, Weather: WeatherSunny, StartedAt: 100, EndedAt: 180,
		Events: []DemonstrationEvent{{
			ID: "tree-1", Kind: EventChopTree, Tick: 180,
			Position: Position{X: 12, Y: 8}, Location: "Farm", TimeOfDay: 920,
			TargetID: "Farm:tree:12:8", TargetKind: "tree", Tool: "Axe", DurationTicks: 80,
			Delta:      StateDelta{EnergyDelta: -10, InventoryDelta: 2},
			ItemDeltas: []ItemDelta{{ItemID: "388", Name: "Wood", Quantity: 14}}, Success: true,
		}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	legacy := valid
	legacy.SchemaVersion = 1
	if err := legacy.Validate(); err == nil || !strings.Contains(err.Error(), "version 2") {
		t.Fatalf("Validate() error = %v, want version 2 error", err)
	}
}

func TestDemonstrationRejectsInvalidActivityEvidence(t *testing.T) {
	valid := Demonstration{
		SchemaVersion: 2,
		ID:            "demo-activity", SaveID: "farm-1", SessionID: "teaching-2",
		Day: 8, Weather: WeatherSunny, StartedAt: 100, EndedAt: 180,
		Events: []DemonstrationEvent{{
			ID: "fish-1", Kind: EventFishCaught, Tick: 180,
			Location: "Beach", TimeOfDay: 920, TargetKind: "fish", DurationTicks: 80,
			ItemDeltas: []ItemDelta{{ItemID: "128", Quantity: 1}}, Success: true,
		}},
	}
	tests := []struct {
		name   string
		mutate func(*Demonstration)
		want   string
	}{
		{name: "zero item quantity", mutate: func(d *Demonstration) { d.Events[0].ItemDeltas[0].Quantity = 0 }, want: "quantity"},
		{name: "duplicate item", mutate: func(d *Demonstration) {
			d.Events[0].ItemDeltas = append(d.Events[0].ItemDeltas, d.Events[0].ItemDeltas[0])
		}, want: "duplicate item"},
		{name: "too many items", mutate: func(d *Demonstration) {
			d.Events[0].ItemDeltas = make([]ItemDelta, 33)
			for i := range d.Events[0].ItemDeltas {
				d.Events[0].ItemDeltas[i] = ItemDelta{ItemID: fmt.Sprintf("item-%d", i), Quantity: 1}
			}
		}, want: "item deltas"},
		{name: "negative duration", mutate: func(d *Demonstration) { d.Events[0].DurationTicks = -1 }, want: "duration"},
		{name: "unknown target kind", mutate: func(d *Demonstration) { d.Events[0].TargetKind = "monster" }, want: "target kind"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valid
			got.Events = append([]DemonstrationEvent(nil), valid.Events...)
			got.Events[0].ItemDeltas = append([]ItemDelta(nil), valid.Events[0].ItemDeltas...)
			tt.mutate(&got)
			if err := got.Validate(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.want)
			}
		})
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

func TestActionResultRejectsUnknownStatusAndMismatchedAction(t *testing.T) {
	action := HighLevelAction{
		SaveID: "farm-1", SessionID: "echo-1", SnapshotVersion: 2,
		Kind: ActionHarvestTarget, TargetID: "crop-1", Reason: "harvest",
	}
	valid := ActionResult{
		SaveID: "farm-1", SessionID: "echo-1", SnapshotVersion: 2,
		Action: action, Status: ActionSucceeded,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	badStatus := valid
	badStatus.Status = "maybe"
	if err := badStatus.Validate(); err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("Validate() error = %v, want status error", err)
	}

	badIdentity := valid
	badIdentity.Action.SessionID = "another-session"
	if err := badIdentity.Validate(); err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Validate() error = %v, want identity error", err)
	}
}

func TestActionProposalValidatesConfidenceCandidatesAndEvidence(t *testing.T) {
	snapshot := validReflectiveSnapshot()
	primary := reflectiveAction(snapshot, ActionHarvestTarget, "crop-1")
	alternative := reflectiveAction(snapshot, ActionDepositItems, "chest-1")
	valid := ActionProposal{
		Primary: primary, Alternatives: []HighLevelAction{alternative}, ModelConfidence: 0.8,
		UncertaintyCodes: []UncertaintyCode{UncertaintyNovelContext}, AppliedExperienceIDs: []string{"exp-1"},
	}
	available := map[string]struct{}{"exp-1": {}}
	if err := valid.Validate(snapshot, available); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*ActionProposal)
		wantErr string
	}{
		{name: "confidence", mutate: func(p *ActionProposal) { p.ModelConfidence = 1.1 }, wantErr: "confidence"},
		{name: "too many alternatives", mutate: func(p *ActionProposal) { p.Alternatives = []HighLevelAction{alternative, primary, alternative} }, wantErr: "two alternatives"},
		{name: "duplicate candidate", mutate: func(p *ActionProposal) { p.Alternatives = []HighLevelAction{primary} }, wantErr: "duplicate"},
		{name: "unknown uncertainty", mutate: func(p *ActionProposal) { p.UncertaintyCodes = []UncertaintyCode{"guessing"} }, wantErr: "uncertainty"},
		{name: "forged experience", mutate: func(p *ActionProposal) { p.AppliedExperienceIDs = []string{"exp-made-up"} }, wantErr: "experience"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proposal := valid
			proposal.Alternatives = append([]HighLevelAction(nil), valid.Alternatives...)
			proposal.UncertaintyCodes = append([]UncertaintyCode(nil), valid.UncertaintyCodes...)
			proposal.AppliedExperienceIDs = append([]string(nil), valid.AppliedExperienceIDs...)
			tt.mutate(&proposal)
			if err := proposal.Validate(snapshot, available); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestExperienceObservationRejectsUnsupportedSignal(t *testing.T) {
	observation := ExperienceObservation{
		Trigger: ExperienceInventoryFull, Context: TraitContextSunny,
		WhenSignals: []SituationSignal{"inventory_almost_full"},
		AvoidAction: ActionHarvestTarget, PreferAction: ActionDepositItems,
		PreferredTargetID: "chest-1", Summary: "deposit before harvesting",
		EvidenceRef: "decision:day-2:7", Strength: 0.7,
	}
	if err := observation.Validate(
		map[string]struct{}{"decision:day-2:7": {}},
		map[string]struct{}{"chest-1": {}},
	); err == nil || !strings.Contains(err.Error(), "signal") {
		t.Fatalf("Validate() error = %v, want signal error", err)
	}
}

func TestPlayerCorrectionRequiresCorrelatedActions(t *testing.T) {
	snapshot := validReflectiveSnapshot()
	rejected := reflectiveAction(snapshot, ActionHarvestTarget, "crop-1")
	preferred := reflectiveAction(snapshot, ActionDepositItems, "chest-1")
	valid := PlayerCorrection{
		ID: "correction-1", SaveID: snapshot.SaveID, SessionID: snapshot.SessionID,
		RejectedDecisionSnapshotVersion: snapshot.SnapshotVersion,
		RejectedAction:                  rejected, Snapshot: snapshot, PreferredAction: preferred, ObservedAtTick: 200,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	invalid := valid
	invalid.PreferredAction.SessionID = "another-session"
	if err := invalid.Validate(); err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Validate() error = %v, want identity error", err)
	}

	invalid = valid
	invalid.PreferredAction.TargetID = "chest-invented"
	if err := invalid.Validate(); err == nil || !strings.Contains(err.Error(), "not present") {
		t.Fatalf("Validate() error = %v, want missing target error", err)
	}
}

func TestActionEnabledHonorsExplicitCapabilitiesAndLegacySnapshots(t *testing.T) {
	legacy := validReflectiveSnapshot()
	if !ActionEnabled(legacy, ActionHarvestTarget) {
		t.Fatal("legacy snapshot unexpectedly disabled harvesting")
	}

	restricted := legacy
	restricted.Capabilities = &ActionCapabilities{Harvest: false}
	if ActionEnabled(restricted, ActionHarvestTarget) {
		t.Fatal("explicit capabilities allowed disabled harvesting")
	}
	if !ActionEnabled(restricted, ActionWaterTarget) || !ActionEnabled(restricted, ActionStopSession) {
		t.Fatal("harvest capability disabled an unrelated safe action")
	}

	restricted.Capabilities.Harvest = true
	if !ActionEnabled(restricted, ActionHarvestTarget) {
		t.Fatal("explicit capabilities did not enable harvesting")
	}
}

func validReflectiveSnapshot() WorldSnapshot {
	return WorldSnapshot{
		SaveID: "farm-1", SessionID: "echo-day-2", SnapshotVersion: 7,
		Day: 2, TimeOfDay: 700, Weather: WeatherSunny, Location: "Farm",
		Energy: 200, MaxEnergy: 270,
		Inventory:   InventorySummary{FreeSlots: 0, Items: []InventoryItem{{ItemID: "parsnip", Name: "Parsnip", Quantity: 1}}},
		WateringCan: ToolState{Name: "Watering Can", Water: 10, Capacity: 40},
		Crops:       []Crop{{ID: "crop-1", Mature: true}}, Chests: []Chest{{ID: "chest-1"}},
	}
}

func reflectiveAction(snapshot WorldSnapshot, kind ActionKind, targetID string) HighLevelAction {
	return HighLevelAction{
		SaveID: snapshot.SaveID, SessionID: snapshot.SessionID, SnapshotVersion: snapshot.SnapshotVersion,
		Kind: kind, TargetID: targetID, Reason: "reflective policy",
	}
}
