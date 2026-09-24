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

func TestSegmentAggregatesExtendedActivitiesAndItemEvidence(t *testing.T) {
	events := []domain.DemonstrationEvent{
		{ID: "tree-1", Kind: domain.EventChopTree, Tick: 80, Position: domain.Position{X: 1, Y: 1},
			DurationTicks: 80, Delta: domain.StateDelta{EnergyDelta: -10},
			ItemDeltas: []domain.ItemDelta{{ItemID: "388", Name: "Wood", Quantity: 10}}, Success: true},
		{ID: "tree-2", Kind: domain.EventChopTree, Tick: 150, Position: domain.Position{X: 2, Y: 1},
			DurationTicks: 70, Delta: domain.StateDelta{EnergyDelta: -8},
			ItemDeltas: []domain.ItemDelta{{ItemID: "92", Name: "Sap", Quantity: 2}, {ItemID: "388", Name: "Wood", Quantity: 4}}, Success: true},
		{ID: "rock-1", Kind: domain.EventBreakRock, Tick: 170, Position: domain.Position{X: 3, Y: 1},
			DurationTicks: 20, Delta: domain.StateDelta{EnergyDelta: -2},
			ItemDeltas: []domain.ItemDelta{{ItemID: "382", Name: "Coal", Quantity: 1}}, Success: true},
		{ID: "floor-1", Kind: domain.EventEnterMineFloor, Tick: 180, Position: domain.Position{X: 4, Y: 1},
			DurationTicks: 10, Delta: domain.StateDelta{HealthDelta: -5, MineFloorDelta: 1}, Success: true},
		{ID: "fish-1", Kind: domain.EventFishCaught, Tick: 260, Position: domain.Position{X: 5, Y: 1},
			DurationTicks: 80, ItemDeltas: []domain.ItemDelta{{ItemID: "128", Name: "Pufferfish", Quantity: 1}}, Success: true},
	}

	segments := Segment(events)

	if got, want := segmentKinds(segments), []domain.BehaviorKind{
		domain.BehaviorWoodcutting, domain.BehaviorMining, domain.BehaviorMineTraversal, domain.BehaviorFishing,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("segment kinds = %v, want %v", got, want)
	}
	if segments[0].DurationTicks != 150 || segments[0].EnergyDelta != -18 {
		t.Fatalf("woodcutting totals = %+v", segments[0])
	}
	if got := segments[0].ItemDeltas; len(got) != 2 || got[0].ItemID != "388" || got[0].Quantity != 14 || got[1].ItemID != "92" {
		t.Fatalf("woodcutting item deltas = %+v", got)
	}
	if segments[2].MineFloorDelta != 1 || segments[2].HealthDelta != -5 {
		t.Fatalf("mine traversal totals = %+v", segments[2])
	}
}

func segmentKinds(segments []domain.BehaviorSegment) []domain.BehaviorKind {
	result := make([]domain.BehaviorKind, 0, len(segments))
	for _, segment := range segments {
		result = append(result, segment.Kind)
	}
	return result
}
