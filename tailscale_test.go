package main

import (
	"reflect"
	"testing"
)

func TestTrustedPeers(t *testing.T) {
	self := tailscalePeer{ID: "self-id", HostName: "macbook", UserID: 1}

	status := tailscaleStatus{
		Self: self,
		Peer: map[string]tailscalePeer{
			"online-same-tailnet": {
				ID: "peer-1", HostName: "mini1", DNSName: "mini1.example.ts.net.",
				UserID: 1, Online: true,
			},
			"offline-same-tailnet": {
				ID: "peer-2", HostName: "mini2", DNSName: "mini2.example.ts.net.",
				UserID: 1, Online: false,
			},
			"online-foreign-tailnet": {
				ID: "peer-3", HostName: "someone-elses-box", DNSName: "shared.example.ts.net.",
				UserID: 2, Online: true,
			},
			"self-somehow-in-peer-map": {
				ID: "self-id", HostName: "macbook", UserID: 1, Online: true,
			},
			"another-online-same-tailnet": {
				ID: "peer-4", HostName: "mini3", DNSName: "mini3.example.ts.net.",
				UserID: 1, Online: true,
			},
		},
	}

	got := trustedPeers(status)

	var gotHosts []string
	for _, p := range got {
		gotHosts = append(gotHosts, p.HostName)
	}

	want := []string{"mini1", "mini3"} // sorted by hostname
	if !reflect.DeepEqual(gotHosts, want) {
		t.Fatalf("trustedPeers() hosts = %v, want %v", gotHosts, want)
	}
}

func TestTrustedPeersEmpty(t *testing.T) {
	status := tailscaleStatus{
		Self: tailscalePeer{ID: "self-id", UserID: 1},
		Peer: map[string]tailscalePeer{},
	}
	if got := trustedPeers(status); len(got) != 0 {
		t.Fatalf("trustedPeers() = %v, want empty", got)
	}
}

func TestPeerSSHTarget(t *testing.T) {
	cases := []struct {
		name string
		peer tailscalePeer
		want string
	}{
		{
			name: "strips trailing dot from DNSName",
			peer: tailscalePeer{HostName: "mini1", DNSName: "mini1.example.ts.net."},
			want: "mini1.example.ts.net",
		},
		{
			name: "falls back to HostName when DNSName is empty",
			peer: tailscalePeer{HostName: "mini1"},
			want: "mini1",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.peer.sshTarget(); got != c.want {
				t.Errorf("sshTarget() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestPeerTabLabel(t *testing.T) {
	p := tailscalePeer{HostName: "mini1"}
	if got, want := p.tabLabel(), "tailscale: mini1"; got != want {
		t.Errorf("tabLabel() = %q, want %q", got, want)
	}
}
