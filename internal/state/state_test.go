package state

import (
	"testing"
	"time"

	"github.com/RCooLeR/Ajax2Prometheus/internal/devicecatalog"
	"github.com/RCooLeR/Ajax2Prometheus/internal/event"
)

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
