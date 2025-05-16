package main

import (
	"Relay/config"
	"Relay/repository"
	"Relay/server"
	"Relay/service"
	"log"
)

func main() {
	cfg := config.LoadConfig("config.json")

	// Initialize repository
	lobbyRepo, err := repository.NewRepository(
		cfg.PostgresURL,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDB,
	)
	userRepo, err := repository.NewRepository(
		cfg.PostgresURL,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDB,
	)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Initialize services
	authService := service.NewAuthService(cfg.JWTSecretKey)
	relayService := service.NewRelayService(userRepo, lobbyRepo, authService)

	// Start UDP relay server
	relayServer := server.NewRelayServer(cfg.UDPPort, relayService)
	log.Printf("Starting UDP relay server on port %d", cfg.UDPPort)
	if err := relayServer.Start(); err != nil {
		log.Fatalf("Relay server failed: %v", err)
	}
}
