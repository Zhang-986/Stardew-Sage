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

type proposalGeneratorStub struct {
	proposal domain.ActionProposal
	prompt   string
	input    any
	calls    int
}

func (s *proposalGeneratorStub) GenerateJSON(_ context.Context, prompt string, input any, output any) error {
	s.calls++
	s.prompt = prompt
	s.input = input
	target, ok := output.(*domain.ActionProposal)
	if !ok {
		return errors.New("unexpected output type")
	}
	*target = s.proposal
	return nil
}

func (s *actionGeneratorStub) GenerateJSON(_ context.Context, prompt string, input any, output any) error {
	s.calls++
	s.prompt = prompt
	s.input = input
	switch target := output.(type) {
	case *domain.HighLevelAction:
		*target = s.action
	case *domain.ActionProposal:
		*target = domain.ActionProposal{Primary: s.action, ModelConfidence: 0.8}
	default:
		return errors.New("unexpected output type")
	}
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

func TestActionGraphProducesRankedProposalWithExperienceEvidence(t *testing.T) {
	input := validActionInput()
	input.ApplicableExperiences = []domain.PolicyExperience{{
		ID: "exp-inventory", SaveID: input.Snapshot.SaveID,
		Trigger: domain.ExperienceInventoryFull, Context: domain.TraitContextSunny,
		WhenSignals: []domain.SituationSignal{domain.SignalInventoryFull},
		AvoidAction: domain.ActionHarvestTarget, PreferAction: domain.ActionDepositItems,
		Confidence: 0.8, ObservationCount: 2, FirstSeenDay: 1, LastSeenDay: 2,
		EvidenceRefs: []string{"decision:day-1:1"}, Source: domain.ExperienceSourceFailure,
	}}
	primary := domain.HighLevelAction{
		SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID, SnapshotVersion: input.Snapshot.SnapshotVersion,
		Kind: domain.ActionWaterTarget, TargetID: "crop-new", Reason: "water current crop",
	}
	alternative := domain.HighLevelAction{
		SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID, SnapshotVersion: input.Snapshot.SnapshotVersion,
		Kind: domain.ActionStopSession, Reason: "safe fallback",
	}
	want := domain.ActionProposal{
		Primary: primary, Alternatives: []domain.HighLevelAction{alternative}, ModelConfidence: 0.76,
		AppliedExperienceIDs: []string{"exp-inventory"},
	}
	generator := &proposalGeneratorStub{proposal: want}
	graph, err := NewActionGraph(generator)
	if err != nil {
		t.Fatal(err)
	}

	got, err := graph.ProposeAction(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if got.ModelConfidence != want.ModelConfidence || len(got.Alternatives) != 1 || generator.calls != 1 {
		t.Fatalf("ProposeAction() = %+v, calls = %d", got, generator.calls)
	}
}

func TestActionGraphRejectsFabricatedExperienceReference(t *testing.T) {
	input := validActionInput()
	action := domain.HighLevelAction{
		SaveID: input.Snapshot.SaveID, SessionID: input.Snapshot.SessionID, SnapshotVersion: input.Snapshot.SnapshotVersion,
		Kind: domain.ActionWaterTarget, TargetID: "crop-new", Reason: "water current crop",
	}
	graph, err := NewActionGraph(&proposalGeneratorStub{proposal: domain.ActionProposal{
		Primary: action, ModelConfidence: 0.8, AppliedExperienceIDs: []string{"made-up"},
	}})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := graph.ProposeAction(context.Background(), input); !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("ProposeAction() error = %v, want ErrInvalidModelOutput", err)
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
