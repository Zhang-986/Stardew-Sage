package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/httpapi"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/learning"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/policy"
)

type config struct {
	Address      string
	DatabasePath string
	ModelMode    string
	ModelBaseURL string
	ModelAPIKey  string
	ModelName    string
	ModelTimeout time.Duration
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	config, err := loadConfig(os.LookupEnv)
	if err != nil {
		return err
	}
	store, err := memory.OpenSQLite(config.DatabasePath)
	if err != nil {
		return err
	}
	defer store.Close()

	var generator intelligence.StructuredGenerator
	if config.ModelMode == "fixture" {
		generator = intelligence.NewFixtureGenerator()
	} else {
		generator, err = intelligence.NewOpenAIGenerator(ctx, intelligence.OpenAIConfig{
			BaseURL: config.ModelBaseURL, APIKey: config.ModelAPIKey,
			Model: config.ModelName, Timeout: config.ModelTimeout,
		})
		if err != nil {
			return err
		}
	}
	learningGraph, err := intelligence.NewLearningGraph(generator)
	if err != nil {
		return err
	}
	actionGraph, err := intelligence.NewActionGraph(generator)
	if err != nil {
		return err
	}
	teacher, err := learning.NewService(store, learningGraph)
	if err != nil {
		return err
	}
	echoPolicy, err := policy.NewService(store, actionGraph)
	if err != nil {
		return err
	}
	handler, err := httpapi.NewHandler(teacher, echoPolicy, store)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              config.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      config.ModelTimeout + 5*time.Second,
		IdleTimeout:       30 * time.Second,
	}
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("EchoFarm core listening on http://%s (mode=%s)", config.Address, config.ModelMode)
	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve EchoFarm API: %w", err)
	}
	<-shutdownDone
	return nil
}

func loadConfig(lookup func(string) (string, bool)) (config, error) {
	get := func(key, fallback string) string {
		if value, ok := lookup(key); ok && value != "" {
			return value
		}
		return fallback
	}
	result := config{
		Address:      get("ECHOFARM_ADDRESS", "127.0.0.1:18471"),
		DatabasePath: get("ECHOFARM_DATABASE_PATH", "echofarm.db"),
		ModelMode:    get("ECHOFARM_MODEL_MODE", "openai"),
		ModelBaseURL: get("ECHOFARM_MODEL_BASE_URL", ""),
		ModelAPIKey:  get("ECHOFARM_MODEL_API_KEY", ""),
		ModelName:    get("ECHOFARM_MODEL_NAME", ""),
		ModelTimeout: 30 * time.Second,
	}
	host, _, err := net.SplitHostPort(result.Address)
	if err != nil {
		return config{}, fmt.Errorf("invalid ECHOFARM_ADDRESS: %w", err)
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return config{}, errors.New("ECHOFARM_ADDRESS must use a loopback host")
	}
	if result.ModelMode != "fixture" && result.ModelMode != "openai" {
		return config{}, errors.New("ECHOFARM_MODEL_MODE must be fixture or openai")
	}
	if result.ModelMode == "openai" && (result.ModelBaseURL == "" || result.ModelAPIKey == "" || result.ModelName == "") {
		return config{}, errors.New("openai mode requires ECHOFARM_MODEL_BASE_URL, ECHOFARM_MODEL_API_KEY, and ECHOFARM_MODEL_NAME")
	}
	return result, nil
}
