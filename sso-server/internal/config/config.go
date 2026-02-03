package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Session  SessionConfig
	Email    EmailConfig
	Security SecurityConfig
	TwoFA    TwoFAConfig
	OAuth2   OAuth2Config
	Log      LogConfig
	Env      string
}

type ServerConfig struct {
	Port    string
	Host    string
	BaseURL string
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

type SessionConfig struct {
	Timeout        time.Duration
	CookieName     string
	CookieSecure   bool
	CookieHTTPOnly bool
}

type EmailConfig struct {
	SMTPHost         string
	SMTPPort         string
	SMTPUser         string
	SMTPPassword     string
	FromEmail        string
	FromName         string
	PasswordResetURL string
}

type SecurityConfig struct {
	PasswordMinLength      int
	PasswordRequireUpper   bool
	PasswordRequireLower   bool
	PasswordRequireNumber  bool
	PasswordRequireSpecial bool
	PasswordExpiryDays     int
	PasswordHistoryCount   int
	MaxLoginAttempts       int
	AccountLockoutDuration time.Duration
}

type TwoFAConfig struct {
	Issuer string
	Period int
	Digits int
}

type OAuth2Config struct {
	AuthCodeExpiry     time.Duration
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	EnforcePKCE        bool
	Issuer             string
}

type LogConfig struct {
	Level  string
	Format string
	Output string
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Don't error if .env file doesn't exist, just use env vars
	_ = viper.ReadInConfig()

	cfg := &Config{
		Server: ServerConfig{
			Port:    viper.GetString("SERVER_PORT"),
			Host:    viper.GetString("SERVER_HOST"),
			BaseURL: viper.GetString("SERVER_BASE_URL"),
		},
		Database: DatabaseConfig{
			Host:         viper.GetString("DB_HOST"),
			Port:         viper.GetString("DB_PORT"),
			User:         viper.GetString("DB_USER"),
			Password:     viper.GetString("DB_PASSWORD"),
			Name:         viper.GetString("DB_NAME"),
			MaxIdleConns: viper.GetInt("DB_MAX_IDLE_CONNS"),
			MaxOpenConns: viper.GetInt("DB_MAX_OPEN_CONNS"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:             viper.GetString("JWT_SECRET"),
			AccessTokenExpiry:  viper.GetDuration("JWT_ACCESS_TOKEN_EXPIRY"),
			RefreshTokenExpiry: viper.GetDuration("JWT_REFRESH_TOKEN_EXPIRY"),
			Issuer:             viper.GetString("JWT_ISSUER"),
		},
		Session: SessionConfig{
			Timeout:        viper.GetDuration("SESSION_TIMEOUT"),
			CookieName:     viper.GetString("SESSION_COOKIE_NAME"),
			CookieSecure:   viper.GetBool("SESSION_COOKIE_SECURE"),
			CookieHTTPOnly: viper.GetBool("SESSION_COOKIE_HTTP_ONLY"),
		},
		Email: EmailConfig{
			SMTPHost:         viper.GetString("smtp.host"),
			SMTPPort:         viper.GetString("smtp.port"),
			SMTPUser:         viper.GetString("smtp.user"),
			SMTPPassword:     viper.GetString("smtp.password"),
			FromEmail:        viper.GetString("email.from"),
			FromName:         viper.GetString("email.from_name"),
			PasswordResetURL: viper.GetString("email.password_reset_url"),
		},
		Security: SecurityConfig{
			PasswordMinLength:      viper.GetInt("PASSWORD_MIN_LENGTH"),
			PasswordRequireUpper:   viper.GetBool("PASSWORD_REQUIRE_UPPERCASE"),
			PasswordRequireLower:   viper.GetBool("PASSWORD_REQUIRE_LOWERCASE"),
			PasswordRequireNumber:  viper.GetBool("PASSWORD_REQUIRE_NUMBER"),
			PasswordRequireSpecial: viper.GetBool("PASSWORD_REQUIRE_SPECIAL"),
			PasswordExpiryDays:     viper.GetInt("PASSWORD_EXPIRY_DAYS"),
			PasswordHistoryCount:   viper.GetInt("PASSWORD_HISTORY_COUNT"),
			MaxLoginAttempts:       viper.GetInt("MAX_LOGIN_ATTEMPTS"),
			AccountLockoutDuration: viper.GetDuration("ACCOUNT_LOCKOUT_DURATION"),
		},
		TwoFA: TwoFAConfig{
			Issuer: viper.GetString("TOTP_ISSUER"),
			Period: viper.GetInt("TOTP_PERIOD"),
			Digits: viper.GetInt("TOTP_DIGITS"),
		},
		OAuth2: OAuth2Config{
			AuthCodeExpiry:     viper.GetDuration("OAUTH2_AUTH_CODE_EXPIRY"),
			AccessTokenExpiry:  viper.GetDuration("OAUTH2_ACCESS_TOKEN_EXPIRY"),
			RefreshTokenExpiry: viper.GetDuration("OAUTH2_REFRESH_TOKEN_EXPIRY"),
			EnforcePKCE:        viper.GetBool("OAUTH2_ENFORCE_PKCE"),
			Issuer:             viper.GetString("OAUTH2_ISSUER"),
		},
		Log: LogConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
			Output: viper.GetString("LOG_OUTPUT"),
		},
		Env: viper.GetString("ENV"),
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if all required configuration values are set
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}
