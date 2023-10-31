package config

import (
	"os"

	"github.com/torwig/user-service/adapters/repository"
	"github.com/torwig/user-service/log"
	"github.com/torwig/user-service/ports/http"
	"github.com/torwig/user-service/ports/http/jwt"
)

const (
	defaultLogLevel       = "info"
	defaultBindAddress    = ":8080"
	envKeyLogLevel        = "USERS_LOG_LEVEL"
	envKeyRepositoryURI   = "USERS_REPOSITORY_URI"
	envKeyJWTSecret       = "USERS_JWT_SECRET" // #nosec G101
	envKeyJWTIssuer       = "USERS_JWT_ISSUER"
	envKeyHTTPBindAddress = "USERS_HTTP_BIND_ADDRESS"
)

type Config struct {
	Log        log.Config
	Repository repository.Config
	JWT        jwt.Config
	HTTP       http.Config
}

func CreateFromEnv() *Config {
	cfg := &Config{
		Log:        createLogConfig(),
		Repository: createRepositoryConfig(),
		JWT:        createJWTConfig(),
		HTTP:       createHTTPConfig(),
	}

	return cfg
}

func createLogConfig() log.Config {
	level := os.Getenv(envKeyLogLevel)
	if level == "" {
		level = defaultLogLevel
	}

	return log.Config{Level: level}
}

func createRepositoryConfig() repository.Config {
	return repository.Config{DSN: os.Getenv(envKeyRepositoryURI)}
}

func createJWTConfig() jwt.Config {
	return jwt.Config{
		SecretKey: []byte(os.Getenv(envKeyJWTSecret)),
		Issuer:    os.Getenv(envKeyJWTIssuer),
	}
}

func createHTTPConfig() http.Config {
	bindAddress := os.Getenv(envKeyHTTPBindAddress)
	if bindAddress == "" {
		bindAddress = defaultBindAddress
	}

	return http.Config{BindAddress: bindAddress}
}
