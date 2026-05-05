package jeedom

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	payloadOn  = "ON"
	payloadOff = "OFF"
)

type MQTTClient interface {
	PublishStateMessage(ctx context.Context, topic string, payload []byte, retain bool) error
	PublishDiscoveryMessage(ctx context.Context, key, topic string, payload []byte, retain bool) error
	AvailabilityTopic() string
}

type PublisherConfig struct {
	StateTopicPrefix string
	Discovery        bool
	DiscoveryPrefix  string
	DiscoveryNode    string
	RetainState      bool
	RetainDiscovery  bool
	Controls         bool
}

type Publisher struct {
	cfg  PublisherConfig
	mqtt MQTTClient
}

type DiscoveryDevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
}

type DiscoveryConfig struct {
	Name                string          `json:"name"`
	UniqueID            string          `json:"unique_id"`
	StateTopic          string          `json:"state_topic,omitempty"`
	CommandTopic        string          `json:"command_topic,omitempty"`
	ValueTemplate       string          `json:"value_template,omitempty"`
	UnitOfMeasurement   string          `json:"unit_of_measurement,omitempty"`
	DeviceClass         string          `json:"device_class,omitempty"`
	StateClass          string          `json:"state_class,omitempty"`
	EntityCategory      string          `json:"entity_category,omitempty"`
	PayloadOn           string          `json:"payload_on,omitempty"`
	PayloadOff          string          `json:"payload_off,omitempty"`
	Optimistic          *bool           `json:"optimistic,omitempty"`
	AvailabilityTopic   string          `json:"availability_topic,omitempty"`
	PayloadAvailable    string          `json:"payload_available,omitempty"`
	PayloadNotAvailable string          `json:"payload_not_available,omitempty"`
	JSONAttributesTopic string          `json:"json_attributes_topic"`
	Device              DiscoveryDevice `json:"device"`
}

func NewPublisher(cfg PublisherConfig, mqtt MQTTClient) *Publisher {
	cfg.StateTopicPrefix = trimTopic(firstNonEmpty(cfg.StateTopicPrefix, "ajaxbridge/jeedom"))
	cfg.DiscoveryPrefix = trimTopic(firstNonEmpty(cfg.DiscoveryPrefix, "homeassistant"))
	cfg.DiscoveryNode = Slug(firstNonEmpty(cfg.DiscoveryNode, "ajaxbridge"))
	return &Publisher{cfg: cfg, mqtt: mqtt}
}

func (p *Publisher) PublishDevice(ctx context.Context, device Device) error {
	if p == nil || p.mqtt == nil {
		return nil
	}

	stateTopic := p.StateTopic(device.DeviceSlug)
	if p.cfg.Discovery {
		if err := p.publishLegacyCleanup(ctx, device); err != nil {
			return err
		}
		commands := sortedCommands(device.RawCommands)
		if device.DiscoveryDisabled {
			for _, command := range commands {
				if err := p.publishCommandDiscoveryCleanup(ctx, command, "jeedom_discovery_cleanup"); err != nil {
					return err
				}
			}
			if p.cfg.Controls {
				key := "jeedom_switch_discovery_cleanup:" + device.DeviceSlug
				topic := strings.Join([]string{p.cfg.DiscoveryPrefix, ComponentSwitch, p.cfg.DiscoveryNode, "jeedom_control_" + Slug(device.DeviceSlug), "config"}, "/")
				if err := p.mqtt.PublishDiscoveryMessage(ctx, key, topic, []byte{}, true); err != nil {
					return err
				}
			}
		} else {
			for _, command := range commands {
				if !commandDiscoverable(command, device) {
					if err := p.publishCommandDiscoveryCleanup(ctx, command, "jeedom_merged_sia_cleanup"); err != nil {
						return err
					}
					continue
				}
				if err := p.publishCommandAlternateDiscoveryCleanup(ctx, command); err != nil {
					return err
				}
				topic, payload, err := p.BuildDiscovery(command, device)
				if err != nil {
					return err
				}
				key := "jeedom_discovery:" + command.Component + "/" + discoveryObjectID(command)
				if err := p.mqtt.PublishDiscoveryMessage(ctx, key, topic, payload, p.cfg.RetainDiscovery); err != nil {
					return err
				}
			}
			if p.cfg.Controls {
				for _, action := range sortedSwitchActions(device.Actions) {
					topic, payload, err := p.BuildSwitchDiscovery(action, device)
					if err != nil {
						return err
					}
					key := "jeedom_switch_discovery:" + action.DeviceSlug
					if err := p.mqtt.PublishDiscoveryMessage(ctx, key, topic, payload, p.cfg.RetainDiscovery); err != nil {
						return err
					}
				}
			}
		}
	}

	payload, err := json.Marshal(StatePayload(device))
	if err != nil {
		return err
	}
	return p.mqtt.PublishStateMessage(ctx, stateTopic, payload, p.cfg.RetainState)
}

func (p *Publisher) publishCommandDiscoveryCleanup(ctx context.Context, command Command, prefix string) error {
	if command.Component == "" || command.Metric == "" {
		return nil
	}
	key := prefix + ":" + command.Component + "/" + discoveryObjectID(command)
	if err := p.mqtt.PublishDiscoveryMessage(ctx, key, p.discoveryTopic(command), []byte{}, true); err != nil {
		return err
	}
	for _, component := range alternateDiscoveryComponents(command.Component) {
		key := prefix + ":" + component + "/" + discoveryObjectID(command)
		if err := p.mqtt.PublishDiscoveryMessage(ctx, key, p.discoveryTopicFor(component, discoveryObjectID(command)), []byte{}, true); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publisher) publishCommandAlternateDiscoveryCleanup(ctx context.Context, command Command) error {
	if command.Component == "" || command.Metric == "" {
		return nil
	}
	for _, component := range alternateDiscoveryComponents(command.Component) {
		key := "jeedom_component_migration_cleanup:" + component + "/" + discoveryObjectID(command)
		if err := p.mqtt.PublishDiscoveryMessage(ctx, key, p.discoveryTopicFor(component, discoveryObjectID(command)), []byte{}, true); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publisher) publishLegacyCleanup(ctx context.Context, device Device) error {
	if len(device.LegacyDeviceSlugs) == 0 {
		return nil
	}
	for _, legacySlug := range device.LegacyDeviceSlugs {
		legacySlug = Slug(legacySlug)
		if legacySlug == "" || legacySlug == device.DeviceSlug {
			continue
		}
		if p.cfg.Controls {
			key := "jeedom_legacy_switch_cleanup:" + legacySlug
			topic := strings.Join([]string{p.cfg.DiscoveryPrefix, ComponentSwitch, p.cfg.DiscoveryNode, "jeedom_control_" + legacySlug, "config"}, "/")
			if err := p.mqtt.PublishDiscoveryMessage(ctx, key, topic, []byte{}, true); err != nil {
				return err
			}
		}
		if err := p.mqtt.PublishStateMessage(ctx, p.StateTopic(legacySlug), []byte{}, true); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publisher) StateTopic(deviceSlug string) string {
	return p.cfg.StateTopicPrefix + "/devices/" + Slug(deviceSlug) + "/state"
}

func (p *Publisher) CommandTopic(deviceSlug string) string {
	return p.cfg.StateTopicPrefix + "/devices/" + Slug(deviceSlug) + "/set"
}

func (p *Publisher) BuildDiscovery(command Command, device Device) (string, []byte, error) {
	if command.Component == "" {
		return "", nil, fmt.Errorf("command %q has no Home Assistant component", command.CommandID)
	}

	stateTopic := p.StateTopic(device.DeviceSlug)
	cfg := DiscoveryConfig{
		Name:                firstNonEmpty(commandName(command), titleName(command.Metric)),
		UniqueID:            discoveryUniqueID(command),
		StateTopic:          stateTopic,
		ValueTemplate:       valueTemplate(command),
		UnitOfMeasurement:   command.Unit,
		DeviceClass:         command.DeviceClass,
		StateClass:          command.StateClass,
		EntityCategory:      command.EntityCategory,
		JSONAttributesTopic: stateTopic,
		Device: DiscoveryDevice{
			Identifiers:  discoveryIdentifiers(device),
			Name:         device.Device,
			Manufacturer: firstNonEmpty(device.HAManufacturer, "Ajax via Jeedom"),
			Model:        firstNonEmpty(device.HAModel, "Jeedom MQTT Bridge"),
		},
	}
	if command.Component == ComponentBinarySensor {
		cfg.PayloadOn = payloadOn
		cfg.PayloadOff = payloadOff
	}
	if p.mqtt != nil && p.mqtt.AvailabilityTopic() != "" {
		cfg.AvailabilityTopic = p.mqtt.AvailabilityTopic()
		cfg.PayloadAvailable = "online"
		cfg.PayloadNotAvailable = "offline"
	}

	payload, err := json.Marshal(cfg)
	if err != nil {
		return "", nil, err
	}
	return p.discoveryTopic(command), payload, nil
}

func (p *Publisher) BuildSwitchDiscovery(action Action, device Device) (string, []byte, error) {
	stateTopic := p.StateTopic(device.DeviceSlug)
	optimistic := action.StateCommandID == ""
	cfg := DiscoveryConfig{
		Name:                "Control",
		UniqueID:            "ajaxbridge_jeedom_control_" + Slug(device.DeviceSlug),
		CommandTopic:        p.CommandTopic(device.DeviceSlug),
		PayloadOn:           payloadOn,
		PayloadOff:          payloadOff,
		Optimistic:          &optimistic,
		JSONAttributesTopic: stateTopic,
		Device: DiscoveryDevice{
			Identifiers:  discoveryIdentifiers(device),
			Name:         device.Device,
			Manufacturer: firstNonEmpty(device.HAManufacturer, "Ajax via Jeedom"),
			Model:        firstNonEmpty(device.HAModel, "Jeedom MQTT Bridge"),
		},
	}
	if action.StateCommandID != "" {
		cfg.StateTopic = stateTopic
		cfg.ValueTemplate = "{{ '" + payloadOn + "' if value_json.state else '" + payloadOff + "' }}"
	}
	if p.mqtt != nil && p.mqtt.AvailabilityTopic() != "" {
		cfg.AvailabilityTopic = p.mqtt.AvailabilityTopic()
		cfg.PayloadAvailable = "online"
		cfg.PayloadNotAvailable = "offline"
	}

	payload, err := json.Marshal(cfg)
	if err != nil {
		return "", nil, err
	}
	topic := strings.Join([]string{p.cfg.DiscoveryPrefix, ComponentSwitch, p.cfg.DiscoveryNode, "jeedom_control_" + Slug(device.DeviceSlug), "config"}, "/")
	return topic, payload, nil
}

func (p *Publisher) discoveryTopic(command Command) string {
	return p.discoveryTopicFor(command.Component, discoveryObjectID(command))
}

func (p *Publisher) discoveryTopicFor(component, objectID string) string {
	return strings.Join([]string{p.cfg.DiscoveryPrefix, component, p.cfg.DiscoveryNode, objectID, "config"}, "/")
}

func alternateDiscoveryComponents(component string) []string {
	switch component {
	case ComponentSensor:
		return []string{ComponentBinarySensor}
	case ComponentBinarySensor:
		return []string{ComponentSensor}
	default:
		return nil
	}
}

func discoveryIdentifiers(device Device) []string {
	if len(device.HAIdentifiers) == 0 {
		return []string{"ajaxbridge_jeedom_" + device.DeviceSlug}
	}
	return append([]string(nil), device.HAIdentifiers...)
}

func sortedSwitchActions(actions map[string]Action) []Action {
	on, hasOn := actions["on"]
	off, hasOff := actions["off"]
	if !hasOn || !hasOff || !on.Allowed || !off.Allowed {
		return nil
	}
	on.StateCommandID = firstNonEmpty(on.StateCommandID, off.StateCommandID)
	return []Action{on}
}

func sortedCommands(commands map[string]Command) []Command {
	out := make([]Command, 0, len(commands))
	for _, command := range commands {
		if command.Component != "" && command.Metric != "" {
			out = append(out, command)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CommandID < out[j].CommandID
	})
	return out
}

func commandDiscoverable(command Command, device Device) bool {
	switch command.Metric {
	case "event_source", "event", "event_code":
		return false
	}
	return !deviceLinkedToSIA(device) || !siaOwnedJeedomMetric(command.Metric)
}

func deviceLinkedToSIA(device Device) bool {
	return strings.EqualFold(strings.TrimSpace(device.LinkedSource), "sia") ||
		(strings.TrimSpace(device.LinkedAccount) != "" && strings.TrimSpace(device.LinkedZone) != "") ||
		hasSIAIdentifier(device.HAIdentifiers)
}

func hasSIAIdentifier(identifiers []string) bool {
	for _, identifier := range identifiers {
		identifier = strings.ToLower(strings.TrimSpace(identifier))
		if strings.HasPrefix(identifier, "ajaxbridge_") && strings.Contains(identifier, "_zone_") {
			return true
		}
	}
	return false
}

func siaOwnedJeedomMetric(metric string) bool {
	switch strings.ToLower(strings.TrimSpace(metric)) {
	case "alarm", "alarm_active",
		"tamper",
		"trouble", "trouble_active",
		"external_power",
		"bypass", "tamper_bypass",
		"battery_low",
		"connectivity", "hardware", "firmware", "fire_detector",
		"fire", "smoke", "co", "gas", "gas_or_co",
		"water_leak",
		"interference", "accelerometer":
		return true
	default:
		return false
	}
}

func discoveryObjectID(command Command) string {
	if command.CommandID != "" {
		return "jeedom_cmd_" + Slug(command.CommandID)
	}
	return "jeedom_" + command.DeviceSlug + "_" + Slug(command.Metric)
}

func discoveryUniqueID(command Command) string {
	if command.CommandID != "" {
		return "ajaxbridge_jeedom_cmd_" + Slug(command.CommandID)
	}
	return "ajaxbridge_jeedom_" + command.DeviceSlug + "_" + Slug(command.Metric)
}

func valueTemplate(command Command) string {
	if command.Component == ComponentBinarySensor {
		return "{{ '" + payloadOn + "' if value_json." + command.Metric + " else '" + payloadOff + "' }}"
	}
	return "{{ value_json." + command.Metric + " }}"
}

func commandName(command Command) string {
	switch command.Metric {
	case "power_w":
		return "Power"
	case "current_a":
		return "Current"
	case "voltage_v":
		return "Voltage"
	case "energy_kwh":
		return "Energy"
	case "temperature_c":
		return "Temperature"
	case "battery_percent":
		return "Battery"
	case "battery_state":
		return "Battery state"
	case "signal_dbm":
		return "Signal"
	case "signal_level":
		return "Signal"
	case "humidity_percent":
		return "Humidity"
	case "state":
		return "State"
	case "event_source":
		return "Event source"
	case "event":
		return "Event"
	case "event_code":
		return "Event code"
	case "online":
		return "Online"
	case "external_power":
		return "External power"
	case "gsm_network_type":
		return "GSM network type"
	case "cellular_data_active":
		return "Cellular data active"
	case "opening":
		return "Opening"
	case "door":
		return "Door"
	case "leak":
		return "Leak"
	case "tamper":
		return "Tamper"
	default:
		return titleName(command.Name)
	}
}

func trimTopic(value string) string {
	return strings.Trim(strings.TrimSpace(value), "/")
}
