package hamqtt

import (
	"testing"

	"github.com/RCooLeR/Ajax2Prometheus/internal/state"
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
