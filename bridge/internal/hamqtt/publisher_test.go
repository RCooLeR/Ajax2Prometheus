package hamqtt

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/state"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
)

func TestZoneEntitiesIncludeBaseSensorsAndDeviceEventSignals(t *testing.T) {
	entities := zoneEntities(state.Zone{
		Account:           "A0F80D",
		Zone:              "3",
		DeviceEvents:      []string{"fire", "tamper", "battery"},
		SignalActive:      map[string]bool{"connectivity": true},
		DeviceEventsLabel: "fire,tamper,battery",
	})

	want := map[string]bool{
		"zone_A0F80D_3_alarm_active":        true,
		"zone_A0F80D_3_tamper_active":       true,
		"zone_A0F80D_3_trouble_active":      true,
		"zone_A0F80D_3_last_event_name":     true,
		"zone_A0F80D_3_last_event_code":     true,
		"zone_A0F80D_3_last_signal":         true,
		"zone_A0F80D_3_last_event_at":       true,
		"zone_A0F80D_3_alarm_signal":        true,
		"zone_A0F80D_3_alarm_action":        true,
		"zone_A0F80D_3_signal_battery":      true,
		"zone_A0F80D_3_signal_connectivity": true,
		"zone_A0F80D_3_signal_fire":         true,
		"zone_A0F80D_3_signal_tamper":       true,
	}
	got := make(map[string]bool, len(entities))
	for _, entity := range entities {
		got[entity.ObjectID] = true
	}
	for objectID := range want {
		if !got[objectID] {
			t.Fatalf("missing entity %q in %#v", objectID, got)
		}
	}
}

func TestZoneStateFillsInactiveSignalsFromDeviceEvents(t *testing.T) {
	payload := zoneState(state.Zone{
		Account:      "A0F80D",
		Zone:         "3",
		DeviceEvents: []string{"fire", "tamper"},
		SignalActive: map[string]bool{"fire": true},
	})

	if !payload.SignalActive["fire"] {
		t.Fatalf("expected fire active: %#v", payload.SignalActive)
	}
	if payload.SignalActive["tamper"] {
		t.Fatalf("expected tamper inactive: %#v", payload.SignalActive)
	}
	if payload.AlarmSignal != "none" || payload.AlarmAction != "none" {
		t.Fatalf("expected inactive alarm labels to default to none: %#v", payload)
	}
}

func TestPublishSnapshotSkipsUnchangedStatePayloads(t *testing.T) {
	client := &stubClient{open: true}
	publisher := New(Config{
		Broker:          "tcp://mqtt.local:1883",
		ClientID:        "ajaxbridge",
		TopicPrefix:     "ajaxbridge",
		Discovery:       true,
		DiscoveryPrefix: "homeassistant",
		Timeout:         time.Second,
		Retain:          true,
	}, zerologNop())
	publisher.client = client

	snapshot := state.Snapshot{
		Accounts: []state.Account{{
			Account: "A0F80D",
			Online:  true,
			Mode:    "armed",
			Armed:   true,
		}},
		Zones: []state.Zone{{
			Account:      "A0F80D",
			Zone:         "3",
			DeviceName:   "Hall fire detector",
			Kind:         "fireprotect",
			Room:         "Hall",
			DeviceEvents: []string{"fire", "tamper", "battery"},
			SignalActive: map[string]bool{"fire": true},
		}},
	}

	if err := publisher.PublishSnapshot(t.Context(), snapshot); err != nil {
		t.Fatal(err)
	}
	firstPublishCount := len(client.publishes)
	if firstPublishCount == 0 {
		t.Fatal("expected first snapshot to publish discovery and state")
	}

	if err := publisher.PublishSnapshot(t.Context(), snapshot); err != nil {
		t.Fatal(err)
	}
	if got := len(client.publishes); got != firstPublishCount {
		t.Fatalf("unchanged snapshot triggered extra publishes: got %d, want %d", got, firstPublishCount)
	}

	changed := snapshot
	changed.Zones = append([]state.Zone(nil), snapshot.Zones...)
	changed.Zones[0].LastEventName = "Fire alarm"
	changed.Zones[0].LastEventCode = "FA"
	changed.Zones[0].LastSignal = "fire"
	changed.Zones[0].LastEventAt = time.Unix(100, 0)
	if err := publisher.PublishSnapshot(t.Context(), changed); err != nil {
		t.Fatal(err)
	}
	if got := len(client.publishes); got != firstPublishCount+1 {
		t.Fatalf("changed zone should publish exactly one updated state payload: got %d, want %d", got, firstPublishCount+1)
	}
}

func TestPublishSnapshotRepublishesAfterCacheReset(t *testing.T) {
	client := &stubClient{open: true}
	publisher := New(Config{
		Broker:      "tcp://mqtt.local:1883",
		ClientID:    "ajaxbridge",
		TopicPrefix: "ajaxbridge",
		Discovery:   true,
		Timeout:     time.Second,
		Retain:      true,
	}, zerologNop())
	publisher.client = client

	snapshot := state.Snapshot{
		Accounts: []state.Account{{Account: "A0F80D", Online: true}},
		Zones:    []state.Zone{{Account: "A0F80D", Zone: "3", DeviceEvents: []string{"fire"}}},
	}

	if err := publisher.PublishSnapshot(t.Context(), snapshot); err != nil {
		t.Fatal(err)
	}
	firstPublishCount := len(client.publishes)

	publisher.resetCaches()

	if err := publisher.PublishSnapshot(t.Context(), snapshot); err != nil {
		t.Fatal(err)
	}
	if got := len(client.publishes); got != firstPublishCount*2 {
		t.Fatalf("cache reset should republish discovery and state: got %d, want %d", got, firstPublishCount*2)
	}
}

func TestPublishUpdatePublishesOnlyProvidedSubset(t *testing.T) {
	client := &stubClient{open: true}
	publisher := New(Config{
		Broker:          "tcp://mqtt.local:1883",
		ClientID:        "ajaxbridge",
		TopicPrefix:     "ajaxbridge",
		Discovery:       true,
		DiscoveryPrefix: "homeassistant",
		Timeout:         time.Second,
		Retain:          true,
	}, zerologNop())
	publisher.client = client

	if err := publisher.PublishUpdate(t.Context(), Update{
		Accounts: []state.Account{{
			Account: "A0F80D",
			Online:  true,
			Mode:    "armed",
			Armed:   true,
		}},
	}); err != nil {
		t.Fatal(err)
	}

	if client.hasTopicContaining("/zones/") {
		t.Fatalf("account-only update should not publish any zone topics: %#v", client.publishes)
	}
	if !client.hasTopic("ajaxbridge/accounts/A0F80D/state") {
		t.Fatalf("missing account state publish: %#v", client.publishes)
	}
}

func TestDiscoveryIncludesStateTopicAsJSONAttributesTopic(t *testing.T) {
	client := &stubClient{open: true}
	publisher := New(Config{
		Broker:          "tcp://mqtt.local:1883",
		ClientID:        "ajaxbridge",
		TopicPrefix:     "ajaxbridge",
		Discovery:       true,
		DiscoveryPrefix: "homeassistant",
		Timeout:         time.Second,
		Retain:          true,
	}, zerologNop())
	publisher.client = client

	if err := publisher.PublishUpdate(t.Context(), Update{
		Zones: []state.Zone{{
			Account:      "A0F80D",
			Zone:         "3",
			DeviceEvents: []string{"fire"},
			SignalActive: map[string]bool{"fire": true},
		}},
	}); err != nil {
		t.Fatal(err)
	}

	payload := client.payloadForTopic("homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_alarm_active/config")
	raw, ok := payload.([]byte)
	if !ok {
		t.Fatalf("discovery payload = %#v, want []byte", payload)
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg["json_attributes_topic"], "ajaxbridge/accounts/A0F80D/zones/3/state"; got != want {
		t.Fatalf("json_attributes_topic = %#v, want %q", got, want)
	}
}

func TestPublishUpdateRefreshesZoneDiscoveryWhenSignalSetChanges(t *testing.T) {
	client := &stubClient{open: true}
	publisher := New(Config{
		Broker:          "tcp://mqtt.local:1883",
		ClientID:        "ajaxbridge",
		TopicPrefix:     "ajaxbridge",
		Discovery:       true,
		DiscoveryPrefix: "homeassistant",
		Timeout:         time.Second,
		Retain:          true,
	}, zerologNop())
	publisher.client = client

	base := state.Zone{
		Account:      "A0F80D",
		Zone:         "3",
		DeviceName:   "Hall fire detector",
		Kind:         "fireprotect",
		Room:         "Hall",
		DeviceEvents: []string{"fire"},
		SignalActive: map[string]bool{"fire": true},
	}
	if err := publisher.PublishUpdate(t.Context(), Update{Zones: []state.Zone{base}}); err != nil {
		t.Fatal(err)
	}
	firstCount := len(client.publishes)

	expanded := base
	expanded.SignalActive = map[string]bool{"fire": true, "connectivity": true}
	if err := publisher.PublishUpdate(t.Context(), Update{Zones: []state.Zone{expanded}}); err != nil {
		t.Fatal(err)
	}

	if got := len(client.publishes); got != firstCount+3 {
		t.Fatalf("adding a new signal should publish legacy cleanup, one discovery config, and one state update: got %d, want %d", got, firstCount+3)
	}
	if !client.hasTopic("homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_signal_connectivity/config") {
		t.Fatalf("missing discovery publish for new connectivity signal: %#v", client.publishes)
	}
	if !client.hasTopic("homeassistant/binary_sensor/ajax2prometheus/zone_a0f80d_3_signal_connectivity/config") {
		t.Fatalf("missing cleanup publish for legacy connectivity signal discovery topic: %#v", client.publishes)
	}
}

func TestPublishSnapshotCleansLegacyAjax2PrometheusDiscoveryTopics(t *testing.T) {
	client := &stubClient{open: true}
	publisher := New(Config{
		Broker:          "tcp://mqtt.local:1883",
		ClientID:        "ajaxbridge",
		TopicPrefix:     "ajaxbridge",
		Discovery:       true,
		DiscoveryPrefix: "homeassistant",
		Timeout:         time.Second,
		Retain:          true,
	}, zerologNop())
	publisher.client = client

	snapshot := state.Snapshot{
		Zones: []state.Zone{{
			Account:      "A0F80D",
			Zone:         "13",
			DeviceName:   "Attic fire detector",
			Kind:         "FireProtect",
			Room:         "Attic",
			DeviceEvents: []string{"fire", "tamper"},
			SignalActive: map[string]bool{},
		}},
	}

	if err := publisher.PublishSnapshot(t.Context(), snapshot); err != nil {
		t.Fatal(err)
	}

	legacyTopic := "homeassistant/binary_sensor/ajax2prometheus/zone_a0f80d_13_alarm_active/config"
	if !client.hasTopic(legacyTopic) {
		t.Fatalf("missing cleanup publish for legacy discovery topic %q: %#v", legacyTopic, client.publishes)
	}
	if payload := client.payloadForTopic(legacyTopic); payload == nil || stringValue(payload) != "" {
		t.Fatalf("legacy cleanup payload = %#v, want empty retained payload", payload)
	}

	currentTopic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_13_alarm_active/config"
	if !client.hasTopic(currentTopic) {
		t.Fatalf("missing current discovery publish for topic %q: %#v", currentTopic, client.publishes)
	}
}

func TestPublishSnapshotDoesNotCleanCurrentDiscoveryNode(t *testing.T) {
	client := &stubClient{open: true}
	publisher := New(Config{
		Broker:          "tcp://mqtt.local:1883",
		ClientID:        "legacy-client",
		TopicPrefix:     "ajax2prometheus",
		Discovery:       true,
		DiscoveryPrefix: "homeassistant",
		Timeout:         time.Second,
		Retain:          true,
	}, zerologNop())
	publisher.client = client

	snapshot := state.Snapshot{
		Zones: []state.Zone{{
			Account:      "A0F80D",
			Zone:         "13",
			DeviceEvents: []string{"fire"},
		}},
	}

	if err := publisher.PublishSnapshot(t.Context(), snapshot); err != nil {
		t.Fatal(err)
	}

	topic := "homeassistant/binary_sensor/ajax2prometheus/zone_a0f80d_13_alarm_active/config"
	if !client.hasTopic(topic) {
		t.Fatalf("missing discovery publish for current node %q: %#v", topic, client.publishes)
	}
	if payload := client.payloadForTopic(topic); stringValue(payload) == "" {
		t.Fatalf("current discovery topic %q was incorrectly cleaned up", topic)
	}
}

func TestCleanupTopicMatcherTargetsOnlyStaleJeedomAndLegacyTopics(t *testing.T) {
	publisher := New(Config{
		Broker:          "tcp://mqtt.local:1883",
		ClientID:        "ajaxbridge",
		TopicPrefix:     "ajaxbridge",
		DiscoveryPrefix: "homeassistant",
	}, zerologNop())
	patterns := publisher.cleanupSubscriptions(CleanupConfig{
		JeedomStateTopicPrefix: "ajaxbridge/jeedom",
	})

	cases := map[string]bool{
		"homeassistant/sensor/ajaxbridge/jeedom_cmd_56/config":            true,
		"homeassistant/switch/ajaxbridge/jeedom_control_serverna/config":  true,
		"ajaxbridge/jeedom/devices/serverna/state":                        true,
		"homeassistant/binary_sensor/ajax2prometheus/zone_1_alarm/config": true,
		"ajax2prometheus/accounts/A0F80D/state":                           true,
		"homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_alarm/config": false,
		"homeassistant/sensor/ajaxbridge/account_a0f80d_mode/config":      false,
		"ajaxbridge/accounts/A0F80D/state":                                false,
		"ajaxbridge/accounts/A0F80D/zones/8/state":                        false,
	}
	for topic, want := range cases {
		if got := cleanupTopicMatches(patterns, topic); got != want {
			t.Fatalf("cleanupTopicMatches(%q) = %t, want %t", topic, got, want)
		}
	}
}

func zerologNop() zerolog.Logger {
	return zerolog.Nop()
}

type stubClient struct {
	open      bool
	publishes []stubPublish
}

type stubPublish struct {
	topic    string
	retained bool
	payload  interface{}
}

func (c *stubClient) IsConnected() bool {
	return c.open
}

func (c *stubClient) IsConnectionOpen() bool {
	return c.open
}

func (c *stubClient) Connect() paho.Token {
	return closedToken(nil)
}

func (c *stubClient) Disconnect(uint) {}

func (c *stubClient) Publish(topic string, _ byte, retained bool, payload interface{}) paho.Token {
	c.publishes = append(c.publishes, stubPublish{
		topic:    topic,
		retained: retained,
		payload:  payload,
	})
	return closedToken(nil)
}

func (c *stubClient) Subscribe(string, byte, paho.MessageHandler) paho.Token {
	return closedToken(nil)
}

func (c *stubClient) SubscribeMultiple(map[string]byte, paho.MessageHandler) paho.Token {
	return closedToken(nil)
}

func (c *stubClient) Unsubscribe(...string) paho.Token {
	return closedToken(nil)
}

func (c *stubClient) AddRoute(string, paho.MessageHandler) {}

func (c *stubClient) OptionsReader() paho.ClientOptionsReader {
	return paho.ClientOptionsReader{}
}

func (c *stubClient) hasTopic(topic string) bool {
	for _, publish := range c.publishes {
		if publish.topic == topic {
			return true
		}
	}
	return false
}

func (c *stubClient) hasTopicContaining(fragment string) bool {
	for _, publish := range c.publishes {
		if strings.Contains(publish.topic, fragment) {
			return true
		}
	}
	return false
}

func (c *stubClient) payloadForTopic(topic string) interface{} {
	for _, publish := range c.publishes {
		if publish.topic == topic {
			return publish.payload
		}
	}
	return nil
}

func stringValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return ""
	}
}

type stubToken struct {
	done chan struct{}
	err  error
}

func closedToken(err error) *stubToken {
	done := make(chan struct{})
	close(done)
	return &stubToken{done: done, err: err}
}

func (t *stubToken) Wait() bool {
	<-t.done
	return true
}

func (t *stubToken) WaitTimeout(time.Duration) bool {
	<-t.done
	return true
}

func (t *stubToken) Done() <-chan struct{} {
	return t.done
}

func (t *stubToken) Error() error {
	return t.err
}
