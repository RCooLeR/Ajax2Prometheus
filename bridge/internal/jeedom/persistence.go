package jeedom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const persistedStoreVersion = 1

type persistedStore struct {
	Version int      `json:"version"`
	Devices []Device `json:"devices"`
}

func LoadStore(ctx context.Context, path, policy string, resolver IdentityResolver) (*Store, error) {
	store := NewStoreWithResolver(policy, resolver)
	store.SetPath(path)
	path = strings.TrimSpace(path)
	if path == "" {
		return store, nil
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store, nil
		}
		return nil, err
	}

	devices, err := decodePersistedDevices(data)
	if err != nil {
		return nil, err
	}
	store.replaceDevices(devices)
	return store, nil
}

func (s *Store) SetPath(path string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.path = strings.TrimSpace(path)
}

func (s *Store) Save(ctx context.Context) error {
	if s == nil {
		return nil
	}
	s.saveMu.Lock()
	defer s.saveMu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.RLock()
	path := s.path
	devices := s.devicesLocked()
	s.mu.RUnlock()
	if strings.TrimSpace(path) == "" {
		return nil
	}
	return writePersistedDevices(ctx, path, devices)
}

func (s *Store) replaceDevices(devices []Device) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = make(map[string]*Device, len(devices))
	for _, device := range devices {
		device = normalizePersistedDevice(device)
		if device.DeviceSlug == "" || device.DeviceSlug == "unknown" {
			continue
		}
		current := device
		s.devices[current.DeviceSlug] = &current
	}
	s.rebuildIndexesLocked()
}

func decodePersistedDevices(data []byte) ([]Device, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil
	}
	if len(data) > 0 && data[0] == '[' {
		var devices []Device
		if err := json.Unmarshal(data, &devices); err != nil {
			return nil, err
		}
		return devices, nil
	}
	var persisted persistedStore
	if err := json.Unmarshal(data, &persisted); err != nil {
		return nil, err
	}
	return persisted.Devices, nil
}

func writePersistedDevices(ctx context.Context, path string, devices []Device) error {
	body, err := json.MarshalIndent(persistedStore{
		Version: persistedStoreVersion,
		Devices: devices,
	}, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return err
	}
	_ = os.Remove(path)
	return os.Rename(tmp, path)
}

func normalizePersistedDevice(device Device) Device {
	device.Source = firstNonEmpty(device.Source, Source)
	device.DeviceSlug = Slug(device.DeviceSlug)
	if device.DeviceSlug == "" || device.DeviceSlug == "unknown" {
		device.DeviceSlug = Slug(device.Device)
	}
	if device.Values == nil {
		device.Values = make(map[string]any)
	}
	if device.RawCommands == nil {
		device.RawCommands = make(map[string]Command)
	}
	if device.Actions == nil {
		device.Actions = make(map[string]Action)
	}
	if device.HAManufacturer == "" {
		device.HAManufacturer = "Ajax via Jeedom"
	}
	if device.HAModel == "" {
		device.HAModel = firstNonEmpty(device.JeedomDeviceType, "Jeedom MQTT Bridge")
	}
	return device
}
