package main

import (
	"sort"
	"strings"
)

// tailscaleStatus mirrors the fields we need from `tailscale status --json`.
// It intentionally only lists what we read — the real payload has many more.
type tailscaleStatus struct {
	Self tailscalePeer            `json:"Self"`
	Peer map[string]tailscalePeer `json:"Peer"`
}

type tailscalePeer struct {
	ID       string `json:"ID"`
	HostName string `json:"HostName"`
	DNSName  string `json:"DNSName"`
	UserID   int64  `json:"UserID"`
	Online   bool   `json:"Online"`
}

// sshTarget is the value passed to `herdr --remote`: the peer's MagicDNS name
// with the trailing dot `tailscale status --json` appends stripped off.
func (p tailscalePeer) sshTarget() string {
	if p.DNSName != "" {
		return strings.TrimSuffix(p.DNSName, ".")
	}
	return p.HostName
}

// tabLabel is the Herdr tab label we look for (and set) to detect that a peer
// already has an attached remote tab. Keep this in sync with attach.go.
func (p tailscalePeer) tabLabel() string {
	return "tailscale: " + p.HostName
}

// trustedPeers is the single seam that decides which Tailscale peers this
// plugin is willing to auto-attach to. The current default — online, owned by
// the same tailnet identity as this machine (Self), excluding this machine
// itself — is deliberately simple so a follow-up can tighten it (e.g. to a
// specific ACL tag) by editing only this function.
func trustedPeers(status tailscaleStatus) []tailscalePeer {
	var out []tailscalePeer
	for _, p := range status.Peer {
		if !p.Online {
			continue
		}
		if p.ID == status.Self.ID {
			continue
		}
		if p.UserID != status.Self.UserID {
			// Owned by a different tailnet identity: a node someone else
			// shared into this tailnet ("shared-in"), not one of ours.
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].HostName < out[j].HostName })
	return out
}
