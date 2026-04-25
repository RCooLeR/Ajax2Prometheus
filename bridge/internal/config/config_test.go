package config

import "testing"

func TestFromEnvPrefersAjaxBridgeVariables(t *testing.T) {
	t.Setenv("AJAXBRIDGE_MQTT_TOPIC_PREFIX", "ajaxbridge-new")
	t.Setenv("AJAX2PROM_MQTT_TOPIC_PREFIX", "ajax2prom-old")

	if got := FromEnv().MQTTTopicPrefix; got != "ajaxbridge-new" {
		t.Fatalf("MQTTTopicPrefix = %q, want %q", got, "ajaxbridge-new")
	}
}

func TestFromEnvSupportsLegacyAjax2PromVariables(t *testing.T) {
	t.Setenv("AJAX2PROM_MQTT_CLIENT_ID", "legacy-client")

	if got := FromEnv().MQTTClientID; got != "legacy-client" {
		t.Fatalf("MQTTClientID = %q, want %q", got, "legacy-client")
	}
}

func TestFromEnvUsesStandaloneDevicesPathDefault(t *testing.T) {
	if got := FromEnv().DevicesPath; got != "data/devices.json" {
		t.Fatalf("DevicesPath = %q, want %q", got, "data/devices.json")
	}
}

func TestForwardAddressesSupportsSlicesAndCSV(t *testing.T) {
	cfg := Config{SIAForwardAddrs: []string{"127.0.0.1:1,127.0.0.1:2", " 127.0.0.1:3 "}}
	got := cfg.ForwardAddresses()
	want := []string{"127.0.0.1:1", "127.0.0.1:2", "127.0.0.1:3"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
