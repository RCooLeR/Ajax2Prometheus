package config

import "testing"

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
