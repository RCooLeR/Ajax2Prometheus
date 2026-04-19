package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	SIAListenAddr   string
	HTTPAddr        string
	SIAForwardAddr  string
	SIAForwardAddrs []string
	DevicesPath     string

	Account       string
	EncryptionKey string
	StrictCRC     bool

	PingInterval      time.Duration
	OfflineGrace      time.Duration
	ReadTimeout       time.Duration
	ForwardTimeout    time.Duration
	ForwardRequireACK bool

	LogLevel  string
	LogPretty bool
}

func FromEnv() Config {
	return Config{
		SIAListenAddr:     envString("AJAX2PROM_SIA_ADDR", ":8099"),
		HTTPAddr:          envString("AJAX2PROM_HTTP_ADDR", ":8080"),
		SIAForwardAddr:    os.Getenv("AJAX2PROM_FORWARD_ADDR"),
		SIAForwardAddrs:   envCSV("AJAX2PROM_FORWARD_ADDR"),
		DevicesPath:       envString("AJAX2PROM_DEVICES_PATH", "data/devices.json"),
		Account:           os.Getenv("AJAX2PROM_ACCOUNT"),
		EncryptionKey:     os.Getenv("AJAX2PROM_ENCRYPTION_KEY"),
		StrictCRC:         envBool("AJAX2PROM_STRICT_CRC", true),
		PingInterval:      envDuration("AJAX2PROM_PING_INTERVAL", time.Minute),
		OfflineGrace:      envDuration("AJAX2PROM_OFFLINE_GRACE", 3*time.Minute),
		ReadTimeout:       envDuration("AJAX2PROM_READ_TIMEOUT", 2*time.Minute),
		ForwardTimeout:    envDuration("AJAX2PROM_FORWARD_TIMEOUT", 5*time.Second),
		ForwardRequireACK: envBool("AJAX2PROM_FORWARD_REQUIRE_ACK", false),
		LogLevel:          envString("AJAX2PROM_LOG_LEVEL", "info"),
		LogPretty:         envBool("AJAX2PROM_LOG_PRETTY", false),
	}
}

func (c Config) Validate() error {
	if c.SIAListenAddr == "" {
		return errors.New("SIA listen address is required")
	}
	if c.HTTPAddr == "" {
		return errors.New("HTTP address is required")
	}
	if c.PingInterval <= 0 {
		return fmt.Errorf("ping interval must be positive: %s", c.PingInterval)
	}
	if c.OfflineGrace <= 0 {
		return fmt.Errorf("offline grace must be positive: %s", c.OfflineGrace)
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive: %s", c.ReadTimeout)
	}
	if len(c.ForwardAddresses()) > 0 && c.ForwardTimeout <= 0 {
		return fmt.Errorf("forward timeout must be positive: %s", c.ForwardTimeout)
	}
	return nil
}

func (c Config) ForwardAddresses() []string {
	if len(c.SIAForwardAddrs) > 0 {
		var out []string
		for _, value := range c.SIAForwardAddrs {
			out = append(out, parseCSV(value)...)
		}
		return compactStrings(out)
	}
	return parseCSV(c.SIAForwardAddr)
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	switch os.Getenv(key) {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func envCSV(key string) []string {
	return parseCSV(os.Getenv(key))
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	return compactStrings(parts)
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
