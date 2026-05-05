package jeedom

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SampleWriter struct {
	dir string
}

type sampleEnvelope struct {
	Topic      string          `json:"topic"`
	CommandID  string          `json:"command_id,omitempty"`
	ReceivedAt string          `json:"received_at"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	PayloadRaw string          `json:"payload_raw,omitempty"`
}

func NewSampleWriter(dir string) *SampleWriter {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil
	}
	dir = filepath.Clean(dir)
	if dir == "." || dir == "" {
		return nil
	}
	return &SampleWriter{dir: dir}
}

func (w *SampleWriter) Write(topic string, payload []byte, receivedAt time.Time) error {
	if w == nil {
		return nil
	}
	if receivedAt.IsZero() {
		receivedAt = time.Now()
	}
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}

	commandID, _ := CommandIDFromTopic(topic)
	envelope := sampleEnvelope{
		Topic:      topic,
		CommandID:  commandID,
		ReceivedAt: receivedAt.Format(time.RFC3339Nano),
	}
	if json.Valid(payload) {
		envelope.Payload = append(json.RawMessage(nil), bytes.TrimSpace(payload)...)
	} else {
		envelope.PayloadRaw = string(payload)
	}

	body, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}

	filename := receivedAt.UTC().Format("20060102T150405.000000000Z") + "_" + Slug(firstNonEmpty(commandID, topic)) + "_" + shortHash(string(payload)) + ".json"
	return os.WriteFile(filepath.Join(w.dir, filename), append(body, '\n'), 0o644)
}
