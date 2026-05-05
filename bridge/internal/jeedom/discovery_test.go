package jeedom

import (
	"encoding/json"
	"testing"
	"time"
)

const relayDiscoveryPayload = `{
  "id": 10,
  "name": "Garage gate",
  "logicalId": "3092B925",
  "eqType_name": "ajaxSystem",
  "isVisible": 1,
  "isEnable": 1,
  "configuration": {"device":"Relay","applyDevice":"Relay","hub_id":"002BAD2F"},
  "cmds": {
    "81": {"id":81,"logicalId":"realState","name":"Etat","type":"info","subType":"binary","unite":"","eqLogic_id":10,"isVisible":1,"value":null},
    "85": {"id":85,"logicalId":"SWITCH_ON","name":"On","type":"action","subType":"other","eqLogic_id":10,"isVisible":1,"value":"81"},
    "86": {"id":86,"logicalId":"SWITCH_OFF","name":"Off","type":"action","subType":"other","eqLogic_id":10,"isVisible":1,"value":"81"}
  }
}`

func TestParseDiscoveryMessageExtractsActions(t *testing.T) {
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/10", []byte(relayDiscoveryPayload), time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	if discovery.EqLogicID != "10" {
		t.Fatalf("EqLogicID = %q, want 10", discovery.EqLogicID)
	}
	if discovery.DeviceType != "Relay" {
		t.Fatalf("DeviceType = %q, want Relay", discovery.DeviceType)
	}
	if discovery.InfoCommands["81"].Subtype != "binary" {
		t.Fatalf("state command = %#v", discovery.InfoCommands["81"])
	}
	if discovery.Actions["85"].StateCommandID != "81" {
		t.Fatalf("on state command = %q, want 81", discovery.Actions["85"].StateCommandID)
	}
}

func TestStoreApplyDiscoveryRegistersSafeActions(t *testing.T) {
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/10", []byte(relayDiscoveryPayload), time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore("keep_last")
	result := store.ApplyDiscovery(discovery)

	if result.Device.DeviceSlug != "garage_gate" {
		t.Fatalf("DeviceSlug = %q, want garage_gate", result.Device.DeviceSlug)
	}
	state := result.Device.RawCommands["81"]
	if state.Metric != "state" || state.Component != ComponentBinarySensor {
		t.Fatalf("state command mapping = %#v", state)
	}
	on := result.Device.Actions["on"]
	if on.CommandID != "85" || !on.Allowed || on.StateCommandID != "81" {
		t.Fatalf("on action = %#v", on)
	}
	if on.Name != "On" || on.RawName != "On" {
		t.Fatalf("translated action name = %#v", on)
	}
}

func TestStoreApplyDiscoveryBlocksWaterStop(t *testing.T) {
	payload := []byte(`{
	  "id":3,
	  "name":"Valve",
	  "configuration":{"device":"WaterStop"},
	  "isVisible":1,
	  "isEnable":1,
	  "cmds":{
	    "26":{"id":26,"logicalId":"SWITCH_ON","name":"On","type":"action","subType":"other","isVisible":1},
	    "27":{"id":27,"logicalId":"SWITCH_OFF","name":"Off","type":"action","subType":"other","isVisible":1}
	  }
	}`)
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/3", payload, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	result := NewStore("keep_last").ApplyDiscovery(discovery)
	if result.Device.Actions["on"].Allowed {
		body, _ := json.Marshal(result.Device.Actions["on"])
		t.Fatalf("WaterStop action unexpectedly allowed: %s", body)
	}
}
