package hamqtt

import (
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

	if got := len(client.publishes); got != firstCount+2 {
		t.Fatalf("adding a new signal should publish one discovery config and one state update: got %d, want %d", got, firstCount+2)
	}
	if !client.hasTopic("homeassistant/binary_sensor/ajaxbridge/zone_a0f80d_3_signal_connectivity/config") {
		t.Fatalf("missing discovery publish for new connectivity signal: %#v", client.publishes)
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
