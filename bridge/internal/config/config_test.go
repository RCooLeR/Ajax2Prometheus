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

func TestFromEnvUsesJeedomDefaults(t *testing.T) {
	cfg := FromEnv()
	if cfg.JeedomEnabled {
		t.Fatal("JeedomEnabled default = true, want false")
	}
	if cfg.JeedomEventTopic != "jeedom/cmd/event/#" {
		t.Fatalf("JeedomEventTopic = %q, want default", cfg.JeedomEventTopic)
	}
	if cfg.JeedomDiscoveryTopic != "jeedom/discovery/eqLogic/#" {
		t.Fatalf("JeedomDiscoveryTopic = %q, want default", cfg.JeedomDiscoveryTopic)
	}
	if cfg.JeedomStateTopicPrefix != "ajaxbridge/jeedom" {
		t.Fatalf("JeedomStateTopicPrefix = %q, want default", cfg.JeedomStateTopicPrefix)
	}
	if cfg.JeedomEmptyValuePolicy != "keep_last" {
		t.Fatalf("JeedomEmptyValuePolicy = %q, want keep_last", cfg.JeedomEmptyValuePolicy)
	}
	if cfg.JeedomSampleDir != "" {
		t.Fatalf("JeedomSampleDir = %q, want empty production default", cfg.JeedomSampleDir)
	}
	if cfg.JeedomDiscoverUnlinked {
		t.Fatal("JeedomDiscoverUnlinked default = true, want false")
	}
	if len(cfg.JeedomAccountNames) != 0 {
		t.Fatalf("JeedomAccountNames = %#v, want empty", cfg.JeedomAccountNames)
	}
	if cfg.JeedomControlsEnabled {
		t.Fatal("JeedomControlsEnabled default = true, want false")
	}
	if cfg.JeedomSetTopicPrefix != "jeedom/cmd/set" {
		t.Fatalf("JeedomSetTopicPrefix = %q, want default", cfg.JeedomSetTopicPrefix)
	}
	if cfg.JeedomControlPayload != "1" {
		t.Fatalf("JeedomControlPayload = %q, want default 1", cfg.JeedomControlPayload)
	}
}

func TestValidateRequiresMQTTBrokerWhenJeedomEnabled(t *testing.T) {
	cfg := FromEnv()
	cfg.JeedomEnabled = true
	cfg.MQTTBroker = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
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
