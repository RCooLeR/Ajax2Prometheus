package jeedom

import (
	"math"
	"testing"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/devicecatalog"
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

func TestStoreApplyDiscoveryRegistersRelayImpulseAction(t *testing.T) {
	payload := []byte(`{
	  "id":11,
	  "name":"Garage pulse",
	  "configuration":{"device":"Relay","applyDevice":"Relay"},
	  "isVisible":1,
	  "isEnable":1,
	  "cmds":{
	    "90":{"id":90,"logicalId":"IMPULSE","name":"Impulsion","type":"action","subType":"other","isVisible":1}
	  }
	}`)
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/11", payload, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	result := NewStore("keep_last").ApplyDiscovery(discovery)
	impulse := result.Device.Actions["impulse"]
	if impulse.CommandID != "90" || !impulse.Allowed {
		t.Fatalf("impulse action = %#v", impulse)
	}
	if impulse.Name != "Impulse" {
		t.Fatalf("impulse name = %q, want Impulse", impulse.Name)
	}
}

func TestStoreApplyDiscoverySeedsCurrentInfoValues(t *testing.T) {
	payload := []byte(`{
	  "id":8,
	  "name":"Server power",
	  "configuration":{"device":"WallSwitch"},
	  "isVisible":1,
	  "isEnable":1,
	  "cmds":{
	    "56":{"id":56,"name":"Puissance","type":"info","subType":"numeric","unite":"W","isVisible":1,"currentValue":"123,4"},
	    "57":{"id":57,"name":"Temp\u00e9rature","type":"info","subType":"numeric","unite":"\u00b0C","isVisible":1,"value":18.6}
	  }
	}`)
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/8", payload, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	result := NewStore("keep_last").ApplyDiscovery(discovery)

	if got := result.Device.Values["power_w"]; got != 123.4 {
		t.Fatalf("power_w = %#v, want 123.4", got)
	}
	if got := result.Device.Values["temperature_c"]; got != 18.6 {
		t.Fatalf("temperature_c = %#v, want 18.6", got)
	}
	if result.Device.RawCommands["56"].LastValueAt.IsZero() {
		t.Fatalf("power command LastValueAt was not seeded")
	}
}

func TestStoreApplyDiscoveryNormalizesRelayVoltageSeed(t *testing.T) {
	payload := []byte(`{
	  "id":6,
	  "name":"Garage gate",
	  "configuration":{"device":"Relay"},
	  "isVisible":1,
	  "isEnable":1,
	  "cmds":{
	    "230":{"id":230,"name":"Voltage","type":"info","subType":"numeric","unite":"V","isVisible":1,"currentValue":"289.02"}
	  }
	}`)
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/6", payload, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	result := NewStore("keep_last").ApplyDiscovery(discovery)

	got, ok := result.Device.Values["voltage_v"].(float64)
	if !ok {
		t.Fatalf("voltage_v = %#v, want float64", result.Device.Values["voltage_v"])
	}
	if math.Abs(got-28.902) > 1e-9 {
		t.Fatalf("voltage_v = %#v, want 28.902", got)
	}
}

func TestStoreApplyDiscoveryAllowsWaterStopToggleActions(t *testing.T) {
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
	on := result.Device.Actions["on"]
	off := result.Device.Actions["off"]
	if !on.Allowed || !off.Allowed {
		t.Fatalf("WaterStop actions should be allowed: on=%#v off=%#v", on, off)
	}
}

func TestStoreApplyDiscoveryAllowsCatalogWaterStopToggleWithoutJeedomDeviceType(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "12",
		Name:             "Water valve",
		Kind:             "WaterStop",
		JeedomCommandIDs: []string{"174", "175"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))
	payload := []byte(`{
	  "id":30,
	  "name":"Water valve",
	  "isVisible":1,
	  "isEnable":1,
	  "cmds":{
	    "174":{"id":174,"logicalId":"SWITCH_ON","name":"On","type":"action","subType":"other","isVisible":1},
	    "175":{"id":175,"logicalId":"SWITCH_OFF","name":"Off","type":"action","subType":"other","isVisible":1}
	  }
	}`)
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/30", payload, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	result := store.ApplyDiscovery(discovery)

	if result.Device.JeedomDeviceType != "WaterStop" {
		t.Fatalf("JeedomDeviceType = %q, want WaterStop", result.Device.JeedomDeviceType)
	}
	on := result.Device.Actions["on"]
	off := result.Device.Actions["off"]
	if on.CommandID != "174" || !on.Allowed {
		t.Fatalf("catalog WaterStop on action = %#v", on)
	}
	if off.CommandID != "175" || !off.Allowed {
		t.Fatalf("catalog WaterStop off action = %#v", off)
	}
}

func TestStoreApplyDiscoveryAllowsHubSecurityActions(t *testing.T) {
	payload := []byte(`{
	  "id":40,
	  "name":"Security hub",
	  "configuration":{"device":"Hub"},
	  "isVisible":1,
	  "isEnable":1,
	  "cmds":{
	    "163":{"id":163,"logicalId":"ARM","name":"Armement","type":"action","subType":"other","isVisible":1},
	    "164":{"id":164,"logicalId":"NIGHT_MODE","name":"Mode nuit","type":"action","subType":"other","isVisible":1},
	    "165":{"id":165,"logicalId":"DISARM","name":"Desarmement","type":"action","subType":"other","isVisible":1},
	    "167":{"id":167,"logicalId":"muteFireDetectors","name":"Arret detection incendie","type":"action","subType":"other","isVisible":1}
	  }
	}`)
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/40", payload, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	result := NewStore("keep_last").ApplyDiscovery(discovery)

	want := map[string]string{
		"arm":                 "163",
		"night_mode":          "164",
		"disarm":              "165",
		"mute_fire_detectors": "167",
	}
	for actionName, commandID := range want {
		action := result.Device.Actions[actionName]
		if action.CommandID != commandID || !action.Allowed {
			t.Fatalf("%s action = %#v, want command %s allowed", actionName, action, commandID)
		}
	}
}
