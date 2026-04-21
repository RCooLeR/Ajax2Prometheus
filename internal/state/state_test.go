package state

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/RCooLeR/Ajax2Prometheus/internal/devicecatalog"
	"github.com/RCooLeR/Ajax2Prometheus/internal/event"
)

func TestNewEngineSeedsCatalogDevicesAsInactiveZones(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices.json")
	data := []byte(`[{
  "account": "0001",
  "zone": "7",
  "partition": "1",
  "group": "garage",
  "device": "ri1",
  "name": "Garage door",
  "room": "Garage",
  "kind": "doorprotect",
  "events": ["burglary", "tamper", "battery", "connectivity"]
}]`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := devicecatalog.Load(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(time.Minute, catalog)
	snapshot := engine.Snapshot()

	if len(snapshot.Accounts) != 0 {
		t.Fatalf("accounts length = %d, want 0 before first event", len(snapshot.Accounts))
	}
	if len(snapshot.Zones) != 1 {
		t.Fatalf("zones length = %d, want 1", len(snapshot.Zones))
	}
	zone := snapshot.Zones[0]
	if zone.Account != "0001" || zone.Zone != "7" || zone.Partition != "1" || zone.Group != "garage" || zone.Device != "ri1" {
		t.Fatalf("unexpected zone identity: %#v", zone)
	}
	if zone.DeviceName != "Garage door" || zone.Room != "Garage" || zone.Kind != "doorprotect" {
		t.Fatalf("unexpected zone labels: %#v", zone)
	}
	if zone.DeviceEventsLabel != "burglary,tamper,battery,connectivity" {
		t.Fatalf("device events label = %q", zone.DeviceEventsLabel)
	}
	for _, signal := range []string{"burglary", "tamper", "battery", "connectivity"} {
		if zone.SignalActive[signal] {
			t.Fatalf("catalog-only signal %q should be inactive: %#v", signal, zone.SignalActive)
		}
	}
	if zone.AlarmActive || zone.TamperActive || zone.TroubleActive || !zone.LastEventAt.IsZero() {
		t.Fatalf("catalog-only zone should be inactive with no event timestamp: %#v", zone)
	}
}

func TestNewEngineSeedsCatalogLastSeenAndLastEventIntoZones(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices.json")
	data := []byte(`[{
  "account": "0001",
  "zone": "7",
  "name": "Garage fire sensor",
  "room": "Garage",
  "kind": "fireprotect",
  "events": ["fire", "smoke", "battery"],
  "last_seen_at": "2026-04-20T12:34:56Z",
  "last_event_code": "TU",
  "last_event_name": "Tamper bypass restored or reactivated",
  "last_signal": "tamper_bypass"
}]`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := devicecatalog.Load(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(time.Minute, catalog)
	snapshot := engine.Snapshot()

	if len(snapshot.Zones) != 1 {
		t.Fatalf("zones length = %d, want 1", len(snapshot.Zones))
	}
	zone := snapshot.Zones[0]
	wantTime := time.Date(2026, 4, 20, 12, 34, 56, 0, time.UTC)
	if !zone.LastEventAt.Equal(wantTime) {
		t.Fatalf("last event at = %s, want %s", zone.LastEventAt, wantTime)
	}
	if zone.LastEventCode != "TU" || zone.LastEventName != "Tamper bypass restored or reactivated" || zone.LastSignal != "tamper_bypass" {
		t.Fatalf("unexpected seeded catalog state: %#v", zone)
	}
}

func TestZoneSignalActiveTracksAlarmTroubleAndRestore(t *testing.T) {
	engine := NewEngine(time.Minute, devicecatalog.Empty())

	engine.Apply(event.Normalized{
		Account:     "0001",
		Zone:        "7",
		EventCode:   "BA",
		EventClass:  event.ClassAlarm,
		EventAction: "burglary_alarm",
		EventName:   "Burglary alarm",
		Signal:      "burglary",
		ReceivedAt:  time.Unix(100, 0),
		ParseStatus: event.ParseStatusOK,
	})
	engine.Apply(event.Normalized{
		Account:     "0001",
		Zone:        "7",
		EventCode:   "XT",
		EventClass:  event.ClassTrouble,
		EventAction: "battery_low",
		EventName:   "Device battery low",
		Signal:      "battery",
		ReceivedAt:  time.Unix(200, 0),
		ParseStatus: event.ParseStatusOK,
	})
	snapshot := engine.Snapshot()
	zone := snapshot.Zones[0]
	if !zone.SignalActive["burglary"] || !zone.SignalActive["battery"] {
		t.Fatalf("expected burglary and battery signals active: %#v", zone.SignalActive)
	}

	engine.Apply(event.Normalized{
		Account:     "0001",
		Zone:        "7",
		EventCode:   "BR",
		EventClass:  event.ClassRestore,
		EventAction: "burglary_restore",
		EventName:   "Burglary alarm restored",
		Signal:      "burglary",
		ReceivedAt:  time.Unix(300, 0),
		ParseStatus: event.ParseStatusOK,
	})
	snapshot = engine.Snapshot()
	zone = snapshot.Zones[0]
	if zone.SignalActive["burglary"] {
		t.Fatalf("expected burglary signal restored: %#v", zone.SignalActive)
	}
	if !zone.SignalActive["battery"] {
		t.Fatalf("battery trouble should remain active: %#v", zone.SignalActive)
	}
}

func TestBypassRestoreClearsTrouble(t *testing.T) {
	for _, tc := range []struct {
		name        string
		restoreCode string
		restoreName string
		restoreSig  string
	}{
		{name: "QU clears QB", restoreCode: "QU", restoreName: "Device bypass restored or reactivated", restoreSig: "bypass"},
		{name: "TU clears QB", restoreCode: "TU", restoreName: "Tamper bypass restored or reactivated", restoreSig: "tamper_bypass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewEngine(time.Minute, devicecatalog.Empty())

			engine.Apply(event.Normalized{
				Account:     "0001",
				EventCode:   "QB",
				EventClass:  event.ClassTrouble,
				EventAction: "device_bypassed",
				EventName:   "Device bypassed or deactivated",
				Signal:      "bypass",
				Zone:        "7",
				ReceivedAt:  time.Unix(100, 0),
				ParseStatus: event.ParseStatusOK,
			})
			snapshot := engine.Snapshot()
			if !snapshot.Accounts[0].TroubleActive {
				t.Fatal("expected account trouble active after QB")
			}
			if !snapshot.Zones[0].TroubleActive {
				t.Fatal("expected zone trouble active after QB")
			}

			engine.Apply(event.Normalized{
				Account:     "0001",
				EventCode:   tc.restoreCode,
				EventClass:  event.ClassRestore,
				EventAction: "bypass_restore",
				EventName:   tc.restoreName,
				Signal:      tc.restoreSig,
				Zone:        "7",
				ReceivedAt:  time.Unix(200, 0),
				ParseStatus: event.ParseStatusOK,
			})
			snapshot = engine.Snapshot()
			if snapshot.Accounts[0].TroubleActive {
				t.Fatalf("expected account trouble cleared after %s", tc.restoreCode)
			}
			if snapshot.Zones[0].TroubleActive {
				t.Fatalf("expected zone trouble cleared after %s", tc.restoreCode)
			}
		})
	}
}

func TestTamperBypassRestoreClearsTrouble(t *testing.T) {
	for _, tc := range []struct {
		name        string
		restoreCode string
		restoreName string
		restoreSig  string
	}{
		{name: "QU clears TB", restoreCode: "QU", restoreName: "Device bypass restored or reactivated", restoreSig: "bypass"},
		{name: "TU clears TB", restoreCode: "TU", restoreName: "Tamper bypass restored or reactivated", restoreSig: "tamper_bypass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewEngine(time.Minute, devicecatalog.Empty())

			engine.Apply(event.Normalized{
				Account:     "0001",
				EventCode:   "TB",
				EventClass:  event.ClassTrouble,
				EventAction: "tamper_bypassed",
				EventName:   "Tamper bypassed or deactivated",
				Signal:      "tamper_bypass",
				Zone:        "7",
				ReceivedAt:  time.Unix(100, 0),
				ParseStatus: event.ParseStatusOK,
			})
			snapshot := engine.Snapshot()
			if !snapshot.Accounts[0].TroubleActive {
				t.Fatal("expected account trouble active after TB")
			}
			if !snapshot.Zones[0].TroubleActive {
				t.Fatal("expected zone trouble active after TB")
			}

			engine.Apply(event.Normalized{
				Account:     "0001",
				EventCode:   tc.restoreCode,
				EventClass:  event.ClassRestore,
				EventAction: "bypass_restore",
				EventName:   tc.restoreName,
				Signal:      tc.restoreSig,
				Zone:        "7",
				ReceivedAt:  time.Unix(200, 0),
				ParseStatus: event.ParseStatusOK,
			})
			snapshot = engine.Snapshot()
			if snapshot.Accounts[0].TroubleActive {
				t.Fatalf("expected account trouble cleared after %s", tc.restoreCode)
			}
			if snapshot.Zones[0].TroubleActive {
				t.Fatalf("expected zone trouble cleared after %s", tc.restoreCode)
			}
		})
	}
}
