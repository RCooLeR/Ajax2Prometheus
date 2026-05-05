package jeedom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrMalformedHumanName = errors.New("malformed Jeedom humanName")
	ErrMissingCommandID   = errors.New("missing Jeedom command id in MQTT topic")
)

type Event struct {
	Topic       string          `json:"topic"`
	CommandID   string          `json:"command_id"`
	ObjectName  string          `json:"object"`
	DeviceName  string          `json:"device"`
	CommandName string          `json:"command"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Subtype     string          `json:"subtype"`
	Unit        string          `json:"unit"`
	Value       json.RawMessage `json:"value"`
	ReceivedAt  time.Time       `json:"received_at"`
	RawPayload  json.RawMessage `json:"raw_payload"`
}

type payload struct {
	Value     json.RawMessage `json:"value"`
	HumanName string          `json:"humanName"`
	Unite     string          `json:"unite"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Subtype   string          `json:"subtype"`
}

func ParseMessage(topic string, body []byte, receivedAt time.Time) (Event, error) {
	commandID, err := CommandIDFromTopic(topic)
	if err != nil {
		return Event{Topic: topic, ReceivedAt: receivedAt, RawPayload: append([]byte(nil), body...)}, err
	}

	var raw payload
	if err := json.Unmarshal(body, &raw); err != nil {
		return Event{Topic: topic, CommandID: commandID, ReceivedAt: receivedAt, RawPayload: append([]byte(nil), body...)}, err
	}

	objectName, deviceName, commandName, err := ParseHumanName(raw.HumanName)
	if err != nil {
		return Event{Topic: topic, CommandID: commandID, ReceivedAt: receivedAt, RawPayload: append([]byte(nil), body...)}, err
	}
	if commandName == "" {
		commandName = raw.Name
	}

	return Event{
		Topic:       topic,
		CommandID:   commandID,
		ObjectName:  RepairText(objectName),
		DeviceName:  RepairText(deviceName),
		CommandName: RepairText(commandName),
		Name:        RepairText(raw.Name),
		Type:        strings.TrimSpace(raw.Type),
		Subtype:     strings.TrimSpace(raw.Subtype),
		Unit:        RepairText(raw.Unite),
		Value:       append(json.RawMessage(nil), raw.Value...),
		ReceivedAt:  receivedAt,
		RawPayload:  append([]byte(nil), body...),
	}, nil
}

func CommandIDFromTopic(topic string) (string, error) {
	topic = strings.Trim(strings.TrimSpace(topic), "/")
	if topic == "" {
		return "", ErrMissingCommandID
	}
	parts := strings.Split(topic, "/")
	id := strings.TrimSpace(parts[len(parts)-1])
	if id == "" || id == "#" || id == "+" {
		return "", ErrMissingCommandID
	}
	return id, nil
}

func IsCommandEventTopic(topic string) bool {
	topic = strings.Trim(strings.TrimSpace(topic), "/")
	if topic == "" {
		return false
	}
	parts := strings.Split(topic, "/")
	if len(parts) < 4 {
		return false
	}
	if parts[len(parts)-3] != "cmd" || parts[len(parts)-2] != "event" {
		return false
	}
	id := strings.TrimSpace(parts[len(parts)-1])
	return id != "" && id != "#" && id != "+"
}

func ParseHumanName(value string) (objectName, deviceName, commandName string, err error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", "", ErrMalformedHumanName
	}

	var parts []string
	rest := value
	for rest != "" {
		if !strings.HasPrefix(rest, "[") {
			return "", "", "", fmt.Errorf("%w: %q", ErrMalformedHumanName, value)
		}
		end := strings.Index(rest, "]")
		if end < 0 {
			return "", "", "", fmt.Errorf("%w: %q", ErrMalformedHumanName, value)
		}
		parts = append(parts, rest[1:end])
		rest = rest[end+1:]
	}
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("%w: %q", ErrMalformedHumanName, value)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), nil
}

func (e Event) EmptyValue() bool {
	return EmptyRawValue(e.Value)
}

func EmptyRawValue(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return true
	}
	var text string
	if err := json.Unmarshal(trimmed, &text); err == nil && strings.TrimSpace(text) == "" {
		return true
	}
	return false
}

func NumericRawValue(value json.RawMessage) (float64, bool) {
	if EmptyRawValue(value) {
		return 0, false
	}
	var number float64
	if err := json.Unmarshal(value, &number); err == nil {
		return number, true
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return 0, false
	}
	text = strings.TrimSpace(strings.ReplaceAll(text, ",", "."))
	if text == "" {
		return 0, false
	}
	number, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, false
	}
	return number, true
}

func BoolRawValue(value json.RawMessage) (bool, bool) {
	if EmptyRawValue(value) {
		return false, false
	}
	var boolean bool
	if err := json.Unmarshal(value, &boolean); err == nil {
		return boolean, true
	}
	var number float64
	if err := json.Unmarshal(value, &number); err == nil {
		return number != 0, true
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(RepairText(text))) {
	case "1", "true", "on", "open", "opened", "ouvert", "ouverte", "oui", "yes", "active", "actif":
		return true, true
	case "0", "false", "off", "closed", "close", "ferme", "fermee", "non", "no", "inactive", "inactif":
		return false, true
	default:
		return false, false
	}
}

func StringRawValue(value json.RawMessage) (string, bool) {
	if EmptyRawValue(value) {
		return "", false
	}
	var text string
	if err := json.Unmarshal(value, &text); err == nil {
		return EnglishValue(text), true
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err == nil {
		return number.String(), true
	}
	var boolean bool
	if err := json.Unmarshal(value, &boolean); err == nil {
		if boolean {
			return "true", true
		}
		return "false", true
	}
	return string(value), true
}
