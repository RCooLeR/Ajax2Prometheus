package jeedom

import (
	"errors"
	"testing"
	"time"
)

func TestParseMessageHumanName(t *testing.T) {
	body := []byte(`{"value":"123.4","humanName":"[None][РЎРµСЂРІРµСЂРЅР°][Puissance]","unite":"","name":"Puissance","type":"info","subtype":"numeric"}`)

	evt, err := ParseMessage("jeedom/cmd/event/56", body, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	if evt.ObjectName != "None" {
		t.Fatalf("ObjectName = %q, want None", evt.ObjectName)
	}
	if evt.DeviceName != "Серверна" {
		t.Fatalf("DeviceName = %q, want repaired Cyrillic name", evt.DeviceName)
	}
	if evt.CommandName != "Puissance" {
		t.Fatalf("CommandName = %q, want Puissance", evt.CommandName)
	}
	if got := Slug(evt.DeviceName); got != "serverna" {
		t.Fatalf("Slug(DeviceName) = %q, want serverna", got)
	}
}

func TestParseHumanNameMalformed(t *testing.T) {
	_, _, _, err := ParseHumanName("[None][OnlyDevice]")
	if !errors.Is(err, ErrMalformedHumanName) {
		t.Fatalf("error = %v, want ErrMalformedHumanName", err)
	}
}

func TestCommandIDFromTopic(t *testing.T) {
	got, err := CommandIDFromTopic("jeedom/cmd/event/56")
	if err != nil {
		t.Fatal(err)
	}
	if got != "56" {
		t.Fatalf("CommandIDFromTopic = %q, want 56", got)
	}
}

func TestIsCommandEventTopic(t *testing.T) {
	tests := []struct {
		topic string
		want  bool
	}{
		{topic: "jeedom/cmd/event/56", want: true},
		{topic: "prefix/jeedom/cmd/event/56", want: true},
		{topic: "jeedom/state", want: false},
		{topic: "jeedom/cmd/action/56", want: false},
		{topic: "jeedom/cmd/event/#", want: false},
	}
	for _, tt := range tests {
		if got := IsCommandEventTopic(tt.topic); got != tt.want {
			t.Fatalf("IsCommandEventTopic(%q) = %t, want %t", tt.topic, got, tt.want)
		}
	}
}

func TestMappingPowerAndCurrent(t *testing.T) {
	power := MappingFor(Event{CommandName: "Puissance", Type: "info", Subtype: "numeric"})
	if power.Metric != "power_w" || power.Unit != "W" || power.DeviceClass != "power" || power.StateClass != "measurement" {
		t.Fatalf("power mapping = %#v", power)
	}

	current := MappingFor(Event{CommandName: "Courant", Type: "info", Subtype: "numeric"})
	if current.Metric != "current_a" || current.Unit != "A" || current.DeviceClass != "current" || current.StateClass != "measurement" {
		t.Fatalf("current mapping = %#v", current)
	}
}

func TestMappingProductionJeedomCommands(t *testing.T) {
	tests := []struct {
		name      string
		subtype   string
		metric    string
		component string
		numeric   bool
		binary    bool
	}{
		{name: "Consommation", subtype: "numeric", metric: "energy_kwh", component: ComponentSensor, numeric: true},
		{name: "Etat", subtype: "binary", metric: "state", component: ComponentBinarySensor, binary: true},
		{name: "En ligne", subtype: "binary", metric: "online", component: ComponentBinarySensor, binary: true},
		{name: "Signal", subtype: "string", metric: "signal_level", component: ComponentSensor},
		{name: "Alimentation secteur", subtype: "binary", metric: "external_power", component: ComponentBinarySensor, binary: true},
	}
	for _, tt := range tests {
		mapping := MappingFor(Event{CommandName: tt.name, Subtype: tt.subtype})
		if mapping.Metric != tt.metric || mapping.Component != tt.component || mapping.Numeric != tt.numeric || mapping.Binary != tt.binary {
			t.Fatalf("MappingFor(%q/%q) = %#v", tt.name, tt.subtype, mapping)
		}
	}
}

func TestEnglishCommandNameTranslatesFrenchMetadata(t *testing.T) {
	mapping := MappingFor(Event{CommandName: "Alimentation secteur", Subtype: "binary"})
	if got := EnglishCommandName("Alimentation secteur", mapping, "12"); got != "External power" {
		t.Fatalf("EnglishCommandName = %q, want External power", got)
	}
	if got := EnglishActionName("night_mode", "Mode nuit"); got != "Night mode" {
		t.Fatalf("EnglishActionName = %q, want Night mode", got)
	}
}
