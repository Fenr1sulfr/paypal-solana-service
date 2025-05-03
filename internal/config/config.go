package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Server  ServerConfig
	MongoDB MongoDBConfig
	PayPal  PayPalConfig
	Solana  SolanaConfig
}

type ServerConfig struct {
	Port    int
	BaseURL string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type PayPalConfig struct {
	ClientID     string
	ClientSecret string
	Sandbox      bool
}

type SolanaConfig struct {
	Endpoint   string
	PrivateKey string
	ProgramID  string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnvAsInt("SERVER_PORT", 50051),
			BaseURL: getEnv("SERVER_BASE_URL", "http://localhost:50051"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "payment_api"),
		},
		PayPal: PayPalConfig{
			ClientID:     getEnv("PAYPAL_CLIENT_ID", ""),
			ClientSecret: getEnv("PAYPAL_CLIENT_SECRET", ""),
			Sandbox:      getEnvAsBool("PAYPAL_SANDBOX", true),
		},
		Solana: SolanaConfig{
			Endpoint:   getEnv("SOLANA_ENDPOINT", "https://api.devnet.solana.com"),
			PrivateKey: getEnv("SOLANA_PRIVATE_KEY", ""),
			ProgramID:  getEnv("SOLANA_PROGRAM_ID", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s, using default", key)
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Warning: Invalid boolean value for %s, using default", key)
	}
	return defaultValue
}
