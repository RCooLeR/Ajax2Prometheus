package sia

import (
	"strings"

	"github.com/RCooLeR/Ajax2Prometheus/internal/event"
)

type eventInfo struct {
	class       event.Class
	action      string
	name        string
	description string
	source      string
	signal      string
	severity    string
	restore     bool
}

var siaCatalog = map[string]eventInfo{
	"AR": {class: event.ClassRestore, action: "power_restored", name: "External power restored", source: "system", signal: "power", severity: "info", restore: true},
	"AT": {class: event.ClassTrouble, action: "power_failure", name: "External power failure", source: "system", signal: "power", severity: "warning"},
	"BA": {class: event.ClassAlarm, action: "burglary_alarm", name: "Burglary alarm", source: "sensor", signal: "burglary", severity: "critical"},
	"BR": {class: event.ClassRestore, action: "burglary_restore", name: "Burglary alarm restored", source: "sensor", signal: "burglary", severity: "info", restore: true},
	"BS": {class: event.ClassTrouble, action: "accelerometer_failure", name: "Accelerometer failure", source: "sensor", signal: "accelerometer", severity: "warning"},
	"CA": {class: event.ClassArm, action: "armed_automatically", name: "Armed automatically", source: "automation", signal: "arming", severity: "info"},
	"CB": {class: event.ClassArm, action: "group_armed_automatically", name: "Group armed automatically", source: "automation", signal: "arming", severity: "info"},
	"CC": {class: event.ClassTrouble, action: "arming_failed", name: "Unsuccessful arming attempt", source: "automation", signal: "arming", severity: "warning"},
	"CD": {class: event.ClassTrouble, action: "group_arming_failed", name: "Unsuccessful group arming attempt", source: "automation", signal: "arming", severity: "warning"},
	"CF": {class: event.ClassArm, action: "armed_with_malfunctions", name: "Armed with malfunctions", source: "system", signal: "arming", severity: "warning"},
	"CG": {class: event.ClassArm, action: "group_armed", name: "Group armed", source: "user_or_device", signal: "arming", severity: "info"},
	"CL": {class: event.ClassArm, action: "armed", name: "Armed", source: "user_or_device", signal: "arming", severity: "info"},
	"FA": {class: event.ClassAlarm, action: "fire_alarm", name: "Fire or smoke alarm", source: "sensor", signal: "fire", severity: "critical"},
	"FH": {class: event.ClassRestore, action: "fire_restore", name: "Fire or smoke alarm restored", source: "sensor", signal: "fire", severity: "info", restore: true},
	"FJ": {class: event.ClassRestore, action: "fire_detector_restore", name: "Fire detector fault restored", source: "sensor", signal: "fire_detector", severity: "info", restore: true},
	"FS": {class: event.ClassTrouble, action: "smoke_chamber_dirty", name: "Smoke chamber dirty or fire detector fault", source: "sensor", signal: "fire_detector", severity: "warning"},
	"FT": {class: event.ClassTrouble, action: "fire_detector_failure", name: "Fire detector hardware failure", source: "sensor", signal: "fire_detector", severity: "warning"},
	"FX": {class: event.ClassRestore, action: "smoke_chamber_restore", name: "Smoke chamber restored", source: "sensor", signal: "fire_detector", severity: "info", restore: true},
	"GA": {class: event.ClassAlarm, action: "gas_or_co_alarm", name: "Gas or CO alarm", source: "sensor", signal: "gas_or_co", severity: "critical"},
	"GH": {class: event.ClassRestore, action: "gas_or_co_restore", name: "Gas or CO alarm restored", source: "sensor", signal: "gas_or_co", severity: "info", restore: true},
	"HA": {class: event.ClassAlarm, action: "duress_disarm", name: "Duress disarm", source: "user_or_keypad", signal: "duress", severity: "critical"},
	"JA": {class: event.ClassTrouble, action: "password_bruteforce", name: "Password guessing attempt", source: "keypad", signal: "access", severity: "warning"},
	"KA": {class: event.ClassAlarm, action: "temperature_alarm", name: "High temperature or rapid temperature rise alarm", source: "sensor", signal: "temperature", severity: "critical"},
	"KH": {class: event.ClassRestore, action: "temperature_restore", name: "Temperature alarm restored", source: "sensor", signal: "temperature", severity: "info", restore: true},
	"MA": {class: event.ClassAlarm, action: "medical_alarm", name: "Medical alarm", source: "button_or_transmitter", signal: "medical", severity: "critical"},
	"MR": {class: event.ClassRestore, action: "medical_restore", name: "Medical alarm restored", source: "button_or_transmitter", signal: "medical", severity: "info", restore: true},
	"NB": {class: event.ClassNight, action: "night_mode_on_with_malfunctions", name: "Night mode activated with malfunctions", source: "automation", signal: "night_mode", severity: "warning"},
	"NC": {class: event.ClassNight, action: "night_mode_on_automatically", name: "Night mode activated automatically", source: "automation", signal: "night_mode", severity: "info"},
	"ND": {class: event.ClassAlarm, action: "night_mode_off_duress", name: "Night mode deactivated by duress", source: "user_or_keypad", signal: "duress", severity: "critical"},
	"NE": {class: event.ClassTrouble, action: "night_mode_failed", name: "Unsuccessful Night mode activation attempt", source: "automation", signal: "night_mode", severity: "warning"},
	"NF": {class: event.ClassNight, action: "night_mode_on_with_malfunctions", name: "Night mode activated with malfunctions", source: "user_or_device", signal: "night_mode", severity: "warning"},
	"NL": {class: event.ClassNight, action: "night_mode_on", name: "Night mode activated", source: "user_or_device", signal: "night_mode", severity: "info"},
	"NO": {class: event.ClassDisarm, action: "night_mode_off_automatically", name: "Night mode deactivated automatically", source: "automation", signal: "night_mode", severity: "info"},
	"NP": {class: event.ClassDisarm, action: "night_mode_off", name: "Night mode deactivated", source: "user_or_device", signal: "night_mode", severity: "info"},
	"OA": {class: event.ClassDisarm, action: "disarmed_automatically", name: "Disarmed automatically", source: "automation", signal: "arming", severity: "info"},
	"OB": {class: event.ClassDisarm, action: "group_disarmed_automatically", name: "Group disarmed automatically", source: "automation", signal: "arming", severity: "info"},
	"OG": {class: event.ClassDisarm, action: "group_disarmed", name: "Group disarmed", source: "user_or_device", signal: "arming", severity: "info"},
	"OP": {class: event.ClassDisarm, action: "disarmed", name: "Disarmed", source: "user_or_device", signal: "arming", severity: "info"},
	"PA": {class: event.ClassAlarm, action: "panic_alarm", name: "Panic button alarm", source: "button_or_user", signal: "panic", severity: "critical"},
	"PF": {class: event.ClassTrouble, action: "photo_channel_lost", name: "Photo channel connection lost", source: "sensor", signal: "connectivity", severity: "warning"},
	"PH": {class: event.ClassRestore, action: "panic_restore", name: "Panic alarm restored", source: "button_or_transmitter", signal: "panic", severity: "info", restore: true},
	"PO": {class: event.ClassRestore, action: "photo_channel_restored", name: "Photo channel connection restored", source: "sensor", signal: "connectivity", severity: "info", restore: true},
	"QA": {class: event.ClassAlarm, action: "emergency_alarm", name: "Emergency alarm", source: "button_or_transmitter", signal: "emergency", severity: "critical"},
	"QB": {class: event.ClassTrouble, action: "device_bypassed", name: "Device bypassed or deactivated", source: "device", signal: "bypass", severity: "warning"},
	"QU": {class: event.ClassRestore, action: "device_bypass_restored", name: "Device bypass restored or reactivated", source: "device", signal: "bypass", severity: "info", restore: true},
	"RB": {class: event.ClassCommon, action: "firmware_update_started", name: "Firmware update started", source: "system_or_device", signal: "firmware", severity: "info"},
	"RP": {class: event.ClassTest, action: "periodic_test", name: "Periodic test or null message", source: "hub", signal: "supervision", severity: "info"},
	"RS": {class: event.ClassCommon, action: "firmware_updated", name: "Firmware updated", source: "system_or_device", signal: "firmware", severity: "info"},
	"RX": {class: event.ClassTest, action: "manual_test", name: "Manual test", source: "hub", signal: "supervision", severity: "info"},
	"RY": {class: event.ClassTest, action: "test_restore", name: "Test restore", source: "hub", signal: "supervision", severity: "info", restore: true},
	"SM": {class: event.ClassTamper, action: "device_moved", name: "Device moved", source: "sensor", signal: "tamper", severity: "warning"},
	"TA": {class: event.ClassTamper, action: "tamper_alarm", name: "Tamper alarm or masking detected", source: "hub_or_device", signal: "tamper", severity: "warning"},
	"TB": {class: event.ClassTrouble, action: "tamper_bypassed", name: "Tamper bypassed or deactivated", source: "device", signal: "tamper_bypass", severity: "warning"},
	"TR": {class: event.ClassRestore, action: "tamper_restore", name: "Tamper restored or masking cleared", source: "hub_or_device", signal: "tamper", severity: "info", restore: true},
	"TU": {class: event.ClassRestore, action: "tamper_bypass_restored", name: "Tamper bypass restored or reactivated", source: "device", signal: "tamper_bypass", severity: "info", restore: true},
	"WA": {class: event.ClassAlarm, action: "water_leak_alarm", name: "Water leak alarm", source: "sensor", signal: "water_leak", severity: "critical"},
	"WH": {class: event.ClassRestore, action: "water_leak_restore", name: "Water leak restored", source: "sensor", signal: "water_leak", severity: "info", restore: true},
	"XC": {class: event.ClassRestore, action: "device_connection_restored", name: "Device connection restored", source: "device", signal: "connectivity", severity: "info", restore: true},
	"XH": {class: event.ClassRestore, action: "interference_restored", name: "Interference level OK", source: "system", signal: "interference", severity: "info", restore: true},
	"XI": {class: event.ClassCommon, action: "factory_reset", name: "Factory reset", source: "system_or_device", signal: "configuration", severity: "warning"},
	"XL": {class: event.ClassTrouble, action: "device_connection_lost", name: "Device connection lost", source: "device", signal: "connectivity", severity: "warning"},
	"XQ": {class: event.ClassTrouble, action: "interference_high", name: "Radio interference high", source: "system", signal: "interference", severity: "warning"},
	"XR": {class: event.ClassRestore, action: "battery_restored", name: "Device battery charged", source: "device", signal: "battery", severity: "info", restore: true},
	"XT": {class: event.ClassTrouble, action: "battery_low", name: "Device battery low", source: "device", signal: "battery", severity: "warning"},
	"YA": {class: event.ClassRestore, action: "battery_connected", name: "Battery connected", source: "hub", signal: "battery", severity: "info", restore: true},
	"YC": {class: event.ClassTrouble, action: "hub_offline", name: "Hub offline", source: "hub", signal: "connectivity", severity: "critical"},
	"YG": {class: event.ClassCommon, action: "settings_changed", name: "Settings changed", source: "system", signal: "configuration", severity: "info"},
	"YK": {class: event.ClassRestore, action: "hub_online", name: "Hub online", source: "hub", signal: "connectivity", severity: "info", restore: true},
	"YM": {class: event.ClassTrouble, action: "battery_missing", name: "Battery missing", source: "hub", signal: "battery", severity: "warning"},
	"YP": {class: event.ClassTrouble, action: "external_power_failure", name: "External power failure", source: "device", signal: "power", severity: "warning"},
	"YQ": {class: event.ClassRestore, action: "external_power_restored", name: "External power restored", source: "device", signal: "power", severity: "info", restore: true},
	"YR": {class: event.ClassRestore, action: "battery_charged", name: "Battery charged", source: "hub_or_detector", signal: "battery", severity: "info", restore: true},
	"YS": {class: event.ClassTrouble, action: "monitoring_connection_lost", name: "Monitoring station connection lost", source: "hub", signal: "connectivity", severity: "warning"},
	"YT": {class: event.ClassTrouble, action: "battery_low", name: "Battery low", source: "hub_or_detector", signal: "battery", severity: "warning"},
	"ZZ": {class: event.ClassCommon, action: "turned_off", name: "Turned off", source: "system_or_device", signal: "power", severity: "warning"},
	"ZY": {class: event.ClassCommon, action: "turned_on", name: "Turned on", source: "system_or_device", signal: "power", severity: "info"},
}

var cidCatalog = map[string]eventInfo{
	"E100": {class: event.ClassAlarm, action: "medical_alarm", name: "Medical alarm", source: "button_or_transmitter", signal: "medical", severity: "critical"},
	"R100": {class: event.ClassRestore, action: "medical_restore", name: "Medical alarm restored", source: "button_or_transmitter", signal: "medical", severity: "info", restore: true},
	"E110": {class: event.ClassAlarm, action: "fire_alarm", name: "Fire alarm", source: "sensor", signal: "fire", severity: "critical"},
	"R110": {class: event.ClassRestore, action: "fire_restore", name: "Fire alarm restored", source: "sensor", signal: "fire", severity: "info", restore: true},
	"E111": {class: event.ClassAlarm, action: "smoke_alarm", name: "Smoke alarm", source: "sensor", signal: "smoke", severity: "critical"},
	"R111": {class: event.ClassRestore, action: "smoke_restore", name: "Smoke alarm restored", source: "sensor", signal: "smoke", severity: "info", restore: true},
	"E114": {class: event.ClassAlarm, action: "temperature_alarm", name: "Temperature alarm", source: "sensor", signal: "temperature", severity: "critical"},
	"R114": {class: event.ClassRestore, action: "temperature_restore", name: "Temperature alarm restored", source: "sensor", signal: "temperature", severity: "info", restore: true},
	"E120": {class: event.ClassAlarm, action: "panic_alarm", name: "Panic button alarm", source: "button_or_user", signal: "panic", severity: "critical"},
	"R120": {class: event.ClassRestore, action: "panic_restore", name: "Panic alarm restored", source: "button_or_transmitter", signal: "panic", severity: "info", restore: true},
	"E130": {class: event.ClassAlarm, action: "burglary_alarm", name: "Burglary alarm", source: "sensor", signal: "burglary", severity: "critical"},
	"R130": {class: event.ClassRestore, action: "burglary_restore", name: "Burglary alarm restored", source: "sensor", signal: "burglary", severity: "info", restore: true},
	"E137": {class: event.ClassTamper, action: "masking_detected", name: "Masking detected", source: "sensor", signal: "tamper", severity: "warning"},
	"R137": {class: event.ClassRestore, action: "masking_restored", name: "Masking cleared", source: "sensor", signal: "tamper", severity: "info", restore: true},
	"E144": {class: event.ClassTamper, action: "tamper_alarm", name: "Tamper alarm", source: "device", signal: "tamper", severity: "warning"},
	"R144": {class: event.ClassRestore, action: "tamper_restore", name: "Tamper restored", source: "device", signal: "tamper", severity: "info", restore: true},
	"E145": {class: event.ClassTamper, action: "hub_tamper_alarm", name: "Hub lid open", source: "hub", signal: "tamper", severity: "warning"},
	"R145": {class: event.ClassRestore, action: "hub_tamper_restore", name: "Hub lid closed", source: "hub", signal: "tamper", severity: "info", restore: true},
	"E151": {class: event.ClassAlarm, action: "gas_alarm", name: "Gas alarm", source: "sensor", signal: "gas", severity: "critical"},
	"R151": {class: event.ClassRestore, action: "gas_restore", name: "Gas alarm restored", source: "sensor", signal: "gas", severity: "info", restore: true},
	"E154": {class: event.ClassAlarm, action: "water_leak_alarm", name: "Water leak alarm", source: "sensor", signal: "water_leak", severity: "critical"},
	"R154": {class: event.ClassRestore, action: "water_leak_restore", name: "Water leak restored", source: "sensor", signal: "water_leak", severity: "info", restore: true},
	"E162": {class: event.ClassAlarm, action: "co_alarm", name: "Carbon monoxide alarm", source: "sensor", signal: "co", severity: "critical"},
	"R162": {class: event.ClassRestore, action: "co_restore", name: "Carbon monoxide restored", source: "sensor", signal: "co", severity: "info", restore: true},
	"E300": {class: event.ClassTrouble, action: "battery_missing", name: "Battery missing", source: "hub", signal: "battery", severity: "warning"},
	"R300": {class: event.ClassRestore, action: "battery_connected", name: "Battery connected", source: "hub", signal: "battery", severity: "info", restore: true},
	"E301": {class: event.ClassTrouble, action: "external_power_failure", name: "External power failure", source: "system_or_device", signal: "power", severity: "warning"},
	"R301": {class: event.ClassRestore, action: "external_power_restored", name: "External power restored", source: "system_or_device", signal: "power", severity: "info", restore: true},
	"E302": {class: event.ClassTrouble, action: "battery_low", name: "Battery low", source: "hub", signal: "battery", severity: "warning"},
	"R302": {class: event.ClassRestore, action: "battery_charged", name: "Battery charged", source: "hub", signal: "battery", severity: "info", restore: true},
	"E337": {class: event.ClassTrouble, action: "device_external_power_failure", name: "Device external power failure", source: "device", signal: "power", severity: "warning"},
	"R337": {class: event.ClassRestore, action: "device_external_power_restored", name: "Device external power restored", source: "device", signal: "power", severity: "info", restore: true},
	"E344": {class: event.ClassTrouble, action: "interference_high", name: "Radio interference high", source: "system", signal: "interference", severity: "warning"},
	"R344": {class: event.ClassRestore, action: "interference_restored", name: "Radio interference restored", source: "system", signal: "interference", severity: "info", restore: true},
	"E350": {class: event.ClassTrouble, action: "hub_offline", name: "Hub offline", source: "hub", signal: "connectivity", severity: "critical"},
	"R350": {class: event.ClassRestore, action: "hub_online", name: "Hub online", source: "hub", signal: "connectivity", severity: "info", restore: true},
	"E377": {class: event.ClassTrouble, action: "accelerometer_failure", name: "Accelerometer failure", source: "sensor", signal: "accelerometer", severity: "warning"},
	"R377": {class: event.ClassRestore, action: "accelerometer_restore", name: "Accelerometer restored", source: "sensor", signal: "accelerometer", severity: "info", restore: true},
	"E381": {class: event.ClassTrouble, action: "device_connection_lost", name: "Device connection lost", source: "device", signal: "connectivity", severity: "warning"},
	"R381": {class: event.ClassRestore, action: "device_connection_restored", name: "Device connection restored", source: "device", signal: "connectivity", severity: "info", restore: true},
	"E384": {class: event.ClassTrouble, action: "device_battery_low", name: "Device battery low", source: "device", signal: "battery", severity: "warning"},
	"R384": {class: event.ClassRestore, action: "device_battery_restored", name: "Device battery restored", source: "device", signal: "battery", severity: "info", restore: true},
	"E389": {class: event.ClassTrouble, action: "hardware_failure", name: "Hardware failure", source: "device", signal: "hardware", severity: "warning"},
	"R389": {class: event.ClassRestore, action: "hardware_restore", name: "Hardware restored", source: "device", signal: "hardware", severity: "info", restore: true},
	"E391": {class: event.ClassTrouble, action: "photo_channel_lost", name: "Photo channel lost", source: "sensor", signal: "connectivity", severity: "warning"},
	"R391": {class: event.ClassRestore, action: "photo_channel_restored", name: "Photo channel restored", source: "sensor", signal: "connectivity", severity: "info", restore: true},
	"E393": {class: event.ClassTrouble, action: "smoke_chamber_dirty", name: "Smoke chamber dirty", source: "sensor", signal: "fire_detector", severity: "warning"},
	"R393": {class: event.ClassRestore, action: "smoke_chamber_restore", name: "Smoke chamber restored", source: "sensor", signal: "fire_detector", severity: "info", restore: true},
	"E400": {class: event.ClassDisarm, action: "disarmed_by_user", name: "Disarmed by user", source: "user", signal: "arming", severity: "info"},
	"R400": {class: event.ClassArm, action: "armed_by_user", name: "Armed by user", source: "user", signal: "arming", severity: "info"},
	"E403": {class: event.ClassDisarm, action: "disarmed_automatically", name: "Disarmed automatically", source: "automation", signal: "arming", severity: "info"},
	"R403": {class: event.ClassArm, action: "armed_automatically", name: "Armed automatically", source: "automation", signal: "arming", severity: "info"},
	"E409": {class: event.ClassDisarm, action: "disarmed_by_device", name: "Disarmed by device", source: "keypad_or_spacecontrol", signal: "arming", severity: "info"},
	"R409": {class: event.ClassArm, action: "armed_by_device", name: "Armed by device", source: "keypad_or_spacecontrol", signal: "arming", severity: "info"},
	"E423": {class: event.ClassAlarm, action: "duress_disarm", name: "Duress disarm", source: "user_or_keypad", signal: "duress", severity: "critical"},
	"E441": {class: event.ClassDisarm, action: "night_mode_off_group", name: "Night mode deactivated for group", source: "user", signal: "night_mode", severity: "info"},
	"R441": {class: event.ClassNight, action: "night_mode_on_group", name: "Night mode activated for group", source: "user", signal: "night_mode", severity: "info"},
	"E442": {class: event.ClassDisarm, action: "night_mode_off_group", name: "Night mode deactivated for group", source: "device", signal: "night_mode", severity: "info"},
	"R442": {class: event.ClassNight, action: "night_mode_on_group", name: "Night mode activated for group", source: "device", signal: "night_mode", severity: "info"},
	"E455": {class: event.ClassTrouble, action: "arming_failed", name: "Unsuccessful arming attempt", source: "system", signal: "arming", severity: "warning"},
	"E461": {class: event.ClassTrouble, action: "password_bruteforce", name: "Password guessing attempt", source: "keypad", signal: "access", severity: "warning"},
	"E627": {class: event.ClassCommon, action: "settings_changed", name: "Settings changed", source: "system", signal: "configuration", severity: "info"},
}

func enrichEvent(evt *event.Normalized) {
	if evt == nil {
		return
	}

	info, ok := lookupEventInfo(evt)
	if ok {
		evt.EventClass = info.class
		evt.EventAction = info.action
		evt.EventName = info.name
		evt.Description = info.description
		evt.Source = info.source
		evt.Signal = info.signal
		evt.Severity = info.severity
		return
	}

	evt.EventClass = fallbackClass(evt.EventCode)
	evt.EventAction = fallbackAction(evt.EventClass)
	evt.EventName = fallbackName(evt.EventCode)
	evt.Source = "unknown"
	evt.Signal = "unknown"
	evt.Severity = fallbackSeverity(evt.EventClass)
}

func lookupEventInfo(evt *event.Normalized) (eventInfo, bool) {
	if evt.ContactID != "" {
		if info, ok := cidCatalog[strings.ToUpper(evt.ContactID)]; ok {
			info = mergeSIAHints(evt.EventCode, info)
			return info, true
		}
	}
	if info, ok := siaCatalog[strings.ToUpper(evt.EventCode)]; ok {
		return info, true
	}
	return eventInfo{}, false
}

func mergeSIAHints(code string, info eventInfo) eventInfo {
	switch strings.ToUpper(code) {
	case "NL", "NC", "NF", "NB":
		info.class = event.ClassNight
		info.signal = "night_mode"
	case "NP", "NO":
		info.class = event.ClassDisarm
		info.signal = "night_mode"
	case "HA", "ND":
		info.class = event.ClassAlarm
		info.signal = "duress"
		info.severity = "critical"
	case "PA":
		info.signal = "panic"
	}
	return info
}

func fallbackClass(code string) event.Class {
	switch {
	case strings.HasPrefix(code, "CID1"):
		return event.ClassAlarm
	case strings.HasPrefix(code, "CID3"):
		return event.ClassRestore
	case strings.HasPrefix(code, "CID4"):
		return event.ClassDisarm
	default:
		return event.ClassUnknown
	}
}

func fallbackAction(class event.Class) string {
	switch class {
	case event.ClassAlarm:
		return "alarm"
	case event.ClassRestore:
		return "restore"
	case event.ClassArm:
		return "arm"
	case event.ClassDisarm:
		return "disarm"
	case event.ClassNight:
		return "night_mode"
	case event.ClassTamper:
		return "tamper"
	case event.ClassTrouble:
		return "trouble"
	case event.ClassTest:
		return "test"
	case event.ClassCommon:
		return "common"
	default:
		return "unknown"
	}
}

func fallbackName(code string) string {
	if code == "" {
		return "Unknown event"
	}
	return "Unknown event code " + code
}

func fallbackSeverity(class event.Class) string {
	switch class {
	case event.ClassAlarm:
		return "critical"
	case event.ClassTamper, event.ClassTrouble:
		return "warning"
	default:
		return "info"
	}
}
