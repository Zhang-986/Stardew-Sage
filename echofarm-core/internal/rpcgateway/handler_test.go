package rpcgateway

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc"
)

const testGatewayToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestGatewayAuthenticatesBeforeCallingCore(t *testing.T) {
	called := false
	gateway, err := New(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}), testGatewayToken, 2<<20, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}

	response, err := gateway.Health(context.Background(), &echofarmrpc.RelayRequest{
		RequestId: "0123456789abcdef0123456789abcdef",
		Token:     strings.Repeat("b", 64),
	})

	if err != nil || response.StatusCode != http.StatusUnauthorized || called {
		t.Fatalf("response=%+v called=%v err=%v", response, called, err)
	}
}

func TestGatewayForwardsTypedMethodToCore(t *testing.T) {
	var gotMethod, gotURI, gotContentType string
	var gotBody []byte
	core := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotURI = r.Method, r.URL.RequestURI()
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"learned":true}`))
	})
	var logs bytes.Buffer
	gateway, err := New(core, testGatewayToken, 2<<20, log.New(&logs, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	query, contentType := "saveId=farm", "application/json"
	request := &echofarmrpc.RelayRequest{
		RequestId:   "0123456789abcdef0123456789abcdef",
		Token:       testGatewayToken,
		Query:       &query,
		Body:        []byte(`{"private":"game-data"}`),
		ContentType: &contentType,
	}

	response, err := gateway.Learn(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost || gotURI != "/v1/demonstrations/learn?saveId=farm" || gotContentType != contentType {
		t.Fatalf("forwarded request = %s %s %s", gotMethod, gotURI, gotContentType)
	}
	if string(gotBody) != string(request.Body) {
		t.Fatalf("forwarded body = %q", gotBody)
	}
	if response.StatusCode != http.StatusCreated || string(response.Body) != `{"learned":true}` || response.RequestId != request.RequestId {
		t.Fatalf("response = %+v", response)
	}
	for _, forbidden := range []string{testGatewayToken, "game-data", "saveId=farm"} {
		if strings.Contains(logs.String(), forbidden) {
			t.Fatalf("log contains %q: %s", forbidden, logs.String())
		}
	}
}

func TestGatewayRejectsOversizedPayload(t *testing.T) {
	called := false
	gateway, err := New(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}), testGatewayToken, 8, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	response, err := gateway.NextAction(context.Background(), &echofarmrpc.RelayRequest{
		RequestId: "0123456789abcdef0123456789abcdef",
		Token:     testGatewayToken,
		Body:      []byte("123456789"),
	})
	if err != nil || response.StatusCode != http.StatusRequestEntityTooLarge || called {
		t.Fatalf("response=%+v called=%v err=%v", response, called, err)
	}
}

func TestGatewayRejectsMalformedIdentityAndQuery(t *testing.T) {
	gateway, err := New(http.NotFoundHandler(), testGatewayToken, 2<<20, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	badQuery := "%zz"
	for _, request := range []*echofarmrpc.RelayRequest{
		{RequestId: "not-valid", Token: testGatewayToken},
		{RequestId: "0123456789abcdef0123456789abcdef", Token: testGatewayToken, Query: &badQuery},
	} {
		response, err := gateway.Memory(context.Background(), request)
		if err != nil || response.StatusCode != http.StatusBadRequest {
			t.Fatalf("response=%+v err=%v", response, err)
		}
	}
}
