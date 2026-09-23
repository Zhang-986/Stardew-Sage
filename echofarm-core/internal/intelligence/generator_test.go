package intelligence

import (
	"context"
	"errors"
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

func TestNewOpenAIGeneratorRequiresConfiguration(t *testing.T) {
	_, err := NewOpenAIGenerator(context.Background(), OpenAIConfig{})
	if err == nil {
		t.Fatal("NewOpenAIGenerator() error = nil, want missing configuration error")
	}
}
