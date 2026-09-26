package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/internal/relay"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	config, err := relay.LoadConfig(os.LookupEnv)
	if err != nil {
		return err
	}
	server, err := newRelayServer(config, log.Default())
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

	log.Printf("EchoFarm relay listening on http://%s -> Kitex/Thrift %s", config.ListenAddress, config.UpstreamAddress)
	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve EchoFarm relay: %w", err)
	}
	<-shutdownDone
	return nil
}

func newRelayServer(config relay.Config, logger *log.Logger) (*http.Server, error) {
	gatewayClient, err := relay.NewKitexClient(config)
	if err != nil {
		return nil, err
	}
	handler, err := relay.NewHandler(config, gatewayClient, logger)
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:              config.ListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      config.RequestTimeout + 5*time.Second,
		IdleTimeout:       30 * time.Second,
	}, nil
}
