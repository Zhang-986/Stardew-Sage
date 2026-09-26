package relay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc"
)

const requestIDHeader = "X-EchoFarm-Request-ID"

type Operation string

const (
	OperationHealth       Operation = "Health"
	OperationLearn        Operation = "Learn"
	OperationNextAction   Operation = "NextAction"
	OperationActionResult Operation = "ActionResult"
	OperationCorrection   Operation = "Correction"
	OperationPlayerModel  Operation = "PlayerModel"
	OperationSkill        Operation = "Skill"
	OperationMemory       Operation = "Memory"
	OperationModelUsage   Operation = "ModelUsage"
)

type route struct {
	operation Operation
	retry     bool
}

var routes = map[string]route{
	http.MethodGet + " /healthz":                        {operation: OperationHealth},
	http.MethodGet + " /v1/player-model":                {operation: OperationPlayerModel},
	http.MethodGet + " /v1/skills/morning-farm-routine": {operation: OperationSkill},
	http.MethodGet + " /v1/echo/memory":                 {operation: OperationMemory},
	http.MethodGet + " /v1/model-usage":                 {operation: OperationModelUsage},
	http.MethodPost + " /v1/demonstrations/learn":       {operation: OperationLearn, retry: true},
	http.MethodPost + " /v1/echo/next-action":           {operation: OperationNextAction},
	http.MethodPost + " /v1/echo/action-result":         {operation: OperationActionResult},
	http.MethodPost + " /v1/echo/corrections":           {operation: OperationCorrection, retry: true},
}

type GatewayClient interface {
	Call(context.Context, Operation, *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error)
}

type Handler struct {
	config Config
	client GatewayClient
	logger *log.Logger
	slots  chan struct{}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func NewHandler(config Config, client GatewayClient, logger *log.Logger) (*Handler, error) {
	if config.ListenAddress == "" || config.UpstreamAddress == "" || config.LANToken == "" {
		return nil, errors.New("relay addresses and token are required")
	}
	if config.MaxRequestBytes <= 0 || config.MaxRequestBytes > absoluteMaxRequestBytes {
		return nil, errors.New("valid request size limit is required")
	}
	if config.MaxInFlight < 1 || config.MaxInFlight > 8 {
		return nil, errors.New("valid concurrency limit is required")
	}
	if client == nil || logger == nil {
		return nil, errors.New("gateway client and logger are required")
	}
	return &Handler{config: config, client: client, logger: logger, slots: make(chan struct{}, config.MaxInFlight)}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	requestID, err := newRequestID()
	if err != nil {
		writeRelayError(w, http.StatusServiceUnavailable, "request_id_unavailable")
		return
	}
	w.Header().Set(requestIDHeader, requestID)
	recorder := &responseRecorder{ResponseWriter: w}
	defer h.logRequest(requestID, r, recorder, started)

	route, ok := routes[r.Method+" "+r.URL.Path]
	if !ok {
		writeRelayError(recorder, http.StatusNotFound, "route_not_found")
		return
	}
	select {
	case h.slots <- struct{}{}:
		defer func() { <-h.slots }()
	default:
		writeRelayError(recorder, http.StatusTooManyRequests, "relay_busy")
		return
	}
	if r.ContentLength > h.config.MaxRequestBytes {
		writeRelayError(recorder, http.StatusRequestEntityTooLarge, "request_too_large")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, h.config.MaxRequestBytes+1))
	if err != nil {
		writeRelayError(recorder, http.StatusBadRequest, "invalid_request")
		return
	}
	if int64(len(body)) > h.config.MaxRequestBytes {
		writeRelayError(recorder, http.StatusRequestEntityTooLarge, "request_too_large")
		return
	}
	var query, contentType *string
	if r.URL.RawQuery != "" {
		value := r.URL.RawQuery
		query = &value
	}
	if value := r.Header.Get("Content-Type"); value != "" {
		contentType = &value
	}
	rpcRequest := &echofarmrpc.RelayRequest{
		RequestId:   requestID,
		Token:       h.config.LANToken,
		Query:       query,
		Body:        body,
		ContentType: contentType,
	}
	rpcResponse, err := h.client.Call(r.Context(), route.operation, rpcRequest)
	if err != nil && route.retry {
		rpcResponse, err = h.client.Call(r.Context(), route.operation, rpcRequest)
	}
	if err != nil {
		writeRelayError(recorder, http.StatusServiceUnavailable, "upstream_unavailable")
		return
	}
	if rpcResponse == nil || rpcResponse.RequestId != requestID || rpcResponse.StatusCode < 100 || rpcResponse.StatusCode > 599 || int64(len(rpcResponse.Body)) > h.config.MaxRequestBytes {
		writeRelayError(recorder, http.StatusBadGateway, "invalid_upstream_response")
		return
	}
	if rpcResponse.ContentType != "" {
		recorder.Header().Set("Content-Type", rpcResponse.ContentType)
	}
	recorder.WriteHeader(int(rpcResponse.StatusCode))
	_, _ = recorder.Write(rpcResponse.Body)
}

func (h *Handler) logRequest(requestID string, request *http.Request, response *responseRecorder, started time.Time) {
	status := response.status
	if status == 0 {
		status = http.StatusOK
	}
	h.logger.Printf(
		"request_id=%s method=%s path=%s status=%d bytes=%d latency_ms=%d",
		requestID,
		request.Method,
		request.URL.EscapedPath(),
		status,
		response.bytes,
		time.Since(started).Milliseconds(),
	)
}

func newRequestID() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate request ID: %w", err)
	}
	return hex.EncodeToString(random), nil
}

func writeRelayError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, "{\"code\":%q,\"message\":\"relay request failed\"}\n", code)
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	written, err := w.ResponseWriter.Write(body)
	w.bytes += written
	return written, err
}
