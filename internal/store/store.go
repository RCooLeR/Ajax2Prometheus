package store

import (
	"context"
	"sync"
	"time"

	"github.com/RCooLeR/Ajax2Prometheus/internal/event"
)

type EventRecord struct {
	ID          uint64             `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	Account     string             `json:"account"`
	Protocol    string             `json:"protocol"`
	Sequence    string             `json:"sequence"`
	Receiver    string             `json:"receiver"`
	Line        string             `json:"line"`
	EventCode   string             `json:"event_code"`
	ContactID   string             `json:"contact_id"`
	EventClass  string             `json:"event_class"`
	EventAction string             `json:"event_action"`
	EventName   string             `json:"event_name"`
	Description string             `json:"description"`
	Source      string             `json:"source"`
	Signal      string             `json:"signal"`
	Severity    string             `json:"severity"`
	Partition   string             `json:"partition"`
	Group       string             `json:"group"`
	Zone        string             `json:"zone"`
	Device      string             `json:"device"`
	User        string             `json:"user"`
	OccurredAt  time.Time          `json:"occurred_at"`
	ReceivedAt  time.Time          `json:"received_at"`
	RawData     string             `json:"raw_data"`
	RawPayload  string             `json:"raw_payload"`
	RawMessage  string             `json:"raw_message"`
	XData       []event.XDataEntry `json:"xdata"`
	ParseStatus string             `json:"parse_status"`
	ParseError  string             `json:"parse_error"`
	Encrypted   bool               `json:"encrypted"`
}

type Store struct {
	mu     sync.RWMutex
	nextID uint64
	events []EventRecord
}

func New() *Store {
	return &Store{events: make([]EventRecord, 0)}
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) AppendEvent(ctx context.Context, evt event.Normalized) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	s.events = append(s.events, EventRecord{
		ID:          s.nextID,
		CreatedAt:   time.Now().UTC(),
		Account:     evt.Account,
		Protocol:    evt.Protocol,
		Sequence:    evt.Sequence,
		Receiver:    evt.Receiver,
		Line:        evt.Line,
		EventCode:   evt.EventCode,
		ContactID:   evt.ContactID,
		EventClass:  string(evt.EventClass),
		EventAction: evt.EventAction,
		EventName:   evt.EventName,
		Description: evt.Description,
		Source:      evt.Source,
		Signal:      evt.Signal,
		Severity:    evt.Severity,
		Partition:   evt.Partition,
		Group:       evt.Group,
		Zone:        evt.Zone,
		Device:      evt.Device,
		User:        evt.User,
		OccurredAt:  evt.OccurredAt,
		ReceivedAt:  evt.ReceivedAt,
		RawData:     evt.RawData,
		RawPayload:  evt.RawPayload,
		RawMessage:  evt.RawMessage,
		XData:       append([]event.XDataEntry(nil), evt.XData...),
		ParseStatus: string(evt.ParseStatus),
		ParseError:  evt.ParseError,
		Encrypted:   evt.Encrypted,
	})
	return nil
}

func (s *Store) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (s *Store) ListEvents(limit int) []EventRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.events) {
		limit = len(s.events)
	}
	start := len(s.events) - limit
	out := make([]EventRecord, limit)
	copy(out, s.events[start:])
	return out
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}
