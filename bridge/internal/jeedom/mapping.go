package jeedom

import (
	"strings"
	"unicode"
)

const (
	ComponentSensor       = "sensor"
	ComponentBinarySensor = "binary_sensor"
	ComponentSwitch       = "switch"
	ComponentButton       = "button"
)

type Mapping struct {
	Metric         string
	Component      string
	EntityName     string
	Unit           string
	DeviceClass    string
	StateClass     string
	EntityCategory string
	Numeric        bool
	Binary         bool
}

func MappingFor(evt Event) Mapping {
	command := firstNonEmpty(evt.CommandName, evt.Name, evt.CommandID)
	unit := strings.TrimSpace(RepairText(evt.Unit))
	switch commandKey(command) {
	case "puissance", "power":
		return sensorMapping("power_w", "Power", firstNonEmpty(unit, "W"), "power", "measurement", true)
	case "consommation", "energy", "consumption":
		return sensorMapping("energy_kwh", "Energy", firstNonEmpty(unit, "kWh"), "energy", "total_increasing", true)
	case "courant", "current":
		return sensorMapping("current_a", "Current", firstNonEmpty(unit, "A"), "current", "measurement", true)
	case "tension", "voltage":
		return sensorMapping("voltage_v", "Voltage", firstNonEmpty(unit, "V"), "voltage", "measurement", true)
	case "temperature", "temp":
		return sensorMapping("temperature_c", "Temperature", firstNonEmpty(unit, "\u00b0C"), "temperature", "measurement", true)
	case "batterie", "battery":
		return sensorMapping("battery_percent", "Battery", firstNonEmpty(unit, "%"), "battery", "measurement", true)
	case "etatdelabatterie", "batterystate":
		return diagnosticSensorMapping("battery_state", "Battery state", unit, "", "", false)
	case "signal", "rssi":
		if strings.EqualFold(strings.TrimSpace(evt.Subtype), "numeric") || strings.EqualFold(unit, "dBm") {
			return sensorMapping("signal_dbm", "Signal", firstNonEmpty(unit, "dBm"), "signal_strength", "measurement", true)
		}
		return diagnosticSensorMapping("signal_level", "Signal", unit, "", "", false)
	case "intensitedusignalcellulaire", "cellularsignalstrength":
		return sensorMapping("signal_dbm", "Signal", firstNonEmpty(unit, "dBm"), "signal_strength", "measurement", true)
	case "humidite", "humidity":
		return sensorMapping("humidity_percent", "Humidity", firstNonEmpty(unit, "%"), "humidity", "measurement", true)
	case "etat", "state", "status":
		if strings.EqualFold(strings.TrimSpace(evt.Subtype), "binary") {
			return binaryMapping("state", "State", "")
		}
		return sensorMapping("state", "State", unit, "", "", false)
	case "sourceevenement", "eventsource":
		return diagnosticSensorMapping("event_source", "Event source", unit, "", "", false)
	case "evenement", "event":
		return diagnosticSensorMapping("event", "Event", unit, "", "", false)
	case "codeevenement", "eventcode":
		return diagnosticSensorMapping("event_code", "Event code", unit, "", "", false)
	case "enligne", "online":
		return diagnosticBinaryMapping("online", "Online", "connectivity")
	case "trafique", "tampered", "tamper", "sabotage":
		return binaryMapping("tamper", "Tamper", "tamper")
	case "alimentationsecteur", "externallypowered", "externalpower", "mainspower":
		return binaryMapping("external_power", "External power", "power")
	case "gsm", "cms", "ethernet":
		return diagnosticBinaryMapping(MetricName(command, "cmd_"+evt.CommandID), firstNonEmpty(translateCommandLabel(command), titleName(command)), "connectivity")
	case "typereseaugsm", "gsmnetworktype":
		return diagnosticSensorMapping("gsm_network_type", "GSM network type", unit, "", "", false)
	case "donneescellulairesactives", "cellulardataactive":
		return diagnosticBinaryMapping("cellular_data_active", "Cellular data active", "connectivity")
	case "ouverture", "opening":
		return binaryMapping("opening", "Opening", "opening")
	case "porte", "door":
		return binaryMapping("door", "Door", "door")
	case "fuite", "leak":
		return binaryMapping("leak", "Leak", "moisture")
	default:
		return fallbackMapping(evt, command, unit)
	}
}

func sensorMapping(metric, name, unit, deviceClass, stateClass string, numeric bool) Mapping {
	return Mapping{
		Metric:      metric,
		Component:   ComponentSensor,
		EntityName:  name,
		Unit:        unit,
		DeviceClass: deviceClass,
		StateClass:  stateClass,
		Numeric:     numeric,
	}
}

func diagnosticSensorMapping(metric, name, unit, deviceClass, stateClass string, numeric bool) Mapping {
	mapping := sensorMapping(metric, name, unit, deviceClass, stateClass, numeric)
	mapping.EntityCategory = "diagnostic"
	return mapping
}

func binaryMapping(metric, name, deviceClass string) Mapping {
	return Mapping{
		Metric:      metric,
		Component:   ComponentBinarySensor,
		EntityName:  name,
		DeviceClass: deviceClass,
		Binary:      true,
	}
}

func diagnosticBinaryMapping(metric, name, deviceClass string) Mapping {
	mapping := binaryMapping(metric, name, deviceClass)
	mapping.EntityCategory = "diagnostic"
	return mapping
}

func fallbackMapping(evt Event, command, unit string) Mapping {
	metric := MetricName(command, "cmd_"+evt.CommandID)
	name := firstNonEmpty(translateCommandLabel(command), titleName(command))
	subtype := strings.ToLower(strings.TrimSpace(evt.Subtype))
	switch subtype {
	case "numeric":
		return sensorMapping(metric+"_value", name, unit, "", "measurement", true)
	case "binary":
		return binaryMapping(metric, name, "")
	default:
		return sensorMapping(metric, name, unit, "", "", false)
	}
}

func commandKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(RepairText(value)))
	var b strings.Builder
	for _, r := range value {
		r = unicode.ToLower(foldLatin(r))
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func titleName(value string) string {
	value = strings.TrimSpace(RepairText(value))
	if value == "" {
		return "Value"
	}
	parts := strings.Fields(strings.ReplaceAll(value, "_", " "))
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
