package rpcgateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc"
)

type Gateway struct {
	core          http.Handler
	expectedToken [sha256.Size]byte
	maxBodyBytes  int
	logger        *log.Logger
}

type captureWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func New(core http.Handler, token string, maxBodyBytes int, logger *log.Logger) (*Gateway, error) {
	if core == nil || logger == nil {
		return nil, errors.New("core handler and logger are required")
	}
	if !validToken(token) {
		return nil, errors.New("LAN token must be exactly 64 hexadecimal characters")
	}
	if maxBodyBytes <= 0 || maxBodyBytes > 2<<20 {
		return nil, errors.New("max body bytes must be between 1 and 2097152")
	}
	return &Gateway{
		core:          core,
		expectedToken: sha256.Sum256([]byte(token)),
		maxBodyBytes:  maxBodyBytes,
		logger:        logger,
	}, nil
}

func (g *Gateway) Health(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "Health", http.MethodGet, "/healthz", request)
}

func (g *Gateway) Learn(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "Learn", http.MethodPost, "/v1/demonstrations/learn", request)
}

func (g *Gateway) NextAction(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "NextAction", http.MethodPost, "/v1/echo/next-action", request)
}

func (g *Gateway) ActionResult_(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "ActionResult", http.MethodPost, "/v1/echo/action-result", request)
}

func (g *Gateway) Correction(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "Correction", http.MethodPost, "/v1/echo/corrections", request)
}

func (g *Gateway) PlayerModel(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "PlayerModel", http.MethodGet, "/v1/player-model", request)
}

func (g *Gateway) Skill(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "Skill", http.MethodGet, "/v1/skills/morning-farm-routine", request)
}

func (g *Gateway) Memory(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "Memory", http.MethodGet, "/v1/echo/memory", request)
}

func (g *Gateway) ModelUsage(ctx context.Context, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	return g.forward(ctx, "ModelUsage", http.MethodGet, "/v1/model-usage", request)
}

func (g *Gateway) forward(ctx context.Context, rpcMethod, method, path string, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	started := time.Now()
	requestID := "invalid"
	status := http.StatusBadRequest
	defer func() {
		g.logger.Printf("request_id=%s rpc_method=%s path=%s status=%d latency_ms=%d", requestID, rpcMethod, path, status, time.Since(started).Milliseconds())
	}()
	if request == nil {
		return gatewayError("", http.StatusBadRequest, "invalid_request"), nil
	}
	requestID = request.RequestId
	if !g.authorized(request.Token) {
		status = http.StatusUnauthorized
		return gatewayError(requestID, status, "unauthorized"), nil
	}
	if !validRequestID(requestID) {
		return gatewayError("", status, "invalid_request_id"), nil
	}
	if len(request.Body) > g.maxBodyBytes {
		status = http.StatusRequestEntityTooLarge
		return gatewayError(requestID, status, "request_too_large"), nil
	}
	query, err := url.ParseQuery(request.GetQuery())
	if err != nil {
		return gatewayError(requestID, status, "invalid_query"), nil
	}
	target := path
	if encoded := query.Encode(); encoded != "" {
		target += "?" + encoded
	}
	httpRequest, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(request.Body))
	if err != nil {
		return gatewayError(requestID, status, "invalid_request"), nil
	}
	if contentType := request.GetContentType(); contentType != "" {
		httpRequest.Header.Set("Content-Type", contentType)
	}
	response := &captureWriter{header: make(http.Header)}
	g.core.ServeHTTP(response, httpRequest)
	status = response.status
	if status == 0 {
		status = http.StatusOK
	}
	if response.body.Len() > g.maxBodyBytes {
		status = http.StatusBadGateway
		return gatewayError(requestID, status, "response_too_large"), nil
	}
	contentType := response.header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return &echofarmrpc.RelayResponse{
		RequestId: requestID, StatusCode: int32(status), ContentType: contentType,
		Body: append([]byte(nil), response.body.Bytes()...),
	}, nil
}

func (g *Gateway) authorized(token string) bool {
	presented := sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(g.expectedToken[:], presented[:]) == 1
}

func validToken(token string) bool {
	if len(token) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(token)
	return err == nil && len(decoded) == 32
}

func validRequestID(requestID string) bool {
	decoded, err := hex.DecodeString(requestID)
	return err == nil && len(decoded) == 16
}

func gatewayError(requestID string, status int, code string) *echofarmrpc.RelayResponse {
	return &echofarmrpc.RelayResponse{
		RequestId: requestID, StatusCode: int32(status), ContentType: "application/json",
		Body: []byte(fmt.Sprintf("{\"code\":%q,\"message\":\"gateway request rejected\"}\n", code)),
	}
}

func (w *captureWriter) Header() http.Header { return w.header }

func (w *captureWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *captureWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(body)
}
