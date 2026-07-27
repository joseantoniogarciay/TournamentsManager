// Package config carga y valida la configuración de ejecución de la API.
package config

import (
	"fmt"
	"net/url"
	"os"
)

const (
	databaseURLEnv = "DATABASE_URL"
	httpAddrEnv    = "HTTP_ADDR"
)

// Config contiene únicamente la configuración necesaria para arrancar la API.
type Config struct {
	DatabaseURL string
	HTTPAddr    string
}

// Load obtiene la configuración desde el entorno y falla antes de abrir puertos
// o conexiones cuando falta un valor obligatorio.
func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	databaseURL := getenv(databaseURLEnv)
	if databaseURL == "" {
		return Config{}, fmt.Errorf("%s debe estar definido", databaseURLEnv)
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return Config{}, fmt.Errorf("%s no es una URL válida: %w", databaseURLEnv, err)
	}
	if (parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql") || parsedURL.Host == "" {
		return Config{}, fmt.Errorf("%s debe usar una URL PostgreSQL con host", databaseURLEnv)
	}

	httpAddr := getenv(httpAddrEnv)
	if httpAddr == "" {
		return Config{}, fmt.Errorf("%s debe estar definido", httpAddrEnv)
	}

	return Config{
		DatabaseURL: databaseURL,
		HTTPAddr:    httpAddr,
	}, nil
}
