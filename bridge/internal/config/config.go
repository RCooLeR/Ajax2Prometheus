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

	MQTTBroker          string
	MQTTUsername        string
	MQTTPassword        string
	MQTTClientID        string
	MQTTTopicPrefix     string
	MQTTDiscovery       bool
	MQTTDiscoveryPrefix string
	MQTTTimeout         time.Duration
	MQTTRetain          bool

	LogLevel  string
	LogPretty bool
}

func FromEnv() Config {
	return Config{
		SIAListenAddr:     envString(":8099", "AJAXBRIDGE_SIA_ADDR", "AJAX2PROM_SIA_ADDR"),
		HTTPAddr:          envString(":8080", "AJAXBRIDGE_HTTP_ADDR", "AJAX2PROM_HTTP_ADDR"),
		SIAForwardAddr:    envValue("AJAXBRIDGE_FORWARD_ADDR", "AJAX2PROM_FORWARD_ADDR"),
		SIAForwardAddrs:   envCSV("AJAXBRIDGE_FORWARD_ADDR", "AJAX2PROM_FORWARD_ADDR"),
		DevicesPath:       envString("data/devices.json", "AJAXBRIDGE_DEVICES_PATH", "AJAX2PROM_DEVICES_PATH"),
		Account:           envValue("AJAXBRIDGE_ACCOUNT", "AJAX2PROM_ACCOUNT"),
		EncryptionKey:     envValue("AJAXBRIDGE_ENCRYPTION_KEY", "AJAX2PROM_ENCRYPTION_KEY"),
		StrictCRC:         envBool(true, "AJAXBRIDGE_STRICT_CRC", "AJAX2PROM_STRICT_CRC"),
		PingInterval:      envDuration(time.Minute, "AJAXBRIDGE_PING_INTERVAL", "AJAX2PROM_PING_INTERVAL"),
		OfflineGrace:      envDuration(3*time.Minute, "AJAXBRIDGE_OFFLINE_GRACE", "AJAX2PROM_OFFLINE_GRACE"),
		ReadTimeout:       envDuration(2*time.Minute, "AJAXBRIDGE_READ_TIMEOUT", "AJAX2PROM_READ_TIMEOUT"),
		ForwardTimeout:    envDuration(5*time.Second, "AJAXBRIDGE_FORWARD_TIMEOUT", "AJAX2PROM_FORWARD_TIMEOUT"),
		ForwardRequireACK: envBool(false, "AJAXBRIDGE_FORWARD_REQUIRE_ACK", "AJAX2PROM_FORWARD_REQUIRE_ACK"),
		MQTTBroker:        envValue("AJAXBRIDGE_MQTT_BROKER", "AJAX2PROM_MQTT_BROKER"),
		MQTTUsername:      envValue("AJAXBRIDGE_MQTT_USERNAME", "AJAX2PROM_MQTT_USERNAME"),
		MQTTPassword:      envValue("AJAXBRIDGE_MQTT_PASSWORD", "AJAX2PROM_MQTT_PASSWORD"),
		MQTTClientID:      envString("ajaxbridge", "AJAXBRIDGE_MQTT_CLIENT_ID", "AJAX2PROM_MQTT_CLIENT_ID"),
		MQTTTopicPrefix:   envString("ajaxbridge", "AJAXBRIDGE_MQTT_TOPIC_PREFIX", "AJAX2PROM_MQTT_TOPIC_PREFIX"),
		MQTTDiscovery:     envBool(true, "AJAXBRIDGE_MQTT_DISCOVERY", "AJAX2PROM_MQTT_DISCOVERY"),
		MQTTDiscoveryPrefix: envString(
			"homeassistant",
			"AJAXBRIDGE_MQTT_DISCOVERY_PREFIX",
			"AJAX2PROM_MQTT_DISCOVERY_PREFIX",
		),
		MQTTTimeout: envDuration(5*time.Second, "AJAXBRIDGE_MQTT_TIMEOUT", "AJAX2PROM_MQTT_TIMEOUT"),
		MQTTRetain:  envBool(true, "AJAXBRIDGE_MQTT_RETAIN", "AJAX2PROM_MQTT_RETAIN"),
		LogLevel:    envString("info", "AJAXBRIDGE_LOG_LEVEL", "AJAX2PROM_LOG_LEVEL"),
		LogPretty:   envBool(false, "AJAXBRIDGE_LOG_PRETTY", "AJAX2PROM_LOG_PRETTY"),
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
	if c.MQTTBroker != "" && c.MQTTTimeout <= 0 {
		return fmt.Errorf("MQTT timeout must be positive: %s", c.MQTTTimeout)
	}
	if c.MQTTBroker != "" && strings.TrimSpace(c.MQTTTopicPrefix) == "" {
		return errors.New("MQTT topic prefix is required when MQTT is enabled")
	}
	if c.MQTTBroker != "" && c.MQTTDiscovery && strings.TrimSpace(c.MQTTDiscoveryPrefix) == "" {
		return errors.New("MQTT discovery prefix is required when MQTT discovery is enabled")
	}
	return nil
}

func (c Config) MQTTEnabled() bool {
	return strings.TrimSpace(c.MQTTBroker) != ""
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

func envValue(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func envString(fallback string, keys ...string) string {
	if value := envValue(keys...); value != "" {
		return value
	}
	return fallback
}

func envBool(fallback bool, keys ...string) bool {
	switch envValue(keys...) {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}

func envDuration(fallback time.Duration, keys ...string) time.Duration {
	value := envValue(keys...)
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func envCSV(keys ...string) []string {
	return parseCSV(envValue(keys...))
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
