package config

import (
	"crypto/tls"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Host         string
	Port         string
	CertFile     string
	KeyFile      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	TLSConfig    *tls.Config
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	return &Config{
		Host:         "127.0.0.1",
		Port:         os.Getenv("PORT"),
		CertFile:     os.Getenv("CERT_FILE"),
		KeyFile:      os.Getenv("KEY_FILE"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 5),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 5),
	}, nil
}

func getDurationEnv(key string, def int) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return time.Duration(def) * time.Second
	}

	intValue, err := strconv.Atoi(val)
	if err != nil {
		return time.Duration(def) * time.Second
	}

	return time.Duration(intValue) * time.Second
}
