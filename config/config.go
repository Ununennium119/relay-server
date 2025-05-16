package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	JWTSecretKey      string        `json:"JWTSecretKey"`
	PostgresURL       string        `json:"PostgresURL"`
	PostgresUser      string        `json:"PostgresUser"`
	PostgresPassword  string        `json:"PostgresPassword"`
	PostgresDB        string        `json:"PostgresDB"`
	UDPPort           int           `json:"UDPPort"`
	ConnectionTimeout time.Duration `json:"ConnectionTimeout"`
}

func LoadConfig(filename string) *Config {
	cfg := &Config{
		JWTSecretKey:      "",
		PostgresURL:       "localhost:5432",
		PostgresUser:      "postgres",
		PostgresPassword:  "postgres",
		PostgresDB:        "relay",
		UDPPort:           8080,
		ConnectionTimeout: 30 * time.Second,
	}

	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Could not open config file %s: %v", filename, err)
		} else {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(cfg); err != nil {
				log.Printf("Could not parse config file %s: %v", filename, err)
			}
		}
	}

	overrideWithEnv(cfg)

	return cfg
}

func overrideWithEnv(cfg *Config) {
	if val := os.Getenv("JWT_SECRET_KEY"); val != "" {
		cfg.JWTSecretKey = val
	}
	if val := os.Getenv("POSTGRES_URL"); val != "" {
		cfg.PostgresURL = val
	}
	if val := os.Getenv("POSTGRES_USER"); val != "" {
		cfg.PostgresUser = val
	}
	if val := os.Getenv("POSTGRES_PASSWORD"); val != "" {
		cfg.PostgresPassword = val
	}
	if val := os.Getenv("POSTGRES_DB"); val != "" {
		cfg.PostgresDB = val
	}
	if val := os.Getenv("UDP_PORT"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.UDPPort = i
		}
	}
	if val := os.Getenv("CONNECTION_TIMEOUT"); val != "" {
		// try as duration string first
		if d, err := time.ParseDuration(val); err == nil {
			cfg.ConnectionTimeout = d
		} else if s, err := strconv.Atoi(val); err == nil {
			cfg.ConnectionTimeout = time.Duration(s) * time.Second
		}
	}
}
