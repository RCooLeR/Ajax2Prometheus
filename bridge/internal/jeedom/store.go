package jeedom

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const Source = "jeedom"

type EmptyValuePolicy string

const (
	EmptyValueKeepLast EmptyValuePolicy = "keep_last"
	EmptyValueUnknown  EmptyValuePolicy = "unknown"
)

type Store struct {
	mu         sync.RWMutex
	policy     EmptyValuePolicy
	resolver   IdentityResolver
	devices    map[string]*Device
	commands   map[string]string
	eqLogics   map[string]string
	baseGroups map[string][]string
	audits     []ControlAudit
}

type Device struct {
	Source            string             `json:"source"`
	ObjectName        string             `json:"object"`
	Device            string             `json:"device"`
	DeviceSlug        string             `json:"device_slug"`
	BaseSlug          string             `json:"base_slug,omitempty"`
	LastUpdate        time.Time          `json:"last_update"`
	Values            map[string]any     `json:"values"`
	RawCommands       map[string]Command `json:"raw_commands"`
	HAIdentifiers     []string           `json:"ha_identifiers,omitempty"`
	HAManufacturer    string             `json:"ha_manufacturer,omitempty"`
	HAModel           string             `json:"ha_model,omitempty"`
	SuggestedArea     string             `json:"suggested_area,omitempty"`
	LegacyDeviceSlugs []string           `json:"legacy_device_slugs,omitempty"`
	LinkedSource      string             `json:"linked_source,omitempty"`
	LinkedAccount     string             `json:"linked_account,omitempty"`
	LinkedZone        string             `json:"linked_zone,omitempty"`
	DiscoveryDisabled bool               `json:"discovery_disabled,omitempty"`
	JeedomID          string             `json:"jeedom_id,omitempty"`
	JeedomLogicalID   string             `json:"jeedom_logical_id,omitempty"`
	JeedomDeviceType  string             `json:"jeedom_device_type,omitempty"`
	JeedomEnabled     bool               `json:"jeedom_enabled,omitempty"`
	JeedomVisible     bool               `json:"jeedom_visible,omitempty"`
	Actions           map[string]Action  `json:"actions,omitempty"`
}

type Command struct {
	CommandID      string    `json:"command_id"`
	ObjectName     string    `json:"object"`
	Device         string    `json:"device"`
	DeviceSlug     string    `json:"device_slug"`
	Name           string    `json:"name"`
	RawName        string    `json:"raw_name,omitempty"`
	Metric         string    `json:"metric"`
	Component      string    `json:"component"`
	Topic          string    `json:"topic"`
	Type           string    `json:"type"`
	Subtype        string    `json:"subtype"`
	Unit           string    `json:"unit,omitempty"`
	DeviceClass    string    `json:"device_class,omitempty"`
	StateClass     string    `json:"state_class,omitempty"`
	EntityCategory string    `json:"entity_category,omitempty"`
	LogicalID      string    `json:"logical_id,omitempty"`
	GenericType    string    `json:"generic_type,omitempty"`
	Visible        bool      `json:"visible,omitempty"`
	Value          any       `json:"value,omitempty"`
	LastUpdate     time.Time `json:"last_update"`
	LastValueAt    time.Time `json:"last_value_at,omitempty"`
	EmptyValue     bool      `json:"empty_value,omitempty"`
}

type Action struct {
	Action            string    `json:"action"`
	CommandID         string    `json:"command_id"`
	Device            string    `json:"device"`
	DeviceSlug        string    `json:"device_slug"`
	DeviceType        string    `json:"device_type,omitempty"`
	EqLogicID         string    `json:"eq_logic_id,omitempty"`
	Name              string    `json:"name"`
	RawName           string    `json:"raw_name,omitempty"`
	LogicalID         string    `json:"logical_id,omitempty"`
	Subtype           string    `json:"subtype,omitempty"`
	StateCommandID    string    `json:"state_command_id,omitempty"`
	Allowed           bool      `json:"allowed"`
	DenyReason        string    `json:"deny_reason,omitempty"`
	LastRequestedAt   time.Time `json:"last_requested_at,omitempty"`
	LastRequestSource string    `json:"last_request_source,omitempty"`
}

type ControlAudit struct {
	Time       time.Time `json:"time"`
	Source     string    `json:"source"`
	Device     string    `json:"device"`
	DeviceSlug string    `json:"device_slug"`
	Action     string    `json:"action"`
	CommandID  string    `json:"command_id"`
	Topic      string    `json:"topic"`
	Result     string    `json:"result"`
	Error      string    `json:"error,omitempty"`
}

type ApplyResult struct {
	Device       Device
	Command      Command
	Mapping      Mapping
	EmptyValue   bool
	UpdatedValue bool
	NumericValue float64
	HasNumeric   bool
}

type ApplyDiscoveryResult struct {
	Device  Device
	Actions []Action
}

type IdentityResolver interface {
	Resolve(evt Event, mapping Mapping) DeviceIdentity
}

type DiscoveryIdentityResolver interface {
	ResolveDiscovery(discovery Discovery) DeviceIdentity
}

type DeviceIdentity struct {
	DeviceSlug        string
	DeviceName        string
	BaseSlug          string
	HAIdentifiers     []string
	HAManufacturer    string
	HAModel           string
	SuggestedArea     string
	LegacyDeviceSlugs []string
	LinkedSource      string
	LinkedAccount     string
	LinkedZone        string
	DiscoveryDisabled bool
}

func NewStore(policy string) *Store {
	return NewStoreWithResolver(policy, nil)
}

func NewStoreWithResolver(policy string, resolver IdentityResolver) *Store {
	return &Store{
		policy:     NormalizeEmptyValuePolicy(policy),
		resolver:   resolver,
		devices:    make(map[string]*Device),
		commands:   make(map[string]string),
		eqLogics:   make(map[string]string),
		baseGroups: make(map[string][]string),
	}
}

func NormalizeEmptyValuePolicy(policy string) EmptyValuePolicy {
	switch EmptyValuePolicy(strings.TrimSpace(policy)) {
	case EmptyValueUnknown:
		return EmptyValueUnknown
	default:
		return EmptyValueKeepLast
	}
}

func (s *Store) Apply(evt Event) ApplyResult {
	mapping := MappingFor(evt)
	now := evt.ReceivedAt
	if now.IsZero() {
		now = time.Now()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	identity := s.identityFor(evt, mapping)
	deviceSlug := identity.DeviceSlug
	device := s.devices[deviceSlug]
	if device == nil {
		device = &Device{
			Source:      Source,
			DeviceSlug:  deviceSlug,
			BaseSlug:    identity.BaseSlug,
			Values:      make(map[string]any),
			RawCommands: make(map[string]Command),
			Actions:     make(map[string]Action),
		}
		s.devices[deviceSlug] = device
		if identity.BaseSlug != "" && !containsString(s.baseGroups[identity.BaseSlug], deviceSlug) {
			s.baseGroups[identity.BaseSlug] = append(s.baseGroups[identity.BaseSlug], deviceSlug)
		}
	}
	device.Source = Source
	device.ObjectName = evt.ObjectName
	device.Device = firstNonEmpty(identity.DeviceName, evt.DeviceName)
	device.BaseSlug = identity.BaseSlug
	device.LastUpdate = now
	device.HAIdentifiers = append([]string(nil), identity.HAIdentifiers...)
	device.HAManufacturer = identity.HAManufacturer
	device.HAModel = identity.HAModel
	device.SuggestedArea = identity.SuggestedArea
	device.LegacyDeviceSlugs = mergeStringLists(device.LegacyDeviceSlugs, identity.LegacyDeviceSlugs)
	device.LinkedSource = identity.LinkedSource
	device.LinkedAccount = identity.LinkedAccount
	device.LinkedZone = identity.LinkedZone
	device.DiscoveryDisabled = identity.DiscoveryDisabled

	command := Command{
		CommandID:      evt.CommandID,
		ObjectName:     evt.ObjectName,
		Device:         evt.DeviceName,
		DeviceSlug:     deviceSlug,
		Name:           EnglishCommandName(firstNonEmpty(evt.CommandName, evt.Name), mapping, evt.CommandID),
		RawName:        firstNonEmpty(evt.CommandName, evt.Name),
		Metric:         mapping.Metric,
		Component:      mapping.Component,
		Topic:          evt.Topic,
		Type:           evt.Type,
		Subtype:        evt.Subtype,
		Unit:           mapping.Unit,
		DeviceClass:    mapping.DeviceClass,
		StateClass:     mapping.StateClass,
		EntityCategory: mapping.EntityCategory,
		LastUpdate:     now,
	}
	if existing, ok := device.RawCommands[evt.CommandID]; ok {
		command.LogicalID = existing.LogicalID
		command.GenericType = existing.GenericType
		command.Visible = existing.Visible
	}

	result := ApplyResult{
		Mapping:    mapping,
		EmptyValue: evt.EmptyValue(),
	}

	if result.EmptyValue {
		command.EmptyValue = true
		if existing, ok := device.RawCommands[evt.CommandID]; ok && s.policy == EmptyValueKeepLast {
			command.Value = existing.Value
			command.LastValueAt = existing.LastValueAt
		}
		if s.policy == EmptyValueUnknown {
			device.Values[mapping.Metric] = nil
			command.Value = nil
			result.UpdatedValue = true
		}
	} else {
		value, ok := mappedValue(evt, mapping)
		if ok {
			device.Values[mapping.Metric] = value
			command.Value = value
			command.LastValueAt = now
			result.UpdatedValue = true
		}
		if number, ok := NumericRawValue(evt.Value); ok {
			result.NumericValue = number
			result.HasNumeric = true
		}
	}

	device.RawCommands[evt.CommandID] = command
	s.commands[evt.CommandID] = deviceSlug
	result.Device = copyDevice(*device)
	result.Command = command
	return result
}

func (s *Store) ApplyDiscovery(discovery Discovery) ApplyDiscoveryResult {
	now := discovery.ReceivedAt
	if now.IsZero() {
		now = time.Now()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	identity := s.identityForDiscovery(discovery)
	deviceSlug := identity.DeviceSlug
	device := s.devices[deviceSlug]
	if device == nil {
		device = &Device{
			Source:      Source,
			DeviceSlug:  deviceSlug,
			BaseSlug:    identity.BaseSlug,
			Values:      make(map[string]any),
			RawCommands: make(map[string]Command),
			Actions:     make(map[string]Action),
		}
		s.devices[deviceSlug] = device
		if identity.BaseSlug != "" && !containsString(s.baseGroups[identity.BaseSlug], deviceSlug) {
			s.baseGroups[identity.BaseSlug] = append(s.baseGroups[identity.BaseSlug], deviceSlug)
		}
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

	device.Source = Source
	device.ObjectName = discovery.ObjectName
	device.Device = firstNonEmpty(identity.DeviceName, discovery.Name)
	device.BaseSlug = identity.BaseSlug
	device.LastUpdate = now
	device.HAIdentifiers = append([]string(nil), identity.HAIdentifiers...)
	device.HAManufacturer = identity.HAManufacturer
	device.HAModel = identity.HAModel
	device.SuggestedArea = identity.SuggestedArea
	device.LegacyDeviceSlugs = mergeStringLists(device.LegacyDeviceSlugs, identity.LegacyDeviceSlugs)
	device.LinkedSource = identity.LinkedSource
	device.LinkedAccount = identity.LinkedAccount
	device.LinkedZone = identity.LinkedZone
	device.DiscoveryDisabled = identity.DiscoveryDisabled
	device.JeedomID = discovery.EqLogicID
	device.JeedomLogicalID = discovery.LogicalID
	device.JeedomDeviceType = firstNonEmpty(discovery.DeviceType, discovery.ApplyDevice)
	device.JeedomEnabled = discovery.Enabled
	device.JeedomVisible = discovery.Visible
	if discovery.EqLogicID != "" {
		s.eqLogics[discovery.EqLogicID] = deviceSlug
	}

	for _, info := range discovery.InfoCommands {
		mapping := MappingFor(Event{
			Topic:       "jeedom/cmd/event/" + info.CommandID,
			CommandID:   info.CommandID,
			ObjectName:  discovery.ObjectName,
			DeviceName:  discovery.Name,
			CommandName: info.Name,
			Name:        info.Name,
			Type:        info.Type,
			Subtype:     info.Subtype,
			Unit:        info.Unit,
			ReceivedAt:  now,
		})
		command := Command{
			CommandID:      info.CommandID,
			ObjectName:     discovery.ObjectName,
			Device:         discovery.Name,
			DeviceSlug:     deviceSlug,
			Name:           EnglishCommandName(info.Name, mapping, info.CommandID),
			RawName:        info.Name,
			Metric:         mapping.Metric,
			Component:      mapping.Component,
			Topic:          "jeedom/cmd/event/" + info.CommandID,
			Type:           info.Type,
			Subtype:        info.Subtype,
			Unit:           mapping.Unit,
			DeviceClass:    mapping.DeviceClass,
			StateClass:     mapping.StateClass,
			EntityCategory: mapping.EntityCategory,
			LogicalID:      info.LogicalID,
			GenericType:    info.GenericType,
			Visible:        info.Visible,
			LastUpdate:     now,
		}
		if existing, ok := device.RawCommands[info.CommandID]; ok {
			command.Value = existing.Value
			command.LastValueAt = existing.LastValueAt
			command.EmptyValue = existing.EmptyValue
			if existing.LastUpdate.After(command.LastUpdate) {
				command.LastUpdate = existing.LastUpdate
			}
		}
		device.RawCommands[info.CommandID] = command
		s.commands[info.CommandID] = deviceSlug
	}

	actions := make([]Action, 0, len(discovery.Actions))
	for _, actionCommand := range discovery.Actions {
		actionName := normalizeDiscoveryAction(actionCommand)
		if actionName == "" {
			actionName = "cmd_" + actionCommand.CommandID
		}
		allowed, denyReason := controlAllowed(discovery, actionCommand, actionName)
		action := Action{
			Action:         actionName,
			CommandID:      actionCommand.CommandID,
			Device:         discovery.Name,
			DeviceSlug:     deviceSlug,
			DeviceType:     device.JeedomDeviceType,
			EqLogicID:      discovery.EqLogicID,
			Name:           EnglishActionName(actionName, actionCommand.Name),
			RawName:        actionCommand.Name,
			LogicalID:      actionCommand.LogicalID,
			Subtype:        actionCommand.Subtype,
			StateCommandID: actionCommand.StateCommandID,
			Allowed:        allowed,
			DenyReason:     denyReason,
		}
		if existing, ok := device.Actions[actionName]; ok {
			action.LastRequestedAt = existing.LastRequestedAt
			action.LastRequestSource = existing.LastRequestSource
		}
		device.Actions[actionName] = action
		actions = append(actions, action)
	}

	return ApplyDiscoveryResult{Device: copyDevice(*device), Actions: actions}
}

func (s *Store) identityFor(evt Event, mapping Mapping) DeviceIdentity {
	baseSlug := Slug(evt.DeviceName)
	identity := DeviceIdentity{
		DeviceName:     evt.DeviceName,
		BaseSlug:       baseSlug,
		HAManufacturer: "Ajax via Jeedom",
		HAModel:        "Jeedom MQTT Bridge",
	}
	identity.DeviceSlug = s.localDeviceSlug(evt, mapping, baseSlug)
	identity.HAIdentifiers = []string{"ajaxbridge_jeedom_" + identity.DeviceSlug}
	if s.resolver == nil {
		return identity
	}

	resolved := s.resolver.Resolve(evt, mapping)
	if resolved.DeviceSlug == "" {
		resolved.DeviceSlug = identity.DeviceSlug
	}
	if resolved.DeviceSlug != identity.DeviceSlug {
		resolved.LegacyDeviceSlugs = mergeStringLists(resolved.LegacyDeviceSlugs, []string{identity.DeviceSlug})
	}
	if resolved.DeviceName == "" {
		resolved.DeviceName = identity.DeviceName
	}
	if resolved.BaseSlug == "" {
		resolved.BaseSlug = identity.BaseSlug
	}
	if len(resolved.HAIdentifiers) == 0 {
		resolved.HAIdentifiers = identity.HAIdentifiers
	}
	if resolved.HAManufacturer == "" {
		resolved.HAManufacturer = identity.HAManufacturer
	}
	if resolved.HAModel == "" {
		resolved.HAModel = identity.HAModel
	}
	return resolved
}

func (s *Store) identityForDiscovery(discovery Discovery) DeviceIdentity {
	baseSlug := Slug(discovery.Name)
	identity := DeviceIdentity{
		DeviceName:     discovery.Name,
		BaseSlug:       baseSlug,
		HAManufacturer: "Ajax via Jeedom",
		HAModel:        firstNonEmpty(discovery.DeviceType, discovery.ApplyDevice, "Jeedom MQTT Bridge"),
	}
	if current, ok := s.eqLogics[discovery.EqLogicID]; ok && current != "" {
		identity.DeviceSlug = current
	} else {
		identity.DeviceSlug = s.localDiscoveryDeviceSlug(discovery, baseSlug)
	}
	identity.HAIdentifiers = []string{"ajaxbridge_jeedom_" + identity.DeviceSlug}
	if s.resolver == nil {
		return identity
	}
	if resolver, ok := s.resolver.(DiscoveryIdentityResolver); ok {
		resolved := resolver.ResolveDiscovery(discovery)
		return mergeIdentity(identity, resolved)
	}
	return mergeIdentity(identity, s.resolver.Resolve(Event{
		Topic:       discovery.Topic,
		CommandID:   firstDiscoveryCommandID(discovery),
		ObjectName:  discovery.ObjectName,
		DeviceName:  discovery.Name,
		CommandName: firstDiscoveryCommandName(discovery),
		ReceivedAt:  discovery.ReceivedAt,
	}, Mapping{}))
}

func (s *Store) localDeviceSlug(evt Event, mapping Mapping, baseSlug string) string {
	if current, ok := s.commands[evt.CommandID]; ok {
		return current
	}
	groups := s.baseGroups[baseSlug]
	if len(groups) == 0 {
		return baseSlug
	}
	for i := len(groups) - 1; i >= 0; i-- {
		device := s.devices[groups[i]]
		if device == nil || !deviceHasMetric(device, mapping.Metric) {
			return groups[i]
		}
	}
	suffix := Slug(evt.CommandID)
	if suffix == "" || suffix == "unknown" {
		suffix = shortHash(evt.Topic)
	}
	return baseSlug + "_" + suffix
}

func deviceHasMetric(device *Device, metric string) bool {
	if device == nil || metric == "" {
		return false
	}
	for _, command := range device.RawCommands {
		if command.Metric == metric {
			return true
		}
	}
	return false
}

func mappedValue(evt Event, mapping Mapping) (any, bool) {
	switch {
	case mapping.Numeric:
		return NumericRawValue(evt.Value)
	case mapping.Binary:
		return BoolRawValue(evt.Value)
	default:
		return StringRawValue(evt.Value)
	}
}

func (s *Store) Devices() []Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, copyDevice(*device))
	}
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].DeviceSlug < devices[j].DeviceSlug
	})
	return devices
}

func (s *Store) Device(slug string) (Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, ok := s.devices[slug]
	if !ok {
		return Device{}, false
	}
	return copyDevice(*device), true
}

func (s *Store) Commands() []Command {
	s.mu.RLock()
	defer s.mu.RUnlock()

	commands := make([]Command, 0, len(s.commands))
	for commandID, deviceSlug := range s.commands {
		device := s.devices[deviceSlug]
		if device == nil {
			continue
		}
		command, ok := device.RawCommands[commandID]
		if ok {
			commands = append(commands, command)
		}
	}
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].CommandID < commands[j].CommandID
	})
	return commands
}

func (s *Store) Actions() []Action {
	s.mu.RLock()
	defer s.mu.RUnlock()

	actions := make([]Action, 0)
	for _, device := range s.devices {
		if device == nil {
			continue
		}
		for _, action := range device.Actions {
			actions = append(actions, action)
		}
	}
	sort.Slice(actions, func(i, j int) bool {
		if actions[i].DeviceSlug == actions[j].DeviceSlug {
			return actions[i].Action < actions[j].Action
		}
		return actions[i].DeviceSlug < actions[j].DeviceSlug
	})
	return actions
}

func (s *Store) Action(deviceSlug, actionName string) (Action, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device := s.devices[Slug(deviceSlug)]
	if device == nil {
		return Action{}, false
	}
	action, ok := device.Actions[NormalizeControlAction(actionName)]
	return action, ok
}

func (s *Store) ActionByCommandID(commandID string) (Action, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	commandID = strings.TrimSpace(commandID)
	if commandID == "" {
		return Action{}, false
	}
	for _, device := range s.devices {
		if device == nil {
			continue
		}
		for _, action := range device.Actions {
			if action.CommandID == commandID {
				return action, true
			}
		}
	}
	return Action{}, false
}

func (s *Store) RecordControl(action Action, source, topic string, err error) {
	if s == nil {
		return
	}
	now := time.Now().UTC()
	result := "published"
	errText := ""
	if err != nil {
		result = "failed"
		errText = err.Error()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if device := s.devices[action.DeviceSlug]; device != nil && device.Actions != nil {
		if current, ok := device.Actions[action.Action]; ok {
			current.LastRequestedAt = now
			current.LastRequestSource = source
			device.Actions[action.Action] = current
		}
	}
	s.audits = append(s.audits, ControlAudit{
		Time:       now,
		Source:     source,
		Device:     action.Device,
		DeviceSlug: action.DeviceSlug,
		Action:     action.Action,
		CommandID:  action.CommandID,
		Topic:      topic,
		Result:     result,
		Error:      errText,
	})
	if len(s.audits) > 200 {
		s.audits = append([]ControlAudit(nil), s.audits[len(s.audits)-200:]...)
	}
}

func (s *Store) HasRecentBridgeControl(commandID string, window time.Duration) bool {
	if s == nil {
		return false
	}
	commandID = strings.TrimSpace(commandID)
	if commandID == "" || window <= 0 {
		return false
	}
	cutoff := time.Now().UTC().Add(-window)

	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.audits) - 1; i >= 0; i-- {
		audit := s.audits[i]
		if audit.Time.Before(cutoff) {
			return false
		}
		if audit.CommandID != commandID {
			continue
		}
		if strings.HasPrefix(audit.Source, "http:") || strings.HasPrefix(audit.Source, "mqtt:") {
			return true
		}
	}
	return false
}

func (s *Store) ControlAudit(limit int) []ControlAudit {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.audits) {
		limit = len(s.audits)
	}
	out := make([]ControlAudit, 0, limit)
	for i := len(s.audits) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, s.audits[i])
	}
	return out
}

func StatePayload(device Device) map[string]any {
	payload := map[string]any{
		"source":       Source,
		"object":       device.ObjectName,
		"device":       device.Device,
		"device_slug":  device.DeviceSlug,
		"last_update":  mqttTime(device.LastUpdate),
		"raw_commands": device.RawCommands,
	}
	if device.JeedomID != "" {
		payload["jeedom_id"] = device.JeedomID
	}
	if device.JeedomLogicalID != "" {
		payload["jeedom_logical_id"] = device.JeedomLogicalID
	}
	if device.JeedomDeviceType != "" {
		payload["jeedom_device_type"] = device.JeedomDeviceType
	}
	if len(device.Actions) > 0 {
		payload["actions"] = device.Actions
	}
	for metric, value := range device.Values {
		payload[metric] = value
	}
	return payload
}

func copyDevice(device Device) Device {
	device.Values = copyAnyMap(device.Values)
	device.RawCommands = copyCommands(device.RawCommands)
	device.HAIdentifiers = append([]string(nil), device.HAIdentifiers...)
	device.LegacyDeviceSlugs = append([]string(nil), device.LegacyDeviceSlugs...)
	device.Actions = copyActions(device.Actions)
	return device
}

func copyAnyMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func copyCommands(commands map[string]Command) map[string]Command {
	if len(commands) == 0 {
		return map[string]Command{}
	}
	out := make(map[string]Command, len(commands))
	for key, value := range commands {
		out[key] = value
	}
	return out
}

func copyActions(actions map[string]Action) map[string]Action {
	if len(actions) == 0 {
		return map[string]Action{}
	}
	out := make(map[string]Action, len(actions))
	for key, value := range actions {
		out[key] = value
	}
	return out
}

func mqttTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func containsString(values []string, value string) bool {
	for _, current := range values {
		if current == value {
			return true
		}
	}
	return false
}

func mergeIdentity(fallback, resolved DeviceIdentity) DeviceIdentity {
	linkedToDifferentSlug := resolved.DeviceSlug != "" && resolved.DeviceSlug != fallback.DeviceSlug
	if resolved.DeviceSlug == "" {
		resolved.DeviceSlug = fallback.DeviceSlug
	}
	if resolved.DeviceName == "" {
		resolved.DeviceName = fallback.DeviceName
	}
	if resolved.BaseSlug == "" {
		resolved.BaseSlug = fallback.BaseSlug
	}
	if len(resolved.HAIdentifiers) == 0 {
		resolved.HAIdentifiers = fallback.HAIdentifiers
	}
	if resolved.HAManufacturer == "" {
		resolved.HAManufacturer = fallback.HAManufacturer
	}
	if resolved.HAModel == "" {
		resolved.HAModel = fallback.HAModel
	}
	if resolved.SuggestedArea == "" {
		resolved.SuggestedArea = fallback.SuggestedArea
	}
	if linkedToDifferentSlug {
		resolved.LegacyDeviceSlugs = mergeStringLists(resolved.LegacyDeviceSlugs, []string{fallback.DeviceSlug})
	}
	return resolved
}

func mergeStringLists(current, extra []string) []string {
	out := make([]string, 0, len(current)+len(extra))
	seen := make(map[string]struct{}, len(current)+len(extra))
	for _, values := range [][]string{current, extra} {
		for _, value := range values {
			value = Slug(value)
			if value == "" || value == "unknown" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			out = append(out, value)
		}
	}
	return out
}

func (s *Store) localDiscoveryDeviceSlug(discovery Discovery, baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "dev_" + shortHash(discovery.Name+discovery.EqLogicID)
	}
	groups := s.baseGroups[baseSlug]
	if len(groups) == 0 {
		return baseSlug
	}
	for _, slug := range groups {
		device := s.devices[slug]
		if device == nil {
			continue
		}
		if device.JeedomID != "" && discovery.EqLogicID != "" && device.JeedomID == discovery.EqLogicID {
			return slug
		}
		if device.JeedomLogicalID != "" && discovery.LogicalID != "" && device.JeedomLogicalID == discovery.LogicalID {
			return slug
		}
	}
	suffix := Slug(discovery.EqLogicID)
	if suffix == "" || suffix == "unknown" {
		suffix = shortHash(discovery.Topic)
	}
	return baseSlug + "_" + suffix
}

func firstDiscoveryCommandID(discovery Discovery) string {
	for _, command := range discovery.InfoCommands {
		if command.CommandID != "" {
			return command.CommandID
		}
	}
	for _, command := range discovery.Actions {
		if command.CommandID != "" {
			return command.CommandID
		}
	}
	return discovery.EqLogicID
}

func firstDiscoveryCommandName(discovery Discovery) string {
	for _, command := range discovery.InfoCommands {
		if command.Name != "" {
			return command.Name
		}
	}
	for _, command := range discovery.Actions {
		if command.Name != "" {
			return command.Name
		}
	}
	return ""
}

func normalizeDiscoveryAction(command DiscoveryCommand) string {
	switch strings.ToUpper(strings.TrimSpace(command.LogicalID)) {
	case "SWITCH_ON", "ON":
		return "on"
	case "SWITCH_OFF", "OFF":
		return "off"
	case "ARM":
		return "arm"
	case "NIGHT_MODE":
		return "night_mode"
	case "DISARM":
		return "disarm"
	case "PANIC":
		return "panic"
	}
	switch commandKey(command.Name) {
	case "on":
		return "on"
	case "off":
		return "off"
	case "armement", "arm":
		return "arm"
	case "modenuit", "nightmode":
		return "night_mode"
	case "desarmement", "disarm":
		return "disarm"
	case "panic":
		return "panic"
	case "arretdetectionincendie", "mutefiredetectors":
		return "mute_fire_detectors"
	default:
		return ""
	}
}

func NormalizeControlAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "1", "true", "on", "open", "enable":
		return "on"
	case "0", "false", "off", "close", "disable":
		return "off"
	default:
		return strings.ToLower(strings.TrimSpace(action))
	}
}

func controlAllowed(discovery Discovery, command DiscoveryCommand, action string) (bool, string) {
	if action != "on" && action != "off" {
		return false, "only on/off device controls are allowlisted"
	}
	switch strings.ToLower(firstNonEmpty(discovery.DeviceType, discovery.ApplyDevice)) {
	case "relay", "socket", "wallswitch", "lightswitch", "outlet":
		return true, ""
	case "waterstop":
		return false, "WaterStop controls are blocked by default"
	case "hub", "hub_2_plus", "hub2plus", "hub_plus", "hubplus":
		return false, "hub/security controls are blocked by default"
	default:
		return false, "device type is not allowlisted for Jeedom control"
	}
}
