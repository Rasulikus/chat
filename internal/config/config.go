package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	keyHTTPHost, defaultHTTPHost = "HTTP_HOST", "localhost"
	keyHTTPPort, defaultHTTPPort = "HTTP_PORT", "8081"

	keyDBHost, defaultDBHost = "DB_HOST", "localhost"
	keyDBPort, defaultDBPort = "DB_PORT", "5432"
	keyDBUser, defaultDBUser = "DB_USER", "admin"
	keyDBPass, defaultDBPass = "DB_PASS", "mypassword"
	keyDBName, defaultDBName = "DB_NAME", "chat"

	LogDefaultValue = "%s is missing, using default value"
)

type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
}

type DBConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

func (cfg *DBConfig) PostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Name)
}

type HTTPConfig struct {
	Host string
	Port string
}

func getEnv(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		log.Printf(LogDefaultValue, key)
		return defaultValue
	}
	return value
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("load .env: %v", err)
	}

	cfg := new(Config)

	cfg.HTTP.Host = getEnv(keyHTTPHost, defaultHTTPHost)
	cfg.HTTP.Port = getEnv(keyHTTPPort, defaultHTTPPort)

	cfg.DB.Host = getEnv(keyDBHost, defaultDBHost)
	cfg.DB.Port = getEnv(keyDBPort, defaultDBPort)
	cfg.DB.User = getEnv(keyDBUser, defaultDBUser)
	cfg.DB.Pass = getEnv(keyDBPass, defaultDBPass)
	cfg.DB.Name = getEnv(keyDBName, defaultDBName)

	return cfg
}
