package app

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/jeedom"
	"github.com/rs/zerolog"
)

func TestControlObserverPublishesUpdatedWallSwitchState(t *testing.T) {
	discovery, err := jeedom.ParseDiscoveryMessage("jeedom/discovery/eqLogic/26", []byte(`{
	  "id":26,
	  "name":"Grid load",
	  "configuration":{"device":"WallSwitch"},
	  "isVisible":1,
	  "isEnable":1,
	  "cmds":{
	    "348":{"id":348,"logicalId":"SWITCH_ON","name":"On","type":"action","subType":"other"},
	    "349":{"id":349,"logicalId":"SWITCH_OFF","name":"Off","type":"action","subType":"other"}
	  }
	}`), time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	store := jeedom.NewStore("keep_last")
	store.ApplyDiscovery(discovery)
	action, ok := store.Action("grid_load", "on")
	if !ok {
		t.Fatal("missing WallSwitch ON action")
	}
	if _, updated := store.RecordOptimisticControlState(action, time.Unix(101, 0)); !updated {
		t.Fatal("WallSwitch state was not recorded")
	}

	mqtt := &recordingJeedomMQTT{}
	application := &App{
		log:       zerolog.Nop(),
		jeedom:    store,
		jeedomPub: jeedom.NewPublisher(jeedom.PublisherConfig{StateTopicPrefix: "ajaxbridge/jeedom", RetainState: true}, mqtt),
	}
	observer := notificationObserver{app: application}
	observer.ObserveJeedomControl(t.Context(), jeedom.ControlResult{
		DeviceSlug:   "grid_load",
		Action:       "on",
		Published:    true,
		StateUpdated: true,
	}, nil)

	body := mqtt.state["ajaxbridge/jeedom/devices/grid_load/state"]
	if body == "" {
		t.Fatal("updated WallSwitch MQTT state was not published")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["state"] != true {
		t.Fatalf("published state = %#v, want true", payload["state"])
	}
}

type recordingJeedomMQTT struct {
	state map[string]string
}

func (m *recordingJeedomMQTT) PublishStateMessage(_ context.Context, topic string, payload []byte, _ bool) error {
	if m.state == nil {
		m.state = make(map[string]string)
	}
	m.state[topic] = string(payload)
	return nil
}

func (*recordingJeedomMQTT) PublishDiscoveryMessage(context.Context, string, string, []byte, bool) error {
	return nil
}

func (*recordingJeedomMQTT) AvailabilityTopic() string {
	return ""
}
