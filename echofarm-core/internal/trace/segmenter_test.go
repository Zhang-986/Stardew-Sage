package trace

import (
	"reflect"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

func TestSegmentCompressesMovementAndGroupsSemanticActions(t *testing.T) {
	events := []domain.DemonstrationEvent{
		{ID: "move-1", Kind: domain.EventMove, Tick: 1, Position: domain.Position{X: 1, Y: 1}, Success: true},
		{ID: "equip-1", Kind: domain.EventEquipTool, Tick: 2, Position: domain.Position{X: 1, Y: 1}, Tool: "Watering Can", Success: true},
		{ID: "water-1", Kind: domain.EventWater, Tick: 3, Position: domain.Position{X: 2, Y: 1}, TargetID: "crop-1", Delta: domain.StateDelta{EnergyDelta: -2, WaterDelta: -1}, Success: true},
		{ID: "move-2", Kind: domain.EventMove, Tick: 4, Position: domain.Position{X: 3, Y: 1}, Success: true},
		{ID: "water-2", Kind: domain.EventWater, Tick: 5, Position: domain.Position{X: 3, Y: 1}, TargetID: "crop-2", Delta: domain.StateDelta{EnergyDelta: -2, WaterDelta: -1}, Success: true},
		{ID: "water-3", Kind: domain.EventWater, Tick: 6, Position: domain.Position{X: 4, Y: 1}, TargetID: "crop-3", Delta: domain.StateDelta{EnergyDelta: -2}, Success: false, ErrorCode: "out_of_water"},
		{ID: "refill-1", Kind: domain.EventRefill, Tick: 7, Position: domain.Position{X: 5, Y: 3}, TargetID: "pond-1", Delta: domain.StateDelta{WaterDelta: 40}, Success: true},
		{ID: "harvest-1", Kind: domain.EventHarvest, Tick: 8, Position: domain.Position{X: 6, Y: 2}, TargetID: "crop-4", Delta: domain.StateDelta{InventoryDelta: 1}, Success: true},
		{ID: "deposit-1", Kind: domain.EventDeposit, Tick: 9, Position: domain.Position{X: 8, Y: 2}, TargetID: "chest-1", Delta: domain.StateDelta{InventoryDelta: -1}, Success: true},
	}

	segments := Segment(events)

	if got, want := segmentKinds(segments), []domain.BehaviorKind{
		domain.BehaviorWatering,
		domain.BehaviorRefilling,
		domain.BehaviorHarvesting,
		domain.BehaviorDepositing,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("segment kinds = %v, want %v", got, want)
	}
	if got, want := segments[0].TargetIDs, []string{"crop-1", "crop-2", "crop-3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("watering targets = %v, want %v", got, want)
	}
	if got, want := segments[0].FailureEventIDs, []string{"water-3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("failure event ids = %v, want %v", got, want)
	}
	if got, want := segments[0].EnergyDelta, -6; got != want {
		t.Fatalf("energy delta = %d, want %d", got, want)
	}
	if got, want := segments[0].WaterDelta, -2; got != want {
		t.Fatalf("water delta = %d, want %d", got, want)
	}
	if got, want := segments[0].Start, (domain.Position{X: 1, Y: 1}); got != want {
		t.Fatalf("watering start = %+v, want %+v", got, want)
	}
}

func TestSegmentReturnsNoSegmentsForMovementOnly(t *testing.T) {
	events := []domain.DemonstrationEvent{
		{ID: "move-1", Kind: domain.EventMove, Tick: 1, Position: domain.Position{X: 1, Y: 1}, Success: true},
		{ID: "move-2", Kind: domain.EventMove, Tick: 2, Position: domain.Position{X: 2, Y: 1}, Success: true},
	}

	if got := Segment(events); len(got) != 0 {
		t.Fatalf("Segment() = %v, want no semantic segments", got)
	}
}

func segmentKinds(segments []domain.BehaviorSegment) []domain.BehaviorKind {
	result := make([]domain.BehaviorKind, 0, len(segments))
	for _, segment := range segments {
		result = append(result, segment.Kind)
	}
	return result
}
