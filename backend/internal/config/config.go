package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	MongoURI               string
	MongoDB                string
	RedisAddr              string
	JWTSecret              string
	JWTExpirationHours     int
	AIConfidenceThreshold  float64
	FollowUpThresholdHours int
	AutoClassifyEnabled    bool
	ClientURL              string
	OpenAIAPIKey           string
	SMTPHost               string
	SMTPPort               int
	SMTPUser               string
	SMTPPass               string
	FromEmail              string
}

func LoadConfig() *Config {
	// Attempt loading .env if present
	_ = godotenv.Load(".env", "/home/mahesh/one-time-call-alert/backend/.env")

	return &Config{
		Port:                   getEnv("PORT", "8085"),
		MongoURI:               getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:                getEnv("MONGO_DB", "lead_followup_db"),
		RedisAddr:              getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:              getEnv("JWT_SECRET", "super-secret-jwt-key-for-lead-followup-system"),
		JWTExpirationHours:     getEnvInt("JWT_EXPIRATION_HOURS", 24),
		AIConfidenceThreshold:  getEnvFloat("AI_CONFIDENCE_THRESHOLD", 0.90),
		FollowUpThresholdHours: getEnvInt("FOLLOWUP_THRESHOLD_HOURS", 24),
		AutoClassifyEnabled:    getEnvBool("AUTO_CLASSIFY_ENABLED", true),
		ClientURL:              getEnv("CLIENT_URL", "http://localhost:5173"),
		OpenAIAPIKey:           getEnv("OPENAI_API_KEY", ""),
		SMTPHost:               getEnv("SMTP_HOST", "smtp.mailtrap.io"),
		SMTPPort:               getEnvInt("SMTP_PORT", 2525),
		SMTPUser:               getEnv("SMTP_USER", ""),
		SMTPPass:               getEnv("SMTP_PASS", ""),
		FromEmail:              getEnv("FROM_EMAIL", "alerts@leadfollowup.com"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}
