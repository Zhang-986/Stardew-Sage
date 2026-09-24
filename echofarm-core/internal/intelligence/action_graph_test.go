package intelligence

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

type actionGeneratorStub struct {
	action domain.HighLevelAction
	prompt string
	input  any
	calls  int
}

func (s *actionGeneratorStub) GenerateJSON(_ context.Context, prompt string, input any, output any) error {
	s.calls++
	s.prompt = prompt
	s.input = input
	target, ok := output.(*domain.HighLevelAction)
	if !ok {
		return errors.New("unexpected output type")
	}
	*target = s.action
	return nil
}

func TestActionGraphChoosesStructuredAction(t *testing.T) {
	input := validActionInput()
	want := domain.HighLevelAction{
		SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID,
		SnapshotVersion: input.Snapshot.SnapshotVersion, Kind: domain.ActionWaterTarget,
		TargetID: "crop-new", Reason: "water a currently dry crop",
	}
	generator := &actionGeneratorStub{action: want}
	graph, err := NewActionGraph(generator)
	if err != nil {
		t.Fatalf("NewActionGraph() error = %v", err)
	}

	got, err := graph.ChooseAction(context.Background(), input)
	if err != nil {
		t.Fatalf("ChooseAction() error = %v", err)
	}
	if got != want || generator.calls != 1 {
		t.Fatalf("ChooseAction() = %+v, calls = %d", got, generator.calls)
	}
	if !strings.Contains(generator.prompt, "current world") {
		t.Fatalf("prompt does not require current-world reasoning: %q", generator.prompt)
	}
	if !strings.Contains(generator.prompt, "claimed") {
		t.Fatalf("prompt does not require player-target coordination: %q", generator.prompt)
	}
}

func TestActionGraphReplansFromFailure(t *testing.T) {
	input := validActionInput()
	want := domain.HighLevelAction{
		SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID,
		SnapshotVersion: input.Snapshot.SnapshotVersion, Kind: domain.ActionMoveTo,
		TargetID: "crop-new", Reason: "route around the blocked tile",
	}
	generator := &actionGeneratorStub{action: want}
	graph, err := NewActionGraph(generator)
	if err != nil {
		t.Fatalf("NewActionGraph() error = %v", err)
	}

	got, err := graph.Replan(context.Background(), ReplanInput{
		ActionInput: input,
		LastResult: domain.ActionResult{
			SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID,
			SnapshotVersion: input.Snapshot.SnapshotVersion,
			Action:          domain.HighLevelAction{Kind: domain.ActionWaterTarget},
			Status:          domain.ActionFailed, ErrorCode: "path_blocked",
		},
	})
	if err != nil {
		t.Fatalf("Replan() error = %v", err)
	}
	if got != want || !strings.Contains(generator.prompt, "failure") {
		t.Fatalf("Replan() = %+v, prompt = %q", got, generator.prompt)
	}
}

func TestActionGraphRejectsMalformedAction(t *testing.T) {
	graph, err := NewActionGraph(&actionGeneratorStub{action: domain.HighLevelAction{Kind: "teleport"}})
	if err != nil {
		t.Fatalf("NewActionGraph() error = %v", err)
	}

	_, err = graph.ChooseAction(context.Background(), validActionInput())
	if !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("ChooseAction() error = %v, want ErrInvalidModelOutput", err)
	}
}

func validActionInput() ActionInput {
	snapshot := domain.WorldSnapshot{
		SaveID: "farm-1", SessionID: "day-2", SnapshotVersion: 7, Day: 2, TimeOfDay: 620,
		Weather: domain.WeatherSunny, Location: "Farm", Energy: 200, MaxEnergy: 270,
		WateringCan: domain.ToolState{Name: "Watering Can", Water: 5, Capacity: 40},
		Crops:       []domain.Crop{{ID: "crop-new", NeedsWater: true}},
	}
	return ActionInput{
		Snapshot:    snapshot,
		PlayerModel: domain.PlayerModel{SaveID: "farm-1", Revision: 1, EnergyReserve: 40},
		Skill: domain.SkillProgram{
			Name: "morning-farm-routine", Revision: 1, Goal: "care for crops", TargetSelector: "actionable_crops",
			Steps:             []domain.SkillStep{{Action: domain.ActionWaterTarget, TargetSelector: "dry_crops"}},
			SuccessConditions: []string{"all crops cared for"}, EvidenceEventIDs: []string{"water-1"},
		},
	}
}
