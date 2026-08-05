package hamqtt

import (
	"context"
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

func TestSIASignalEntitiesUseAlarmSpecificIDsForJeedomMeasurements(t *testing.T) {
	entities := zoneEntities(state.Zone{
		Account:      "A0F80D",
		Zone:         "3",
		DeviceEvents: []string{"power", "temperature"},
	})

	got := make(map[string]entity, len(entities))
	for _, entity := range entities {
		got[entity.ObjectID] = entity
	}
	if _, ok := got["zone_A0F80D_3_signal_power"]; ok {
		t.Fatal("old generic SIA power signal object id should not be used")
	}
	if _, ok := got["zone_A0F80D_3_signal_temperature"]; ok {
		t.Fatal("old generic SIA temperature signal object id should not be used")
	}
	if got["zone_A0F80D_3_signal_power_failure"].Name != "Power failure" {
		t.Fatalf("power signal entity = %#v", got["zone_A0F80D_3_signal_power_failure"])
	}
	if got["zone_A0F80D_3_signal_temperature_alarm"].Name != "Temperature alarm" {
		t.Fatalf("temperature signal entity = %#v", got["zone_A0F80D_3_signal_temperature_alarm"])
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

func TestRetainedDiscoveryRepublishesAfterCleanupPayload(t *testing.T) {
	client := &stubClient{open: true}
	publisher := New(Config{
		Broker:   "tcp://mqtt.local:1883",
		ClientID: "ajaxbridge",
		Timeout:  time.Second,
	}, zerologNop())
	publisher.client = client

	topic := "homeassistant/sensor/ajaxbridge/jeedom_cmd_345/config"
	payload := []byte(`{"name":"Voltage","unique_id":"ajaxbridge_jeedom_cmd_345"}`)

	if err := publisher.PublishDiscoveryMessage(t.Context(), "jeedom_discovery:sensor/jeedom_cmd_345", topic, payload, true); err != nil {
		t.Fatal(err)
	}
	if err := publisher.PublishDiscoveryMessage(t.Context(), "jeedom_discovery:sensor/jeedom_cmd_345", topic, payload, true); err != nil {
		t.Fatal(err)
	}
	if err := publisher.PublishDiscoveryMessage(t.Context(), "jeedom_discovery_cleanup:sensor/jeedom_cmd_345", topic, []byte{}, true); err != nil {
		t.Fatal(err)
	}
	if err := publisher.PublishDiscoveryMessage(t.Context(), "jeedom_discovery:sensor/jeedom_cmd_345", topic, payload, true); err != nil {
		t.Fatal(err)
	}

	var topicPayloads []string
	for _, publish := range client.publishes {
		if publish.topic == topic {
			topicPayloads = append(topicPayloads, stringValue(publish.payload))
		}
	}
	if got, want := len(topicPayloads), 3; got != want {
		t.Fatalf("discovery publishes for %s = %d, want %d: %#v", topic, got, want, topicPayloads)
	}
	if topicPayloads[0] == "" || topicPayloads[1] != "" || topicPayloads[2] == "" {
		t.Fatalf("discovery payload sequence = %#v, want config, cleanup, config", topicPayloads)
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

func TestDiscoveryUsesStableMetadataAsJSONAttributesTopic(t *testing.T) {
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
			Kind:         "FireProtect",
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
	if got, want := cfg["json_attributes_topic"], "ajaxbridge/accounts/A0F80D/zones/3/attributes"; got != want {
		t.Fatalf("json_attributes_topic = %#v, want %q", got, want)
	}

	attributesRaw, ok := client.payloadForTopic("ajaxbridge/accounts/A0F80D/zones/3/attributes").([]byte)
	if !ok {
		t.Fatalf("attributes payload = %#v, want []byte", client.payloadForTopic("ajaxbridge/accounts/A0F80D/zones/3/attributes"))
	}
	var attributes map[string]any
	if err := json.Unmarshal(attributesRaw, &attributes); err != nil {
		t.Fatal(err)
	}
	if got := attributes["zone"]; got != "3" {
		t.Fatalf("zone attribute = %#v, want 3", got)
	}
	if got := attributes["kind"]; got != "FireProtect" {
		t.Fatalf("kind attribute = %#v, want FireProtect", got)
	}
	for _, volatile := range []string{"alarm_active", "last_event_at", "last_event_unix", "signal_active"} {
		if _, ok := attributes[volatile]; ok {
			t.Fatalf("volatile field %q leaked into stable attributes: %#v", volatile, attributes)
		}
	}
}

func TestZoneDiscoveryPayloadKeepsStableHAContract(t *testing.T) {
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
			Zone:         "8",
			DeviceName:   "Server room detector",
			Kind:         "FireProtect",
			Room:         "Server Room",
			DeviceEvents: []string{"fire"},
		}},
	}); err != nil {
		t.Fatal(err)
	}

	topic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_8_alarm_active/config"
	raw, ok := client.payloadForTopic(topic).([]byte)
	if !ok {
		t.Fatalf("discovery payload for %s = %#v, want []byte", topic, client.payloadForTopic(topic))
	}
	var cfg discoveryConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.UniqueID != "ajaxbridge_zone_a0f80d_8_alarm_active" {
		t.Fatalf("UniqueID = %q", cfg.UniqueID)
	}
	if cfg.StateTopic != "ajaxbridge/accounts/A0F80D/zones/8/state" {
		t.Fatalf("StateTopic = %q", cfg.StateTopic)
	}
	if cfg.JSONAttributesTopic != "ajaxbridge/accounts/A0F80D/zones/8/attributes" {
		t.Fatalf("JSONAttributesTopic = %q, want stable attributes topic", cfg.JSONAttributesTopic)
	}
	if cfg.ValueTemplate != "{{ 'ON' if value_json.alarm_active else 'OFF' }}" {
		t.Fatalf("ValueTemplate = %q", cfg.ValueTemplate)
	}
	if cfg.Device.Name != "Server room detector" || cfg.Device.Model != "FireProtect" || cfg.Device.SuggestedArea != "Server Room" {
		t.Fatalf("device metadata = %#v", cfg.Device)
	}
	if len(cfg.Device.Identifiers) != 1 || cfg.Device.Identifiers[0] != "ajaxbridge_A0F80D_zone_8" {
		t.Fatalf("device identifiers = %#v", cfg.Device.Identifiers)
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

	if got := len(client.publishes); got <= firstCount {
		t.Fatalf("adding a new signal should publish additional cleanup, discovery, and state payloads: got %d, first %d", got, firstCount)
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

func TestPublishSnapshotCleansLegacySIAObjectIDDiscoveryTopics(t *testing.T) {
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

	if err := publisher.PublishSnapshot(t.Context(), state.Snapshot{
		Zones: []state.Zone{{
			Account:      "A0F80D",
			Zone:         "13",
			DeviceName:   "Attic fire detector",
			Kind:         "FireProtect",
			Room:         "Attic",
			DeviceEvents: []string{"firmware"},
		}},
	}); err != nil {
		t.Fatal(err)
	}

	oldAlarmSignalTopic := "homeassistant/sensor/ajaxbridge/zone_13_alarm_signal/config"
	if !client.hasTopic(oldAlarmSignalTopic) || stringValue(client.payloadForTopic(oldAlarmSignalTopic)) != "" {
		t.Fatalf("missing retained cleanup for old alarm signal object id: %#v", client.publishes)
	}
	oldFirmwareTopic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_13_firmware/config"
	if !client.hasTopic(oldFirmwareTopic) || stringValue(client.payloadForTopic(oldFirmwareTopic)) != "" {
		t.Fatalf("missing retained cleanup for old firmware object id: %#v", client.publishes)
	}
	oldShortFirmwareTopic := "homeassistant/binary_sensor/ajaxbridge/zone_13_signal_firmware/config"
	if !client.hasTopic(oldShortFirmwareTopic) || stringValue(client.payloadForTopic(oldShortFirmwareTopic)) != "" {
		t.Fatalf("missing retained cleanup for old short firmware object id: %#v", client.publishes)
	}
	oldLegacyFirmwareTopic := "homeassistant/binary_sensor/ajax2prometheus/zone_a0f80d_13_signal_firmware/config"
	if !client.hasTopic(oldLegacyFirmwareTopic) || stringValue(client.payloadForTopic(oldLegacyFirmwareTopic)) != "" {
		t.Fatalf("missing retained cleanup for old legacy firmware object id: %#v", client.publishes)
	}

	currentAlarmSignalTopic := "homeassistant/sensor/ajaxbridge/zone_a0f80d_13_alarm_signal/config"
	if !client.hasTopic(currentAlarmSignalTopic) || stringValue(client.payloadForTopic(currentAlarmSignalTopic)) == "" {
		t.Fatalf("missing current discovery for alarm signal object id: %#v", client.publishes)
	}
	currentFirmwareTopic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_13_signal_firmware/config"
	if !client.hasTopic(currentFirmwareTopic) || stringValue(client.payloadForTopic(currentFirmwareTopic)) == "" {
		t.Fatalf("missing current discovery for firmware object id: %#v", client.publishes)
	}
}

func TestPublishSnapshotCleansRenamedSIASignalDiscoveryTopics(t *testing.T) {
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

	if err := publisher.PublishSnapshot(t.Context(), state.Snapshot{
		Zones: []state.Zone{{
			Account:      "A0F80D",
			Zone:         "3",
			DeviceEvents: []string{"power", "temperature"},
		}},
	}); err != nil {
		t.Fatal(err)
	}

	oldTemperatureTopic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_signal_temperature/config"
	if !client.hasTopic(oldTemperatureTopic) || stringValue(client.payloadForTopic(oldTemperatureTopic)) != "" {
		t.Fatalf("missing retained cleanup for old temperature signal topic: %#v", client.publishes)
	}
	oldLegacyTemperatureTopic := "homeassistant/binary_sensor/ajax2prometheus/zone_a0f80d_3_signal_temperature/config"
	if !client.hasTopic(oldLegacyTemperatureTopic) || stringValue(client.payloadForTopic(oldLegacyTemperatureTopic)) != "" {
		t.Fatalf("missing retained cleanup for old legacy temperature signal topic: %#v", client.publishes)
	}
	olderTemperatureTopic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_temperature/config"
	if !client.hasTopic(olderTemperatureTopic) || stringValue(client.payloadForTopic(olderTemperatureTopic)) != "" {
		t.Fatalf("missing retained cleanup for older temperature signal topic: %#v", client.publishes)
	}
	newTemperatureTopic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_signal_temperature_alarm/config"
	if !client.hasTopic(newTemperatureTopic) || stringValue(client.payloadForTopic(newTemperatureTopic)) == "" {
		t.Fatalf("missing discovery for new temperature alarm signal topic: %#v", client.publishes)
	}
	oldPowerTopic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_signal_power/config"
	if !client.hasTopic(oldPowerTopic) || stringValue(client.payloadForTopic(oldPowerTopic)) != "" {
		t.Fatalf("missing retained cleanup for old power signal topic: %#v", client.publishes)
	}
	newPowerTopic := "homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_signal_power_failure/config"
	if !client.hasTopic(newPowerTopic) || stringValue(client.payloadForTopic(newPowerTopic)) == "" {
		t.Fatalf("missing discovery for new power failure signal topic: %#v", client.publishes)
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

func TestConnectHandlerRunsOnReconnectAndImmediateRegistration(t *testing.T) {
	publisher := New(Config{
		Broker:      "tcp://mqtt.local:1883",
		ClientID:    "ajaxbridge",
		TopicPrefix: "ajaxbridge",
		Timeout:     time.Second,
	}, zerologNop())
	first := make(chan struct{}, 1)
	publisher.AddConnectHandler(func(_ context.Context) {
		first <- struct{}{}
	})
	publisher.notifyConnected()
	select {
	case <-first:
	case <-time.After(time.Second):
		t.Fatal("connect handler did not run on reconnect notification")
	}

	publisher.client = &stubClient{open: true}
	second := make(chan struct{}, 1)
	publisher.AddConnectHandler(func(_ context.Context) {
		second <- struct{}{}
	})
	select {
	case <-second:
	case <-time.After(time.Second):
		t.Fatal("connect handler did not run when registered after connection")
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
