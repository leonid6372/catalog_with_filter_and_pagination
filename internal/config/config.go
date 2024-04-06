package config

import (
	"log"
	"os"
)

type Config struct {
	SQLDriver         string
	SQLConnectionInfo string
	SQLMigrationInfo  string
	HTTPServerAddress string
}

func MustLoad() *Config {
	return &Config{
		SQLDriver:         getEnv("SQL_DRIVER"),
		SQLConnectionInfo: getEnv("SQL_CONNECTION_INFO"),
		SQLMigrationInfo:  getEnv("SQL_MIGRATION_INFO"),
		HTTPServerAddress: getEnv("HTTP_SERVER_ADDRESS"),
	}
}

// / Simple helper function to read an environment or return a default value
func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal("config parameter does not set in environment")
	}
	return value
}
