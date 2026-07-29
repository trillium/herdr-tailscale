package main

import (
	"fmt"
	"testing"
)

// Fixtures below are trimmed from real `herdr tab list` / `herdr tab create`
// output, captured against an isolated `herdr server` instance (separate
// HERDR_SOCKET_PATH/HERDR_CONFIG_PATH, own workspace) during development —
// see README.md's "Verification" section.

const tabListFixture = `{"id":"cli:tab:list","result":{"tabs":[` +
	`{"agent_status":"unknown","focused":false,"label":"1","number":1,"pane_count":1,"tab_id":"wX:t1","workspace_id":"wX"},` +
	`{"agent_status":"unknown","focused":false,"label":"tailscale: mini1","number":2,"pane_count":1,"tab_id":"wX:t2","workspace_id":"wX"}` +
	`],"type":"tab_list"}}`

const tabCreateFixture = `{"id":"cli:tab:create","result":{"root_pane":{"agent_status":"unknown","cwd":"/Users/x","focused":false,` +
	`"foreground_cwd":"/Users/x","pane_id":"wX:p2","revision":0,"scroll":{"max_offset_from_bottom":0,"offset_from_bottom":0,"viewport_rows":23},` +
	`"tab_id":"wX:t2","terminal_id":"term_657b4a3b162bf1e","workspace_id":"wX"},` +
	`"tab":{"agent_status":"unknown","focused":false,"label":"tailscale: mini2","number":2,"pane_count":1,"tab_id":"wX:t2","workspace_id":"wX"},` +
	`"type":"tab_created"}}`

const errorFixture = `{"id":"cli:tab:create","error":{"code":"not_found","message":"workspace wZ not found"}}`

const paneRunErrorFixture = `{"error":{"code":"pane_not_found","message":"pane wX:p999 not found"},"id":"cli:request"}`

type exitError struct{ msg string }

func (e *exitError) Error() string { return e.msg }

func fakeCLI(response string, callErr error) *herdrCLI {
	return &herdrCLI{
		workspaceID: "wX",
		run: func(args ...string) ([]byte, error) {
			if callErr != nil {
				return nil, callErr
			}
			return []byte(response), nil
		},
	}
}

func TestListTabLabels(t *testing.T) {
	cli := fakeCLI(tabListFixture, nil)
	labels, err := cli.listTabLabels()
	if err != nil {
		t.Fatalf("listTabLabels() error = %v", err)
	}
	want := map[string]bool{"1": true, "tailscale: mini1": true}
	if len(labels) != len(want) {
		t.Fatalf("listTabLabels() = %v, want %v", labels, want)
	}
	for k := range want {
		if !labels[k] {
			t.Errorf("listTabLabels() missing label %q", k)
		}
	}
}

func TestCreateTab(t *testing.T) {
	cli := fakeCLI(tabCreateFixture, nil)
	tabID, paneID, err := cli.createTab("tailscale: mini2")
	if err != nil {
		t.Fatalf("createTab() error = %v", err)
	}
	if tabID != "wX:t2" {
		t.Errorf("createTab() tabID = %q, want %q", tabID, "wX:t2")
	}
	if paneID != "wX:p2" {
		t.Errorf("createTab() paneID = %q, want %q", paneID, "wX:p2")
	}
}

func TestCallSurfacesHerdrError(t *testing.T) {
	cli := fakeCLI(errorFixture, nil)
	_, _, err := cli.createTab("tailscale: mini2")
	if err == nil {
		t.Fatal("createTab() error = nil, want error for workspace not found")
	}
}

func TestCallSurfacesExecError(t *testing.T) {
	cli := fakeCLI("", fmt.Errorf("exec: \"herdr\": executable file not found in $PATH"))
	_, _, err := cli.createTab("tailscale: mini2")
	if err == nil {
		t.Fatal("createTab() error = nil, want error when herdr binary is missing")
	}
}

// runInPane has to be treated differently from the other subcommands: on
// success `herdr pane run` prints nothing at all (confirmed empirically
// against a real herdr server), so an empty response is not a parse failure.
func TestRunInPaneEmptyResponseIsSuccess(t *testing.T) {
	cli := &herdrCLI{
		workspaceID: "wX",
		run: func(args ...string) ([]byte, error) {
			return []byte{}, nil
		},
	}
	if err := cli.runInPane("wX:p2", "herdr --remote mini1"); err != nil {
		t.Fatalf("runInPane() error = %v, want nil for empty-but-successful response", err)
	}
}

func TestRunInPaneSurfacesHerdrError(t *testing.T) {
	cli := &herdrCLI{
		workspaceID: "wX",
		run: func(args ...string) ([]byte, error) {
			return []byte(paneRunErrorFixture), &exitError{"exit status 1"}
		},
	}
	err := cli.runInPane("wX:p999", "herdr --remote mini1")
	if err == nil {
		t.Fatal("runInPane() error = nil, want error for unknown pane id")
	}
}
