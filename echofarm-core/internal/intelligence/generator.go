package intelligence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type OpenAIConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type ChatGenerator struct {
	model   model.BaseChatModel
	timeout time.Duration
}

func NewChatGenerator(chatModel model.BaseChatModel, timeout time.Duration) (*ChatGenerator, error) {
	if chatModel == nil {
		return nil, errors.New("chat model is required")
	}
	if timeout <= 0 {
		return nil, errors.New("model timeout must be positive")
	}
	return &ChatGenerator{model: chatModel, timeout: timeout}, nil
}

func NewOpenAIGenerator(ctx context.Context, config OpenAIConfig) (*ChatGenerator, error) {
	if config.BaseURL == "" || config.APIKey == "" || config.Model == "" {
		return nil, errors.New("model base URL, API key, and name are required")
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	temperature := float32(0)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     config.BaseURL,
		APIKey:      config.APIKey,
		Model:       config.Model,
		Timeout:     config.Timeout,
		Temperature: &temperature,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: create OpenAI-compatible model: %v", ErrModelUnavailable, err)
	}
	return NewChatGenerator(chatModel, config.Timeout)
}

func (g *ChatGenerator) GenerateJSON(ctx context.Context, systemPrompt string, input any, output any) error {
	_, err := g.GenerateJSONWithUsage(ctx, systemPrompt, input, output)
	return err
}

func (g *ChatGenerator) GenerateJSONWithUsage(ctx context.Context, systemPrompt string, input any, output any) (GenerationUsage, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return GenerationUsage{}, fmt.Errorf("marshal model input: %w", err)
	}

	requestCtx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()
	response, err := g.model.Generate(requestCtx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(string(payload)),
	})
	if err != nil {
		return GenerationUsage{}, fmt.Errorf("%w: %v", ErrModelUnavailable, err)
	}
	if response == nil {
		return GenerationUsage{}, fmt.Errorf("%w: empty response", ErrInvalidModelOutput)
	}
	usage := generationUsage(response)

	decoder := json.NewDecoder(bytes.NewBufferString(response.Content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(output); err != nil {
		return usage, fmt.Errorf("%w: decode JSON: %v", ErrInvalidModelOutput, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return usage, fmt.Errorf("%w: response contains trailing content", ErrInvalidModelOutput)
	}
	return usage, nil
}

func generationUsage(response *schema.Message) GenerationUsage {
	if response.ResponseMeta == nil || response.ResponseMeta.Usage == nil {
		return GenerationUsage{}
	}
	usage := response.ResponseMeta.Usage
	return GenerationUsage{
		Reported:         true,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
}
