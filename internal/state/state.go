package state

import (
	"strings"
	"sync"
	"time"

	"github.com/RCooLeR/Ajax2Prometheus/internal/devicecatalog"
	"github.com/RCooLeR/Ajax2Prometheus/internal/event"
)

type Account struct {
	Account        string    `json:"account"`
	Online         bool      `json:"online"`
	Mode           string    `json:"mode"`
	Armed          bool      `json:"armed"`
	NightMode      bool      `json:"night_mode"`
	PartiallyArmed bool      `json:"partially_armed"`
	AlarmActive    bool      `json:"alarm_active"`
	TamperActive   bool      `json:"tamper_active"`
	TroubleActive  bool      `json:"trouble_active"`
	LastEventCode  string    `json:"last_event_code"`
	LastEventName  string    `json:"last_event_name"`
	LastSignal     string    `json:"last_signal"`
	LastEventAt    time.Time `json:"last_event_at"`
	LastPingAt     time.Time `json:"last_ping_at"`
}

type Zone struct {
	Account           string    `json:"account"`
	Partition         string    `json:"partition"`
	Group             string    `json:"group"`
	Zone              string    `json:"zone"`
	Device            string    `json:"device"`
	DeviceName        string    `json:"device_name"`
	Room              string    `json:"room"`
	Kind              string    `json:"kind"`
	DeviceEvents      []string  `json:"device_events"`
	DeviceEventsLabel string    `json:"device_events_label"`
	AlarmActive       bool      `json:"alarm_active"`
	AlarmSignal       string    `json:"alarm_signal"`
	AlarmAction       string    `json:"alarm_action"`
	AlarmEventCode    string    `json:"alarm_event_code"`
	AlarmEventName    string    `json:"alarm_event_name"`
	AlarmStartedAt    time.Time `json:"alarm_started_at"`
	TamperActive      bool      `json:"tamper_active"`
	TroubleActive     bool      `json:"trouble_active"`
	LastEventCode     string    `json:"last_event_code"`
	LastEventName     string    `json:"last_event_name"`
	LastSignal        string    `json:"last_signal"`
	LastEventAt       time.Time `json:"last_event_at"`
}

type Snapshot struct {
	Accounts []Account `json:"accounts"`
	Zones    []Zone    `json:"zones"`
}

type Engine struct {
	mu           sync.RWMutex
	offlineGrace time.Duration
	devices      *devicecatalog.Catalog
	accounts     map[string]*Account
	zones        map[string]*Zone
}

func NewEngine(offlineGrace time.Duration, devices *devicecatalog.Catalog) *Engine {
	return &Engine{
		offlineGrace: offlineGrace,
		devices:      devices,
		accounts:     make(map[string]*Account),
		zones:        make(map[string]*Zone),
	}
}

func (e *Engine) Apply(evt event.Normalized) Snapshot {
	e.mu.Lock()
	defer e.mu.Unlock()

	account := e.account(evt.Account)
	account.LastEventAt = maxTime(account.LastEventAt, evt.ReceivedAt)
	account.LastEventCode = evt.EventCode
	account.LastEventName = evt.EventName
	account.LastSignal = evt.Signal
	account.Online = true

	if evt.EventClass == event.ClassTest {
		account.LastPingAt = maxTime(account.LastPingAt, evt.ReceivedAt)
	}

	switch evt.EventClass {
	case event.ClassAlarm:
		account.AlarmActive = true
	case event.ClassRestore:
		applyAccountRestore(account, evt)
	case event.ClassArm:
		account.Armed = true
		account.NightMode = false
		account.PartiallyArmed = evt.EventCode == "CG"
	case event.ClassNight:
		account.Armed = true
		account.NightMode = true
		account.PartiallyArmed = true
	case event.ClassDisarm:
		if evt.Signal == "night_mode" {
			account.NightMode = false
			account.PartiallyArmed = false
		} else {
			account.Armed = false
			account.NightMode = false
			account.PartiallyArmed = false
			account.AlarmActive = false
		}
	case event.ClassTamper:
		account.TamperActive = true
	case event.ClassTrouble:
		account.TroubleActive = true
	}

	if evt.Zone != "" {
		zone := e.zone(evt.Account, evt.Zone)
		e.updateZoneMetadata(zone, evt)
		zone.LastEventAt = maxTime(zone.LastEventAt, evt.ReceivedAt)
		zone.LastEventCode = evt.EventCode
		zone.LastEventName = evt.EventName
		zone.LastSignal = evt.Signal
		switch evt.EventClass {
		case event.ClassAlarm:
			zone.AlarmActive = true
			zone.AlarmSignal = evt.Signal
			zone.AlarmAction = evt.EventAction
			zone.AlarmEventCode = evt.EventCode
			zone.AlarmEventName = evt.EventName
			zone.AlarmStartedAt = evt.ReceivedAt
		case event.ClassRestore:
			applyZoneRestore(zone, evt)
		case event.ClassTamper:
			zone.TamperActive = true
		case event.ClassTrouble:
			zone.TroubleActive = true
		}
	}

	e.refreshOnlineLocked(time.Now().UTC())
	return e.snapshotLocked()
}

func (e *Engine) RefreshOnline(now time.Time) Snapshot {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.refreshOnlineLocked(now)
	return e.snapshotLocked()
}

func (e *Engine) Snapshot() Snapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.snapshotLocked()
}

func (e *Engine) account(accountID string) *Account {
	if accountID == "" {
		accountID = "unknown"
	}
	if current, ok := e.accounts[accountID]; ok {
		return current
	}
	current := &Account{Account: accountID, Mode: "unknown"}
	e.accounts[accountID] = current
	return current
}

func (e *Engine) zone(accountID, zoneID string) *Zone {
	key := accountID + "/" + zoneID
	if current, ok := e.zones[key]; ok {
		return current
	}
	current := &Zone{Account: accountID, Zone: zoneID}
	e.zones[key] = current
	return current
}

func (e *Engine) refreshOnlineLocked(now time.Time) {
	for _, account := range e.accounts {
		reference := account.LastPingAt
		if reference.IsZero() {
			reference = account.LastEventAt
		}
		account.Online = !reference.IsZero() && now.Sub(reference) <= e.offlineGrace
		switch {
		case account.NightMode:
			account.Mode = "night"
		case account.Armed:
			account.Mode = "armed"
		default:
			account.Mode = "disarmed"
		}
	}
}

func (e *Engine) snapshotLocked() Snapshot {
	snapshot := Snapshot{
		Accounts: make([]Account, 0, len(e.accounts)),
		Zones:    make([]Zone, 0, len(e.zones)),
	}
	for _, account := range e.accounts {
		snapshot.Accounts = append(snapshot.Accounts, *account)
	}
	for _, zone := range e.zones {
		snapshot.Zones = append(snapshot.Zones, *zone)
	}
	return snapshot
}

func maxTime(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}

func (e *Engine) updateZoneMetadata(zone *Zone, evt event.Normalized) {
	if evt.Partition != "" {
		zone.Partition = evt.Partition
	}
	if evt.Group != "" {
		zone.Group = evt.Group
	}
	if evt.Device != "" {
		zone.Device = evt.Device
	}
	if e.devices == nil {
		return
	}
	if device, ok := e.devices.Lookup(evt.Account, evt.Zone, evt.Device); ok {
		if device.Partition != "" {
			zone.Partition = device.Partition
		}
		if device.Group != "" {
			zone.Group = device.Group
		}
		if device.Device != "" {
			zone.Device = device.Device
		}
		zone.DeviceName = device.Name
		zone.Room = device.Room
		zone.Kind = device.Kind
		zone.DeviceEvents = append([]string(nil), device.Events...)
		zone.DeviceEventsLabel = strings.Join(device.Events, ",")
	}
}

func applyAccountRestore(account *Account, evt event.Normalized) {
	switch evt.Signal {
	case "tamper":
		account.TamperActive = false
	case "battery", "connectivity", "power", "interference", "fire_detector", "hardware", "accelerometer", "bypass", "tamper_bypass":
		account.TroubleActive = false
	default:
		account.AlarmActive = false
	}
}

func applyZoneRestore(zone *Zone, evt event.Normalized) {
	switch evt.Signal {
	case "tamper":
		zone.TamperActive = false
	case "battery", "connectivity", "power", "interference", "fire_detector", "hardware", "accelerometer", "bypass", "tamper_bypass":
		zone.TroubleActive = false
	default:
		zone.AlarmActive = false
		zone.AlarmSignal = ""
		zone.AlarmAction = ""
		zone.AlarmEventCode = ""
		zone.AlarmEventName = ""
		zone.AlarmStartedAt = time.Time{}
	}
}
