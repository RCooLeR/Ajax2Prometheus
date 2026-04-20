package hamqtt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/RCooLeR/Ajax2Prometheus/internal/state"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
)

const (
	payloadOnline  = "online"
	payloadOffline = "offline"
	payloadOn      = "ON"
	payloadOff     = "OFF"
)

type Config struct {
	Broker          string
	Username        string
	Password        string
	ClientID        string
	TopicPrefix     string
	Discovery       bool
	DiscoveryPrefix string
	Timeout         time.Duration
	Retain          bool
}

type Publisher struct {
	cfg        Config
	client     paho.Client
	log        zerolog.Logger
	mu         sync.Mutex
	discovered map[string]struct{}
}

type deviceInfo struct {
	Identifiers   []string `json:"identifiers"`
	Name          string   `json:"name"`
	Manufacturer  string   `json:"manufacturer"`
	Model         string   `json:"model,omitempty"`
	SuggestedArea string   `json:"suggested_area,omitempty"`
}

type discoveryConfig struct {
	Name                string     `json:"name"`
	UniqueID            string     `json:"unique_id"`
	StateTopic          string     `json:"state_topic"`
	ValueTemplate       string     `json:"value_template"`
	PayloadOn           string     `json:"payload_on,omitempty"`
	PayloadOff          string     `json:"payload_off,omitempty"`
	AvailabilityTopic   string     `json:"availability_topic"`
	PayloadAvailable    string     `json:"payload_available"`
	PayloadNotAvailable string     `json:"payload_not_available"`
	DeviceClass         string     `json:"device_class,omitempty"`
	EntityCategory      string     `json:"entity_category,omitempty"`
	Icon                string     `json:"icon,omitempty"`
	Device              deviceInfo `json:"device"`
}

type entity struct {
	Component      string
	ObjectID       string
	Name           string
	ValueTemplate  string
	DeviceClass    string
	EntityCategory string
	Icon           string
	Binary         bool
}

type zonePayload struct {
	Account          string          `json:"account"`
	Partition        string          `json:"partition"`
	Group            string          `json:"group"`
	Zone             string          `json:"zone"`
	Device           string          `json:"device"`
	DeviceName       string          `json:"device_name"`
	Room             string          `json:"room"`
	Kind             string          `json:"kind"`
	DeviceEvents     []string        `json:"device_events"`
	AlarmActive      bool            `json:"alarm_active"`
	AlarmSignal      string          `json:"alarm_signal"`
	AlarmAction      string          `json:"alarm_action"`
	AlarmEventCode   string          `json:"alarm_event_code"`
	AlarmEventName   string          `json:"alarm_event_name"`
	TamperActive     bool            `json:"tamper_active"`
	TroubleActive    bool            `json:"trouble_active"`
	SignalActive     map[string]bool `json:"signal_active"`
	LastEventCode    string          `json:"last_event_code"`
	LastEventName    string          `json:"last_event_name"`
	LastSignal       string          `json:"last_signal"`
	LastEventAt      string          `json:"last_event_at"`
	LastEventUnix    int64           `json:"last_event_unix"`
	AlarmStartedAt   string          `json:"alarm_started_at"`
	AlarmStartedUnix int64           `json:"alarm_started_unix"`
}

type accountPayload struct {
	Account        string `json:"account"`
	Online         bool   `json:"online"`
	Mode           string `json:"mode"`
	Armed          bool   `json:"armed"`
	NightMode      bool   `json:"night_mode"`
	PartiallyArmed bool   `json:"partially_armed"`
	AlarmActive    bool   `json:"alarm_active"`
	TamperActive   bool   `json:"tamper_active"`
	TroubleActive  bool   `json:"trouble_active"`
	LastEventCode  string `json:"last_event_code"`
	LastEventName  string `json:"last_event_name"`
	LastSignal     string `json:"last_signal"`
	LastEventAt    string `json:"last_event_at"`
	LastEventUnix  int64  `json:"last_event_unix"`
	LastPingAt     string `json:"last_ping_at"`
	LastPingUnix   int64  `json:"last_ping_unix"`
}

func New(cfg Config, log zerolog.Logger) *Publisher {
	cfg.Broker = strings.TrimSpace(cfg.Broker)
	cfg.ClientID = fallback(strings.TrimSpace(cfg.ClientID), "ajax2prometheus")
	cfg.TopicPrefix = trimTopic(fallback(strings.TrimSpace(cfg.TopicPrefix), "ajax2prometheus"))
	cfg.DiscoveryPrefix = trimTopic(fallback(strings.TrimSpace(cfg.DiscoveryPrefix), "homeassistant"))
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &Publisher{
		cfg:        cfg,
		log:        log,
		discovered: make(map[string]struct{}),
	}
}

func (p *Publisher) Enabled() bool {
	return p != nil && p.cfg.Broker != ""
}

func (p *Publisher) Connect(ctx context.Context) error {
	if !p.Enabled() {
		return nil
	}

	opts := paho.NewClientOptions()
	opts.AddBroker(p.cfg.Broker)
	opts.SetClientID(p.cfg.ClientID)
	opts.SetUsername(p.cfg.Username)
	opts.SetPassword(p.cfg.Password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetCleanSession(true)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetWill(p.availabilityTopic(), payloadOffline, 1, true)
	opts.OnConnect = func(client paho.Client) {
		token := client.Publish(p.availabilityTopic(), 1, true, payloadOnline)
		token.WaitTimeout(p.cfg.Timeout)
		p.log.Info().Str("broker", p.cfg.Broker).Msg("MQTT connected")
	}
	opts.OnConnectionLost = func(_ paho.Client, err error) {
		p.log.Warn().Err(err).Str("broker", p.cfg.Broker).Msg("MQTT connection lost")
	}

	p.client = paho.NewClient(opts)
	return p.wait(ctx, p.client.Connect())
}

func (p *Publisher) Close() {
	if !p.Enabled() || p.client == nil {
		return
	}
	if p.client.IsConnected() {
		_ = p.publish(context.Background(), p.availabilityTopic(), payloadOffline, true)
	}
	p.client.Disconnect(250)
}

func (p *Publisher) PublishSnapshot(ctx context.Context, snapshot state.Snapshot) error {
	if !p.Enabled() || p.client == nil {
		return nil
	}
	if token := ctx.Err(); token != nil {
		return token
	}
	if !p.client.IsConnectionOpen() {
		return errors.New("MQTT client is not connected")
	}

	var errs []error
	accounts := append([]state.Account(nil), snapshot.Accounts...)
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Account < accounts[j].Account
	})
	for _, account := range accounts {
		if err := p.publishAccount(ctx, account); err != nil {
			errs = append(errs, err)
		}
	}

	zones := append([]state.Zone(nil), snapshot.Zones...)
	sort.Slice(zones, func(i, j int) bool {
		if zones[i].Account == zones[j].Account {
			return zones[i].Zone < zones[j].Zone
		}
		return zones[i].Account < zones[j].Account
	})
	for _, zone := range zones {
		if err := p.publishZone(ctx, zone); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (p *Publisher) publishAccount(ctx context.Context, account state.Account) error {
	topic := p.accountStateTopic(account.Account)
	if p.cfg.Discovery {
		for _, ent := range accountEntities(account.Account) {
			if err := p.publishDiscovery(ctx, ent, topic, accountDevice(account), p.cfg.ClientID); err != nil {
				return err
			}
		}
	}
	payload, err := json.Marshal(accountState(account))
	if err != nil {
		return err
	}
	return p.publish(ctx, topic, payload, p.cfg.Retain)
}

func (p *Publisher) publishZone(ctx context.Context, zone state.Zone) error {
	topic := p.zoneStateTopic(zone.Account, zone.Zone)
	if p.cfg.Discovery {
		device := zoneDevice(zone)
		for _, ent := range zoneEntities(zone) {
			if err := p.publishDiscovery(ctx, ent, topic, device, p.cfg.ClientID); err != nil {
				return err
			}
		}
	}
	payload, err := json.Marshal(zoneState(zone))
	if err != nil {
		return err
	}
	return p.publish(ctx, topic, payload, p.cfg.Retain)
}

func (p *Publisher) publishDiscovery(ctx context.Context, ent entity, stateTopic string, device deviceInfo, node string) error {
	key := ent.Component + "/" + ent.ObjectID
	p.mu.Lock()
	if _, ok := p.discovered[key]; ok {
		p.mu.Unlock()
		return nil
	}
	p.mu.Unlock()

	cfg := discoveryConfig{
		Name:                ent.Name,
		UniqueID:            p.uniqueID(ent.ObjectID),
		StateTopic:          stateTopic,
		ValueTemplate:       ent.ValueTemplate,
		AvailabilityTopic:   p.availabilityTopic(),
		PayloadAvailable:    payloadOnline,
		PayloadNotAvailable: payloadOffline,
		DeviceClass:         ent.DeviceClass,
		EntityCategory:      ent.EntityCategory,
		Icon:                ent.Icon,
		Device:              device,
	}
	if ent.Binary {
		cfg.PayloadOn = payloadOn
		cfg.PayloadOff = payloadOff
	}

	payload, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	topic := p.discoveryTopic(ent.Component, node, ent.ObjectID)
	if err := p.publish(ctx, topic, payload, true); err != nil {
		return err
	}

	p.mu.Lock()
	p.discovered[key] = struct{}{}
	p.mu.Unlock()
	return nil
}

func (p *Publisher) publish(ctx context.Context, topic string, payload interface{}, retain bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.client == nil || !p.client.IsConnectionOpen() {
		return fmt.Errorf("MQTT publish %s: client is not connected", topic)
	}
	return p.wait(ctx, p.client.Publish(topic, 1, retain, payload))
}

func (p *Publisher) wait(ctx context.Context, token paho.Token) error {
	done := make(chan struct{})
	go func() {
		token.Wait()
		close(done)
	}()
	timer := time.NewTimer(p.cfg.Timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return fmt.Errorf("MQTT operation timed out after %s", p.cfg.Timeout)
	case <-done:
		return token.Error()
	}
}

func (p *Publisher) availabilityTopic() string {
	return p.cfg.TopicPrefix + "/status"
}

func (p *Publisher) accountStateTopic(account string) string {
	return p.cfg.TopicPrefix + "/accounts/" + topicPart(account) + "/state"
}

func (p *Publisher) zoneStateTopic(account, zone string) string {
	return p.cfg.TopicPrefix + "/accounts/" + topicPart(account) + "/zones/" + topicPart(zone) + "/state"
}

func (p *Publisher) discoveryTopic(component, node, objectID string) string {
	return strings.Join([]string{
		p.cfg.DiscoveryPrefix,
		component,
		slug(node),
		slug(objectID),
		"config",
	}, "/")
}

func (p *Publisher) uniqueID(objectID string) string {
	return slug(p.cfg.ClientID + "_" + objectID)
}

func accountDevice(account state.Account) deviceInfo {
	name := "Ajax account " + account.Account
	return deviceInfo{
		Identifiers:  []string{"ajax2prometheus_account_" + account.Account},
		Name:         name,
		Manufacturer: "Ajax Systems",
		Model:        "Ajax account",
	}
}

func zoneDevice(zone state.Zone) deviceInfo {
	name := fallback(zone.DeviceName, "Ajax zone "+zone.Zone)
	model := fallback(zone.Kind, "Ajax device")
	device := deviceInfo{
		Identifiers:  []string{"ajax2prometheus_" + zone.Account + "_zone_" + zone.Zone},
		Name:         name,
		Manufacturer: "Ajax Systems",
		Model:        model,
	}
	if zone.Room != "" && zone.Room != "unknown" {
		device.SuggestedArea = zone.Room
	}
	return device
}

func accountEntities(account string) []entity {
	base := "account_" + account + "_"
	return []entity{
		binaryEntity(base+"online", "Online", "connectivity", "", "{{ '"+payloadOn+"' if value_json.online else '"+payloadOff+"' }}"),
		binaryEntity(base+"armed", "Armed", "", "mdi:shield-lock", "{{ '"+payloadOn+"' if value_json.armed else '"+payloadOff+"' }}"),
		binaryEntity(base+"night_mode", "Night mode", "", "mdi:weather-night", "{{ '"+payloadOn+"' if value_json.night_mode else '"+payloadOff+"' }}"),
		binaryEntity(base+"partially_armed", "Partially armed", "", "mdi:shield-half-full", "{{ '"+payloadOn+"' if value_json.partially_armed else '"+payloadOff+"' }}"),
		binaryEntity(base+"alarm_active", "Alarm active", "safety", "", "{{ '"+payloadOn+"' if value_json.alarm_active else '"+payloadOff+"' }}"),
		binaryEntity(base+"tamper_active", "Tamper active", "tamper", "", "{{ '"+payloadOn+"' if value_json.tamper_active else '"+payloadOff+"' }}"),
		binaryEntity(base+"trouble_active", "Trouble active", "problem", "", "{{ '"+payloadOn+"' if value_json.trouble_active else '"+payloadOff+"' }}"),
		sensorEntity(base+"mode", "Mode", "", "mdi:shield-home", "{{ value_json.mode }}"),
		sensorEntity(base+"last_event_name", "Last event", "", "mdi:message-alert", "{{ value_json.last_event_name }}"),
		sensorEntity(base+"last_event_code", "Last event code", "", "mdi:identifier", "{{ value_json.last_event_code }}"),
		sensorEntity(base+"last_signal", "Last signal", "", "mdi:signal", "{{ value_json.last_signal }}"),
		sensorEntity(base+"last_event_at", "Last event time", "timestamp", "", "{{ value_json.last_event_at }}"),
		sensorEntity(base+"last_ping_at", "Last ping time", "timestamp", "", "{{ value_json.last_ping_at }}"),
	}
}

func zoneEntities(zone state.Zone) []entity {
	base := "zone_" + zone.Account + "_" + zone.Zone + "_"
	entities := []entity{
		binaryEntity(base+"alarm_active", "Alarm active", "safety", "", "{{ '"+payloadOn+"' if value_json.alarm_active else '"+payloadOff+"' }}"),
		binaryEntity(base+"tamper_active", "Tamper active", "tamper", "", "{{ '"+payloadOn+"' if value_json.tamper_active else '"+payloadOff+"' }}"),
		binaryEntity(base+"trouble_active", "Trouble active", "problem", "", "{{ '"+payloadOn+"' if value_json.trouble_active else '"+payloadOff+"' }}"),
		sensorEntity(base+"last_event_name", "Last event", "", "mdi:message-alert", "{{ value_json.last_event_name }}"),
		sensorEntity(base+"last_event_code", "Last event code", "", "mdi:identifier", "{{ value_json.last_event_code }}"),
		sensorEntity(base+"last_signal", "Last signal", "", "mdi:signal", "{{ value_json.last_signal }}"),
		sensorEntity(base+"last_event_at", "Last event time", "timestamp", "", "{{ value_json.last_event_at }}"),
		sensorEntity(base+"alarm_signal", "Alarm signal", "", "mdi:alarm-light", "{{ value_json.alarm_signal }}"),
		sensorEntity(base+"alarm_action", "Alarm action", "", "mdi:alarm-light-outline", "{{ value_json.alarm_action }}"),
	}
	for _, signal := range sortedSignals(zone.DeviceEvents, zone.SignalActive) {
		entities = append(entities, signalEntity(base, signal))
	}
	return entities
}

func signalEntity(base, signal string) entity {
	deviceClass, icon := signalPresentation(signal)
	name := signalName(signal)
	return binaryEntity(
		base+"signal_"+signal,
		name,
		deviceClass,
		icon,
		"{{ '"+payloadOn+"' if value_json.signal_active.get('"+signal+"', false) else '"+payloadOff+"' }}",
	)
}

func binaryEntity(objectID, name, deviceClass, icon, template string) entity {
	return entity{
		Component:     "binary_sensor",
		ObjectID:      objectID,
		Name:          name,
		ValueTemplate: template,
		DeviceClass:   deviceClass,
		Icon:          icon,
		Binary:        true,
	}
}

func sensorEntity(objectID, name, deviceClass, icon, template string) entity {
	return entity{
		Component:     "sensor",
		ObjectID:      objectID,
		Name:          name,
		ValueTemplate: template,
		DeviceClass:   deviceClass,
		Icon:          icon,
	}
}

func zoneState(zone state.Zone) zonePayload {
	signals := make(map[string]bool, len(zone.DeviceEvents)+len(zone.SignalActive))
	for _, signal := range zone.DeviceEvents {
		if signal != "" {
			signals[signal] = false
		}
	}
	for signal, active := range zone.SignalActive {
		if signal != "" {
			signals[signal] = active
		}
	}
	return zonePayload{
		Account:          zone.Account,
		Partition:        zone.Partition,
		Group:            zone.Group,
		Zone:             zone.Zone,
		Device:           fallback(zone.Device, zone.Zone),
		DeviceName:       zone.DeviceName,
		Room:             zone.Room,
		Kind:             zone.Kind,
		DeviceEvents:     append([]string(nil), zone.DeviceEvents...),
		AlarmActive:      zone.AlarmActive,
		AlarmSignal:      fallback(zone.AlarmSignal, "none"),
		AlarmAction:      fallback(zone.AlarmAction, "none"),
		AlarmEventCode:   zone.AlarmEventCode,
		AlarmEventName:   zone.AlarmEventName,
		TamperActive:     zone.TamperActive,
		TroubleActive:    zone.TroubleActive,
		SignalActive:     signals,
		LastEventCode:    zone.LastEventCode,
		LastEventName:    zone.LastEventName,
		LastSignal:       zone.LastSignal,
		LastEventAt:      mqttTime(zone.LastEventAt),
		LastEventUnix:    unixTime(zone.LastEventAt),
		AlarmStartedAt:   mqttTime(zone.AlarmStartedAt),
		AlarmStartedUnix: unixTime(zone.AlarmStartedAt),
	}
}

func accountState(account state.Account) accountPayload {
	return accountPayload{
		Account:        account.Account,
		Online:         account.Online,
		Mode:           account.Mode,
		Armed:          account.Armed,
		NightMode:      account.NightMode,
		PartiallyArmed: account.PartiallyArmed,
		AlarmActive:    account.AlarmActive,
		TamperActive:   account.TamperActive,
		TroubleActive:  account.TroubleActive,
		LastEventCode:  account.LastEventCode,
		LastEventName:  account.LastEventName,
		LastSignal:     account.LastSignal,
		LastEventAt:    mqttTime(account.LastEventAt),
		LastEventUnix:  unixTime(account.LastEventAt),
		LastPingAt:     mqttTime(account.LastPingAt),
		LastPingUnix:   unixTime(account.LastPingAt),
	}
}

func sortedSignals(events []string, active map[string]bool) []string {
	seen := make(map[string]struct{}, len(events)+len(active))
	for _, signal := range events {
		if signal != "" {
			seen[signal] = struct{}{}
		}
	}
	for signal := range active {
		if signal != "" {
			seen[signal] = struct{}{}
		}
	}
	signals := make([]string, 0, len(seen))
	for signal := range seen {
		signals = append(signals, signal)
	}
	sort.Strings(signals)
	return signals
}

func signalPresentation(signal string) (string, string) {
	switch signal {
	case "battery":
		return "battery", ""
	case "co", "gas", "gas_or_co":
		return "gas", ""
	case "connectivity", "hardware", "fire_detector", "interference", "accelerometer", "bypass", "tamper_bypass":
		return "problem", ""
	case "duress", "emergency", "fire", "medical", "panic", "temperature":
		return "safety", ""
	case "power":
		return "problem", "mdi:power-plug-off"
	case "smoke":
		return "smoke", ""
	case "tamper":
		return "tamper", ""
	case "water_leak":
		return "moisture", ""
	case "arming":
		return "", "mdi:shield-lock"
	case "burglary":
		return "safety", "mdi:shield-alert"
	case "configuration":
		return "", "mdi:cog"
	case "firmware":
		return "", "mdi:update"
	case "night_mode":
		return "", "mdi:weather-night"
	default:
		return "problem", ""
	}
}

func signalName(signal string) string {
	switch signal {
	case "battery":
		return "Battery low"
	case "burglary":
		return "Burglary"
	case "bypass":
		return "Bypassed"
	case "co":
		return "Carbon monoxide"
	case "connectivity":
		return "Connection lost"
	case "fire_detector":
		return "Fire detector fault"
	case "gas_or_co":
		return "Gas or CO"
	case "hardware":
		return "Hardware fault"
	case "night_mode":
		return "Night mode"
	case "power":
		return "Power failure"
	case "tamper_bypass":
		return "Tamper bypassed"
	case "water_leak":
		return "Water leak"
	default:
		return titleLabel(signal)
	}
}

func titleLabel(value string) string {
	words := strings.Fields(strings.ReplaceAll(value, "_", " "))
	for i, word := range words {
		if word == "" {
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

func mqttTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func unixTime(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.Unix()
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "unknown"
	}
	return out
}

func topicPart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	return strings.NewReplacer("#", "_", "+", "_", " ", "_").Replace(value)
}

func trimTopic(value string) string {
	return strings.Trim(strings.TrimSpace(value), "/")
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}
