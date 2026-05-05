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

	"github.com/RCooLeR/AjaxBridge/internal/state"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
)

const (
	payloadOnline  = "online"
	payloadOffline = "offline"
	payloadOn      = "ON"
	payloadOff     = "OFF"
)

var legacyDiscoveryNodes = []string{"ajax2prometheus"}

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
	cfg             Config
	client          paho.Client
	log             zerolog.Logger
	mu              sync.Mutex
	discovered      map[string]struct{}
	publishedStates map[string]string
	accountPlans    map[string]accountPlan
	zonePlans       map[string]zonePlan
	subscriptions   map[string]MessageHandler
}

type Update struct {
	Accounts []state.Account
	Zones    []state.Zone
}

type MessageHandler = func(topic string, payload []byte)

type CleanupConfig struct {
	DiscoveryPrefix        string
	DiscoveryNode          string
	TopicPrefix            string
	JeedomStateTopicPrefix string
	Wait                   time.Duration
}

type CleanupResult struct {
	Received int
	Matched  int
	Cleared  int
	Topics   []string
}

type accountPlan struct {
	stateTopic string
	cleanup    []discoveryMessage
	discovery  []discoveryMessage
}

type zonePlan struct {
	signature  string
	stateTopic string
	cleanup    []discoveryMessage
	discovery  []discoveryMessage
}

type discoveryMessage struct {
	key     string
	topic   string
	payload []byte
}

type cleanupSubscription struct {
	Topic string
	Match func(topic string) bool
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
	JSONAttributesTopic string     `json:"json_attributes_topic,omitempty"`
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
	cfg.ClientID = fallback(strings.TrimSpace(cfg.ClientID), "ajaxbridge")
	cfg.TopicPrefix = trimTopic(fallback(strings.TrimSpace(cfg.TopicPrefix), "ajaxbridge"))
	cfg.DiscoveryPrefix = trimTopic(fallback(strings.TrimSpace(cfg.DiscoveryPrefix), "homeassistant"))
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &Publisher{
		cfg:             cfg,
		log:             log,
		discovered:      make(map[string]struct{}),
		publishedStates: make(map[string]string),
		accountPlans:    make(map[string]accountPlan),
		zonePlans:       make(map[string]zonePlan),
		subscriptions:   make(map[string]MessageHandler),
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
		p.resetCaches()
		token := client.Publish(p.availabilityTopic(), 1, true, payloadOnline)
		token.WaitTimeout(p.cfg.Timeout)
		p.resubscribe(client)
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
	return p.PublishUpdate(ctx, Update{
		Accounts: snapshot.Accounts,
		Zones:    snapshot.Zones,
	})
}

func (p *Publisher) PublishUpdate(ctx context.Context, update Update) error {
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
	for _, account := range update.Accounts {
		if err := p.publishAccount(ctx, account); err != nil {
			errs = append(errs, err)
		}
	}

	for _, zone := range update.Zones {
		if err := p.publishZone(ctx, zone); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (p *Publisher) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	if !p.Enabled() || p.client == nil {
		return nil
	}
	topic = strings.TrimSpace(topic)
	if topic == "" || handler == nil {
		return nil
	}

	p.mu.Lock()
	p.subscriptions[topic] = handler
	p.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}
	if !p.client.IsConnectionOpen() {
		return nil
	}
	return p.wait(ctx, p.client.Subscribe(topic, 1, wrapMessageHandler(handler)))
}

func (p *Publisher) CleanupRetained(ctx context.Context, cfg CleanupConfig) (CleanupResult, error) {
	var result CleanupResult
	if !p.Enabled() || p.client == nil {
		return result, nil
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if !p.client.IsConnectionOpen() {
		return result, errors.New("MQTT client is not connected")
	}

	if cfg.Wait <= 0 {
		cfg.Wait = 5 * time.Second
	}
	patterns := p.cleanupSubscriptions(cfg)
	matches := make(map[string]struct{})
	var mu sync.Mutex
	handler := func(_ paho.Client, message paho.Message) {
		topic := message.Topic()
		mu.Lock()
		result.Received++
		if cleanupTopicMatches(patterns, topic) {
			if _, ok := matches[topic]; !ok {
				matches[topic] = struct{}{}
				result.Matched++
			}
		}
		mu.Unlock()
	}

	for _, pattern := range patterns {
		if err := p.wait(ctx, p.client.Subscribe(pattern.Topic, 1, handler)); err != nil {
			return result, err
		}
	}
	timer := time.NewTimer(cfg.Wait)
	select {
	case <-ctx.Done():
		timer.Stop()
		return result, ctx.Err()
	case <-timer.C:
	}
	topics := make([]string, 0, len(matches))
	mu.Lock()
	for topic := range matches {
		topics = append(topics, topic)
	}
	mu.Unlock()
	sort.Strings(topics)

	for _, pattern := range patterns {
		_ = p.client.Unsubscribe(pattern.Topic).WaitTimeout(p.cfg.Timeout)
	}
	for _, topic := range topics {
		if err := p.publish(ctx, topic, []byte{}, true); err != nil {
			return result, err
		}
		result.Cleared++
	}
	result.Topics = topics
	return result, nil
}

func (p *Publisher) cleanupSubscriptions(cfg CleanupConfig) []cleanupSubscription {
	discoveryPrefix := trimTopic(cfg.DiscoveryPrefix)
	if discoveryPrefix == "" {
		discoveryPrefix = p.cfg.DiscoveryPrefix
	}
	if discoveryPrefix == "" {
		discoveryPrefix = "homeassistant"
	}
	discoveryNode := strings.TrimSpace(cfg.DiscoveryNode)
	if discoveryNode == "" {
		discoveryNode = p.discoveryNode()
	}
	discoveryNode = slug(discoveryNode)
	if discoveryNode == "" || discoveryNode == "unknown" {
		discoveryNode = "ajaxbridge"
	}
	topicPrefix := trimTopic(cfg.TopicPrefix)
	if topicPrefix == "" {
		topicPrefix = p.cfg.TopicPrefix
	}
	if topicPrefix == "" {
		topicPrefix = "ajaxbridge"
	}
	jeedomStatePrefix := trimTopic(cfg.JeedomStateTopicPrefix)
	if jeedomStatePrefix == "" {
		jeedomStatePrefix = topicPrefix + "/jeedom"
	}

	return []cleanupSubscription{
		{
			Topic: strings.Join([]string{discoveryPrefix, "+", discoveryNode, "+", "config"}, "/"),
			Match: func(topic string) bool {
				return matchAjaxBridgeDiscoveryTopic(discoveryPrefix, discoveryNode, topic)
			},
		},
		{
			Topic: topicPrefix + "/accounts/+/state",
			Match: func(topic string) bool {
				return matchSIAStateTopic(topicPrefix, topic)
			},
		},
		{
			Topic: topicPrefix + "/accounts/+/zones/+/state",
			Match: func(topic string) bool {
				return matchSIAStateTopic(topicPrefix, topic)
			},
		},
		{
			Topic: discoveryPrefix + "/+/ajax2prometheus/#",
			Match: func(topic string) bool {
				return matchTopicPrefix(topic, discoveryPrefix+"/") && strings.Contains(trimTopic(topic), "/ajax2prometheus/")
			},
		},
		{
			Topic: jeedomStatePrefix + "/devices/+/state",
			Match: func(topic string) bool {
				return matchJeedomStateTopic(jeedomStatePrefix, topic)
			},
		},
		{
			Topic: "ajax2prometheus/#",
			Match: func(topic string) bool {
				return matchTopicPrefix(topic, "ajax2prometheus")
			},
		},
	}
}

func cleanupTopicMatches(patterns []cleanupSubscription, topic string) bool {
	for _, pattern := range patterns {
		if pattern.Match != nil && pattern.Match(topic) {
			return true
		}
	}
	return false
}

func matchAjaxBridgeDiscoveryTopic(discoveryPrefix, discoveryNode, topic string) bool {
	parts := strings.Split(trimTopic(topic), "/")
	if len(parts) != 5 {
		return false
	}
	if parts[0] != discoveryPrefix || parts[2] != discoveryNode || parts[4] != "config" {
		return false
	}
	return strings.HasPrefix(parts[3], "account_") ||
		strings.HasPrefix(parts[3], "zone_") ||
		strings.HasPrefix(parts[3], "jeedom_cmd_") ||
		strings.HasPrefix(parts[3], "jeedom_control_")
}

func matchSIAStateTopic(prefix, topic string) bool {
	prefix = trimTopic(prefix)
	topic = trimTopic(topic)
	if prefix == "" || topic == "" {
		return false
	}
	prefixParts := strings.Split(prefix, "/")
	topicParts := strings.Split(topic, "/")
	if len(topicParts) < len(prefixParts) {
		return false
	}
	for i, part := range prefixParts {
		if topicParts[i] != part {
			return false
		}
	}
	rest := topicParts[len(prefixParts):]
	if len(rest) == 3 {
		return rest[0] == "accounts" && rest[1] != "" && rest[2] == "state"
	}
	if len(rest) == 5 {
		return rest[0] == "accounts" && rest[1] != "" && rest[2] == "zones" && rest[3] != "" && rest[4] == "state"
	}
	return false
}

func matchJeedomStateTopic(prefix, topic string) bool {
	prefix = trimTopic(prefix)
	topic = trimTopic(topic)
	if prefix == "" || !strings.HasPrefix(topic, prefix+"/devices/") {
		return false
	}
	return strings.HasSuffix(topic, "/state")
}

func matchTopicPrefix(topic, prefix string) bool {
	topic = trimTopic(topic)
	prefix = trimTopic(prefix)
	return topic == prefix || strings.HasPrefix(topic, prefix+"/")
}

func (p *Publisher) PublishStateMessage(ctx context.Context, topic string, payload []byte, retain bool) error {
	if !p.Enabled() || p.client == nil {
		return nil
	}
	return p.publishState(ctx, topic, payload, retain)
}

func (p *Publisher) PublishDiscoveryMessage(ctx context.Context, key, topic string, payload []byte, retain bool) error {
	if !p.Enabled() || p.client == nil {
		return nil
	}
	if retain {
		return p.publishDiscoveryMessage(ctx, discoveryMessage{key: key, topic: topic, payload: payload})
	}
	return p.publish(ctx, topic, payload, false)
}

func (p *Publisher) PublishCommandMessage(ctx context.Context, topic string, payload []byte) error {
	if !p.Enabled() || p.client == nil {
		return nil
	}
	return p.publish(ctx, topic, payload, false)
}

func (p *Publisher) AvailabilityTopic() string {
	if p == nil {
		return ""
	}
	return p.availabilityTopic()
}

func (p *Publisher) publishAccount(ctx context.Context, account state.Account) error {
	plan, err := p.accountPlanFor(account)
	if err != nil {
		return err
	}
	if p.cfg.Discovery {
		for _, message := range plan.cleanup {
			if err := p.publishDiscoveryMessage(ctx, message); err != nil {
				return err
			}
		}
		for _, message := range plan.discovery {
			if err := p.publishDiscoveryMessage(ctx, message); err != nil {
				return err
			}
		}
	}
	payload, err := json.Marshal(accountState(account))
	if err != nil {
		return err
	}
	return p.publishState(ctx, plan.stateTopic, payload, p.cfg.Retain)
}

func (p *Publisher) publishZone(ctx context.Context, zone state.Zone) error {
	plan, err := p.zonePlanFor(zone)
	if err != nil {
		return err
	}
	if p.cfg.Discovery {
		for _, message := range plan.cleanup {
			if err := p.publishDiscoveryMessage(ctx, message); err != nil {
				return err
			}
		}
		for _, message := range plan.discovery {
			if err := p.publishDiscoveryMessage(ctx, message); err != nil {
				return err
			}
		}
	}
	payload, err := json.Marshal(zoneState(zone))
	if err != nil {
		return err
	}
	return p.publishState(ctx, plan.stateTopic, payload, p.cfg.Retain)
}

func (p *Publisher) publishDiscoveryMessage(ctx context.Context, message discoveryMessage) error {
	if !p.markDiscoveredPending(message.key) {
		return nil
	}
	if err := p.publish(ctx, message.topic, message.payload, true); err != nil {
		p.clearDiscovered(message.key)
		return err
	}
	return nil
}

func (p *Publisher) publishState(ctx context.Context, topic string, payload []byte, retain bool) error {
	if !p.shouldPublishState(topic, payload) {
		return nil
	}
	if err := p.publish(ctx, topic, payload, retain); err != nil {
		p.clearPublishedState(topic)
		return err
	}
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
	timer := time.NewTimer(p.cfg.Timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return fmt.Errorf("MQTT operation timed out after %s", p.cfg.Timeout)
	case <-token.Done():
		return token.Error()
	}
}

func (p *Publisher) resetCaches() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.discovered = make(map[string]struct{})
	p.publishedStates = make(map[string]string)
}

func (p *Publisher) resubscribe(client paho.Client) {
	p.mu.Lock()
	subscriptions := make(map[string]MessageHandler, len(p.subscriptions))
	for topic, handler := range p.subscriptions {
		subscriptions[topic] = handler
	}
	p.mu.Unlock()

	for topic, handler := range subscriptions {
		token := client.Subscribe(topic, 1, wrapMessageHandler(handler))
		if ok := token.WaitTimeout(p.cfg.Timeout); !ok {
			p.log.Warn().Str("topic", topic).Msg("MQTT subscription timed out")
			continue
		}
		if err := token.Error(); err != nil {
			p.log.Warn().Err(err).Str("topic", topic).Msg("MQTT subscription failed")
		}
	}
}

func (p *Publisher) markDiscoveredPending(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.discovered[key]; ok {
		return false
	}
	p.discovered[key] = struct{}{}
	return true
}

func (p *Publisher) clearDiscovered(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.discovered, key)
}

func (p *Publisher) shouldPublishState(topic string, payload []byte) bool {
	value := string(payload)
	p.mu.Lock()
	defer p.mu.Unlock()
	if cached, ok := p.publishedStates[topic]; ok && cached == value {
		return false
	}
	p.publishedStates[topic] = value
	return true
}

func (p *Publisher) clearPublishedState(topic string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.publishedStates, topic)
}

func (p *Publisher) accountPlanFor(account state.Account) (accountPlan, error) {
	key := account.Account
	p.mu.Lock()
	plan, ok := p.accountPlans[key]
	p.mu.Unlock()
	if ok {
		return plan, nil
	}

	discovery, err := p.buildDiscoveryMessages(
		accountEntities(account.Account),
		p.accountStateTopic(account.Account),
		accountDevice(account),
		p.discoveryNode(),
	)
	if err != nil {
		return accountPlan{}, err
	}
	plan = accountPlan{
		stateTopic: p.accountStateTopic(account.Account),
		cleanup:    p.legacyCleanupMessages(accountEntities(account.Account)),
		discovery:  discovery,
	}

	p.mu.Lock()
	if cached, ok := p.accountPlans[key]; ok {
		p.mu.Unlock()
		return cached, nil
	}
	p.accountPlans[key] = plan
	p.mu.Unlock()
	return plan, nil
}

func (p *Publisher) zonePlanFor(zone state.Zone) (zonePlan, error) {
	key := zonePlanKey(zone.Account, zone.Zone)
	signature := zoneDiscoverySignature(zone)

	p.mu.Lock()
	plan, ok := p.zonePlans[key]
	p.mu.Unlock()
	if ok && plan.signature == signature {
		return plan, nil
	}

	entities := zoneEntities(zone)
	discovery, err := p.buildDiscoveryMessages(
		entities,
		p.zoneStateTopic(zone.Account, zone.Zone),
		zoneDevice(zone),
		p.discoveryNode(),
	)
	if err != nil {
		return zonePlan{}, err
	}
	plan = zonePlan{
		signature:  signature,
		stateTopic: p.zoneStateTopic(zone.Account, zone.Zone),
		cleanup:    append(p.legacyCleanupMessages(entities), p.renamedSignalCleanupMessages(zone)...),
		discovery:  discovery,
	}

	p.mu.Lock()
	cached, ok := p.zonePlans[key]
	if ok && cached.signature == signature {
		p.mu.Unlock()
		return cached, nil
	}
	p.zonePlans[key] = plan
	p.mu.Unlock()
	return plan, nil
}

func (p *Publisher) buildDiscoveryMessages(entities []entity, stateTopic string, device deviceInfo, node string) ([]discoveryMessage, error) {
	messages := make([]discoveryMessage, 0, len(entities))
	for _, ent := range entities {
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
			JSONAttributesTopic: stateTopic,
			Device:              device,
		}
		if ent.Binary {
			cfg.PayloadOn = payloadOn
			cfg.PayloadOff = payloadOff
		}

		payload, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		messages = append(messages, discoveryMessage{
			key:     "discover:" + slug(node) + ":" + ent.Component + "/" + ent.ObjectID,
			topic:   p.discoveryTopic(ent.Component, node, ent.ObjectID),
			payload: payload,
		})
	}
	return messages, nil
}

func (p *Publisher) legacyCleanupMessages(entities []entity) []discoveryMessage {
	currentNode := slug(p.discoveryNode())
	messages := make([]discoveryMessage, 0, len(entities)*len(legacyDiscoveryNodes))
	for _, legacyNode := range legacyDiscoveryNodes {
		if slug(legacyNode) == currentNode {
			continue
		}
		for _, ent := range entities {
			messages = append(messages, discoveryMessage{
				key:     "cleanup:" + slug(legacyNode) + ":" + ent.Component + "/" + ent.ObjectID,
				topic:   p.discoveryTopic(ent.Component, legacyNode, ent.ObjectID),
				payload: []byte{},
			})
		}
	}
	return messages
}

func (p *Publisher) renamedSignalCleanupMessages(zone state.Zone) []discoveryMessage {
	base := "zone_" + zone.Account + "_" + zone.Zone + "_"
	signals := sortedSignals(zone.DeviceEvents, zone.SignalActive)
	messages := make([]discoveryMessage, 0, len(signals))
	for _, signal := range signals {
		for _, suffix := range legacySignalObjectSuffixes(signal) {
			objectID := base + suffix
			messages = append(messages, discoveryMessage{
				key:     "cleanup:renamed_signal:" + p.discoveryNode() + ":binary_sensor/" + objectID,
				topic:   p.discoveryTopic("binary_sensor", p.discoveryNode(), objectID),
				payload: []byte{},
			})
		}
	}
	return messages
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
	return slug(p.discoveryNode() + "_" + objectID)
}

func (p *Publisher) discoveryNode() string {
	if strings.TrimSpace(p.cfg.TopicPrefix) != "" {
		return p.cfg.TopicPrefix
	}
	return p.cfg.ClientID
}

func accountDevice(account state.Account) deviceInfo {
	name := "Ajax account " + account.Account
	return deviceInfo{
		Identifiers:  []string{"ajaxbridge_account_" + account.Account},
		Name:         name,
		Manufacturer: "Ajax Systems",
		Model:        "Ajax account",
	}
}

func zoneDevice(zone state.Zone) deviceInfo {
	name := fallback(zone.DeviceName, "Ajax zone "+zone.Zone)
	model := fallback(zone.Kind, "Ajax device")
	device := deviceInfo{
		Identifiers:  []string{"ajaxbridge_" + zone.Account + "_zone_" + zone.Zone},
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
		base+signalObjectSuffix(signal),
		name,
		deviceClass,
		icon,
		"{{ '"+payloadOn+"' if value_json.signal_active.get('"+signal+"', false) else '"+payloadOff+"' }}",
	)
}

func signalObjectSuffix(signal string) string {
	switch signal {
	case "power":
		return "signal_power_failure"
	case "temperature":
		return "signal_temperature_alarm"
	default:
		return "signal_" + signal
	}
}

func legacySignalObjectSuffixes(signal string) []string {
	switch signal {
	case "power":
		return []string{"signal_power"}
	case "temperature":
		return []string{"signal_temperature"}
	default:
		return nil
	}
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
	case "temperature":
		return "Temperature alarm"
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

func zonePlanKey(account, zone string) string {
	return account + "/" + zone
}

func zoneDiscoverySignature(zone state.Zone) string {
	signals := sortedSignals(zone.DeviceEvents, zone.SignalActive)
	var b strings.Builder
	b.Grow(len(zone.Account) + len(zone.Zone) + len(zone.DeviceName) + len(zone.Room) + len(zone.Kind) + len(signals)*16 + 8)
	b.WriteString(zone.Account)
	b.WriteByte('|')
	b.WriteString(zone.Zone)
	b.WriteByte('|')
	b.WriteString(zone.DeviceName)
	b.WriteByte('|')
	b.WriteString(zone.Room)
	b.WriteByte('|')
	b.WriteString(zone.Kind)
	for _, signal := range signals {
		b.WriteByte('|')
		b.WriteString(signal)
	}
	return b.String()
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

func wrapMessageHandler(handler MessageHandler) paho.MessageHandler {
	return func(_ paho.Client, message paho.Message) {
		handler(message.Topic(), append([]byte(nil), message.Payload()...))
	}
}
