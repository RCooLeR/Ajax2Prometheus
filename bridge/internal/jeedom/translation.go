package jeedom

import "strings"

func EnglishCommandName(command string, mapping Mapping, fallback string) string {
	if mapping.EntityName != "" {
		return mapping.EntityName
	}
	if translated := translateCommandLabel(command); translated != "" {
		return translated
	}
	return titleName(firstNonEmpty(command, fallback, "Value"))
}

func EnglishActionName(action, rawName string) string {
	switch NormalizeControlAction(action) {
	case "on":
		return "On"
	case "off":
		return "Off"
	case "impulse":
		return "Impulse"
	case "arm":
		return "Arm"
	case "night_mode":
		return "Night mode"
	case "disarm":
		return "Disarm"
	case "panic":
		return "Panic"
	case "mute_fire_detectors":
		return "Mute fire detectors"
	}
	if translated := translateCommandLabel(rawName); translated != "" {
		return translated
	}
	return titleName(firstNonEmpty(rawName, action, "Action"))
}

func translateCommandLabel(value string) string {
	switch commandKey(value) {
	case "puissance", "power":
		return "Power"
	case "consommation", "energy", "consumption":
		return "Energy"
	case "courant", "current":
		return "Current"
	case "tension", "voltage":
		return "Voltage"
	case "temperature", "temp":
		return "Temperature"
	case "batterie", "battery":
		return "Battery"
	case "etatdelabatterie", "batterystate":
		return "Battery state"
	case "signal", "rssi":
		return "Signal"
	case "intensitedusignalcellulaire", "cellularsignalstrength":
		return "Cellular signal strength"
	case "humidite", "humidity":
		return "Humidity"
	case "etat", "state", "status":
		return "State"
	case "sourceevenement", "eventsource":
		return "Event source"
	case "evenement", "event":
		return "Event"
	case "codeevenement", "eventcode":
		return "Event code"
	case "enligne", "online":
		return "Online"
	case "trafique", "tampered", "tamper", "sabotage":
		return "Tamper"
	case "alimentationsecteur", "externallypowered", "externalpower", "mainspower":
		return "External power"
	case "gsm":
		return "GSM"
	case "cms":
		return "CMS"
	case "ethernet":
		return "Ethernet"
	case "typereseaugsm", "gsmnetworktype":
		return "GSM network type"
	case "donneescellulairesactives", "cellulardataactive":
		return "Cellular data active"
	case "ouverture", "opening":
		return "Opening"
	case "porte", "door":
		return "Door"
	case "fuite", "leak":
		return "Leak"
	case "armement", "arm":
		return "Arm"
	case "modenuit", "nightmode":
		return "Night mode"
	case "desarmement", "disarm":
		return "Disarm"
	case "panic":
		return "Panic"
	case "impulse", "impulsion", "pulse", "toggle":
		return "Impulse"
	case "arretdetectionincendie", "mutefiredetectors":
		return "Mute fire detectors"
	default:
		return ""
	}
}

func EnglishValue(value string) string {
	switch strings.ToUpper(strings.TrimSpace(RepairText(value))) {
	case "ARME", "ARMEMENT", "ARMED":
		return "ARMED"
	case "DESARME", "DESARMEMENT", "DISARMED":
		return "DISARMED"
	case "MODE_NUIT", "NIGHT_MODE":
		return "NIGHT_MODE"
	case "PASSIF", "PASSIVE":
		return "PASSIVE"
	case "ACTIF", "ACTIVE":
		return "ACTIVE"
	case "FORT", "STRONG":
		return "STRONG"
	case "FAIBLE", "WEAK":
		return "WEAK"
	case "MOYEN", "AVERAGE", "MEDIUM":
		return "MEDIUM"
	case "CHARGE", "CHARGED":
		return "CHARGED"
	case "DECHARGE", "DISCHARGED":
		return "DISCHARGED"
	default:
		return RepairText(value)
	}
}
