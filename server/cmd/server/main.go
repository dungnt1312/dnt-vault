package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dnt/vault-server/internal/api"
	"github.com/dnt/vault-server/internal/auth"
	"github.com/dnt/vault-server/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}

	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "/var/lib/dnt-vault/data"
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/dnt-vault"
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
	log.Printf("DNT-Vault server starting on %s", addr)
	log.Printf("Data path: %s", dataPath)
	log.Printf("Config path: %s", configPath)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
