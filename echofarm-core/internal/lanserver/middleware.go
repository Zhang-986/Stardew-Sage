package lanserver

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const requestIDHeader = "X-EchoFarm-Request-ID"

type middleware struct {
	next          http.Handler
	expectedToken [sha256.Size]byte
	logger        *log.Logger
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func New(next http.Handler, token string, logger *log.Logger) (http.Handler, error) {
	if next == nil {
		return nil, errors.New("next handler is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	if !validHexToken(token) {
		return nil, errors.New("LAN token must be exactly 64 hexadecimal characters")
	}
	return &middleware{next: next, expectedToken: sha256.Sum256([]byte(token)), logger: logger}, nil
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	requestID, err := normalizedRequestID(r.Header.Get(requestIDHeader))
	if err != nil {
		writeLocalError(w, http.StatusServiceUnavailable, "request_id_unavailable")
		return
	}
	w.Header().Set(requestIDHeader, requestID)
	recorder := &responseRecorder{ResponseWriter: w}

	if !m.authorized(r.Header.Get("Authorization")) {
		writeLocalError(recorder, http.StatusUnauthorized, "unauthorized")
		m.logRequest(requestID, r, recorder, started)
		return
	}
	r.Header.Del("Authorization")
	r.Header.Set(requestIDHeader, requestID)
	m.next.ServeHTTP(recorder, r)
	m.logRequest(requestID, r, recorder, started)
}

func (m *middleware) authorized(header string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) || len(header) <= len(prefix) {
		return false
	}
	presented := sha256.Sum256([]byte(header[len(prefix):]))
	return subtle.ConstantTimeCompare(m.expectedToken[:], presented[:]) == 1
}

func (m *middleware) logRequest(requestID string, r *http.Request, recorder *responseRecorder, started time.Time) {
	status := recorder.status
	if status == 0 {
		status = http.StatusOK
	}
	m.logger.Printf(
		"request_id=%s method=%s path=%s status=%d bytes=%d latency_ms=%d",
		requestID,
		r.Method,
		r.URL.EscapedPath(),
		status,
		recorder.bytes,
		time.Since(started).Milliseconds(),
	)
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

func normalizedRequestID(candidate string) (string, error) {
	if decoded, err := hex.DecodeString(candidate); err == nil && len(decoded) == 16 {
		return strings.ToLower(candidate), nil
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate request ID: %w", err)
	}
	return hex.EncodeToString(random), nil
}

func validHexToken(token string) bool {
	if len(token) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(token)
	return err == nil && len(decoded) == 32
}

func writeLocalError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, "{\"code\":%q,\"message\":\"request rejected\"}\n", code)
}
