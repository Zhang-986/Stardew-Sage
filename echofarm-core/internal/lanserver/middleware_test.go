package lanserver

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

const testToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type observedReader struct {
	read bool
}

func (r *observedReader) Read([]byte) (int, error) {
	r.read = true
	return 0, io.EOF
}

func (r *observedReader) Close() error { return nil }

func TestMiddlewareAuthenticatesBeforeReadingBody(t *testing.T) {
	called := false
	body := &observedReader{}
	handler, err := New(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}), testToken, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/demonstrations/learn", body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || called || body.read {
		t.Fatalf("code=%d called=%v bodyRead=%v", response.Code, called, body.read)
	}
}

func TestMiddlewareAcceptsValidTokenAndPreservesValidRequestID(t *testing.T) {
	var logs bytes.Buffer
	const requestID = "0123456789abcdef0123456789abcdef"
	handler, err := New(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), testToken, log.New(&logs, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/healthz?private=value", nil)
	request.Header.Set("Authorization", "Bearer "+testToken)
	request.Header.Set("X-EchoFarm-Request-ID", requestID)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	if got := response.Header().Get("X-EchoFarm-Request-ID"); got != requestID {
		t.Fatalf("request ID = %q", got)
	}
	text := logs.String()
	for _, forbidden := range []string{testToken, "private=value", "Authorization"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("log contains forbidden value %q: %s", forbidden, text)
		}
	}
	for _, required := range []string{"request_id=" + requestID, "method=GET", "path=/healthz", "status=204"} {
		if !strings.Contains(text, required) {
			t.Fatalf("log missing %q: %s", required, text)
		}
	}
}

func TestMiddlewareReplacesMalformedRequestID(t *testing.T) {
	handler, err := New(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), testToken, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Authorization", "Bearer "+testToken)
	request.Header.Set("X-EchoFarm-Request-ID", "../../not-valid")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	requestID := response.Header().Get("X-EchoFarm-Request-ID")
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(requestID) {
		t.Fatalf("generated request ID = %q", requestID)
	}
}

func TestNewRejectsInvalidToken(t *testing.T) {
	for _, token := range []string{"", strings.Repeat("a", 63), strings.Repeat("z", 64)} {
		if _, err := New(http.NotFoundHandler(), token, log.New(io.Discard, "", 0)); err == nil {
			t.Fatalf("New() accepted token %q", token)
		}
	}
}
