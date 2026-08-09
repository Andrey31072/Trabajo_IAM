package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sena-iam-api/internal/application"
	"sena-iam-api/internal/config"
	"sena-iam-api/internal/infrastructure/email"
	"sena-iam-api/internal/infrastructure/postgres"
	"sena-iam-api/internal/infrastructure/security"
	"sena-iam-api/internal/interfaces/httpapi"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	repo, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	defer repo.Close()
	if err := waitForDatabase(ctx, repo); err != nil {
		log.Fatalf("database unavailable: %v", err)
	}
	tokens, err := security.NewTokenManager(cfg)
	if err != nil {
		log.Fatalf("jwt init failed: %v", err)
	}
	mailer := email.NewSender(cfg)
	app := application.New(repo, tokens, mailer, cfg.AppPublicURL, cfg.RefreshTokenTTLDays, cfg.DemoPassword, cfg.DemoTrainingCenterID)
	if err := app.Bootstrap(ctx); err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}
	if err := mailer.Verify(); err != nil {
		log.Fatalf("smtp verification failed: %v", err)
	}
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.New(app, tokens, cfg.CORSOrigins).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("iam-api Go listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func waitForDatabase(ctx context.Context, repo interface{ Ping(context.Context) error }) error {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if err := repo.Ping(ctx); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return repo.Ping(ctx)
}