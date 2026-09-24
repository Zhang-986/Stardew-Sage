package memory

import (
	"errors"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
)

const reflectionLeaseTimeLayout = "2006-01-02T15:04:05.000000000Z"

const (
	reflectionJobPending    = "pending"
	reflectionJobProcessing = "processing"
	reflectionJobCompleted  = "completed"

	ReflectionFailureModelUnavailable = "model_unavailable"
	ReflectionFailureCanceled         = "canceled"
	ReflectionFailureInternal         = "reflection_failed"
)

var ErrReflectionLeaseLost = errors.New("reflection job lease lost")

type ReflectionJob struct {
	SaveID       string
	SourceID     string
	Snapshot     domain.WorldSnapshot
	Result       domain.ActionResult
	AttemptCount int
}

type ReflectionJobLease struct {
	Job   ReflectionJob
	Token string
}

type reflectionJobPayload struct {
	Snapshot domain.WorldSnapshot `json:"snapshot"`
	Result   domain.ActionResult  `json:"result"`
}

func validateReflectionLease(lease ReflectionJobLease) error {
	if lease.Job.SaveID == "" || lease.Job.SourceID == "" || lease.Token == "" {
		return errors.New("reflection job lease identity is required")
	}
	return nil
}

func validReflectionFailureCode(code string) bool {
	switch code {
	case ReflectionFailureModelUnavailable, ReflectionFailureCanceled, ReflectionFailureInternal:
		return true
	default:
		return false
	}
}

func formatReflectionLeaseTime(value time.Time) string {
	return value.UTC().Format(reflectionLeaseTimeLayout)
}
