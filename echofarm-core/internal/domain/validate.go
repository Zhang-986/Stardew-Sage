package domain

import (
	"errors"
	"fmt"
	"math"
)

func (s WorldSnapshot) Validate() error {
	if s.SaveID == "" {
		return errors.New("save_id is required")
	}
	if s.SessionID == "" {
		return errors.New("session_id is required")
	}
	if s.SnapshotVersion < 0 {
		return errors.New("snapshot_version cannot be negative")
	}
	if s.Tick < 0 {
		return errors.New("tick cannot be negative")
	}
	if s.Day <= 0 {
		return errors.New("day must be positive")
	}
	if s.MaxEnergy <= 0 || s.Energy < 0 || s.Energy > s.MaxEnergy {
		return errors.New("energy must be between zero and max_energy")
	}
	if s.WateringCan.Water < 0 || s.WateringCan.Capacity < 0 || s.WateringCan.Water > s.WateringCan.Capacity {
		return errors.New("watering_can water must be between zero and capacity")
	}
	if !validWeather(s.Weather) {
		return fmt.Errorf("unsupported weather %q", s.Weather)
	}
	for i, activity := range s.RecentPlayerActions {
		if err := activity.Validate(); err != nil {
			return fmt.Errorf("recent player action %d: %w", i, err)
		}
	}

	ids := make(map[string]struct{}, len(s.Crops)+len(s.WaterSources)+len(s.Chests))
	for _, item := range []struct {
		kind string
		ids  []string
	}{
		{kind: "crop", ids: cropIDs(s.Crops)},
		{kind: "water_source", ids: waterSourceIDs(s.WaterSources)},
		{kind: "chest", ids: chestIDs(s.Chests)},
	} {
		for _, id := range item.ids {
			if id == "" {
				return fmt.Errorf("%s id is required", item.kind)
			}
			if _, exists := ids[id]; exists {
				return fmt.Errorf("duplicate target id %q", id)
			}
			ids[id] = struct{}{}
		}
	}
	return nil
}

func (d Demonstration) Validate() error {
	if d.ID == "" || d.SaveID == "" || d.SessionID == "" {
		return errors.New("demonstration id, save_id, and session_id are required")
	}
	if len(d.Events) == 0 {
		return errors.New("demonstration events are required")
	}
	if d.EndedAt < d.StartedAt {
		return errors.New("demonstration ended_at precedes started_at")
	}
	if d.Day < 0 {
		return errors.New("demonstration day cannot be negative")
	}
	if d.Weather != "" && !validWeather(d.Weather) {
		return fmt.Errorf("unsupported demonstration weather %q", d.Weather)
	}
	seen := make(map[string]struct{}, len(d.Events))
	for _, event := range d.Events {
		if event.ID == "" {
			return errors.New("demonstration event id is required")
		}
		if _, ok := seen[event.ID]; ok {
			return fmt.Errorf("duplicate demonstration event id %q", event.ID)
		}
		seen[event.ID] = struct{}{}
		if !validEventKind(event.Kind) {
			return fmt.Errorf("unsupported event kind %q", event.Kind)
		}
	}
	return nil
}

func (m PlayerModel) Validate() error {
	if m.SaveID == "" {
		return errors.New("save_id is required")
	}
	if m.Revision <= 0 {
		return errors.New("revision must be positive")
	}
	if m.EnergyReserve < 0 {
		return errors.New("energy_reserve cannot be negative")
	}
	if m.LearnedThroughDay < 0 {
		return errors.New("learned_through_day cannot be negative")
	}
	for i, preference := range m.Preferences {
		if preference.Key == "" || preference.Value == "" {
			return fmt.Errorf("preference %d requires key and value", i)
		}
		if preference.Confidence < 0 || preference.Confidence > 1 {
			return fmt.Errorf("preference %d confidence must be between zero and one", i)
		}
		if preference.ObservationCount <= 0 || len(preference.EvidenceEventIDs) == 0 {
			return fmt.Errorf("preference %d requires observations and evidence", i)
		}
	}
	for i, trait := range m.Traits {
		if err := trait.Validate(); err != nil {
			return fmt.Errorf("trait %d: %w", i, err)
		}
	}
	return nil
}

func (o TraitObservation) Validate(evidence map[string]struct{}) error {
	if !validPreferenceKey(o.Key) {
		return fmt.Errorf("unsupported trait key %q", o.Key)
	}
	if o.Value == "" {
		return errors.New("trait value is required")
	}
	if !validTraitContext(o.Context) {
		return fmt.Errorf("unsupported trait context %q", o.Context)
	}
	if math.IsNaN(o.Strength) || math.IsInf(o.Strength, 0) || o.Strength < 0 || o.Strength > 1 {
		return errors.New("trait strength must be between zero and one")
	}
	if len(o.SupportingEventIDs) == 0 {
		return errors.New("trait evidence is required")
	}
	for _, id := range o.SupportingEventIDs {
		if _, ok := evidence[id]; !ok {
			return fmt.Errorf("trait evidence %q is not in the demonstration", id)
		}
	}
	return nil
}

func (m TraitMemory) Validate() error {
	if !validPreferenceKey(m.Key) || m.Value == "" {
		return errors.New("trait key and value are required")
	}
	if !validTraitContext(m.Context) {
		return fmt.Errorf("unsupported trait context %q", m.Context)
	}
	if math.IsNaN(m.Confidence) || math.IsInf(m.Confidence, 0) || m.Confidence < 0 || m.Confidence > 1 {
		return errors.New("trait confidence must be between zero and one")
	}
	if m.ObservationCount <= 0 || len(m.EvidenceRefs) == 0 {
		return errors.New("trait observations and evidence are required")
	}
	if m.ContradictionCount < 0 || m.FirstSeenDay < 0 || m.LastSeenDay < m.FirstSeenDay {
		return errors.New("trait history is invalid")
	}
	return nil
}

func (a PlayerActivity) Validate() error {
	switch a.Kind {
	case EventWater, EventRefill, EventHarvest, EventDeposit:
	default:
		return fmt.Errorf("unsupported player activity kind %q", a.Kind)
	}
	if a.TargetID == "" {
		return errors.New("player activity target is required")
	}
	if a.Tick < 0 {
		return errors.New("player activity tick cannot be negative")
	}
	return nil
}

func (s SkillProgram) Validate() error {
	if s.Name == "" || s.Goal == "" {
		return errors.New("skill name and goal are required")
	}
	if len(s.Steps) == 0 {
		return errors.New("skill steps are required")
	}
	if len(s.SuccessConditions) == 0 {
		return errors.New("skill success_conditions are required")
	}
	for i, step := range s.Steps {
		if _, ok := AllowedActionKinds[step.Action]; !ok {
			return fmt.Errorf("skill step %d has unsupported action kind %q", i, step.Action)
		}
	}
	for i, recovery := range s.RecoveryStrategies {
		for _, action := range recovery.Actions {
			if _, ok := AllowedActionKinds[action]; !ok {
				return fmt.Errorf("recovery strategy %d has unsupported action kind %q", i, action)
			}
		}
	}
	return nil
}

func (a HighLevelAction) Validate() error {
	if a.SaveID == "" || a.SessionID == "" {
		return errors.New("save_id and session_id are required")
	}
	if _, ok := AllowedActionKinds[a.Kind]; !ok {
		return fmt.Errorf("unsupported action kind %q", a.Kind)
	}
	if a.Kind != ActionStopSession && a.TargetID == "" && a.Destination == nil {
		return errors.New("action requires a target or destination")
	}
	return nil
}

func (r ActionResult) Validate() error {
	if r.Status != ActionSucceeded && r.Status != ActionFailed {
		return fmt.Errorf("unsupported action status %q", r.Status)
	}
	if err := r.Action.Validate(); err != nil {
		return fmt.Errorf("action: %w", err)
	}
	if r.SaveID == "" || r.SessionID == "" || r.SaveID != r.Action.SaveID ||
		r.SessionID != r.Action.SessionID || r.SnapshotVersion != r.Action.SnapshotVersion {
		return errors.New("action result identity does not match action")
	}
	if r.Status == ActionFailed && r.ErrorCode == "" {
		return errors.New("failed action result requires an error code")
	}
	return nil
}

func validWeather(weather Weather) bool {
	switch weather {
	case WeatherSunny, WeatherRainy, WeatherStorm, WeatherSnow:
		return true
	default:
		return false
	}
}

func validEventKind(kind EventKind) bool {
	switch kind {
	case EventMove, EventEquipTool, EventWater, EventRefill, EventHarvest, EventDeposit:
		return true
	default:
		return false
	}
}

func validPreferenceKey(key PreferenceKey) bool {
	switch key {
	case PreferenceTaskOrder, PreferencePreferredChest, PreferenceEnergyReserve, PreferenceRouteStyle:
		return true
	default:
		return false
	}
}

func validTraitContext(context TraitContext) bool {
	switch context {
	case TraitContextAny, TraitContextSunny, TraitContextRainy, TraitContextStorm, TraitContextSnow:
		return true
	default:
		return false
	}
}

func cropIDs(crops []Crop) []string {
	ids := make([]string, 0, len(crops))
	for _, crop := range crops {
		ids = append(ids, crop.ID)
	}
	return ids
}

func waterSourceIDs(sources []WaterSource) []string {
	ids := make([]string, 0, len(sources))
	for _, source := range sources {
		ids = append(ids, source.ID)
	}
	return ids
}

func chestIDs(chests []Chest) []string {
	ids := make([]string, 0, len(chests))
	for _, chest := range chests {
		ids = append(ids, chest.ID)
	}
	return ids
}
