package main

import (
	"Grippy/internal/transport/http/server"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env.dev"
	}
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			slog.Warn(fmt.Sprintf("Failed to load %s: %v", envFile, err))
		}
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "localhost"
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid PORT value '%s': %v", portStr, err)
	}

	srv, err := server.Init(addr, port)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
