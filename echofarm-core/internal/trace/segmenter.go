package trace

import (
	"sort"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

// Segment removes frame-level navigation noise and groups adjacent semantic
// actions. It deliberately does not infer intent; that remains the AI's job.
func Segment(events []domain.DemonstrationEvent) []domain.BehaviorSegment {
	segments := make([]domain.BehaviorSegment, 0)
	var routeStart *domain.Position

	for _, event := range events {
		kind, semantic := behaviorFor(event.Kind)
		if !semantic {
			if routeStart == nil {
				start := event.Position
				routeStart = &start
			}
			continue
		}

		if len(segments) == 0 || segments[len(segments)-1].Kind != kind {
			start := event.Position
			if routeStart != nil {
				start = *routeStart
			}
			segments = append(segments, domain.BehaviorSegment{
				Kind:      kind,
				Start:     start,
				StartedAt: event.Tick,
			})
		}

		segment := &segments[len(segments)-1]
		segment.EventIDs = append(segment.EventIDs, event.ID)
		if event.TargetID != "" {
			segment.TargetIDs = append(segment.TargetIDs, event.TargetID)
		}
		segment.End = event.Position
		segment.EndedAt = event.Tick
		segment.EnergyDelta += event.Delta.EnergyDelta
		segment.WaterDelta += event.Delta.WaterDelta
		segment.InventoryDelta += event.Delta.InventoryDelta
		segment.HealthDelta += event.Delta.HealthDelta
		segment.MineFloorDelta += event.Delta.MineFloorDelta
		segment.DurationTicks += event.DurationTicks
		segment.ItemDeltas = mergeItemDeltas(segment.ItemDeltas, event.ItemDeltas)
		if !event.Success {
			segment.FailureEventIDs = append(segment.FailureEventIDs, event.ID)
		}
		routeStart = nil
	}

	return segments
}

func behaviorFor(kind domain.EventKind) (domain.BehaviorKind, bool) {
	switch kind {
	case domain.EventWater:
		return domain.BehaviorWatering, true
	case domain.EventRefill:
		return domain.BehaviorRefilling, true
	case domain.EventHarvest:
		return domain.BehaviorHarvesting, true
	case domain.EventDeposit:
		return domain.BehaviorDepositing, true
	case domain.EventChopTree:
		return domain.BehaviorWoodcutting, true
	case domain.EventBreakRock:
		return domain.BehaviorMining, true
	case domain.EventEnterMineFloor:
		return domain.BehaviorMineTraversal, true
	case domain.EventFishCaught, domain.EventFishEscaped:
		return domain.BehaviorFishing, true
	default:
		return "", false
	}
}

func mergeItemDeltas(existing, incoming []domain.ItemDelta) []domain.ItemDelta {
	type total struct {
		name     string
		quantity int
	}
	values := make(map[string]total, len(existing)+len(incoming))
	for _, item := range append(append([]domain.ItemDelta(nil), existing...), incoming...) {
		current := values[item.ItemID]
		if current.name == "" {
			current.name = item.Name
		}
		current.quantity += item.Quantity
		values[item.ItemID] = current
	}
	ids := make([]string, 0, len(values))
	for id, value := range values {
		if value.quantity != 0 {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	result := make([]domain.ItemDelta, 0, len(ids))
	for _, id := range ids {
		value := values[id]
		result = append(result, domain.ItemDelta{ItemID: id, Name: value.name, Quantity: value.quantity})
	}
	return result
}
