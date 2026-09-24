using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class EchoMemoryPresenterTests
{
    [Fact]
    public void BuildLinesExplainsStableTraitsAndLatestCoordination()
    {
        var view = new EchoMemoryView
        {
            SaveId = "farm-1",
            ModelRevision = 3,
            LearnedThroughDay = 3,
            StableTraits = new[]
            {
                new TraitMemory
                {
                    Key = PreferenceKey.TaskOrder, Value = "watering,harvesting", Context = TraitContext.Sunny,
                    Confidence = 0.91, ObservationCount = 3, EvidenceRefs = new[] { "demo-1:water-1" }
                },
                new TraitMemory
                {
                    Key = PreferenceKey.PreferredChest, Value = "east", Context = TraitContext.Any,
                    Confidence = 0.82, ObservationCount = 2, EvidenceRefs = new[] { "demo-1:deposit-1" }
                }
            },
            RecentLearningChange = new LearningChange { ModelRevision = 3, Kind = LearningChangeKind.Strengthened, Summary = "task order strengthened" },
            LastDecision = new DecisionRecord
            {
                SaveId = "farm-1",
                SessionId = "echo-4",
                SnapshotVersion = 1,
                Day = 4,
                ModelRevision = 3,
                InferredIntent = PlayerIntent.Watering,
                PlayerClaimedTargets = new[] { "crop-1", "crop-2" },
                FinalAction = new HighLevelAction { Kind = ActionKind.HarvestTarget, TargetId = "crop-3", Reason = "complement player work" },
                PolicyConfidence = 0.91,
                SafeAlternatives = new[] { new HighLevelAction { Kind = ActionKind.StopSession, Reason = "safe fallback" } },
                Proposal = new ActionProposal { AppliedExperienceIds = new[] { "exp-correction" } }
            },
            Experiences = new[]
            {
                new PolicyExperience
                {
                    Id = "exp-correction", SaveId = "farm-1", Trigger = ExperienceTrigger.PlayerCorrection,
                    Context = TraitContext.Sunny, WhenSignals = new[] { SituationSignal.InventoryHasItems },
                    PreferAction = ActionKind.DepositItems, PreferredTargetId = "chest-west",
                    Summary = "use the player's demonstrated chest", Confidence = 0.85,
                    ObservationCount = 1, FirstSeenDay = 3, LastSeenDay = 3,
                    EvidenceRefs = new[] { "correction-1" }, Source = ExperienceSource.Correction
                }
            }
        };

        IReadOnlyList<string> lines = EchoMemoryPresenter.BuildLines(view);

        Assert.InRange(lines.Count, 5, 8);
        Assert.Contains(lines, line => line.Contains("Echo v3", StringComparison.Ordinal));
        Assert.Contains(lines, line => line.Contains("浇水", StringComparison.Ordinal));
        Assert.Contains(lines, line => line.Contains("crop-3", StringComparison.Ordinal));
        Assert.Contains(lines, line => line.Contains("2", StringComparison.Ordinal) && line.Contains("避开", StringComparison.Ordinal));
        Assert.Contains(lines, line => line.Contains("91%", StringComparison.Ordinal) && line.Contains("备选", StringComparison.Ordinal));
        Assert.Contains(lines, line => line.Contains("玩家纠正", StringComparison.Ordinal) && line.Contains("chest-west", StringComparison.Ordinal));
    }

    [Fact]
    public void BuildLinesHandlesEmptyOptionalMemory()
    {
        IReadOnlyList<string> lines = EchoMemoryPresenter.BuildLines(new EchoMemoryView
        {
            SaveId = "farm-1",
            ModelRevision = 1,
            LearnedThroughDay = 1
        });

        Assert.Single(lines);
        Assert.Contains("Echo v1", lines[0], StringComparison.Ordinal);
    }
}
