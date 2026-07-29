package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// peersNeedingAttach is the idempotency check, isolated from any herdr or
// tailscale I/O: given the trusted peers and the tab labels already open in
// the workspace, it returns only the peers that still need a tab. Re-running
// attach-trusted-peers is safe because a peer whose label is already present
// is skipped rather than given a second tab.
func peersNeedingAttach(peers []tailscalePeer, existingLabels map[string]bool) []tailscalePeer {
	var need []tailscalePeer
	for _, p := range peers {
		if !existingLabels[p.tabLabel()] {
			need = append(need, p)
		}
	}
	return need
}

func runAttachTrustedPeers() error {
	workspaceID := os.Getenv("HERDR_WORKSPACE_ID")
	if workspaceID == "" {
		return fmt.Errorf("HERDR_WORKSPACE_ID is not set; run this as a herdr plugin action (herdr plugin action invoke attach-trusted-peers --plugin trillium.herdr-tailscale)")
	}

	statusBytes, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return fmt.Errorf("tailscale status --json: %w", err)
	}
	var status tailscaleStatus
	if err := json.Unmarshal(statusBytes, &status); err != nil {
		return fmt.Errorf("parsing tailscale status --json: %w", err)
	}

	trusted := trustedPeers(status)
	if len(trusted) == 0 {
		fmt.Println("herdr-tailscale: no trusted Tailscale peers online")
		return nil
	}

	cli := newHerdrCLI(workspaceID)
	existingLabels, err := cli.listTabLabels()
	if err != nil {
		return fmt.Errorf("listing existing tabs: %w", err)
	}

	toAttach := peersNeedingAttach(trusted, existingLabels)
	if len(toAttach) == 0 {
		fmt.Println("herdr-tailscale: all trusted peers already have an attached tab")
		return nil
	}

	for _, p := range toAttach {
		tabID, paneID, err := cli.createTab(p.tabLabel())
		if err != nil {
			fmt.Fprintf(os.Stderr, "herdr-tailscale: creating tab for %s: %v\n", p.HostName, err)
			continue
		}
		if err := cli.runInPane(paneID, "herdr --remote "+p.sshTarget()); err != nil {
			fmt.Fprintf(os.Stderr, "herdr-tailscale: attaching remote for %s: %v\n", p.HostName, err)
			continue
		}
		fmt.Printf("herdr-tailscale: attached %s (tab %s)\n", p.HostName, tabID)
	}
	return nil
}
