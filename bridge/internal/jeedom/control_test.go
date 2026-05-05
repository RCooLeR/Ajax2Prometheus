package jeedom

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestControllerPublishesJeedomSetCommand(t *testing.T) {
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/10", []byte(relayDiscoveryPayload), time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore("keep_last")
	store.ApplyDiscovery(discovery)
	mqtt := &fakeCommandPublisher{}
	controller := NewController(ControllerConfig{
		Enabled:              true,
		StateTopicPrefix:     "ajaxbridge/jeedom",
		JeedomSetTopicPrefix: "jeedom/cmd/set",
	}, store, mqtt, zerolog.Nop())

	result, err := controller.Execute(context.Background(), "garage_gate", "ON", "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Published || result.CommandID != "85" {
		t.Fatalf("result = %#v", result)
	}
	if mqtt.topic != "jeedom/cmd/set/85" {
		t.Fatalf("published topic = %q, want jeedom/cmd/set/85", mqtt.topic)
	}
	audits := store.ControlAudit(1)
	if len(audits) != 1 || audits[0].Result != "published" {
		t.Fatalf("audit = %#v", audits)
	}
}

func TestControllerRejectsDeniedAction(t *testing.T) {
	payload := []byte(`{"id":3,"name":"Valve","configuration":{"device":"WaterStop"},"isVisible":1,"isEnable":1,"cmds":{"26":{"id":26,"logicalId":"SWITCH_ON","name":"On","type":"action","subType":"other","isVisible":1}}}`)
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/3", payload, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore("keep_last")
	store.ApplyDiscovery(discovery)
	controller := NewController(ControllerConfig{Enabled: true}, store, &fakeCommandPublisher{}, zerolog.Nop())

	if _, err := controller.Execute(context.Background(), "valve", "on", "test"); err == nil {
		t.Fatal("expected denied action error")
	}
}

type fakeCommandPublisher struct {
	topic   string
	payload string
}

func (f *fakeCommandPublisher) PublishCommandMessage(_ context.Context, topic string, payload []byte) error {
	f.topic = topic
	f.payload = string(payload)
	return nil
}
