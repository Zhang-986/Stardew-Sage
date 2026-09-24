package intelligence

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type stubChatModel struct {
	response *schema.Message
	err      error
	messages []*schema.Message
}

func (s *stubChatModel) Generate(_ context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	s.messages = input
	return s.response, s.err
}

func (s *stubChatModel) Stream(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, errors.New("stream is not used")
}

func TestChatGeneratorDecodesStrictJSON(t *testing.T) {
	chat := &stubChatModel{response: schema.AssistantMessage(`{"playerModel":{"saveId":"farm-1"}}`, nil)}
	generator, err := NewChatGenerator(chat, time.Second)
	if err != nil {
		t.Fatalf("NewChatGenerator() error = %v", err)
	}
	var output struct {
		PlayerModel struct {
			SaveID string `json:"saveId"`
		} `json:"playerModel"`
	}

	err = generator.GenerateJSON(context.Background(), "system rules", map[string]string{"saveId": "farm-1"}, &output)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}
	if output.PlayerModel.SaveID != "farm-1" {
		t.Fatalf("decoded save id = %q", output.PlayerModel.SaveID)
	}
	if len(chat.messages) != 2 || chat.messages[0].Role != schema.System || chat.messages[1].Role != schema.User {
		t.Fatalf("messages = %+v, want system then user", chat.messages)
	}
}

func TestChatGeneratorRejectsProseAroundJSON(t *testing.T) {
	chat := &stubChatModel{response: schema.AssistantMessage("result: {\"ok\":true}", nil)}
	generator, err := NewChatGenerator(chat, time.Second)
	if err != nil {
		t.Fatalf("NewChatGenerator() error = %v", err)
	}
	var output struct {
		OK bool `json:"ok"`
	}

	err = generator.GenerateJSON(context.Background(), "system rules", map[string]string{"input": "value"}, &output)
	if !errors.Is(err, ErrInvalidModelOutput) {
		t.Fatalf("GenerateJSON() error = %v, want ErrInvalidModelOutput", err)
	}
}

func TestChatGeneratorMapsModelFailure(t *testing.T) {
	chat := &stubChatModel{err: errors.New("connection refused")}
	generator, err := NewChatGenerator(chat, time.Second)
	if err != nil {
		t.Fatalf("NewChatGenerator() error = %v", err)
	}

	err = generator.GenerateJSON(context.Background(), "system rules", struct{}{}, &struct{}{})
	if !errors.Is(err, ErrModelUnavailable) {
		t.Fatalf("GenerateJSON() error = %v, want ErrModelUnavailable", err)
	}
}

func TestChatGeneratorReportsProviderTokenUsage(t *testing.T) {
	response := schema.AssistantMessage(`{"ok":true}`, nil)
	response.ResponseMeta = &schema.ResponseMeta{Usage: &schema.TokenUsage{
		PromptTokens: 41, CompletionTokens: 7, TotalTokens: 48,
	}}
	generator, err := NewChatGenerator(&stubChatModel{response: response}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		OK bool `json:"ok"`
	}

	usage, err := generator.GenerateJSONWithUsage(context.Background(), "return JSON", struct{}{}, &output)
	if err != nil {
		t.Fatal(err)
	}
	if !output.OK || !usage.Reported || usage.PromptTokens != 41 || usage.CompletionTokens != 7 || usage.TotalTokens != 48 {
		t.Fatalf("output/usage = %+v / %+v", output, usage)
	}
}

func TestChatGeneratorLeavesProviderTokenUsageUnknownWhenAbsent(t *testing.T) {
	generator, err := NewChatGenerator(
		&stubChatModel{response: schema.AssistantMessage(`{"ok":true}`, nil)},
		time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}

	usage, err := generator.GenerateJSONWithUsage(context.Background(), "return JSON", struct{}{}, &struct {
		OK bool `json:"ok"`
	}{})
	if err != nil {
		t.Fatal(err)
	}
	if usage.Reported || usage.PromptTokens != 0 || usage.CompletionTokens != 0 || usage.TotalTokens != 0 {
		t.Fatalf("usage = %+v, want explicitly unknown zero values", usage)
	}
}

func TestNewOpenAIGeneratorRequiresConfiguration(t *testing.T) {
	_, err := NewOpenAIGenerator(context.Background(), OpenAIConfig{})
	if err == nil {
		t.Fatal("NewOpenAIGenerator() error = nil, want missing configuration error")
	}
}

func TestOpenAIGeneratorRequestsJSONMode(t *testing.T) {
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"completion-1","object":"chat.completion","created":1,"model":"test-model",
			"choices":[{"index":0,"message":{"role":"assistant","content":"{\"ok\":true}"},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}
		}`))
	}))
	t.Cleanup(server.Close)
	generator, err := NewOpenAIGenerator(context.Background(), OpenAIConfig{
		BaseURL: server.URL + "/v1", APIKey: "test-key", Model: "test-model", Timeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		OK bool `json:"ok"`
	}
	if err := generator.GenerateJSON(context.Background(), "return JSON", struct{}{}, &output); err != nil {
		t.Fatal(err)
	}
	responseFormat, ok := requestBody["response_format"].(map[string]any)
	if !ok || responseFormat["type"] != "json_object" {
		t.Fatalf("response_format = %#v, want json_object", requestBody["response_format"])
	}
}
