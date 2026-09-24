package domain

import "time"

type Weather string

const (
	WeatherSunny Weather = "sunny"
	WeatherRainy Weather = "rainy"
	WeatherStorm Weather = "storm"
	WeatherSnow  Weather = "snow"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Crop struct {
	ID          string   `json:"id"`
	Position    Position `json:"position"`
	Kind        string   `json:"kind,omitempty"`
	GrowthStage int      `json:"growthStage"`
	Mature      bool     `json:"mature"`
	NeedsWater  bool     `json:"needsWater"`
}

type WaterSource struct {
	ID       string   `json:"id"`
	Position Position `json:"position"`
}

type Chest struct {
	ID       string   `json:"id"`
	Label    string   `json:"label,omitempty"`
	Position Position `json:"position"`
}

type ToolState struct {
	Name     string `json:"name"`
	Level    int    `json:"level"`
	Water    int    `json:"water"`
	Capacity int    `json:"capacity"`
}

type InventoryItem struct {
	ItemID   string `json:"itemId"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type InventorySummary struct {
	FreeSlots int             `json:"freeSlots"`
	Items     []InventoryItem `json:"items,omitempty"`
}

type ActionCapabilities struct {
	Harvest bool `json:"harvest"`
}

type WorldSnapshot struct {
	SaveID              string              `json:"saveId"`
	SessionID           string              `json:"sessionId"`
	SnapshotVersion     int64               `json:"snapshotVersion"`
	Tick                int64               `json:"tick,omitempty"`
	Day                 int                 `json:"day"`
	TimeOfDay           int                 `json:"timeOfDay"`
	Weather             Weather             `json:"weather"`
	Location            string              `json:"location"`
	PlayerPosition      Position            `json:"playerPosition"`
	Energy              int                 `json:"energy"`
	MaxEnergy           int                 `json:"maxEnergy"`
	Inventory           InventorySummary    `json:"inventory"`
	WateringCan         ToolState           `json:"wateringCan"`
	Crops               []Crop              `json:"crops,omitempty"`
	WaterSources        []WaterSource       `json:"waterSources,omitempty"`
	Chests              []Chest             `json:"chests,omitempty"`
	Obstacles           []Position          `json:"obstacles,omitempty"`
	RecentPlayerActions []PlayerActivity    `json:"recentPlayerActions,omitempty"`
	Capabilities        *ActionCapabilities `json:"capabilities,omitempty"`
}

type EventKind string

const (
	EventMove           EventKind = "move"
	EventEquipTool      EventKind = "equip_tool"
	EventWater          EventKind = "water_target"
	EventRefill         EventKind = "refill_can"
	EventHarvest        EventKind = "harvest_target"
	EventDeposit        EventKind = "deposit_items"
	EventChopTree       EventKind = "chop_tree"
	EventBreakRock      EventKind = "break_rock"
	EventEnterMineFloor EventKind = "enter_mine_floor"
	EventFishCaught     EventKind = "fish_caught"
	EventFishEscaped    EventKind = "fish_escaped"
)

type StateDelta struct {
	EnergyDelta    int `json:"energyDelta"`
	WaterDelta     int `json:"waterDelta"`
	InventoryDelta int `json:"inventoryDelta"`
	HealthDelta    int `json:"healthDelta,omitempty"`
	MineFloorDelta int `json:"mineFloorDelta,omitempty"`
}

type ItemDelta struct {
	ItemID   string `json:"itemId"`
	Name     string `json:"name,omitempty"`
	Quantity int    `json:"quantity"`
}

type DemonstrationEvent struct {
	ID            string      `json:"id"`
	Kind          EventKind   `json:"kind"`
	Tick          int64       `json:"tick"`
	Position      Position    `json:"position"`
	Location      string      `json:"location,omitempty"`
	TimeOfDay     int         `json:"timeOfDay,omitempty"`
	TargetID      string      `json:"targetId,omitempty"`
	TargetKind    string      `json:"targetKind,omitempty"`
	Tool          string      `json:"tool,omitempty"`
	DurationTicks int64       `json:"durationTicks,omitempty"`
	Delta         StateDelta  `json:"delta"`
	ItemDeltas    []ItemDelta `json:"itemDeltas,omitempty"`
	Success       bool        `json:"success"`
	ErrorCode     string      `json:"errorCode,omitempty"`
}

type Demonstration struct {
	SchemaVersion int                  `json:"schemaVersion,omitempty"`
	ID            string               `json:"id"`
	SaveID        string               `json:"saveId"`
	SessionID     string               `json:"sessionId"`
	Day           int                  `json:"day,omitempty"`
	Weather       Weather              `json:"weather,omitempty"`
	StartedAt     int64                `json:"startedAt"`
	EndedAt       int64                `json:"endedAt"`
	Events        []DemonstrationEvent `json:"events"`
}

type BehaviorKind string

const (
	BehaviorWatering   BehaviorKind = "watering"
	BehaviorRefilling  BehaviorKind = "refilling"
	BehaviorHarvesting BehaviorKind = "harvesting"
	BehaviorDepositing BehaviorKind = "depositing"
)

type BehaviorSegment struct {
	Kind            BehaviorKind `json:"kind"`
	EventIDs        []string     `json:"eventIds"`
	TargetIDs       []string     `json:"targetIds,omitempty"`
	Start           Position     `json:"start"`
	End             Position     `json:"end"`
	StartedAt       int64        `json:"startedAt"`
	EndedAt         int64        `json:"endedAt"`
	EnergyDelta     int          `json:"energyDelta"`
	WaterDelta      int          `json:"waterDelta"`
	InventoryDelta  int          `json:"inventoryDelta"`
	FailureEventIDs []string     `json:"failureEventIds,omitempty"`
}

type PreferenceKey string

const (
	PreferenceTaskOrder      PreferenceKey = "task_order"
	PreferencePreferredChest PreferenceKey = "preferred_chest"
	PreferenceEnergyReserve  PreferenceKey = "energy_reserve"
	PreferenceRouteStyle     PreferenceKey = "route_style"
)

type ObservedPreference struct {
	Key              PreferenceKey `json:"key"`
	Value            string        `json:"value"`
	EvidenceEventIDs []string      `json:"evidenceEventIds"`
	ObservationCount int           `json:"observationCount"`
	Confidence       float64       `json:"confidence"`
}

type TraitContext string

const (
	TraitContextAny   TraitContext = "any"
	TraitContextSunny TraitContext = "sunny"
	TraitContextRainy TraitContext = "rainy"
	TraitContextStorm TraitContext = "storm"
	TraitContextSnow  TraitContext = "snow"
)

type TraitObservation struct {
	Key                PreferenceKey `json:"key"`
	Value              string        `json:"value"`
	Context            TraitContext  `json:"context"`
	SupportingEventIDs []string      `json:"supportingEventIds"`
	Strength           float64       `json:"strength"`
}

type TraitMemory struct {
	Key                PreferenceKey `json:"key"`
	Value              string        `json:"value"`
	Context            TraitContext  `json:"context"`
	Confidence         float64       `json:"confidence"`
	ObservationCount   int           `json:"observationCount"`
	ContradictionCount int           `json:"contradictionCount"`
	FirstSeenDay       int           `json:"firstSeenDay"`
	LastSeenDay        int           `json:"lastSeenDay"`
	EvidenceRefs       []string      `json:"evidenceRefs"`
}

type LearningChangeKind string

const (
	LearningChangeAdded        LearningChangeKind = "added"
	LearningChangeStrengthened LearningChangeKind = "strengthened"
	LearningChangeWeakened     LearningChangeKind = "weakened"
	LearningChangeUnchanged    LearningChangeKind = "unchanged"
)

type LearningChange struct {
	ModelRevision      int                `json:"modelRevision"`
	Kind               LearningChangeKind `json:"kind"`
	Key                PreferenceKey      `json:"key,omitempty"`
	Value              string             `json:"value,omitempty"`
	PreviousConfidence float64            `json:"previousConfidence,omitempty"`
	Confidence         float64            `json:"confidence,omitempty"`
	Summary            string             `json:"summary"`
}

type LearningOutcome struct {
	Demonstration Demonstration  `json:"demonstration,omitempty"`
	PlayerModel   PlayerModel    `json:"playerModel"`
	Skill         SkillProgram   `json:"skill"`
	Change        LearningChange `json:"learningChange"`
}

type PlayerModel struct {
	SaveID            string               `json:"saveId"`
	Revision          int                  `json:"revision"`
	LearnedThroughDay int                  `json:"learnedThroughDay,omitempty"`
	CommonTaskOrder   []BehaviorKind       `json:"commonTaskOrder,omitempty"`
	PreferredChestID  string               `json:"preferredChestId,omitempty"`
	EnergyReserve     int                  `json:"energyReserve"`
	RouteStyle        string               `json:"routeStyle,omitempty"`
	Preferences       []ObservedPreference `json:"preferences,omitempty"`
	Traits            []TraitMemory        `json:"traits,omitempty"`
}

type PlayerActivity struct {
	Kind     EventKind `json:"kind"`
	TargetID string    `json:"targetId"`
	Tick     int64     `json:"tick"`
	Success  bool      `json:"success"`
}

type PlayerIntent string

const (
	PlayerIntentUnknown    PlayerIntent = "unknown"
	PlayerIntentWatering   PlayerIntent = "watering"
	PlayerIntentHarvesting PlayerIntent = "harvesting"
	PlayerIntentDepositing PlayerIntent = "depositing"
)

type CoordinationContext struct {
	InferredIntent       PlayerIntent `json:"inferredIntent"`
	PlayerClaimedTargets []string     `json:"playerClaimedTargets,omitempty"`
	AvailableGoals       []string     `json:"availableGoals,omitempty"`
	ModelRevision        int          `json:"modelRevision"`
}

type ActionKind string

const (
	ActionMoveTo        ActionKind = "move_to"
	ActionEquipTool     ActionKind = "equip_tool"
	ActionWaterTarget   ActionKind = "water_target"
	ActionRefillCan     ActionKind = "refill_can"
	ActionHarvestTarget ActionKind = "harvest_target"
	ActionDepositItems  ActionKind = "deposit_items"
	ActionStopSession   ActionKind = "stop_session"
)

var AllowedActionKinds = map[ActionKind]struct{}{
	ActionMoveTo: {}, ActionEquipTool: {}, ActionWaterTarget: {},
	ActionRefillCan: {}, ActionHarvestTarget: {}, ActionDepositItems: {},
	ActionStopSession: {},
}

type SkillStep struct {
	Action         ActionKind `json:"action"`
	TargetSelector string     `json:"targetSelector,omitempty"`
}

type RecoveryStrategy struct {
	FailureCode string       `json:"failureCode"`
	Actions     []ActionKind `json:"actions"`
}

type SkillProgram struct {
	Name               string             `json:"name"`
	Revision           int                `json:"revision"`
	Goal               string             `json:"goal"`
	Preconditions      []string           `json:"preconditions,omitempty"`
	TargetSelector     string             `json:"targetSelector"`
	PreferredOrder     []BehaviorKind     `json:"preferredOrder,omitempty"`
	Steps              []SkillStep        `json:"steps"`
	SuccessConditions  []string           `json:"successConditions"`
	StopConditions     []string           `json:"stopConditions,omitempty"`
	RecoveryStrategies []RecoveryStrategy `json:"recoveryStrategies,omitempty"`
	EvidenceEventIDs   []string           `json:"evidenceEventIds"`
}

type HighLevelAction struct {
	SaveID          string     `json:"saveId"`
	SessionID       string     `json:"sessionId"`
	SnapshotVersion int64      `json:"snapshotVersion"`
	Kind            ActionKind `json:"kind"`
	TargetID        string     `json:"targetId,omitempty"`
	Destination     *Position  `json:"destination,omitempty"`
	Reason          string     `json:"reason"`
}

type ActionStatus string

const (
	ActionSucceeded ActionStatus = "succeeded"
	ActionFailed    ActionStatus = "failed"
)

type ActionResult struct {
	SaveID          string          `json:"saveId"`
	SessionID       string          `json:"sessionId"`
	SnapshotVersion int64           `json:"snapshotVersion"`
	Action          HighLevelAction `json:"action"`
	Status          ActionStatus    `json:"status"`
	ErrorCode       string          `json:"errorCode,omitempty"`
}

type UncertaintyCode string

const (
	UncertaintyMissingExperience   UncertaintyCode = "missing_experience"
	UncertaintyConflictingEvidence UncertaintyCode = "conflicting_evidence"
	UncertaintyNovelContext        UncertaintyCode = "novel_context"
	UncertaintyAmbiguousTarget     UncertaintyCode = "ambiguous_target"
)

type ActionProposal struct {
	Primary              HighLevelAction   `json:"primary"`
	Alternatives         []HighLevelAction `json:"alternatives,omitempty"`
	ModelConfidence      float64           `json:"modelConfidence"`
	UncertaintyCodes     []UncertaintyCode `json:"uncertaintyCodes,omitempty"`
	AppliedExperienceIDs []string          `json:"appliedExperienceIds,omitempty"`
}

type ActionDecision struct {
	Action               HighLevelAction   `json:"action"`
	Confidence           float64           `json:"confidence"`
	Alternatives         []HighLevelAction `json:"alternatives,omitempty"`
	AppliedExperienceIDs []string          `json:"appliedExperienceIds,omitempty"`
}

type ExperienceTrigger string

const (
	ExperienceInventoryFull    ExperienceTrigger = "inventory_full"
	ExperienceOutOfWater       ExperienceTrigger = "out_of_water"
	ExperiencePathBlocked      ExperienceTrigger = "path_blocked"
	ExperienceChestFull        ExperienceTrigger = "chest_full"
	ExperiencePlayerCorrection ExperienceTrigger = "player_correction"
)

type SituationSignal string

const (
	SignalInventoryFull     SituationSignal = "inventory_full"
	SignalInventoryHasItems SituationSignal = "inventory_has_items"
	SignalCanEmpty          SituationSignal = "can_empty"
	SignalRaining           SituationSignal = "raining"
	SignalTargetBlocked     SituationSignal = "target_blocked"
)

type ExperienceSource string

const (
	ExperienceSourceFailure    ExperienceSource = "failure"
	ExperienceSourceCorrection ExperienceSource = "correction"
)

type ExperienceObservation struct {
	Trigger           ExperienceTrigger `json:"trigger"`
	Context           TraitContext      `json:"context"`
	WhenSignals       []SituationSignal `json:"whenSignals"`
	AvoidAction       ActionKind        `json:"avoidAction,omitempty"`
	PreferAction      ActionKind        `json:"preferAction"`
	PreferredTargetID string            `json:"preferredTargetId,omitempty"`
	Summary           string            `json:"summary"`
	EvidenceRef       string            `json:"evidenceRef"`
	Strength          float64           `json:"strength"`
}

type PolicyExperience struct {
	ID                  string            `json:"id"`
	SaveID              string            `json:"saveId"`
	Trigger             ExperienceTrigger `json:"trigger"`
	Context             TraitContext      `json:"context"`
	WhenSignals         []SituationSignal `json:"whenSignals"`
	AvoidAction         ActionKind        `json:"avoidAction,omitempty"`
	PreferAction        ActionKind        `json:"preferAction"`
	PreferredTargetID   string            `json:"preferredTargetId,omitempty"`
	Summary             string            `json:"summary"`
	Confidence          float64           `json:"confidence"`
	EffectiveConfidence float64           `json:"effectiveConfidence,omitempty"`
	SuccessCount        int               `json:"successCount,omitempty"`
	FailureCount        int               `json:"failureCount,omitempty"`
	NeutralCount        int               `json:"neutralCount,omitempty"`
	ObservationCount    int               `json:"observationCount"`
	ContradictionCount  int               `json:"contradictionCount"`
	FirstSeenDay        int               `json:"firstSeenDay"`
	LastSeenDay         int               `json:"lastSeenDay"`
	EvidenceRefs        []string          `json:"evidenceRefs"`
	Source              ExperienceSource  `json:"source"`
}

type PlayerCorrection struct {
	ID                              string          `json:"id"`
	SaveID                          string          `json:"saveId"`
	SessionID                       string          `json:"sessionId"`
	RejectedDecisionSnapshotVersion int64           `json:"rejectedDecisionSnapshotVersion"`
	RejectedAction                  HighLevelAction `json:"rejectedAction"`
	Snapshot                        WorldSnapshot   `json:"snapshot"`
	PreferredAction                 HighLevelAction `json:"preferredAction"`
	ObservedAtTick                  int64           `json:"observedAtTick"`
}

type ExperienceOutcome struct {
	SourceID           string                `json:"sourceId"`
	Source             ExperienceSource      `json:"source"`
	Observation        ExperienceObservation `json:"observation"`
	Experience         PolicyExperience      `json:"experience"`
	UpdatedExperiences []PolicyExperience    `json:"updatedExperiences,omitempty"`
	Correction         *PlayerCorrection     `json:"correction,omitempty"`
}

type DecisionRecord struct {
	SaveID               string            `json:"saveId"`
	SessionID            string            `json:"sessionId"`
	SnapshotVersion      int64             `json:"snapshotVersion"`
	Day                  int               `json:"day"`
	ModelRevision        int               `json:"modelRevision"`
	InferredIntent       PlayerIntent      `json:"inferredIntent"`
	PlayerClaimedTargets []string          `json:"playerClaimedTargets,omitempty"`
	CandidateAction      HighLevelAction   `json:"candidateAction"`
	FinalAction          HighLevelAction   `json:"finalAction"`
	Proposal             *ActionProposal   `json:"proposal,omitempty"`
	PolicyConfidence     float64           `json:"policyConfidence,omitempty"`
	SelectedCandidate    int               `json:"selectedCandidate,omitempty"`
	SafeAlternatives     []HighLevelAction `json:"safeAlternatives"`
	Result               *ActionResult     `json:"result,omitempty"`
}

type EchoSessionMemory struct {
	SessionID string `json:"sessionId"`
	Day       int    `json:"day"`
	Status    string `json:"status"`
}

type EchoMemoryView struct {
	SaveID               string             `json:"saveId"`
	ModelRevision        int                `json:"modelRevision"`
	LearnedThroughDay    int                `json:"learnedThroughDay"`
	StableTraits         []TraitMemory      `json:"stableTraits,omitempty"`
	RecentLearningChange *LearningChange    `json:"recentLearningChange,omitempty"`
	ActiveSession        *EchoSessionMemory `json:"activeSession,omitempty"`
	LastDecision         *DecisionRecord    `json:"lastDecision,omitempty"`
	Experiences          []PolicyExperience `json:"experiences,omitempty"`
	ModelUsage           *ModelUsageSummary `json:"modelUsage,omitempty"`
}

type ModelCallPurpose string

const (
	ModelCallLearning   ModelCallPurpose = "learning"
	ModelCallIntent     ModelCallPurpose = "intent"
	ModelCallAction     ModelCallPurpose = "action"
	ModelCallRecovery   ModelCallPurpose = "recovery"
	ModelCallReflection ModelCallPurpose = "reflection"
)

type ModelCallStatus string

const (
	ModelCallStarted   ModelCallStatus = "started"
	ModelCallSucceeded ModelCallStatus = "succeeded"
	ModelCallFailed    ModelCallStatus = "failed"
)

const (
	ModelErrorUnavailable   = "model_unavailable"
	ModelErrorInvalidOutput = "invalid_model_output"
	ModelErrorDeadline      = "deadline_exceeded"
	ModelErrorCanceled      = "cancelled"
	ModelErrorInternal      = "internal"
)

type ModelBudgetLimits struct {
	MaxCallsPerSession          int `json:"maxCallsPerSession"`
	MaxReportedTokensPerSession int `json:"maxReportedTokensPerSession"`
}

type ModelCallRecord struct {
	RequestID        string           `json:"requestId"`
	SaveID           string           `json:"saveId"`
	SessionID        string           `json:"sessionId"`
	Day              int              `json:"day"`
	Purpose          ModelCallPurpose `json:"purpose"`
	StartedAt        time.Time        `json:"startedAt"`
	FinishedAt       time.Time        `json:"finishedAt,omitempty"`
	Status           ModelCallStatus  `json:"status"`
	PromptTokens     *int             `json:"promptTokens,omitempty"`
	CompletionTokens *int             `json:"completionTokens,omitempty"`
	TotalTokens      *int             `json:"totalTokens,omitempty"`
	LatencyMS        int64            `json:"latencyMs,omitempty"`
	ErrorClass       string           `json:"errorClass,omitempty"`
	CallBudget       int              `json:"callBudget"`
	TokenBudget      int              `json:"tokenBudget"`
}

type ModelUsageTotals struct {
	Calls              int   `json:"calls"`
	Succeeded          int   `json:"succeeded"`
	Failed             int   `json:"failed"`
	ReportedTokenCalls int   `json:"reportedTokenCalls"`
	PromptTokens       int   `json:"promptTokens"`
	CompletionTokens   int   `json:"completionTokens"`
	TotalTokens        int   `json:"totalTokens"`
	TokensKnown        bool  `json:"tokensKnown"`
	LastLatencyMS      int64 `json:"lastLatencyMs"`
	AverageLatencyMS   int64 `json:"averageLatencyMs"`
}

type ModelUsageSummary struct {
	SaveID          string           `json:"saveId"`
	SessionID       string           `json:"sessionId"`
	Day             int              `json:"day"`
	Session         ModelUsageTotals `json:"session"`
	DayTotals       ModelUsageTotals `json:"dayTotals"`
	CallBudget      int              `json:"callBudget"`
	TokenBudget     int              `json:"tokenBudget"`
	BudgetExhausted bool             `json:"budgetExhausted"`
}
