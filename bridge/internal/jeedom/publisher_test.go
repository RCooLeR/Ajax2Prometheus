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
}

func TestSwitchDiscoveryUsesBridgeCommandTopic(t *testing.T) {
	device := Device{
		Source:     Source,
		Device:     "Garage gate",
		DeviceSlug: "garage_gate",
		Actions: map[string]Action{
			"on":  {Action: "on", CommandID: "85", DeviceSlug: "garage_gate", StateCommandID: "81", Allowed: true},
			"off": {Action: "off", CommandID: "86", DeviceSlug: "garage_gate", StateCommandID: "81", Allowed: true},
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
	if topic != "homeassistant/switch/ajaxbridge/jeedom_control_garage_gate/config" {
		t.Fatalf("discovery topic = %q", topic)
	}
	var payload DiscoveryConfig
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.CommandTopic != "ajaxbridge/jeedom/devices/garage_gate/set" {
		t.Fatalf("command_topic = %q", payload.CommandTopic)
	}
	if payload.StateTopic != "ajaxbridge/jeedom/devices/garage_gate/state" {
		t.Fatalf("state_topic = %q", payload.StateTopic)
	}
	if payload.Optimistic == nil || *payload.Optimistic {
		t.Fatalf("optimistic = %#v, want false", payload.Optimistic)
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
