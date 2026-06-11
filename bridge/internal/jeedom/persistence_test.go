package jeedom

import (
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestServicePersistsJeedomDiscoveryForRestartRepublish(t *testing.T) {
	path := t.TempDir() + "/jeedom.json"
	store, err := LoadStore(t.Context(), path, "keep_last", nil)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(ServiceConfig{EventTopic: "jeedom/cmd/event/#"}, store, nil, nil, nil, nil, zerolog.Nop())

	service.HandleMessage(t.Context(), "jeedom/discovery/eqLogic/8", []byte(`{
	  "id":8,
	  "name":"Server power",
	  "configuration":{"device":"WallSwitch"},
	  "isVisible":1,
	  "isEnable":1,
	  "cmds":{
	    "56":{"id":56,"name":"Puissance","type":"info","subType":"numeric","unite":"W","isVisible":1,"currentValue":"123,4"},
	    "57":{"id":57,"name":"Temp\u00e9rature","type":"info","subType":"numeric","unite":"\u00b0C","isVisible":1,"value":18.6}
	  }
	}`))

	restarted, err := LoadStore(t.Context(), path, "keep_last", nil)
	if err != nil {
		t.Fatal(err)
	}
	device, ok := restarted.Device("server_power")
	if !ok {
		t.Fatal("missing persisted Jeedom device after restart")
	}
	if got := device.Values["temperature_c"]; got != 18.6 {
		t.Fatalf("temperature_c = %#v, want 18.6", got)
	}

	mqtt := &recordingMQTT{}
	publisher := NewPublisher(PublisherConfig{
		StateTopicPrefix: "ajaxbridge/jeedom",
		Discovery:        true,
		DiscoveryPrefix:  "homeassistant",
		DiscoveryNode:    "ajaxbridge",
		RetainState:      true,
		RetainDiscovery:  true,
	}, mqtt)
	if err := publisher.PublishDevice(t.Context(), device); err != nil {
		t.Fatal(err)
	}
	if got := mqtt.discovery["homeassistant/sensor/ajaxbridge/jeedom_cmd_57/config"]; got == "" {
		t.Fatalf("temperature discovery missing after restart: %#v", mqtt.discovery)
	}
	if got := mqtt.state["ajaxbridge/jeedom/devices/server_power/state"]; !strings.Contains(got, `"temperature_c":18.6`) {
		t.Fatalf("persisted Jeedom state = %q, want temperature_c", got)
	}
}

func TestLoadStoreAcceptsMissingFileAsEmptyCache(t *testing.T) {
	store, err := LoadStore(t.Context(), t.TempDir()+"/missing.json", "keep_last", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(store.Devices()); got != 0 {
		t.Fatalf("devices = %d, want empty cache", got)
	}
}

func TestSaveStoreRestoresActionMetadata(t *testing.T) {
	path := t.TempDir() + "/jeedom.json"
	store, err := LoadStore(t.Context(), path, "keep_last", nil)
	if err != nil {
		t.Fatal(err)
	}
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/10", []byte(relayDiscoveryPayload), time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	store.ApplyDiscovery(discovery)
	if err := store.Save(t.Context()); err != nil {
		t.Fatal(err)
	}

	restarted, err := LoadStore(t.Context(), path, "keep_last", nil)
	if err != nil {
		t.Fatal(err)
	}
	action, ok := restarted.Action("garage_gate", "on")
	if !ok {
		t.Fatal("missing persisted on action")
	}
	if action.CommandID != "85" || action.StateCommandID != "81" {
		t.Fatalf("action = %#v, want command/state ids", action)
	}
}
