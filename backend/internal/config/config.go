package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	CORSOrigins           []string
	JWTIssuer             string
	JWTKID                string
	JWTPrivateKeyB64      string
	JWTPublicKeyB64       string
	AccessTokenTTL        time.Duration
	RefreshTokenTTLDays   int
	DemoPassword          string
	DemoTrainingCenterID  string
	AppPublicURL          string
	SMTPHost              string
	SMTPPort              int
	SMTPSecure            bool
	SMTPUser              string
	SMTPPass              string
	EmailFrom             string
	EmailDeliveryRequired bool
}

func Load() Config {
	port := getenv("PORT", "3000")
	smtpPort := getenvInt("SMTP_PORT", 587)
	appURL := strings.TrimRight(getenv("APP_PUBLIC_URL", "http://localhost:8080"), "/")
	return Config{
		Port:                  port,
		DatabaseURL:           getenv("DATABASE_URL", "postgres://iam_app:iam_app_password@localhost:5432/iam_db"),
		CORSOrigins:           splitCSV(getenv("CORS_ORIGIN", "")),
		JWTIssuer:             getenv("JWT_ISSUER", "sena-iam-service"),
		JWTKID:                getenv("JWT_KID", "local-dev-key"),
		JWTPrivateKeyB64:      getenv("JWT_PRIVATE_KEY_B64", ""),
		JWTPublicKeyB64:       getenv("JWT_PUBLIC_KEY_B64", ""),
		AccessTokenTTL:        time.Duration(getenvInt("ACCESS_TOKEN_TTL_SECONDS", 900)) * time.Second,
		RefreshTokenTTLDays:   getenvInt("REFRESH_TOKEN_TTL_DAYS", 7),
		DemoPassword:          getenv("DEMO_PASSWORD", "Demo2026!"),
		DemoTrainingCenterID:  getenv("DEMO_TRAINING_CENTER_ID", "11111111-1111-1111-1111-111111111111"),
		AppPublicURL:          appURL,
		SMTPHost:              strings.TrimSpace(getenv("SMTP_HOST", "")),
		SMTPPort:              smtpPort,
		SMTPSecure:            getenvBool("SMTP_SECURE", smtpPort == 465),
		SMTPUser:              strings.TrimSpace(getenv("SMTP_USER", "")),
		SMTPPass:              strings.TrimSpace(getenv("SMTP_PASS", "")),
		EmailFrom:             strings.TrimSpace(getenv("EMAIL_FROM", "")),
		EmailDeliveryRequired: getenvBool("EMAIL_DELIVERY_REQUIRED", false),
	}
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes" || value == "y" || value == "si"
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}