package jeedom

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestDiscoveryPayloadUsesStableCommandIDAndDeviceIdentifier(t *testing.T) {
	device := Device{
		Source:     Source,
		Device:     "Серверна",
		DeviceSlug: "serverna",
		RawCommands: map[string]Command{
			"56": {
				CommandID:   "56",
				Device:      "Серверна",
				DeviceSlug:  "serverna",
				Name:        "Puissance",
				Metric:      "power_w",
				Component:   ComponentSensor,
				DeviceClass: "power",
				StateClass:  "measurement",
				Unit:        "W",
				LastUpdate:  time.Unix(100, 0),
			},
		},
	}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
	}, fakeMQTT{})

	topic, body, err := publisher.BuildDiscovery(device.RawCommands["56"], device)
	if err != nil {
		t.Fatal(err)
	}
	if topic != "homeassistant/sensor/ajaxbridge/jeedom_cmd_56/config" {
		t.Fatalf("discovery topic = %q", topic)
	}

	var payload DiscoveryConfig
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.UniqueID != "ajaxbridge_jeedom_cmd_56" {
		t.Fatalf("UniqueID = %q", payload.UniqueID)
	}
	if len(payload.Device.Identifiers) != 1 || payload.Device.Identifiers[0] != "ajaxbridge_jeedom_serverna" {
		t.Fatalf("identifiers = %#v", payload.Device.Identifiers)
	}
	if payload.StateTopic != "ajaxbridge/jeedom/devices/serverna/state" {
		t.Fatalf("state_topic = %q", payload.StateTopic)
	}
	if payload.JSONAttributesTopic != "ajaxbridge/jeedom/devices/serverna/attributes" {
		t.Fatalf("json_attributes_topic = %q", payload.JSONAttributesTopic)
	}
}

func TestSwitchDiscoveryUsesBridgeCommandTopic(t *testing.T) {
	device := Device{
		Source:           Source,
		Device:           "Server outlet",
		DeviceSlug:       "server_outlet",
		JeedomDeviceType: "Outlet",
		Actions: map[string]Action{
			"on":  {Action: "on", CommandID: "85", DeviceSlug: "server_outlet", StateCommandID: "81", Allowed: true},
			"off": {Action: "off", CommandID: "86", DeviceSlug: "server_outlet", StateCommandID: "81", Allowed: true},
		},
	}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		Controls:         true,
	}, fakeMQTT{})

	topic, body, err := publisher.BuildSwitchDiscovery(device.Actions["on"], device)
	if err != nil {
		t.Fatal(err)
	}
	if topic != "homeassistant/switch/ajaxbridge/jeedom_control_server_outlet/config" {
		t.Fatalf("discovery topic = %q", topic)
	}
	var payload DiscoveryConfig
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.CommandTopic != "ajaxbridge/jeedom/devices/server_outlet/set" {
		t.Fatalf("command_topic = %q", payload.CommandTopic)
	}
	if payload.StateTopic != "ajaxbridge/jeedom/devices/server_outlet/state" {
		t.Fatalf("state_topic = %q", payload.StateTopic)
	}
	if payload.JSONAttributesTopic != "ajaxbridge/jeedom/devices/server_outlet/attributes" {
		t.Fatalf("json_attributes_topic = %q", payload.JSONAttributesTopic)
	}
	if payload.Optimistic == nil || *payload.Optimistic {
		t.Fatalf("optimistic = %#v, want false", payload.Optimistic)
	}
}

func TestButtonDiscoveryUsesImpulsePayload(t *testing.T) {
	device := Device{
		Source:           Source,
		Device:           "Garage gate",
		DeviceSlug:       "garage_gate",
		JeedomDeviceType: "Relay",
		Actions: map[string]Action{
			"on": {Action: "on", CommandID: "85", DeviceSlug: "garage_gate", StateCommandID: "81", Allowed: true},
		},
	}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		Controls:         true,
	}, fakeMQTT{})

	topic, body, err := publisher.BuildButtonDiscovery(device.Actions["on"], device)
	if err != nil {
		t.Fatal(err)
	}
	if topic != "homeassistant/button/ajaxbridge/jeedom_control_garage_gate_impulse/config" {
		t.Fatalf("discovery topic = %q", topic)
	}
	var payload DiscoveryConfig
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Name != "Impulse" {
		t.Fatalf("name = %q, want Impulse", payload.Name)
	}
	if payload.PayloadPress != "ON" {
		t.Fatalf("payload_press = %q, want ON", payload.PayloadPress)
	}
	if payload.CommandTopic != "ajaxbridge/jeedom/devices/garage_gate/set" {
		t.Fatalf("command_topic = %q", payload.CommandTopic)
	}
	if payload.StateTopic != "" {
		t.Fatalf("button state_topic = %q, want empty", payload.StateTopic)
	}
}

func TestPublishDevicePublishesToggleForOutletAndImpulseForRelay(t *testing.T) {
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		RetainDiscovery:  true,
		Controls:         true,
	}, nil)

	outletMQTT := &recordingMQTT{}
	publisher.mqtt = outletMQTT
	outlet := Device{
		Source:           Source,
		Device:           "Server outlet",
		DeviceSlug:       "server_outlet",
		JeedomDeviceType: "Outlet",
		Values:           map[string]any{"state": false},
		RawCommands:      map[string]Command{},
		Actions: map[string]Action{
			"on":  {Action: "on", CommandID: "85", DeviceSlug: "server_outlet", StateCommandID: "81", Allowed: true},
			"off": {Action: "off", CommandID: "86", DeviceSlug: "server_outlet", StateCommandID: "81", Allowed: true},
		},
	}
	if err := publisher.PublishDevice(context.Background(), outlet); err != nil {
		t.Fatal(err)
	}
	if got := outletMQTT.discovery["homeassistant/switch/ajaxbridge/jeedom_control_server_outlet/config"]; got == "" {
		t.Fatalf("outlet toggle switch discovery missing")
	}
	if got := outletMQTT.discovery["homeassistant/button/ajaxbridge/jeedom_control_server_outlet_impulse/config"]; got != "" {
		t.Fatalf("outlet impulse button discovery = %q, want cleanup/empty", got)
	}

	relayMQTT := &recordingMQTT{}
	publisher.mqtt = relayMQTT
	relay := Device{
		Source:           Source,
		Device:           "Garage gate",
		DeviceSlug:       "garage_gate",
		JeedomDeviceType: "Relay",
		Values:           map[string]any{"state": false},
		RawCommands:      map[string]Command{},
		Actions: map[string]Action{
			"on":  {Action: "on", CommandID: "85", DeviceSlug: "garage_gate", StateCommandID: "81", Allowed: true},
			"off": {Action: "off", CommandID: "86", DeviceSlug: "garage_gate", StateCommandID: "81", Allowed: true},
		},
	}
	if err := publisher.PublishDevice(context.Background(), relay); err != nil {
		t.Fatal(err)
	}
	if got := relayMQTT.discovery["homeassistant/switch/ajaxbridge/jeedom_control_garage_gate/config"]; got != "" {
		t.Fatalf("relay toggle switch discovery = %q, want cleanup/empty", got)
	}
	if got := relayMQTT.discovery["homeassistant/button/ajaxbridge/jeedom_control_garage_gate_impulse/config"]; got == "" {
		t.Fatalf("relay impulse button discovery missing")
	}

	waterStopMQTT := &recordingMQTT{}
	publisher.mqtt = waterStopMQTT
	waterStop := Device{
		Source:           Source,
		Device:           "Water valve",
		DeviceSlug:       "water_valve",
		JeedomDeviceType: "WaterStop",
		Values:           map[string]any{"state": false},
		RawCommands:      map[string]Command{},
		Actions: map[string]Action{
			"on":  {Action: "on", CommandID: "26", DeviceSlug: "water_valve", StateCommandID: "23", Allowed: true},
			"off": {Action: "off", CommandID: "27", DeviceSlug: "water_valve", StateCommandID: "23", Allowed: true},
		},
	}
	if err := publisher.PublishDevice(context.Background(), waterStop); err != nil {
		t.Fatal(err)
	}
	if got := waterStopMQTT.discovery["homeassistant/switch/ajaxbridge/jeedom_control_water_valve/config"]; got == "" {
		t.Fatalf("WaterStop toggle switch discovery missing")
	}
}

func TestPublishDevicePublishesHubSecurityButtons(t *testing.T) {
	mqtt := &recordingMQTT{}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		RetainDiscovery:  true,
		Controls:         true,
	}, mqtt)

	hub := Device{
		Source:           Source,
		Device:           "Security hub",
		DeviceSlug:       "security_hub",
		JeedomDeviceType: "Hub",
		Values:           map[string]any{},
		RawCommands:      map[string]Command{},
		Actions: map[string]Action{
			"arm":                 {Action: "arm", CommandID: "163", DeviceSlug: "security_hub", Name: "Arm", Allowed: true},
			"night_mode":          {Action: "night_mode", CommandID: "164", DeviceSlug: "security_hub", Name: "Night mode", Allowed: true},
			"disarm":              {Action: "disarm", CommandID: "165", DeviceSlug: "security_hub", Name: "Disarm", Allowed: true},
			"mute_fire_detectors": {Action: "mute_fire_detectors", CommandID: "167", DeviceSlug: "security_hub", Name: "Mute fire detectors", Allowed: true},
		},
	}
	if err := publisher.PublishDevice(context.Background(), hub); err != nil {
		t.Fatal(err)
	}

	for _, actionSlug := range []string{"arm", "night_mode", "disarm", "mute_fire_detectors"} {
		topic := "homeassistant/button/ajaxbridge/jeedom_control_security_hub_" + actionSlug + "/config"
		if got := mqtt.discovery[topic]; got == "" {
			t.Fatalf("hub %s button discovery missing at %s", actionSlug, topic)
		}
	}
	if got := mqtt.discovery["homeassistant/switch/ajaxbridge/jeedom_control_security_hub/config"]; got != "" {
		t.Fatalf("hub switch discovery = %q, want cleanup/empty", got)
	}
}

func TestPublishDeviceClearsLegacyUnlinkedSwitchAndState(t *testing.T) {
	mqtt := &recordingMQTT{}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		RetainState:      true,
		RetainDiscovery:  true,
		Controls:         true,
	}, mqtt)

	device := Device{
		Source:            Source,
		Device:            "Server power",
		DeviceSlug:        "sia_a0f80d_zone_8",
		LegacyDeviceSlugs: []string{"serverna"},
		Values:            map[string]any{},
		RawCommands:       map[string]Command{},
		Actions:           map[string]Action{},
	}
	if err := publisher.PublishDevice(context.Background(), device); err != nil {
		t.Fatal(err)
	}

	if got := mqtt.discovery["homeassistant/switch/ajaxbridge/jeedom_control_serverna/config"]; got != "" {
		t.Fatalf("legacy switch cleanup payload = %q, want empty", got)
	}
	if got := mqtt.state["ajaxbridge/jeedom/devices/serverna/state"]; got != "" {
		t.Fatalf("legacy state cleanup payload = %q, want empty", got)
	}
	if got := mqtt.state["ajaxbridge/jeedom/devices/serverna/attributes"]; got != "" {
		t.Fatalf("legacy attributes cleanup payload = %q, want empty", got)
	}
}

func TestPublishDeviceCleansSIAMergedDuplicateCommands(t *testing.T) {
	mqtt := &recordingMQTT{}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		RetainDiscovery:  true,
	}, mqtt)

	device := Device{
		Source:        Source,
		Device:        "Garage fire detector",
		DeviceSlug:    "sia_a0f80d_zone_4",
		LinkedSource:  "sia",
		LinkedAccount: "A0F80D",
		LinkedZone:    "4",
		HAIdentifiers: []string{"ajaxbridge_A0F80D_zone_4"},
		Values:        map[string]any{},
		RawCommands: map[string]Command{
			"134": {
				CommandID: "134",
				Name:      "Bypassed",
				Metric:    "bypass",
				Component: ComponentBinarySensor,
			},
			"136": {
				CommandID:   "136",
				Name:        "Temperature",
				Metric:      "temperature_c",
				Component:   ComponentSensor,
				Unit:        "\u00b0C",
				DeviceClass: "temperature",
				StateClass:  "measurement",
			},
		},
	}

	if err := publisher.PublishDevice(context.Background(), device); err != nil {
		t.Fatal(err)
	}

	if got := mqtt.discovery["homeassistant/binary_sensor/ajaxbridge/jeedom_cmd_134/config"]; got != "" {
		t.Fatalf("SIA-owned bypass command discovery = %q, want retained cleanup", got)
	}
	if got := mqtt.discovery["homeassistant/sensor/ajaxbridge/jeedom_cmd_134/config"]; got != "" {
		t.Fatalf("SIA-owned bypass sensor migration cleanup = %q, want empty", got)
	}
	if got := mqtt.discovery["homeassistant/sensor/ajaxbridge/jeedom_cmd_136/config"]; got == "" {
		t.Fatalf("Jeedom temperature measurement discovery missing")
	}
	if got := mqtt.discovery["homeassistant/binary_sensor/ajaxbridge/jeedom_cmd_136/config"]; got != "" {
		t.Fatalf("stale Jeedom temperature binary discovery = %q, want retained cleanup", got)
	}
}

func TestPublishDeviceCleansLegacyNameBasedDiscoveryForAllSensorTypes(t *testing.T) {
	mqtt := &recordingMQTT{}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		RetainDiscovery:  true,
	}, mqtt)

	device := Device{
		Source:            Source,
		Device:            "Ajax account A0F80D",
		DeviceSlug:        "account_a0f80d",
		BaseSlug:          "budinok",
		LegacyDeviceSlugs: []string{"budinok"},
		LinkedSource:      "sia",
		LinkedAccount:     "A0F80D",
		HAIdentifiers:     []string{"ajaxbridge_account_A0F80D"},
		Values:            map[string]any{},
		RawCommands: map[string]Command{
			"10": {
				CommandID:  "10",
				ObjectName: "Zahidna 20",
				Device:     "Будинок",
				DeviceSlug: "account_a0f80d",
				Name:       "Battery",
				RawName:    "Batterie",
				Metric:     "battery_percent",
				Component:  ComponentSensor,
			},
			"11": {
				CommandID:  "11",
				ObjectName: "Zahidna 20",
				Device:     "Будинок",
				DeviceSlug: "account_a0f80d",
				Name:       "External power",
				RawName:    "Alimentation secteur",
				Metric:     "external_power",
				Component:  ComponentBinarySensor,
			},
		},
	}

	if err := publisher.PublishDevice(context.Background(), device); err != nil {
		t.Fatal(err)
	}

	for _, topic := range []string{
		"homeassistant/sensor/ajaxbridge/budinok_battery/config",
		"homeassistant/binary_sensor/ajaxbridge/budinok_battery/config",
		"homeassistant/binary_sensor/ajaxbridge/budinok_external_power/config",
		"homeassistant/sensor/ajaxbridge/budinok_external_power/config",
	} {
		if got := mqtt.discovery[topic]; got != "" {
			t.Fatalf("legacy topic %s cleanup payload = %q, want empty", topic, got)
		}
		if _, ok := mqtt.discovery[topic]; !ok {
			t.Fatalf("missing legacy cleanup for %s: %#v", topic, mqtt.discovery)
		}
	}
	if got := mqtt.discovery["homeassistant/sensor/ajaxbridge/jeedom_cmd_10/config"]; got == "" {
		t.Fatalf("current battery discovery missing")
	}
	if got := mqtt.discovery["homeassistant/binary_sensor/ajaxbridge/jeedom_cmd_11/config"]; got != "" {
		t.Fatalf("SIA-owned external power command should be cleaned by command id, got %q", got)
	}
}

func TestPublishDeviceKeepsControlStateForLinkedSIADevice(t *testing.T) {
	mqtt := &recordingMQTT{}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		RetainDiscovery:  true,
	}, mqtt)

	device := Device{
		Source:        Source,
		Device:        "Server power",
		DeviceSlug:    "sia_a0f80d_zone_8",
		LinkedSource:  "sia",
		LinkedAccount: "A0F80D",
		LinkedZone:    "8",
		HAIdentifiers: []string{"ajaxbridge_A0F80D_zone_8"},
		Values:        map[string]any{},
		RawCommands: map[string]Command{
			"52": {
				CommandID: "52",
				Name:      "State",
				Metric:    "state",
				Component: ComponentBinarySensor,
			},
		},
	}

	if err := publisher.PublishDevice(context.Background(), device); err != nil {
		t.Fatal(err)
	}

	if got := mqtt.discovery["homeassistant/binary_sensor/ajaxbridge/jeedom_cmd_52/config"]; got == "" {
		t.Fatalf("Jeedom control state discovery should stay for linked SIA device")
	}
}

func TestPublishDeviceSeparatesStableAttributesFromFullState(t *testing.T) {
	mqtt := &recordingMQTT{}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		RetainState:      true,
	}, mqtt)
	device := Device{
		Source:           Source,
		ObjectName:       "House",
		Device:           "Server power",
		DeviceSlug:       "sia_a0f80d_zone_8",
		JeedomID:         "7",
		JeedomLogicalID:  "30E81A2B",
		JeedomDeviceType: "WallSwitch",
		LastUpdate:       time.Unix(100, 0),
		Values:           map[string]any{"power_w": 120.5, "state": true},
		RawCommands: map[string]Command{
			"56": {
				CommandID:  "56",
				Metric:     "power_w",
				Value:      120.5,
				LastUpdate: time.Unix(100, 0),
			},
		},
		Actions: map[string]Action{
			"on": {Action: "on", CommandID: "58", Allowed: true},
		},
	}

	if err := publisher.PublishDevice(context.Background(), device); err != nil {
		t.Fatal(err)
	}

	var statePayload map[string]any
	if err := json.Unmarshal([]byte(mqtt.state["ajaxbridge/jeedom/devices/sia_a0f80d_zone_8/state"]), &statePayload); err != nil {
		t.Fatal(err)
	}
	if _, ok := statePayload["raw_commands"]; !ok {
		t.Fatalf("existing MQTT state contract lost raw_commands: %#v", statePayload)
	}
	if _, ok := statePayload["last_update"]; !ok {
		t.Fatalf("existing MQTT state contract lost last_update: %#v", statePayload)
	}

	var attributes map[string]any
	if err := json.Unmarshal([]byte(mqtt.state["ajaxbridge/jeedom/devices/sia_a0f80d_zone_8/attributes"]), &attributes); err != nil {
		t.Fatal(err)
	}
	if got := attributes["jeedom_device_type"]; got != "WallSwitch" {
		t.Fatalf("jeedom_device_type = %#v, want WallSwitch", got)
	}
	for _, volatile := range []string{"raw_commands", "actions", "last_update", "last_value_at", "power_w", "state"} {
		if _, ok := attributes[volatile]; ok {
			t.Fatalf("volatile field %q leaked into stable attributes: %#v", volatile, attributes)
		}
	}
}

type fakeMQTT struct{}

func (fakeMQTT) PublishStateMessage(context.Context, string, []byte, bool) error {
	return nil
}

func (fakeMQTT) PublishDiscoveryMessage(context.Context, string, string, []byte, bool) error {
	return nil
}

func (fakeMQTT) AvailabilityTopic() string {
	return "ajaxbridge/status"
}

type recordingMQTT struct {
	state     map[string]string
	discovery map[string]string
}

func (m *recordingMQTT) PublishStateMessage(_ context.Context, topic string, payload []byte, _ bool) error {
	if m.state == nil {
		m.state = make(map[string]string)
	}
	m.state[topic] = string(payload)
	return nil
}

func (m *recordingMQTT) PublishDiscoveryMessage(_ context.Context, _ string, topic string, payload []byte, _ bool) error {
	if m.discovery == nil {
		m.discovery = make(map[string]string)
	}
	m.discovery[topic] = string(payload)
	return nil
}

func (m *recordingMQTT) AvailabilityTopic() string {
	return "ajaxbridge/status"
}
