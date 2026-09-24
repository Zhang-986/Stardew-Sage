package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
)

const maxRequestBytes = 2 << 20

type teacher interface {
	TeachOutcome(context.Context, domain.Demonstration) (domain.LearningOutcome, error)
}

type echoPolicy interface {
	NextAction(context.Context, string, domain.WorldSnapshot) (domain.HighLevelAction, error)
	HandleResult(context.Context, string, domain.WorldSnapshot, domain.ActionResult) (domain.HighLevelAction, error)
}

type memoryReader interface {
	GetPlayerModel(context.Context, string) (domain.PlayerModel, error)
	GetSkill(context.Context, string, string) (domain.SkillProgram, error)
}

type memoryViewReader interface {
	Get(context.Context, string) (domain.EchoMemoryView, error)
}

type Handler struct {
	teacher teacher
	policy  echoPolicy
	memory  memoryReader
	views   memoryViewReader
	mux     *http.ServeMux
}

type LearnResponse struct {
	PlayerModel    domain.PlayerModel    `json:"playerModel"`
	Skill          domain.SkillProgram   `json:"skill"`
	LearningChange domain.LearningChange `json:"learningChange"`
}

type ActionResponse struct {
	Action domain.HighLevelAction `json:"action"`
}

type ActionResultRequest struct {
	SaveID   string               `json:"saveId"`
	Snapshot domain.WorldSnapshot `json:"snapshot"`
	Result   domain.ActionResult  `json:"result"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewHandler(teacher teacher, policy echoPolicy, memory memoryReader, views memoryViewReader) (*Handler, error) {
	if teacher == nil || policy == nil || memory == nil || views == nil {
		return nil, errors.New("teacher, policy, memory, and memory views are required")
	}
	h := &Handler{teacher: teacher, policy: policy, memory: memory, views: views, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /healthz", h.health)
	h.mux.HandleFunc("POST /v1/demonstrations/learn", h.learn)
	h.mux.HandleFunc("POST /v1/echo/next-action", h.nextAction)
	h.mux.HandleFunc("POST /v1/echo/action-result", h.actionResult)
	h.mux.HandleFunc("GET /v1/player-model", h.playerModel)
	h.mux.HandleFunc("GET /v1/skills/morning-farm-routine", h.skill)
	h.mux.HandleFunc("GET /v1/echo/memory", h.echoMemory)
	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) learn(w http.ResponseWriter, r *http.Request) {
	var demonstration domain.Demonstration
	if err := decodeJSON(w, r, &demonstration); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "request body is not a valid demonstration")
		return
	}
	outcome, err := h.teacher.TeachOutcome(r.Context(), demonstration)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, LearnResponse{
		PlayerModel: outcome.PlayerModel, Skill: outcome.Skill, LearningChange: outcome.Change,
	})
}

func (h *Handler) nextAction(w http.ResponseWriter, r *http.Request) {
	var snapshot domain.WorldSnapshot
	if err := decodeJSON(w, r, &snapshot); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "request body is not a valid world snapshot")
		return
	}
	action, err := h.policy.NextAction(r.Context(), snapshot.SaveID, snapshot)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ActionResponse{Action: action})
}

func (h *Handler) actionResult(w http.ResponseWriter, r *http.Request) {
	var request ActionResultRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "request body is not a valid action result")
		return
	}
	if request.SaveID == "" || request.SaveID != request.Snapshot.SaveID || request.SaveID != request.Result.SaveID {
		writeAPIError(w, http.StatusBadRequest, "save_mismatch", "saveId must match snapshot and result")
		return
	}
	action, err := h.policy.HandleResult(r.Context(), request.SaveID, request.Snapshot, request.Result)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ActionResponse{Action: action})
}

func (h *Handler) playerModel(w http.ResponseWriter, r *http.Request) {
	saveID := r.URL.Query().Get("saveId")
	if saveID == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_save_id", "saveId is required")
		return
	}
	model, err := h.memory.GetPlayerModel(r.Context(), saveID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, model)
}

func (h *Handler) skill(w http.ResponseWriter, r *http.Request) {
	saveID := r.URL.Query().Get("saveId")
	if saveID == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_save_id", "saveId is required")
		return
	}
	skill, err := h.memory.GetSkill(r.Context(), saveID, "morning-farm-routine")
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skill)
}

func (h *Handler) echoMemory(w http.ResponseWriter, r *http.Request) {
	saveID := r.URL.Query().Get("saveId")
	if saveID == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_save_id", "saveId is required")
		return
	}
	view, err := h.views.Get(r.Context(), saveID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, output any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(output); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain exactly one JSON value")
	}
	return nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, intelligence.ErrModelUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "model_unavailable", "Echo is unavailable; the game may continue normally")
	case errors.Is(err, memory.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "memory_not_found", "no learned Echo memory exists for this save")
	default:
		writeAPIError(w, http.StatusUnprocessableEntity, "request_rejected", "the request could not be safely processed")
	}
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
