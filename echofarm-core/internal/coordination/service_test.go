package coordination

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
)

type intentStub struct {
	intent domain.PlayerIntent
	input  intelligence.IntentInput
	calls  int
}

func (s *intentStub) InferIntent(_ context.Context, input intelligence.IntentInput) (domain.PlayerIntent, error) {
	s.calls++
	s.input = input
	return s.intent, nil
}

func TestPrepareInfersIntentFromRecentSuccessfulUniqueTargets(t *testing.T) {
	inferer := &intentStub{intent: domain.PlayerIntentWatering}
	service, err := NewService(inferer)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := collaborationSnapshot()
	snapshot.Tick = 2_000
	snapshot.RecentPlayerActions = []domain.PlayerActivity{
		{Kind: domain.EventWater, TargetID: "stale", Tick: 700, Success: true},
		{Kind: domain.EventWater, TargetID: "crop-1", Tick: 1_900, Success: true},
		{Kind: domain.EventWater, TargetID: "crop-1", Tick: 1_950, Success: true},
		{Kind: domain.EventHarvest, TargetID: "failed", Tick: 1_980, Success: false},
	}

	got, err := service.Prepare(context.Background(), snapshot, domain.PlayerModel{SaveID: "farm-1", Revision: 4, EnergyReserve: 40})
	if err != nil {
		t.Fatal(err)
	}
	if inferer.calls != 1 || len(inferer.input.Activities) != 2 {
		t.Fatalf("inferer calls/input = %d/%+v", inferer.calls, inferer.input)
	}
	if got.InferredIntent != domain.PlayerIntentWatering || !reflect.DeepEqual(got.PlayerClaimedTargets, []string{"crop-1"}) {
		t.Fatalf("context = %+v", got)
	}
	if !reflect.DeepEqual(got.AvailableGoals, []string{"harvest:crop-2", "water:crop-2"}) {
		t.Fatalf("available goals = %v", got.AvailableGoals)
	}
}

func TestPrepareWithoutRecentActivityDoesNotInvokeModel(t *testing.T) {
	inferer := &intentStub{intent: domain.PlayerIntentWatering}
	service, err := NewService(inferer)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := collaborationSnapshot()
	snapshot.Tick = 2_000
	snapshot.RecentPlayerActions = []domain.PlayerActivity{{Kind: domain.EventWater, TargetID: "crop-1", Tick: 700, Success: true}}

	got, err := service.Prepare(context.Background(), snapshot, domain.PlayerModel{SaveID: "farm-1", Revision: 4, EnergyReserve: 40})
	if err != nil {
		t.Fatal(err)
	}
	if inferer.calls != 0 || got.InferredIntent != domain.PlayerIntentUnknown || len(got.PlayerClaimedTargets) != 0 {
		t.Fatalf("context/calls = %+v/%d", got, inferer.calls)
	}
}

func TestValidateChoiceRejectsPlayerClaimedTarget(t *testing.T) {
	context := domain.CoordinationContext{PlayerClaimedTargets: []string{"crop-1"}}
	claimed := domain.HighLevelAction{Kind: domain.ActionWaterTarget, TargetID: "crop-1"}
	if err := ValidateChoice(context, claimed); !errors.Is(err, ErrTargetClaimed) {
		t.Fatalf("ValidateChoice() error = %v, want ErrTargetClaimed", err)
	}
	unclaimed := domain.HighLevelAction{Kind: domain.ActionHarvestTarget, TargetID: "crop-2"}
	if err := ValidateChoice(context, unclaimed); err != nil {
		t.Fatalf("ValidateChoice() error = %v", err)
	}
}

func collaborationSnapshot() domain.WorldSnapshot {
	return domain.WorldSnapshot{
		SaveID: "farm-1", SessionID: "echo-4", SnapshotVersion: 1,
		Day: 4, TimeOfDay: 700, Weather: domain.WeatherSunny, Location: "Farm",
		Energy: 200, MaxEnergy: 270,
		WateringCan: domain.ToolState{Name: "Watering Can", Water: 20, Capacity: 40},
		Crops: []domain.Crop{
			{ID: "crop-1", Mature: false, NeedsWater: true},
			{ID: "crop-2", Mature: true, NeedsWater: true},
		},
	}
}
