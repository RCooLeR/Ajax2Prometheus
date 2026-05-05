package jeedom

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/RCooLeR/AjaxBridge/internal/devicecatalog"
)

func TestCatalogResolverLinksByCommandID(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{
		Account:          "A0F80D",
		Zone:             "8",
		Name:             "Живлення сервера",
		Room:             "Котельна",
		Kind:             "WallSwitch",
		JeedomCommandIDs: []string{"55", "56", "57"},
	})

	resolver := NewCatalogResolver(catalog, CatalogResolverConfig{})
	identity := resolver.Resolve(Event{CommandID: "56", DeviceName: "Серверна", CommandName: "Puissance"}, MappingFor(Event{CommandName: "Puissance"}))

	if identity.LinkedZone != "8" {
		t.Fatalf("LinkedZone = %q, want 8", identity.LinkedZone)
	}
	if len(identity.HAIdentifiers) != 1 || identity.HAIdentifiers[0] != "ajaxbridge_A0F80D_zone_8" {
		t.Fatalf("HAIdentifiers = %#v", identity.HAIdentifiers)
	}
	if identity.DeviceSlug != "sia_a0f80d_zone_8" {
		t.Fatalf("DeviceSlug = %q", identity.DeviceSlug)
	}
}

func TestCatalogResolverDisablesUnlinkedDiscoveryByDefault(t *testing.T) {
	resolver := NewCatalogResolver(devicecatalog.Empty(), CatalogResolverConfig{})
	identity := resolver.Resolve(Event{CommandID: "999", DeviceName: "Unknown", CommandName: "Voltage"}, MappingFor(Event{CommandName: "Voltage"}))

	if !identity.DiscoveryDisabled {
		t.Fatal("expected unlinked discovery disabled")
	}
}

func TestCatalogResolverLinksConfiguredAccountName(t *testing.T) {
	catalog := testCatalog(t, devicecatalog.Device{Account: "A0F80D", Zone: "1", Name: "Relay"})
	resolver := NewCatalogResolver(catalog, CatalogResolverConfig{AccountNames: []string{"Будинок"}})
	identity := resolver.Resolve(Event{CommandID: "1", DeviceName: "Будинок", CommandName: "Etat"}, MappingFor(Event{CommandName: "Etat"}))

	if identity.LinkedAccount != "A0F80D" || identity.LinkedZone != "" {
		t.Fatalf("identity = %#v", identity)
	}
	if len(identity.HAIdentifiers) != 1 || identity.HAIdentifiers[0] != "ajaxbridge_account_A0F80D" {
		t.Fatalf("HAIdentifiers = %#v", identity.HAIdentifiers)
	}
}

func testCatalog(t *testing.T, devices ...devicecatalog.Device) *devicecatalog.Catalog {
	t.Helper()
	path := filepath.Join(t.TempDir(), "devices.json")
	body, err := json.Marshal(devices)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := devicecatalog.Load(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
