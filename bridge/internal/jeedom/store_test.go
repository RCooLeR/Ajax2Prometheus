package jeedom

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/devicecatalog"
)

func TestStoreEmptyValueKeepsLastNumericValue(t *testing.T) {
	store := NewStore("keep_last")
	now := time.Unix(100, 0)

	first := Event{
		Topic:       "jeedom/cmd/event/56",
		CommandID:   "56",
		ObjectName:  "None",
		DeviceName:  "Серверна",
		CommandName: "Puissance",
		Type:        "info",
		Subtype:     "numeric",
		Value:       json.RawMessage(`123.4`),
		ReceivedAt:  now,
	}
	store.Apply(first)

	empty := first
	empty.Value = json.RawMessage(`""`)
	empty.ReceivedAt = now.Add(time.Second)
	result := store.Apply(empty)

	if !result.EmptyValue {
		t.Fatal("expected empty value result")
	}
	device, ok := store.Device("serverna")
	if !ok {
		t.Fatal("missing device state")
	}
	if got := device.Values["power_w"]; got != 123.4 {
		t.Fatalf("power_w = %#v, want 123.4", got)
	}
}

func TestStoreKeepsEnglishCommandNameAndRawFrenchName(t *testing.T) {
	store := NewStore("keep_last")
	result := store.Apply(Event{
		Topic:       "jeedom/cmd/event/12",
		CommandID:   "12",
		DeviceName:  "Hub",
		CommandName: "Alimentation secteur",
		Subtype:     "binary",
		Value:       json.RawMessage(`1`),
		ReceivedAt:  time.Unix(100, 0),
	})

	if result.Command.Name != "External power" {
		t.Fatalf("command name = %q, want External power", result.Command.Name)
	}
	if result.Command.RawName != "Alimentation secteur" {
		t.Fatalf("raw name = %q, want original French", result.Command.RawName)
	}
}

func TestStoreDisambiguatesDuplicateDeviceNamesByCommandGroup(t *testing.T) {
	store := NewStore("keep_last")
	now := time.Unix(100, 0)

	events := []Event{
		{Topic: "jeedom/cmd/event/60", CommandID: "60", DeviceName: "Дим", CommandName: "Etat", Value: json.RawMessage(`"PASSIVE"`), ReceivedAt: now},
		{Topic: "jeedom/cmd/event/64", CommandID: "64", DeviceName: "Дим", CommandName: "Température", Value: json.RawMessage(`15`), ReceivedAt: now},
		{Topic: "jeedom/cmd/event/96", CommandID: "96", DeviceName: "Дим", CommandName: "Etat", Value: json.RawMessage(`"PASSIVE"`), ReceivedAt: now},
		{Topic: "jeedom/cmd/event/100", CommandID: "100", DeviceName: "Дим", CommandName: "Température", Value: json.RawMessage(`24`), ReceivedAt: now},
	}
	for _, evt := range events {
		store.Apply(evt)
	}

	devices := store.Devices()
	if len(devices) != 2 {
		t.Fatalf("devices length = %d, want 2: %#v", len(devices), devices)
	}
	if _, ok := store.Device("dym"); !ok {
		t.Fatal("missing first duplicate group slug dym")
	}
	if _, ok := store.Device("dym_96"); !ok {
		t.Fatal("missing second duplicate group slug dym_96")
	}
}

func TestStoreTracksLegacyUnlinkedSlugWhenCatalogLinksDevice(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "8",
		Name:             "Server power",
		Kind:             "WallSwitch",
		JeedomNames:      []string{"Serverna"},
		JeedomCommandIDs: []string{"56"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	result := store.Apply(Event{
		Topic:       "jeedom/cmd/event/56",
		CommandID:   "56",
		DeviceName:  "Serverna",
		CommandName: "Puissance",
		Type:        "info",
		Subtype:     "numeric",
		Value:       json.RawMessage(`100`),
		ReceivedAt:  time.Unix(100, 0),
	})

	if result.Device.DeviceSlug != "sia_a0f80d_zone_8" {
		t.Fatalf("DeviceSlug = %q, want linked SIA slug", result.Device.DeviceSlug)
	}
	if !containsString(result.Device.LegacyDeviceSlugs, "serverna") {
		t.Fatalf("LegacyDeviceSlugs = %#v, want serverna", result.Device.LegacyDeviceSlugs)
	}
}

func TestStoreNormalizesRelayVoltageFromCatalog(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "6",
		Name:             "Garage gate relay",
		Kind:             "Relay",
		JeedomCommandIDs: []string{"230"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	result := store.Apply(Event{
		Topic:       "jeedom/cmd/event/230",
		CommandID:   "230",
		DeviceName:  "Garage gate",
		CommandName: "Voltage",
		Type:        "info",
		Subtype:     "numeric",
		Value:       json.RawMessage(`289.02`),
		ReceivedAt:  time.Unix(100, 0),
	})

	got, ok := result.Device.Values["voltage_v"].(float64)
	if !ok {
		t.Fatalf("voltage_v = %#v, want float64", result.Device.Values["voltage_v"])
	}
	if math.Abs(got-28.902) > 1e-9 {
		t.Fatalf("voltage_v = %#v, want 28.902", got)
	}
	if math.Abs(result.NumericValue-28.902) > 1e-9 {
		t.Fatalf("NumericValue = %#v, want 28.902", result.NumericValue)
	}
}

func TestStoreKeepsWallSwitchVoltageUnscaled(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "8",
		Name:             "Server power",
		Kind:             "WallSwitch",
		JeedomCommandIDs: []string{"203"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	result := store.Apply(Event{
		Topic:       "jeedom/cmd/event/203",
		CommandID:   "203",
		DeviceName:  "Server power",
		CommandName: "Voltage",
		Type:        "info",
		Subtype:     "numeric",
		Value:       json.RawMessage(`238`),
		ReceivedAt:  time.Unix(100, 0),
	})

	if got := result.Device.Values["voltage_v"]; got != 238.0 {
		t.Fatalf("voltage_v = %#v, want 238", got)
	}
}

func TestStoreDerivesWallSwitchStateFromEventCode(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "8",
		Name:             "Server power",
		Kind:             "WallSwitch",
		JeedomCommandIDs: []string{"204"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	on := store.Apply(Event{
		Topic:       "jeedom/cmd/event/204",
		CommandID:   "204",
		DeviceName:  "Server power",
		CommandName: "Code evenement",
		Type:        "info",
		Subtype:     "string",
		Value:       json.RawMessage(`"M_1F_37"`),
		ReceivedAt:  time.Unix(100, 0),
	})
	if got := on.Device.Values["state"]; got != true {
		t.Fatalf("state after M_1F_37 = %#v, want true", got)
	}
	if got := on.Device.Values["event_code"]; got != "M_1F_37" {
		t.Fatalf("event_code = %#v, want M_1F_37", got)
	}

	off := store.Apply(Event{
		Topic:       "jeedom/cmd/event/204",
		CommandID:   "204",
		DeviceName:  "Server power",
		CommandName: "Code evenement",
		Type:        "info",
		Subtype:     "string",
		Value:       json.RawMessage(`"M_1F_46"`),
		ReceivedAt:  time.Unix(101, 0),
	})
	if got := off.Device.Values["state"]; got != false {
		t.Fatalf("state after M_1F_46 = %#v, want false", got)
	}
}

func TestStoreDoesNotDeriveRelayStateFromWallSwitchEventCode(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "6",
		Name:             "Garage relay",
		Kind:             "Relay",
		JeedomCommandIDs: []string{"205"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	result := store.Apply(Event{
		Topic:       "jeedom/cmd/event/205",
		CommandID:   "205",
		DeviceName:  "Garage relay",
		CommandName: "Code evenement",
		Type:        "info",
		Subtype:     "string",
		Value:       json.RawMessage(`"M_1F_37"`),
		ReceivedAt:  time.Unix(100, 0),
	})
	if _, ok := result.Device.Values["state"]; ok {
		t.Fatalf("relay state = %#v, want no derived state", result.Device.Values["state"])
	}
}

func TestStoreDerivesWaterStopStateFromEventCode(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "12",
		Name:             "Water valve",
		Kind:             "WaterStop",
		JeedomCommandIDs: []string{"206"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	open := store.Apply(Event{
		Topic:       "jeedom/cmd/event/206",
		CommandID:   "206",
		DeviceName:  "Water valve",
		CommandName: "Code evenement",
		Type:        "info",
		Subtype:     "string",
		Value:       json.RawMessage(`"M_48_37"`),
		ReceivedAt:  time.Unix(100, 0),
	})
	if got := open.Device.Values["state"]; got != true {
		t.Fatalf("state after M_48_37 = %#v, want true", got)
	}
	if got := open.Device.Values["event_code"]; got != "M_48_37" {
		t.Fatalf("event_code = %#v, want M_48_37", got)
	}

	closed := store.Apply(Event{
		Topic:       "jeedom/cmd/event/206",
		CommandID:   "206",
		DeviceName:  "Water valve",
		CommandName: "Code evenement",
		Type:        "info",
		Subtype:     "string",
		Value:       json.RawMessage(`"M_48_46"`),
		ReceivedAt:  time.Unix(101, 0),
	})
	if got := closed.Device.Values["state"]; got != false {
		t.Fatalf("state after M_48_46 = %#v, want false", got)
	}
}

func TestStoreDoesNotDeriveWaterStopStateFromCommonEvent(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "12",
		Name:             "Water valve",
		Kind:             "WaterStop",
		JeedomCommandIDs: []string{"207"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	result := store.Apply(Event{
		Topic:       "jeedom/cmd/event/207",
		CommandID:   "207",
		DeviceName:  "Water valve",
		CommandName: "Evenement",
		Type:        "info",
		Subtype:     "string",
		Value:       json.RawMessage(`"COMMON"`),
		ReceivedAt:  time.Unix(100, 0),
	})
	if _, ok := result.Device.Values["state"]; ok {
		t.Fatalf("state after COMMON event = %#v, want no derived state", result.Device.Values["state"])
	}
}

func TestStoreDerivesTransmitterGridPowerFromEventCode(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "14",
		Name:             "Grid detector",
		Kind:             "Transmitter",
		JeedomCommandIDs: []string{"208"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	result := store.Apply(Event{
		Topic:       "jeedom/cmd/event/208",
		CommandID:   "208",
		DeviceName:  "Grid detector",
		CommandName: "Code evenement",
		Type:        "info",
		Subtype:     "string",
		Value:       json.RawMessage(`"M_11_40"`),
		ReceivedAt:  time.Unix(100, 0),
	})
	if got := result.Device.Values["grid_power"]; got != true {
		t.Fatalf("grid_power after M_11_40 = %#v, want true", got)
	}
	if _, ok := result.Device.Values["state"]; ok {
		t.Fatalf("state after M_11_40 = %#v, want no switch state", result.Device.Values["state"])
	}
	command := result.Device.RawCommands["grid_power"]
	if command.Metric != "grid_power" || command.Component != ComponentBinarySensor || command.DeviceClass != "power" {
		t.Fatalf("synthetic grid power command = %#v", command)
	}
	if command.Value != true {
		t.Fatalf("synthetic grid power command value = %#v, want true", command.Value)
	}
}

func TestStoreDoesNotDeriveMultiTransmitterGridPowerFromTransmitterEventCode(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "15",
		Name:             "Multi input",
		Kind:             "MultiTransmitter",
		JeedomCommandIDs: []string{"209"},
	})
	store := NewStoreWithResolver("keep_last", NewCatalogResolver(catalog, CatalogResolverConfig{}))

	result := store.Apply(Event{
		Topic:       "jeedom/cmd/event/209",
		CommandID:   "209",
		DeviceName:  "Multi input",
		CommandName: "Code evenement",
		Type:        "info",
		Subtype:     "string",
		Value:       json.RawMessage(`"M_11_40"`),
		ReceivedAt:  time.Unix(100, 0),
	})
	if _, ok := result.Device.Values["grid_power"]; ok {
		t.Fatalf("multitransmitter grid_power = %#v, want no derived value", result.Device.Values["grid_power"])
	}
}
