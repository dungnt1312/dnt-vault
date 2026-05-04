package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dnt/vault-server/internal/api"
	"github.com/dnt/vault-server/internal/auth"
	"github.com/dnt/vault-server/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = filepath.Join(homeDir, "dnt-vault", "data")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = filepath.Join(homeDir, "dnt-vault", "config")
	}

	if err := os.MkdirAll(dataPath, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	jwtSecretFile := filepath.Join(configPath, "jwt-secret")
	usersFile := filepath.Join(configPath, "users.json")

	authService, err := auth.NewAuthService(jwtSecretFile, usersFile)
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}

	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		log.Println("Creating default admin user...")
		if err := authService.CreateUser("admin", "admin"); err != nil {
			log.Fatalf("Failed to create default user: %v", err)
		}
		log.Println("Default user created: admin/admin")
		log.Println("⚠️  Please change the default password!")
	}

	store, err := storage.NewFilesystemStorage(dataPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	handler := api.NewHandler(store, authService)
	middleware := api.NewMiddleware(authService)
	router := api.NewRouter(handler, middleware)

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("DNT-Vault server %s (built %s, commit %s) starting on %s", Version, BuildTime, CommitSHA, addr)
		log.Printf("Data path: %s", dataPath)
		log.Printf("Config path: %s", configPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
