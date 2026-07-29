package main

import (
	"reflect"
	"testing"
)

func TestPeersNeedingAttach(t *testing.T) {
	peers := []tailscalePeer{
		{HostName: "mini1"},
		{HostName: "mini2"},
		{HostName: "mini3"},
	}

	existingLabels := map[string]bool{
		"tailscale: mini2": true, // already attached
	}

	got := peersNeedingAttach(peers, existingLabels)

	var gotHosts []string
	for _, p := range got {
		gotHosts = append(gotHosts, p.HostName)
	}

	want := []string{"mini1", "mini3"}
	if !reflect.DeepEqual(gotHosts, want) {
		t.Fatalf("peersNeedingAttach() hosts = %v, want %v", gotHosts, want)
	}
}

func TestPeersNeedingAttachIdempotent(t *testing.T) {
	peers := []tailscalePeer{{HostName: "mini1"}}

	// First run: nothing attached yet, mini1 needs a tab.
	first := peersNeedingAttach(peers, map[string]bool{})
	if len(first) != 1 {
		t.Fatalf("first run: got %d peers needing attach, want 1", len(first))
	}

	// Simulate the tab that would have been created, then re-run: nothing
	// should need attaching a second time.
	existingLabels := map[string]bool{first[0].tabLabel(): true}
	second := peersNeedingAttach(peers, existingLabels)
	if len(second) != 0 {
		t.Fatalf("second run: got %d peers needing attach, want 0 (idempotent)", len(second))
	}
}

func TestPeersNeedingAttachAllAlreadyPresent(t *testing.T) {
	peers := []tailscalePeer{{HostName: "mini1"}, {HostName: "mini2"}}
	existingLabels := map[string]bool{
		"tailscale: mini1": true,
		"tailscale: mini2": true,
	}
	got := peersNeedingAttach(peers, existingLabels)
	if len(got) != 0 {
		t.Fatalf("peersNeedingAttach() = %v, want empty", got)
	}
}

func TestPeersNeedingAttachNoneExisting(t *testing.T) {
	peers := []tailscalePeer{{HostName: "mini1"}, {HostName: "mini2"}}
	got := peersNeedingAttach(peers, map[string]bool{})
	if len(got) != 2 {
		t.Fatalf("peersNeedingAttach() = %v, want 2 peers", got)
	}
}
