package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/coordination"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/domain"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/experience"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/httpapi"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/intelligence"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/learning"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memory"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/memoryview"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/policy"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/rpcgateway"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc/echofarmgateway"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	kitexserver "github.com/cloudwego/kitex/server"
)

type config struct {
	Address                     string
	DatabasePath                string
	ModelMode                   string
	ModelBaseURL                string
	ModelAPIKey                 string
	ModelName                   string
	ModelTimeout                time.Duration
	MaxModelCallsPerSession     int
	MaxReportedTokensPerSession int
	AllowLAN                    bool
	LANToken                    string
	TLSCertFile                 string
	TLSKeyFile                  string
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
	generator, err = intelligence.NewTrackingGenerator(generator, store, domain.ModelBudgetLimits{
		MaxCallsPerSession:          config.MaxModelCallsPerSession,
		MaxReportedTokensPerSession: config.MaxReportedTokensPerSession,
	})
	if err != nil {
		return err
	}
	handler, err := buildHandler(ctx, store, generator)
	if err != nil {
		return err
	}
	if config.AllowLAN {
		return runRPCServer(ctx, config, handler, log.Default())
	}

	server, err := newHTTPServer(config, handler)
	if err != nil {
		return err
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

func newHTTPServer(cfg config, handler http.Handler) (*http.Server, error) {
	if cfg.AllowLAN {
		return nil, errors.New("LAN mode must use the Kitex RPC server")
	}
	return &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      cfg.ModelTimeout + 5*time.Second,
		IdleTimeout:       30 * time.Second,
	}, nil
}

func runRPCServer(ctx context.Context, cfg config, handler http.Handler, logger *log.Logger) error {
	certificate, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return fmt.Errorf("load LAN TLS identity: %w", err)
	}
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("listen for LAN RPC: %w", err)
	}
	tlsListener := tls.NewListener(listener, &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	})
	gateway, err := rpcgateway.New(handler, cfg.LANToken, 2<<20, logger)
	if err != nil {
		_ = listener.Close()
		return err
	}
	rpcServer := echofarmgateway.NewServer(
		gateway,
		kitexserver.WithListener(tlsListener),
		kitexserver.WithTransServerFactory(gonet.NewTransServerFactory()),
		kitexserver.WithTransHandlerFactory(gonet.NewSvrTransHandlerFactory()),
		kitexserver.WithReadWriteTimeout(cfg.ModelTimeout+5*time.Second),
		kitexserver.WithExitWaitTime(5*time.Second),
		kitexserver.WithLimit(&limit.Option{MaxConnections: 8, MaxQPS: 16}),
	)
	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		<-ctx.Done()
		_ = rpcServer.Stop()
	}()
	logger.Printf("EchoFarm Kitex/Thrift gateway listening on %s (mode=%s)", cfg.Address, cfg.ModelMode)
	err = rpcServer.Run()
	if ctx.Err() != nil {
		<-stopDone
		return nil
	}
	return fmt.Errorf("serve EchoFarm Kitex gateway: %w", err)
}

func buildHandler(ctx context.Context, store *memory.SQLite, generator intelligence.StructuredGenerator) (http.Handler, error) {
	learningGraph, err := intelligence.NewLearningGraph(generator)
	if err != nil {
		return nil, err
	}
	actionGraph, err := intelligence.NewActionGraph(generator)
	if err != nil {
		return nil, err
	}
	reflectionGraph, err := intelligence.NewReflectionGraph(generator)
	if err != nil {
		return nil, err
	}
	intentGraph, err := intelligence.NewIntentGraph(generator)
	if err != nil {
		return nil, err
	}
	coordinator, err := coordination.NewService(intentGraph)
	if err != nil {
		return nil, err
	}
	teacher, err := learning.NewService(store, learningGraph)
	if err != nil {
		return nil, err
	}
	experienceService, err := experience.NewService(store, reflectionGraph)
	if err != nil {
		return nil, err
	}
	echoPolicy, err := policy.NewReflectiveService(store, actionGraph, coordinator, experienceService)
	if err != nil {
		return nil, err
	}
	views, err := memoryview.NewService(store)
	if err != nil {
		return nil, err
	}
	return httpapi.NewHandler(teacher, echoPolicy, experienceService, store, views, store)
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
	}
	var err error
	modelTimeoutSeconds, err := positiveIntSetting(get("ECHOFARM_MODEL_TIMEOUT_SECONDS", "30"), "ECHOFARM_MODEL_TIMEOUT_SECONDS")
	if err != nil {
		return config{}, err
	}
	result.ModelTimeout = time.Duration(modelTimeoutSeconds) * time.Second
	result.MaxModelCallsPerSession, err = positiveIntSetting(get("ECHOFARM_MAX_MODEL_CALLS_PER_SESSION", "32"), "ECHOFARM_MAX_MODEL_CALLS_PER_SESSION")
	if err != nil {
		return config{}, err
	}
	result.MaxReportedTokensPerSession, err = positiveIntSetting(get("ECHOFARM_MAX_REPORTED_TOKENS_PER_SESSION", "100000"), "ECHOFARM_MAX_REPORTED_TOKENS_PER_SESSION")
	if err != nil {
		return config{}, err
	}
	result.AllowLAN, err = strconv.ParseBool(get("ECHOFARM_ALLOW_LAN", "false"))
	if err != nil {
		return config{}, errors.New("ECHOFARM_ALLOW_LAN must be true or false")
	}
	result.LANToken = get("ECHOFARM_LAN_TOKEN", "")
	result.TLSCertFile = get("ECHOFARM_TLS_CERT_FILE", "")
	result.TLSKeyFile = get("ECHOFARM_TLS_KEY_FILE", "")
	host, _, err := net.SplitHostPort(result.Address)
	if err != nil {
		return config{}, fmt.Errorf("invalid ECHOFARM_ADDRESS: %w", err)
	}
	ip := net.ParseIP(host)
	loopback := host == "localhost" || (ip != nil && ip.IsLoopback())
	if !loopback && !result.AllowLAN {
		return config{}, errors.New("non-loopback ECHOFARM_ADDRESS requires ECHOFARM_ALLOW_LAN=true")
	}
	if result.AllowLAN {
		if !validHexSecret(result.LANToken, 32) {
			return config{}, errors.New("ECHOFARM_LAN_TOKEN must be exactly 64 hexadecimal characters")
		}
		if result.TLSCertFile == "" {
			return config{}, errors.New("ECHOFARM_TLS_CERT_FILE is required in LAN mode")
		}
		if result.TLSKeyFile == "" {
			return config{}, errors.New("ECHOFARM_TLS_KEY_FILE is required in LAN mode")
		}
	}
	if result.ModelMode != "fixture" && result.ModelMode != "openai" {
		return config{}, errors.New("ECHOFARM_MODEL_MODE must be fixture or openai")
	}
	if result.ModelMode == "openai" && (result.ModelBaseURL == "" || result.ModelAPIKey == "" || result.ModelName == "") {
		return config{}, errors.New("openai mode requires ECHOFARM_MODEL_BASE_URL, ECHOFARM_MODEL_API_KEY, and ECHOFARM_MODEL_NAME")
	}
	return result, nil
}

func validHexSecret(raw string, size int) bool {
	if len(raw) != size*2 {
		return false
	}
	decoded, err := hex.DecodeString(raw)
	return err == nil && len(decoded) == size
}

func positiveIntSetting(raw, key string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}
