package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config encapsulates all configuration settings for the LivePoll backend service.
type Config struct {
	Port         string
	GinMode      string
	ClientOrigin string

	MongoURI    string
	MongoDBName string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret          string
	JWTExpirationHours int
}

// LoadConfig retrieves configuration parameters from environment variables with safe defaults.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found; loading configuration from environment and defaults")
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	jwtExpHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		jwtExpHours = 24
	}

	redisAddr := getEnv("REDIS_ADDR", "")
	if redisAddr == "" {
		redisAddr = getEnv("REDIS_URL", "localhost:6379")
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		ClientOrigin:       getEnv("CLIENT_ORIGIN", "http://localhost:5173"),
		MongoURI:           getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:        getEnv("MONGO_DB_NAME", "livepoll"),
		RedisAddr:          redisAddr,
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		JWTSecret:          getEnv("JWT_SECRET", "default_dev_secret_change_in_production"),
		JWTExpirationHours: jwtExpHours,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
