package jeedom

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestServiceInvalidJSONIncrementsParseErrorAndDoesNotPanic(t *testing.T) {
	metrics := &fakeMetrics{}
	service := NewService(
		ServiceConfig{EventTopic: "jeedom/cmd/event/#"},
		NewStore("keep_last"),
		nil,
		nil,
		metrics,
		nil,
		zerolog.Nop(),
	)

	service.HandleMessage(t.Context(), "jeedom/cmd/event/56", []byte("{"))

	if metrics.messages != 1 {
		t.Fatalf("messages = %d, want 1", metrics.messages)
	}
	if metrics.parseErrors != 1 {
		t.Fatalf("parseErrors = %d, want 1", metrics.parseErrors)
	}
}

func TestServiceIgnoresNonEventTopicsAfterCapturing(t *testing.T) {
	metrics := &fakeMetrics{}
	service := NewService(
		ServiceConfig{EventTopic: "jeedom/#"},
		NewStore("keep_last"),
		nil,
		nil,
		metrics,
		nil,
		zerolog.Nop(),
	)

	service.HandleMessage(t.Context(), "jeedom/state", []byte("online"))

	if metrics.messages != 1 {
		t.Fatalf("messages = %d, want 1", metrics.messages)
	}
	if metrics.parseErrors != 0 {
		t.Fatalf("parseErrors = %d, want 0", metrics.parseErrors)
	}
}

func TestServiceEmptyValueIncrementsCounter(t *testing.T) {
	metrics := &fakeMetrics{}
	service := NewService(
		ServiceConfig{EventTopic: "jeedom/cmd/event/#"},
		NewStore("keep_last"),
		nil,
		nil,
		metrics,
		nil,
		zerolog.Nop(),
	)

	service.HandleMessage(t.Context(), "jeedom/cmd/event/56", []byte(`{"value":"","humanName":"[None][Server][Puissance]","unite":"","name":"Puissance","type":"info","subtype":"numeric"}`))

	if metrics.emptyValues != 1 {
		t.Fatalf("emptyValues = %d, want 1", metrics.emptyValues)
	}
}

func TestServiceNumericValueUpdatesMetrics(t *testing.T) {
	metrics := &fakeMetrics{}
	service := NewService(
		ServiceConfig{EventTopic: "jeedom/cmd/event/#"},
		NewStore("keep_last"),
		nil,
		nil,
		metrics,
		nil,
		zerolog.Nop(),
	)

	service.HandleMessage(t.Context(), "jeedom/cmd/event/56", []byte(`{"value":"123.4","humanName":"[None][Server][Puissance]","unite":"","name":"Puissance","type":"info","subtype":"numeric"}`))

	if metrics.commands != 1 {
		t.Fatalf("commands = %d, want 1", metrics.commands)
	}
}

func TestServiceObservesExternalJeedomSetCommand(t *testing.T) {
	store := NewStore("keep_last")
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/10", []byte(relayDiscoveryPayload), time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	store.ApplyDiscovery(discovery)
	observer := &fakeObserver{}
	service := NewService(
		ServiceConfig{EventTopic: "jeedom/cmd/event/#", SetTopicPrefix: "jeedom/cmd/set"},
		store,
		nil,
		nil,
		nil,
		nil,
		zerolog.Nop(),
	)
	service.SetObserver(observer)

	service.HandleMessage(t.Context(), "jeedom/cmd/set/85", nil)

	if observer.controls != 1 {
		t.Fatalf("controls = %d, want 1", observer.controls)
	}
	if observer.last.Action != "on" || observer.last.DeviceSlug != "garage_gate" {
		t.Fatalf("last control = %#v", observer.last)
	}
}

func TestServiceSkipsJeedomSetCommandRecentlyIssuedByBridge(t *testing.T) {
	store := NewStore("keep_last")
	discovery, err := ParseDiscoveryMessage("jeedom/discovery/eqLogic/10", []byte(relayDiscoveryPayload), time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	store.ApplyDiscovery(discovery)
	action, ok := store.ActionByCommandID("85")
	if !ok {
		t.Fatal("missing action")
	}
	store.RecordControl(action, "http:127.0.0.1", "jeedom/cmd/set/85", nil)
	observer := &fakeObserver{}
	service := NewService(
		ServiceConfig{EventTopic: "jeedom/cmd/event/#", SetTopicPrefix: "jeedom/cmd/set"},
		store,
		nil,
		nil,
		nil,
		nil,
		zerolog.Nop(),
	)
	service.SetObserver(observer)

	service.HandleMessage(t.Context(), "jeedom/cmd/set/85", nil)

	if observer.controls != 0 {
		t.Fatalf("controls = %d, want 0", observer.controls)
	}
}

type fakeMetrics struct {
	messages    int
	parseErrors int
	emptyValues int
	commands    int
}

func (m *fakeMetrics) ObserveJeedomMessage() {
	m.messages++
}

func (m *fakeMetrics) ObserveJeedomParseError() {
	m.parseErrors++
}

func (m *fakeMetrics) ObserveJeedomEmptyValue() {
	m.emptyValues++
}

func (m *fakeMetrics) ObserveJeedomCommand(string, string, string, string, float64, time.Time) {
	m.commands++
}

type fakeObserver struct {
	updates  int
	controls int
	last     ControlResult
}

func (o *fakeObserver) ObserveJeedomUpdate(context.Context, ApplyResult) {
	o.updates++
}

func (o *fakeObserver) ObserveJeedomControl(_ context.Context, result ControlResult, _ error) {
	o.controls++
	o.last = result
}
