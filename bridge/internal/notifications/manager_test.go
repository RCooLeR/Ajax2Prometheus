package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/jeedom"
	"github.com/RCooLeR/AjaxBridge/internal/state"
	"github.com/rs/zerolog"
)

func TestManagerSendsThresholdAlert(t *testing.T) {
	received := make(chan Alert, 1)
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		var alert Alert
		if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
			t.Errorf("decode alert: %v", err)
			return
		}
		received <- alert
	}))
	defer server.Close()

	store := &Store{cfg: Config{
		Enabled: true,
		Channels: []Channel{{
			ID:   "hook",
			Type: "webhook",
			URL:  server.URL,
		}},
		Rules: []Rule{{
			ID:         "power_high",
			Name:       "Power high",
			Enabled:    true,
			DeviceSlug: "sia_a0f80d_zone_8",
			Metric:     "power_w",
			Condition:  "above",
			Threshold:  1000,
			ArmModes:   []string{"armed"},
			Channels:   []string{"hook"},
			Cooldown:   "0s",
		}},
	}}
	manager := NewManager(store, zerolog.Nop())
	manager.ObserveJeedomUpdate(context.Background(), jeedom.ApplyResult{
		UpdatedValue: true,
		Device: jeedom.Device{
			Device:        "Server power",
			DeviceSlug:    "sia_a0f80d_zone_8",
			LinkedAccount: "A0F80D",
			LinkedZone:    "8",
		},
		Command: jeedom.Command{
			Metric: "power_w",
			Value:  1200.0,
		},
	}, state.Snapshot{Accounts: []state.Account{{Account: "A0F80D", Mode: "armed"}}})

	select {
	case alert := <-received:
		if alert.Rule.ID != "power_high" || alert.Event.NumericValue != 1200 {
			t.Fatalf("alert = %#v", alert)
		}
	case <-time.After(time.Second):
		t.Fatal("expected alert")
	}
}

func TestManagerSendsSwitchChangeOnlyAfterPreviousValue(t *testing.T) {
	store := &Store{cfg: Config{
		Enabled:  true,
		Channels: []Channel{{ID: "log", Type: "log"}},
		Rules: []Rule{{
			ID:         "turned_on",
			Name:       "Turned on",
			Enabled:    true,
			DeviceSlug: "relay",
			Metric:     "state",
			Condition:  "changed_to_on",
			Channels:   []string{"log"},
			Cooldown:   "0s",
		}},
	}}
	manager := NewManager(store, zerolog.Nop())
	snapshot := state.Snapshot{Accounts: []state.Account{{Account: "A0F80D", Mode: "disarmed"}}}

	manager.ObserveJeedomUpdate(context.Background(), jeedom.ApplyResult{
		UpdatedValue: true,
		Device:       jeedom.Device{Device: "Relay", DeviceSlug: "relay"},
		Command:      jeedom.Command{Metric: "state", Value: false},
	}, snapshot)
	if history := manager.History(10); len(history) != 0 {
		t.Fatalf("initial value should not alert: %#v", history)
	}

	manager.ObserveJeedomUpdate(context.Background(), jeedom.ApplyResult{
		UpdatedValue: true,
		Device:       jeedom.Device{Device: "Relay", DeviceSlug: "relay"},
		Command:      jeedom.Command{Metric: "state", Value: true},
	}, snapshot)
	if history := manager.History(10); len(history) != 1 || history[0].RuleID != "turned_on" {
		t.Fatalf("history = %#v", history)
	}
}

func TestManagerSendsControlAlert(t *testing.T) {
	store := &Store{cfg: Config{
		Enabled:  true,
		Channels: []Channel{{ID: "log", Type: "log"}},
		Rules: []Rule{{
			ID:         "control_on",
			Name:       "Control on",
			Enabled:    true,
			DeviceSlug: "relay",
			Condition:  "control_on",
			Channels:   []string{"log"},
			Cooldown:   "0s",
		}},
	}}
	manager := NewManager(store, zerolog.Nop())
	manager.ObserveJeedomControl(context.Background(), jeedom.ControlResult{
		DeviceSlug: "relay",
		Device:     "Relay",
		Action:     "on",
		Published:  true,
	}, nil, state.Snapshot{})
	if history := manager.History(10); len(history) != 1 || history[0].RuleID != "control_on" {
		t.Fatalf("history = %#v", history)
	}
}
