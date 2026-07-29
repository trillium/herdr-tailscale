package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// herdrCLI talks to the running herdr instance via the `herdr` CLI binary,
// scoped to one workspace. herdr's socket API is JSON, and its CLI wraps that
// API and prints the same JSON envelope to stdout, so we shell out rather than
// speak the socket protocol directly — this ties us to the stable, documented
// CLI surface (`herdr tab create`, `herdr tab list`, `herdr pane run`) instead
// of the socket wire format.
type herdrCLI struct {
	workspaceID string

	// run executes `herdr <args...>` and returns its stdout. Overridable in
	// tests so the idempotency/attach logic can be exercised without a real
	// herdr server.
	run func(args ...string) ([]byte, error)
}

func newHerdrCLI(workspaceID string) *herdrCLI {
	return &herdrCLI{
		workspaceID: workspaceID,
		run: func(args ...string) ([]byte, error) {
			return exec.Command("herdr", args...).Output()
		},
	}
}

type herdrTab struct {
	TabID string `json:"tab_id"`
	Label string `json:"label"`
}

type herdrPane struct {
	PaneID string `json:"pane_id"`
}

type herdrErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// herdrEnvelope is the `{"id": ..., "result": ...}` / `{"id": ..., "error":
// ...}` shape every herdr CLI subcommand that talks to the socket API prints.
type herdrEnvelope struct {
	Result json.RawMessage `json:"result"`
	Error  *herdrErrorBody `json:"error"`
}

func (h *herdrCLI) call(result any, args ...string) error {
	out, err := h.run(args...)
	if err != nil {
		return fmt.Errorf("herdr %v: %w", args, err)
	}
	var env herdrEnvelope
	if err := json.Unmarshal(out, &env); err != nil {
		return fmt.Errorf("herdr %v: unparseable output: %w", args, err)
	}
	if env.Error != nil {
		return fmt.Errorf("herdr %v: %s: %s", args, env.Error.Code, env.Error.Message)
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(env.Result, result)
}

// listTabLabels returns the labels of every tab already open in the plugin's
// workspace, used to decide which trusted peers still need a tab.
func (h *herdrCLI) listTabLabels() (map[string]bool, error) {
	var r struct {
		Tabs []herdrTab `json:"tabs"`
	}
	if err := h.call(&r, "tab", "list", "--workspace", h.workspaceID); err != nil {
		return nil, err
	}
	labels := make(map[string]bool, len(r.Tabs))
	for _, t := range r.Tabs {
		labels[t.Label] = true
	}
	return labels, nil
}

// createTab opens a new, unfocused tab in the workspace and returns its
// root pane id — the pane `runInPane` runs the peer's remote-attach command
// in. Unfocused so attaching several peers in the background (e.g. from a
// shell-profile hook) doesn't steal focus from whatever the user is doing.
func (h *herdrCLI) createTab(label string) (tabID, paneID string, err error) {
	var r struct {
		Tab      herdrTab  `json:"tab"`
		RootPane herdrPane `json:"root_pane"`
	}
	if err := h.call(&r, "tab", "create", "--workspace", h.workspaceID, "--label", label, "--no-focus"); err != nil {
		return "", "", err
	}
	return r.Tab.TabID, r.RootPane.PaneID, nil
}

// runInPane runs command in paneID as if typed at the prompt (text + enter).
// Unlike the other subcommands here, `herdr pane run` prints nothing on
// success (confirmed empirically — see README.md's "Verification" section) but
// still prints a JSON error envelope and exits non-zero on failure (e.g. an
// unknown pane id), so on error we try to surface that message.
func (h *herdrCLI) runInPane(paneID, command string) error {
	out, err := h.run("pane", "run", paneID, command)
	if err == nil {
		return nil
	}
	var env herdrEnvelope
	if jsonErr := json.Unmarshal(out, &env); jsonErr == nil && env.Error != nil {
		return fmt.Errorf("herdr pane run %s: %s: %s", paneID, env.Error.Code, env.Error.Message)
	}
	return fmt.Errorf("herdr pane run %s: %w", paneID, err)
}
