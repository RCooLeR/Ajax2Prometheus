package jeedom

import (
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
