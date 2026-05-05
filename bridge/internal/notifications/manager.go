package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/jeedom"
	"github.com/RCooLeR/AjaxBridge/internal/state"
	"github.com/rs/zerolog"
)

type Manager struct {
	store      *Store
	client     *http.Client
	log        zerolog.Logger
	mu         sync.Mutex
	lastValues map[string]any
	lastSent   map[string]time.Time
	history    []Delivery
}

type Event struct {
	Kind          string    `json:"kind"`
	Source        string    `json:"source"`
	DeviceSlug    string    `json:"device_slug"`
	Device        string    `json:"device"`
	DeviceType    string    `json:"device_type,omitempty"`
	Account       string    `json:"account,omitempty"`
	Zone          string    `json:"zone,omitempty"`
	Metric        string    `json:"metric,omitempty"`
	Action        string    `json:"action,omitempty"`
	Value         any       `json:"value,omitempty"`
	PreviousValue any       `json:"previous_value,omitempty"`
	NumericValue  float64   `json:"numeric_value,omitempty"`
	HasNumeric    bool      `json:"has_numeric,omitempty"`
	BoolValue     bool      `json:"bool_value,omitempty"`
	HasBool       bool      `json:"has_bool,omitempty"`
	ArmMode       string    `json:"arm_mode,omitempty"`
	Time          time.Time `json:"time"`
}

type Alert struct {
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Rule      Rule      `json:"rule"`
	Event     Event     `json:"event"`
	CreatedAt time.Time `json:"created_at"`
}

type Delivery struct {
	Time      time.Time `json:"time"`
	RuleID    string    `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	ChannelID string    `json:"channel_id"`
	EventKind string    `json:"event_kind"`
	Device    string    `json:"device"`
	Metric    string    `json:"metric,omitempty"`
	Action    string    `json:"action,omitempty"`
	Message   string    `json:"message"`
	Result    string    `json:"result"`
	Error     string    `json:"error,omitempty"`
}

func NewManager(store *Store, log zerolog.Logger) *Manager {
	return &Manager{
		store:      store,
		client:     &http.Client{Timeout: 5 * time.Second},
		log:        log,
		lastValues: make(map[string]any),
		lastSent:   make(map[string]time.Time),
	}
}

func (m *Manager) Store() *Store {
	if m == nil {
		return nil
	}
	return m.store
}

func (m *Manager) Config() Config {
	if m == nil || m.store == nil {
		return defaultConfig()
	}
	return m.store.Config()
}

func (m *Manager) SaveConfig(ctx context.Context, cfg Config) (Config, error) {
	if m == nil || m.store == nil {
		return Config{}, nil
	}
	return m.store.Save(ctx, cfg)
}

func (m *Manager) History(limit int) []Delivery {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if limit <= 0 || limit > len(m.history) {
		limit = len(m.history)
	}
	out := make([]Delivery, 0, limit)
	for i := len(m.history) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, m.history[i])
	}
	return out
}

func (m *Manager) ObserveJeedomUpdate(ctx context.Context, result jeedom.ApplyResult, snapshot state.Snapshot) {
	if m == nil || result.Command.Metric == "" || result.Device.DeviceSlug == "" || !result.UpdatedValue {
		return
	}
	key := result.Device.DeviceSlug + "/" + result.Command.Metric
	value := result.Command.Value
	m.mu.Lock()
	previous, hadPrevious := m.lastValues[key]
	m.lastValues[key] = value
	m.mu.Unlock()

	number, hasNumber := numberValue(value)
	boolValue, hasBool := boolValue(value)
	event := Event{
		Kind:          "metric",
		Source:        "jeedom",
		DeviceSlug:    result.Device.DeviceSlug,
		Device:        result.Device.Device,
		DeviceType:    firstNonEmpty(result.Device.JeedomDeviceType, result.Device.HAModel),
		Account:       result.Device.LinkedAccount,
		Zone:          result.Device.LinkedZone,
		Metric:        result.Command.Metric,
		Value:         value,
		PreviousValue: previous,
		NumericValue:  number,
		HasNumeric:    hasNumber,
		BoolValue:     boolValue,
		HasBool:       hasBool,
		Time:          time.Now().UTC(),
	}
	event.ArmMode = armMode(snapshot, firstNonEmpty(event.Account, result.Command.ObjectName))
	m.evaluate(ctx, event, hadPrevious)
}

func (m *Manager) ObserveJeedomControl(ctx context.Context, result jeedom.ControlResult, err error, snapshot state.Snapshot) {
	if m == nil || result.DeviceSlug == "" {
		return
	}
	event := Event{
		Kind:       "control",
		Source:     "jeedom",
		DeviceSlug: result.DeviceSlug,
		Device:     result.Device,
		DeviceType: result.DeviceType,
		Account:    result.Account,
		Zone:       result.Zone,
		Action:     result.Action,
		Time:       time.Now().UTC(),
	}
	event.ArmMode = armMode(snapshot, event.Account)
	if err != nil {
		event.Kind = "control_failed"
	}
	m.evaluate(ctx, event, true)
}

func (m *Manager) evaluate(ctx context.Context, event Event, hadPrevious bool) {
	cfg := m.Config()
	if !cfg.Enabled {
		return
	}
	channels := channelsByID(cfg.Channels)
	for _, rule := range cfg.Rules {
		if !rule.Enabled || !ruleMatchesEvent(rule, event, hadPrevious) || !ruleMatchesArm(rule, event.ArmMode) {
			continue
		}
		if !m.cooldownReady(rule, event.Time) {
			continue
		}
		alert := Alert{
			Title:     "AjaxBridge: " + rule.Name,
			Message:   alertMessage(rule, event),
			Rule:      rule,
			Event:     event,
			CreatedAt: event.Time,
		}
		targets := rule.Channels
		if len(targets) == 0 {
			for _, channel := range cfg.Channels {
				targets = append(targets, channel.ID)
			}
		}
		for _, channelID := range targets {
			channel, ok := channels[channelID]
			if !ok {
				m.record(rule, Channel{ID: channelID}, event, alert.Message, "failed", "channel not found")
				continue
			}
			if err := m.deliver(ctx, channel, alert); err != nil {
				m.record(rule, channel, event, alert.Message, "failed", err.Error())
				continue
			}
			m.record(rule, channel, event, alert.Message, "sent", "")
		}
		m.markSent(rule, event.Time)
	}
}

func (m *Manager) cooldownReady(rule Rule, now time.Time) bool {
	cooldown, err := time.ParseDuration(firstNonEmpty(rule.Cooldown, "30m"))
	if err != nil || cooldown < 0 {
		cooldown = 30 * time.Minute
	}
	key := rule.ID
	m.mu.Lock()
	defer m.mu.Unlock()
	last, ok := m.lastSent[key]
	return !ok || cooldown == 0 || now.Sub(last) >= cooldown
}

func (m *Manager) markSent(rule Rule, now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastSent[rule.ID] = now
}

func (m *Manager) deliver(ctx context.Context, channel Channel, alert Alert) error {
	switch strings.ToLower(strings.TrimSpace(channel.Type)) {
	case "", "log":
		m.log.Info().
			Str("rule", alert.Rule.ID).
			Str("device", alert.Event.DeviceSlug).
			Str("metric", alert.Event.Metric).
			Str("action", alert.Event.Action).
			Msg(alert.Message)
		return nil
	case "ntfy":
		return m.deliverNtfy(ctx, channel, alert)
	default:
		return m.deliverWebhook(ctx, channel, alert)
	}
}

func (m *Manager) deliverWebhook(ctx context.Context, channel Channel, alert Alert) error {
	if strings.TrimSpace(channel.URL) == "" {
		return fmt.Errorf("webhook channel %q has no URL", channel.ID)
	}
	body, err := json.Marshal(alert)
	if err != nil {
		return err
	}
	method := firstNonEmpty(channel.Method, http.MethodPost)
	req, err := http.NewRequestWithContext(ctx, method, channel.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range channel.Headers {
		req.Header.Set(key, value)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (m *Manager) deliverNtfy(ctx context.Context, channel Channel, alert Alert) error {
	if strings.TrimSpace(channel.URL) == "" {
		return fmt.Errorf("ntfy channel %q has no URL", channel.ID)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, channel.URL, strings.NewReader(alert.Message))
	if err != nil {
		return err
	}
	req.Header.Set("Title", alert.Title)
	req.Header.Set("Tags", "warning")
	for key, value := range channel.Headers {
		req.Header.Set(key, value)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (m *Manager) record(rule Rule, channel Channel, event Event, message, result, errText string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.history = append(m.history, Delivery{
		Time:      time.Now().UTC(),
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		ChannelID: channel.ID,
		EventKind: event.Kind,
		Device:    firstNonEmpty(event.Device, event.DeviceSlug),
		Metric:    event.Metric,
		Action:    event.Action,
		Message:   message,
		Result:    result,
		Error:     errText,
	})
	if len(m.history) > 200 {
		m.history = append([]Delivery(nil), m.history[len(m.history)-200:]...)
	}
}

func ruleMatchesEvent(rule Rule, event Event, hadPrevious bool) bool {
	if rule.DeviceSlug != "" && rule.DeviceSlug != event.DeviceSlug {
		return false
	}
	if rule.Account != "" && rule.Account != event.Account {
		return false
	}
	if rule.Zone != "" && rule.Zone != event.Zone {
		return false
	}
	condition := strings.ToLower(strings.TrimSpace(rule.Condition))
	switch condition {
	case "above":
		return event.Kind == "metric" && rule.Metric == event.Metric && event.HasNumeric && event.NumericValue > rule.Threshold
	case "below":
		return event.Kind == "metric" && rule.Metric == event.Metric && event.HasNumeric && event.NumericValue < rule.Threshold
	case "changed":
		return event.Kind == "metric" && rule.Metric == event.Metric && hadPrevious && !equalValue(event.PreviousValue, event.Value)
	case "changed_to_on":
		return event.Kind == "metric" && rule.Metric == event.Metric && hadPrevious && !equalValue(event.PreviousValue, event.Value) && event.HasBool && event.BoolValue
	case "changed_to_off":
		return event.Kind == "metric" && rule.Metric == event.Metric && hadPrevious && !equalValue(event.PreviousValue, event.Value) && event.HasBool && !event.BoolValue
	case "control":
		return event.Kind == "control"
	case "control_on":
		return event.Kind == "control" && event.Action == "on"
	case "control_off":
		return event.Kind == "control" && event.Action == "off"
	default:
		return false
	}
}

func ruleMatchesArm(rule Rule, mode string) bool {
	if len(rule.ArmModes) == 0 {
		return true
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	for _, allowed := range rule.ArmModes {
		switch strings.ToLower(strings.TrimSpace(allowed)) {
		case "", "any", "all":
			return true
		case mode:
			return true
		case "armed":
			if mode == "armed" || mode == "night" {
				return true
			}
		}
	}
	return false
}

func alertMessage(rule Rule, event Event) string {
	device := firstNonEmpty(event.Device, event.DeviceSlug, "device")
	switch event.Kind {
	case "control":
		return fmt.Sprintf("%s control requested: %s", device, strings.ToUpper(event.Action))
	case "control_failed":
		return fmt.Sprintf("%s control failed: %s", device, strings.ToUpper(event.Action))
	}
	switch rule.Condition {
	case "above":
		return fmt.Sprintf("%s %s is %.2f, above %.2f", device, event.Metric, event.NumericValue, rule.Threshold)
	case "below":
		return fmt.Sprintf("%s %s is %.2f, below %.2f", device, event.Metric, event.NumericValue, rule.Threshold)
	case "changed_to_on":
		return fmt.Sprintf("%s %s turned ON", device, event.Metric)
	case "changed_to_off":
		return fmt.Sprintf("%s %s turned OFF", device, event.Metric)
	case "changed":
		return fmt.Sprintf("%s %s changed from %v to %v", device, event.Metric, event.PreviousValue, event.Value)
	default:
		return fmt.Sprintf("%s %s matched %s", device, event.Metric, rule.Condition)
	}
}

func channelsByID(channels []Channel) map[string]Channel {
	out := make(map[string]Channel, len(channels))
	for _, channel := range channels {
		out[channel.ID] = channel
	}
	return out
}

func armMode(snapshot state.Snapshot, accountID string) string {
	if accountID != "" {
		for _, account := range snapshot.Accounts {
			if account.Account == accountID {
				return account.Mode
			}
		}
	}
	if len(snapshot.Accounts) == 1 {
		return snapshot.Accounts[0].Mode
	}
	return "unknown"
}

func numberValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(typed), ",", "."), 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func boolValue(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "1", "true", "on", "open", "opened", "active":
			return true, true
		case "0", "false", "off", "closed", "inactive":
			return false, true
		}
	case float64:
		return typed != 0, true
	case int:
		return typed != 0, true
	}
	return false, false
}

func equalValue(a, b any) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
