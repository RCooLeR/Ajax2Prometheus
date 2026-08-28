package jeedom

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var (
	ErrControlDisabled = errors.New("Jeedom controls are disabled")
	ErrActionNotFound  = errors.New("Jeedom action not found")
	ErrActionDenied    = errors.New("Jeedom action denied")
)

type CommandPublisher interface {
	PublishCommandMessage(ctx context.Context, topic string, payload []byte) error
}

type ControlObserver interface {
	ObserveJeedomControl(ctx context.Context, result ControlResult, err error)
}

type ControllerConfig struct {
	Enabled              bool
	StateTopicPrefix     string
	JeedomSetTopicPrefix string
	CommandPayload       string
}

type Controller struct {
	cfg   ControllerConfig
	store *Store
	mqtt  CommandPublisher
	log   zerolog.Logger
	obs   ControlObserver
}

type ControlResult struct {
	DeviceSlug   string `json:"device_slug"`
	Device       string `json:"device"`
	DeviceType   string `json:"device_type,omitempty"`
	Account      string `json:"account,omitempty"`
	Zone         string `json:"zone,omitempty"`
	Action       string `json:"action"`
	CommandID    string `json:"command_id"`
	Topic        string `json:"topic"`
	Published    bool   `json:"published"`
	StateUpdated bool   `json:"state_updated,omitempty"`
}

func NewController(cfg ControllerConfig, store *Store, mqtt CommandPublisher, log zerolog.Logger) *Controller {
	cfg.StateTopicPrefix = trimTopic(firstNonEmpty(cfg.StateTopicPrefix, "ajaxbridge/jeedom"))
	cfg.JeedomSetTopicPrefix = trimTopic(firstNonEmpty(cfg.JeedomSetTopicPrefix, "jeedom/cmd/set"))
	cfg.CommandPayload = firstNonEmpty(cfg.CommandPayload, "1")
	return &Controller{cfg: cfg, store: store, mqtt: mqtt, log: log}
}

func (c *Controller) Enabled() bool {
	return c != nil && c.cfg.Enabled
}

func (c *Controller) SetObserver(observer ControlObserver) {
	if c != nil {
		c.obs = observer
	}
}

func (c *Controller) CommandTopicPattern() string {
	if c == nil {
		return ""
	}
	return c.cfg.StateTopicPrefix + "/devices/+/set"
}

func (c *Controller) CommandTopic(deviceSlug string) string {
	return c.cfg.StateTopicPrefix + "/devices/" + Slug(deviceSlug) + "/set"
}

func (c *Controller) Execute(ctx context.Context, deviceSlug, actionName, source string) (ControlResult, error) {
	if c == nil || !c.cfg.Enabled {
		return ControlResult{}, ErrControlDisabled
	}
	if c.store == nil {
		return ControlResult{}, ErrActionNotFound
	}

	actionName = NormalizeControlAction(actionName)
	action, ok := c.store.Action(deviceSlug, actionName)
	if !ok {
		return ControlResult{}, fmt.Errorf("%w: %s/%s", ErrActionNotFound, deviceSlug, actionName)
	}
	result := ControlResult{
		DeviceSlug: action.DeviceSlug,
		Device:     action.Device,
		Action:     action.Action,
		CommandID:  action.CommandID,
		Topic:      c.cfg.JeedomSetTopicPrefix + "/" + Slug(action.CommandID),
	}
	if device, ok := c.store.Device(action.DeviceSlug); ok {
		result.DeviceType = firstNonEmpty(device.JeedomDeviceType, device.HAModel)
		result.Account = device.LinkedAccount
		result.Zone = device.LinkedZone
	}
	if !action.Allowed {
		err := fmt.Errorf("%w: %s", ErrActionDenied, action.DenyReason)
		c.store.RecordControl(action, source, result.Topic, err)
		c.observe(ctx, result, err)
		return result, err
	}
	if c.mqtt == nil {
		err := errors.New("MQTT command publisher is not configured")
		c.store.RecordControl(action, source, result.Topic, err)
		c.observe(ctx, result, err)
		return result, err
	}

	err := c.mqtt.PublishCommandMessage(ctx, result.Topic, []byte(c.cfg.CommandPayload))
	c.store.RecordControl(action, source, result.Topic, err)
	if err != nil {
		c.observe(ctx, result, err)
		return result, err
	}
	result.Published = true
	if _, updated := c.store.RecordOptimisticControlState(action, time.Now().UTC()); updated {
		result.StateUpdated = true
		if saveErr := c.store.Save(ctx); saveErr != nil {
			c.log.Warn().Err(saveErr).Str("device", result.DeviceSlug).Msg("persist optimistic Jeedom control state")
		}
	}
	c.observe(ctx, result, nil)
	c.log.Info().
		Str("device", result.DeviceSlug).
		Str("action", result.Action).
		Str("command_id", result.CommandID).
		Str("topic", result.Topic).
		Str("source", source).
		Msg("Jeedom control command published")
	return result, nil
}

func (c *Controller) observe(ctx context.Context, result ControlResult, err error) {
	if c != nil && c.obs != nil {
		c.obs.ObserveJeedomControl(ctx, result, err)
	}
}

func (c *Controller) HandleMQTTCommand(ctx context.Context, topic string, payload []byte) (ControlResult, error) {
	deviceSlug, ok := c.deviceSlugFromCommandTopic(topic)
	if !ok {
		return ControlResult{}, ErrActionNotFound
	}
	action := strings.TrimSpace(string(payload))
	return c.Execute(ctx, deviceSlug, action, "mqtt:"+topic)
}

func (c *Controller) deviceSlugFromCommandTopic(topic string) (string, bool) {
	if c == nil {
		return "", false
	}
	prefixParts := strings.Split(trimTopic(c.cfg.StateTopicPrefix), "/")
	parts := strings.Split(trimTopic(topic), "/")
	if len(parts) != len(prefixParts)+3 {
		return "", false
	}
	for i := range prefixParts {
		if parts[i] != prefixParts[i] {
			return "", false
		}
	}
	if parts[len(prefixParts)] != "devices" || parts[len(prefixParts)+2] != "set" {
		return "", false
	}
	deviceSlug := strings.TrimSpace(parts[len(prefixParts)+1])
	return deviceSlug, deviceSlug != "" && deviceSlug != "+" && deviceSlug != "#"
}

func (c *Controller) IsCommandTopic(topic string) bool {
	_, ok := c.deviceSlugFromCommandTopic(topic)
	return ok
}
