package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	Enabled  bool      `json:"enabled"`
	Channels []Channel `json:"channels"`
	Rules    []Rule    `json:"rules"`
}

type Channel struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	URL     string            `json:"url,omitempty"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type Rule struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Enabled    bool     `json:"enabled"`
	DeviceSlug string   `json:"device_slug"`
	Account    string   `json:"account,omitempty"`
	Zone       string   `json:"zone,omitempty"`
	Metric     string   `json:"metric"`
	Condition  string   `json:"condition"`
	Threshold  float64  `json:"threshold,omitempty"`
	ArmModes   []string `json:"arm_modes,omitempty"`
	Channels   []string `json:"channels,omitempty"`
	Cooldown   string   `json:"cooldown,omitempty"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

func Load(ctx context.Context, path string) (*Store, error) {
	store := &Store{path: path, cfg: defaultConfig()}
	if strings.TrimSpace(path) == "" {
		return store, nil
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	store.cfg = normalizeConfig(cfg)
	return store, nil
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.path
}

func (s *Store) Config() Config {
	if s == nil {
		return defaultConfig()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return copyConfig(s.cfg)
}

func (s *Store) Save(ctx context.Context, cfg Config) (Config, error) {
	if s == nil {
		return Config{}, nil
	}
	select {
	case <-ctx.Done():
		return Config{}, ctx.Err()
	default:
	}
	cfg = normalizeConfig(cfg)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	if strings.TrimSpace(s.path) == "" {
		return copyConfig(s.cfg), nil
	}
	if err := s.saveLocked(); err != nil {
		return Config{}, err
	}
	return copyConfig(s.cfg), nil
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Dir(s.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	_ = os.Remove(s.path)
	return os.Rename(tmp, s.path)
}

func defaultConfig() Config {
	return Config{
		Enabled: true,
		Channels: []Channel{{
			ID:   "log",
			Type: "log",
		}},
		Rules: []Rule{},
	}
}

func normalizeConfig(cfg Config) Config {
	if len(cfg.Channels) == 0 {
		cfg.Channels = defaultConfig().Channels
	}
	for i := range cfg.Channels {
		cfg.Channels[i].ID = slug(strings.TrimSpace(cfg.Channels[i].ID))
		if cfg.Channels[i].ID == "" {
			cfg.Channels[i].ID = "channel_" + shortHash(cfg.Channels[i].URL+cfg.Channels[i].Type)
		}
		cfg.Channels[i].Type = strings.ToLower(strings.TrimSpace(cfg.Channels[i].Type))
		if cfg.Channels[i].Type == "" {
			cfg.Channels[i].Type = "log"
		}
		cfg.Channels[i].Method = strings.ToUpper(strings.TrimSpace(cfg.Channels[i].Method))
		if cfg.Channels[i].Method == "" {
			cfg.Channels[i].Method = "POST"
		}
	}
	for i := range cfg.Rules {
		cfg.Rules[i].ID = slug(strings.TrimSpace(cfg.Rules[i].ID))
		if cfg.Rules[i].ID == "" {
			cfg.Rules[i].ID = "rule_" + shortHash(cfg.Rules[i].Name+cfg.Rules[i].DeviceSlug+cfg.Rules[i].Metric+cfg.Rules[i].Condition)
		}
		cfg.Rules[i].Name = strings.TrimSpace(cfg.Rules[i].Name)
		if cfg.Rules[i].Name == "" {
			cfg.Rules[i].Name = cfg.Rules[i].ID
		}
		cfg.Rules[i].DeviceSlug = slug(strings.TrimSpace(cfg.Rules[i].DeviceSlug))
		cfg.Rules[i].Metric = strings.ToLower(strings.TrimSpace(cfg.Rules[i].Metric))
		cfg.Rules[i].Condition = strings.ToLower(strings.TrimSpace(cfg.Rules[i].Condition))
		if cfg.Rules[i].Cooldown == "" {
			cfg.Rules[i].Cooldown = "30m"
		}
		cfg.Rules[i].ArmModes = normalizeList(cfg.Rules[i].ArmModes)
		cfg.Rules[i].Channels = normalizeList(cfg.Rules[i].Channels)
	}
	return cfg
}

func copyConfig(cfg Config) Config {
	out := cfg
	out.Channels = append([]Channel(nil), cfg.Channels...)
	for i := range out.Channels {
		if len(out.Channels[i].Headers) > 0 {
			headers := make(map[string]string, len(out.Channels[i].Headers))
			for k, v := range out.Channels[i].Headers {
				headers[k] = v
			}
			out.Channels[i].Headers = headers
		}
	}
	out.Rules = append([]Rule(nil), cfg.Rules...)
	for i := range out.Rules {
		out.Rules[i].ArmModes = append([]string(nil), cfg.Rules[i].ArmModes...)
		out.Rules[i].Channels = append([]string(nil), cfg.Rules[i].Channels...)
	}
	return out
}

func normalizeList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
