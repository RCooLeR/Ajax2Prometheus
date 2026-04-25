package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/app"
	"github.com/RCooLeR/AjaxBridge/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cfg := config.FromEnv()
	forwardAddrs := cli.NewStringSlice(cfg.ForwardAddresses()...)
	healthcheckURL := ""
	healthcheckPath := "/readyz"
	healthcheckTimeout := 3 * time.Second
	healthcheckHTTPAddr := cfg.HTTPAddr
	cliApp := &cli.App{
		Name:    "ajaxbridge",
		Usage:   "Receive Ajax SIA DC-09 events, maintain alarm state, and expose Prometheus metrics.",
		Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "sia-addr", Value: cfg.SIAListenAddr, Usage: "TCP listen address for SIA DC-09", Destination: &cfg.SIAListenAddr, EnvVars: []string{"AJAXBRIDGE_SIA_ADDR", "AJAX2PROM_SIA_ADDR"}},
			&cli.StringFlag{Name: "http-addr", Value: cfg.HTTPAddr, Usage: "HTTP listen address for /metrics", Destination: &cfg.HTTPAddr, EnvVars: []string{"AJAXBRIDGE_HTTP_ADDR", "AJAX2PROM_HTTP_ADDR"}},
			&cli.StringFlag{Name: "devices-path", Value: cfg.DevicesPath, Usage: "Path to optional device catalog JSON file; empty disables loading", Destination: &cfg.DevicesPath, EnvVars: []string{"AJAXBRIDGE_DEVICES_PATH", "AJAX2PROM_DEVICES_PATH"}},
			&cli.StringSliceFlag{Name: "forward-addr", Value: forwardAddrs, Usage: "Optional upstream SIA TCP address. Repeat flag or use comma-separated AJAXBRIDGE_FORWARD_ADDR for multiple receivers.", Destination: forwardAddrs, EnvVars: []string{"AJAXBRIDGE_FORWARD_ADDR", "AJAX2PROM_FORWARD_ADDR"}},
			&cli.DurationFlag{Name: "forward-timeout", Value: cfg.ForwardTimeout, Usage: "Timeout for forwarding one SIA frame upstream", Destination: &cfg.ForwardTimeout, EnvVars: []string{"AJAXBRIDGE_FORWARD_TIMEOUT", "AJAX2PROM_FORWARD_TIMEOUT"}},
			&cli.BoolFlag{Name: "forward-require-ack", Value: cfg.ForwardRequireACK, Usage: "Return NAK to Ajax if upstream SIA receiver does not ACK", Destination: &cfg.ForwardRequireACK, EnvVars: []string{"AJAXBRIDGE_FORWARD_REQUIRE_ACK", "AJAX2PROM_FORWARD_REQUIRE_ACK"}},
			&cli.StringFlag{Name: "mqtt-broker", Value: cfg.MQTTBroker, Usage: "Optional MQTT broker URL for Home Assistant, for example tcp://homeassistant.local:1883", Destination: &cfg.MQTTBroker, EnvVars: []string{"AJAXBRIDGE_MQTT_BROKER", "AJAX2PROM_MQTT_BROKER"}},
			&cli.StringFlag{Name: "mqtt-username", Value: cfg.MQTTUsername, Usage: "MQTT username", Destination: &cfg.MQTTUsername, EnvVars: []string{"AJAXBRIDGE_MQTT_USERNAME", "AJAX2PROM_MQTT_USERNAME"}},
			&cli.StringFlag{Name: "mqtt-password", Value: cfg.MQTTPassword, Usage: "MQTT password", Destination: &cfg.MQTTPassword, EnvVars: []string{"AJAXBRIDGE_MQTT_PASSWORD", "AJAX2PROM_MQTT_PASSWORD"}},
			&cli.StringFlag{Name: "mqtt-client-id", Value: cfg.MQTTClientID, Usage: "MQTT client ID", Destination: &cfg.MQTTClientID, EnvVars: []string{"AJAXBRIDGE_MQTT_CLIENT_ID", "AJAX2PROM_MQTT_CLIENT_ID"}},
			&cli.StringFlag{Name: "mqtt-topic-prefix", Value: cfg.MQTTTopicPrefix, Usage: "MQTT state topic prefix", Destination: &cfg.MQTTTopicPrefix, EnvVars: []string{"AJAXBRIDGE_MQTT_TOPIC_PREFIX", "AJAX2PROM_MQTT_TOPIC_PREFIX"}},
			&cli.BoolFlag{Name: "mqtt-discovery", Value: cfg.MQTTDiscovery, Usage: "Publish Home Assistant MQTT discovery configs", Destination: &cfg.MQTTDiscovery, EnvVars: []string{"AJAXBRIDGE_MQTT_DISCOVERY", "AJAX2PROM_MQTT_DISCOVERY"}},
			&cli.StringFlag{Name: "mqtt-discovery-prefix", Value: cfg.MQTTDiscoveryPrefix, Usage: "Home Assistant MQTT discovery prefix", Destination: &cfg.MQTTDiscoveryPrefix, EnvVars: []string{"AJAXBRIDGE_MQTT_DISCOVERY_PREFIX", "AJAX2PROM_MQTT_DISCOVERY_PREFIX"}},
			&cli.DurationFlag{Name: "mqtt-timeout", Value: cfg.MQTTTimeout, Usage: "MQTT connect and publish timeout", Destination: &cfg.MQTTTimeout, EnvVars: []string{"AJAXBRIDGE_MQTT_TIMEOUT", "AJAX2PROM_MQTT_TIMEOUT"}},
			&cli.BoolFlag{Name: "mqtt-retain", Value: cfg.MQTTRetain, Usage: "Retain MQTT state messages", Destination: &cfg.MQTTRetain, EnvVars: []string{"AJAXBRIDGE_MQTT_RETAIN", "AJAX2PROM_MQTT_RETAIN"}},
			&cli.StringFlag{Name: "account", Value: cfg.Account, Usage: "Expected Ajax account/object number; empty allows all accounts", Destination: &cfg.Account, EnvVars: []string{"AJAXBRIDGE_ACCOUNT", "AJAX2PROM_ACCOUNT"}},
			&cli.StringFlag{Name: "encryption-key", Value: cfg.EncryptionKey, Usage: "AES key as 32/48/64 hex characters or raw 16/24/32 bytes", Destination: &cfg.EncryptionKey, EnvVars: []string{"AJAXBRIDGE_ENCRYPTION_KEY", "AJAX2PROM_ENCRYPTION_KEY"}},
			&cli.BoolFlag{Name: "strict-crc", Value: cfg.StrictCRC, Usage: "Reject frames with invalid CRC", Destination: &cfg.StrictCRC, EnvVars: []string{"AJAXBRIDGE_STRICT_CRC", "AJAX2PROM_STRICT_CRC"}},
			&cli.DurationFlag{Name: "ping-interval", Value: cfg.PingInterval, Usage: "Ajax monitoring station ping interval", Destination: &cfg.PingInterval, EnvVars: []string{"AJAXBRIDGE_PING_INTERVAL", "AJAX2PROM_PING_INTERVAL"}},
			&cli.DurationFlag{Name: "offline-grace", Value: cfg.OfflineGrace, Usage: "Grace window before account is marked offline", Destination: &cfg.OfflineGrace, EnvVars: []string{"AJAXBRIDGE_OFFLINE_GRACE", "AJAX2PROM_OFFLINE_GRACE"}},
			&cli.DurationFlag{Name: "read-timeout", Value: cfg.ReadTimeout, Usage: "Per-connection SIA read timeout", Destination: &cfg.ReadTimeout, EnvVars: []string{"AJAXBRIDGE_READ_TIMEOUT", "AJAX2PROM_READ_TIMEOUT"}},
			&cli.StringFlag{Name: "log-level", Value: cfg.LogLevel, Usage: "Log level: trace, debug, info, warn, error", Destination: &cfg.LogLevel, EnvVars: []string{"AJAXBRIDGE_LOG_LEVEL", "AJAX2PROM_LOG_LEVEL"}},
			&cli.BoolFlag{Name: "log-pretty", Value: cfg.LogPretty, Usage: "Enable console log output", Destination: &cfg.LogPretty, EnvVars: []string{"AJAXBRIDGE_LOG_PRETTY", "AJAX2PROM_LOG_PRETTY"}},
		},
		Action: func(_ *cli.Context) error {
			cfg.SIAForwardAddrs = forwardAddrs.Value()
			logger, err := configureLogger(cfg)
			if err != nil {
				return err
			}
			return app.Run(context.Background(), cfg, logger)
		},
		Commands: []*cli.Command{
			{
				Name:  "healthcheck",
				Usage: "Probe the local HTTP readiness endpoint and exit non-zero on failure.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "url", Usage: "Full healthcheck URL to probe", Destination: &healthcheckURL},
					&cli.StringFlag{Name: "http-addr", Value: cfg.HTTPAddr, Usage: "HTTP listen address used to derive the probe URL when --url is not set", Destination: &healthcheckHTTPAddr, EnvVars: []string{"AJAXBRIDGE_HTTP_ADDR", "AJAX2PROM_HTTP_ADDR"}},
					&cli.StringFlag{Name: "path", Value: healthcheckPath, Usage: "HTTP path to probe when --url is not set", Destination: &healthcheckPath},
					&cli.DurationFlag{Name: "timeout", Value: healthcheckTimeout, Usage: "HTTP probe timeout", Destination: &healthcheckTimeout},
				},
				Action: func(_ *cli.Context) error {
					probeURL, err := resolveHealthcheckURL(healthcheckURL, healthcheckHTTPAddr, healthcheckPath)
					if err != nil {
						return err
					}
					return runHealthcheck(context.Background(), probeURL, healthcheckTimeout)
				},
			},
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func configureLogger(cfg config.Config) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return zerolog.Logger{}, err
	}
	zerolog.SetGlobalLevel(level)
	if cfg.LogPretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339, NoColor: true})
	}
	return log.With().Timestamp().Logger(), nil
}

func runHealthcheck(ctx context.Context, probeURL string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("healthcheck %s returned HTTP %d", probeURL, resp.StatusCode)
	}
	return nil
}

func resolveHealthcheckURL(rawURL, httpAddr, path string) (string, error) {
	if rawURL != "" {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return "", err
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return "", fmt.Errorf("healthcheck URL must include scheme and host: %q", rawURL)
		}
		return parsed.String(), nil
	}

	normalizedPath := path
	if normalizedPath == "" {
		normalizedPath = "/readyz"
	}
	if !strings.HasPrefix(normalizedPath, "/") {
		normalizedPath = "/" + normalizedPath
	}

	host, port, err := splitHTTPAddr(httpAddr)
	if err != nil {
		return "", err
	}
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "127.0.0.1"
	}
	return (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
		Path:   normalizedPath,
	}).String(), nil
}

func splitHTTPAddr(addr string) (string, string, error) {
	if addr == "" {
		return "", "", fmt.Errorf("http address is empty")
	}
	if strings.HasPrefix(addr, ":") {
		return "", strings.TrimPrefix(addr, ":"), nil
	}

	host, port, err := net.SplitHostPort(addr)
	if err == nil {
		return host, port, nil
	}

	if !strings.Contains(addr, ":") {
		return "", "", fmt.Errorf("http address must include a port: %q", addr)
	}
	return "", "", fmt.Errorf("invalid http address %q: %w", addr, err)
}
