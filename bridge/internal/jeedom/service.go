package jeedom

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler func(topic string, payload []byte)) error
}

type MetricsRecorder interface {
	ObserveJeedomMessage()
	ObserveJeedomParseError()
	ObserveJeedomEmptyValue()
	ObserveJeedomCommand(device, command, commandID, metric string, value float64, lastUpdate time.Time)
}

type ServiceConfig struct {
	EventTopic     string
	DiscoveryTopic string
}

type Service struct {
	cfg        ServiceConfig
	store      *Store
	subscriber Subscriber
	publisher  *Publisher
	controller *Controller
	metrics    MetricsRecorder
	samples    *SampleWriter
	log        zerolog.Logger
}

func NewService(cfg ServiceConfig, store *Store, subscriber Subscriber, publisher *Publisher, metrics MetricsRecorder, samples *SampleWriter, log zerolog.Logger) *Service {
	cfg.EventTopic = firstNonEmpty(cfg.EventTopic, "jeedom/cmd/event/#")
	cfg.DiscoveryTopic = firstNonEmpty(cfg.DiscoveryTopic, "jeedom/discovery/eqLogic/#")
	return &Service{
		cfg:        cfg,
		store:      store,
		subscriber: subscriber,
		publisher:  publisher,
		metrics:    metrics,
		samples:    samples,
		log:        log,
	}
}

func (s *Service) SetController(controller *Controller) {
	if s != nil {
		s.controller = controller
	}
}

func (s *Service) Start(ctx context.Context) error {
	if s == nil || s.subscriber == nil {
		return nil
	}
	var topics []string
	topics = appendUniqueTopic(topics, s.cfg.EventTopic)
	topics = appendUniqueTopic(topics, s.cfg.DiscoveryTopic)
	if s.controller != nil && s.controller.Enabled() {
		topics = appendUniqueTopic(topics, s.controller.CommandTopicPattern())
	}
	for _, topic := range topics {
		if err := s.subscriber.Subscribe(ctx, topic, func(topic string, payload []byte) {
			select {
			case <-ctx.Done():
				return
			default:
			}
			go s.HandleMessage(ctx, topic, payload)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) HandleMessage(ctx context.Context, topic string, payload []byte) {
	if s == nil || s.store == nil {
		return
	}
	receivedAt := time.Now()
	if s.metrics != nil {
		s.metrics.ObserveJeedomMessage()
	}
	if err := s.samples.Write(topic, payload, receivedAt); err != nil {
		s.log.Warn().Err(err).Str("topic", topic).Msg("capture Jeedom MQTT sample")
	}
	if s.controller != nil && s.controller.IsCommandTopic(topic) {
		if _, err := s.controller.HandleMQTTCommand(ctx, topic, payload); err != nil {
			s.log.Warn().Err(err).Str("topic", topic).Msg("execute Jeedom MQTT control command")
		}
		return
	}
	if IsDiscoveryEqLogicTopic(topic) {
		discovery, err := ParseDiscoveryMessage(topic, payload, receivedAt)
		if err != nil {
			if s.metrics != nil {
				s.metrics.ObserveJeedomParseError()
			}
			s.log.Warn().Err(err).Str("topic", topic).Msg("parse Jeedom MQTT discovery")
			return
		}
		result := s.store.ApplyDiscovery(discovery)
		if s.publisher != nil {
			if err := s.publisher.PublishDevice(ctx, result.Device); err != nil {
				s.log.Debug().Err(err).Str("device", result.Device.DeviceSlug).Msg("publish Jeedom MQTT discovery state")
			}
		}
		return
	}
	if !IsCommandEventTopic(topic) {
		s.log.Debug().Str("topic", topic).Msg("captured non-event Jeedom MQTT message")
		return
	}

	evt, err := ParseMessage(topic, payload, receivedAt)
	if err != nil {
		if s.metrics != nil {
			s.metrics.ObserveJeedomParseError()
		}
		s.log.Warn().Err(err).Str("topic", topic).Msg("parse Jeedom MQTT event")
		return
	}

	result := s.store.Apply(evt)
	if result.EmptyValue && s.metrics != nil {
		s.metrics.ObserveJeedomEmptyValue()
	}
	if result.HasNumeric && s.metrics != nil {
		s.metrics.ObserveJeedomCommand(result.Device.DeviceSlug, result.Command.Name, result.Command.CommandID, result.Mapping.Metric, result.NumericValue, evt.ReceivedAt)
	}

	if s.publisher != nil {
		if err := s.publisher.PublishDevice(ctx, result.Device); err != nil {
			s.log.Debug().Err(err).Str("device", result.Device.DeviceSlug).Msg("publish Jeedom MQTT state")
		}
	}
}

func appendUniqueTopic(topics []string, topic string) []string {
	topic = trimTopic(topic)
	if topic == "" {
		return topics
	}
	for _, current := range topics {
		if current == topic {
			return topics
		}
	}
	return append(topics, topic)
}

func (s *Service) Store() *Store {
	if s == nil {
		return nil
	}
	return s.store
}
