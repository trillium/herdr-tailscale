# herdr-tailscale

A [Herdr](https://herdr.dev) plugin that auto-attaches remote tabs for
trusted peers on your Tailscale network, so every machine's Herdr session
can reach every other machine's session without any manual setup.

## What it does

The plugin registers one action, `attach-trusted-peers`:

1. Runs `tailscale status --json` and parses the peer list.
2. Filters to **trusted peers**: online, owned by the same tailnet identity
   as this machine, excluding this machine itself and any node someone else
   has shared into the tailnet from elsewhere ("shared-in"). This filter is
   isolated in a single function, `trustedPeers` in
   [`tailscale.go`](./tailscale.go), so it can be tightened later (e.g. to a
   specific Tailscale ACL tag) without touching anything else.
3. For each trusted peer that doesn't already have an attached tab in the
   current workspace, opens a new (unfocused) tab labeled `tailscale:
   <hostname>` and runs `herdr --remote <peer>` in it.
4. Is **idempotent** — it detects an already-attached peer by its tab label
   before creating anything, so re-running the action (including on every
   new shell, see below) never creates a duplicate tab. See
   `peersNeedingAttach` in [`attach.go`](./attach.go).

## Install

```bash
herdr plugin install trillium/herdr-tailscale
```

For local development, build and link a checkout instead:

```bash
go build -o bin/herdr-tailscale .
herdr plugin link .
```

## Running it automatically on every new session

Herdr has no native "on session start" plugin event today — only
`worktree.created`, `worktree.opened`, and `terminal.closed` exist. To get
the "boot into shared mode every time" behavior the captain wants, invoke
the action explicitly from each machine's shell profile
(`~/.zshrc`/`~/.bashrc`), so it runs on every new shell/Herdr session:

```sh
herdr plugin action invoke attach-trusted-peers --plugin trillium.herdr-tailscale >/dev/null 2>&1
```

Add that one line to the shell profile on **macbook, mini1, mini2, and
mini3**. It's a no-op (and safe to run repeatedly) outside a herdr pane —
the action requires `$HERDR_WORKSPACE_ID`, which herdr only sets for
processes it runs itself, so a plain terminal shell simply gets a "not
running inside herdr" error rather than doing anything. Redirect stdout so
that error doesn't clutter a non-herdr shell's startup.

## Verification

The peer-filtering (`trustedPeers`) and idempotency (`peersNeedingAttach`)
logic, plus the `herdr` CLI response parsing, are covered by `go test ./...`
— no running herdr instance required.

The end-to-end mechanism — `herdr tab create` → find the new tab's pane via
its `root_pane.pane_id` → `herdr pane run <pane_id> "herdr --remote <peer>"`
→ `herdr tab list` to detect an already-attached peer — was verified against
a **real, isolated `herdr server`** instance (a separate `HERDR_SOCKET_PATH`
and `HERDR_CONFIG_PATH`, started and stopped for testing only), using this
machine's real `tailscale status --json` peers, in a workspace created and
torn down solely for the test. That confirmed:

- `herdr tab create` has no `--cmd` flag; the way to run a startup command in
  a newly created tab is `herdr pane run <pane_id> <command>`, using the
  `root_pane.pane_id` the `tab.create` response already returns.
- `herdr pane run` prints nothing on success and a JSON error envelope on
  failure — asymmetric with every other CLI subcommand used here, which
  print a JSON envelope either way. (This was caught as a real bug during
  end-to-end testing: `runInPane` originally tried to parse the empty
  success output as JSON and failed. Fixed in `herdrcli.go`.)
- Running the action twice against the same workspace attached all 5 online
  trusted peers on the first run and created zero additional tabs on the
  second, confirming idempotency end-to-end, not just in the unit tests.
- `$HERDR_WORKSPACE_ID` (along with `$HERDR_TAB_ID`/`$HERDR_PANE_ID`) is
  exported into a pane's environment by herdr itself, so the action can
  target "the current workspace" without needing a `--workspace` flag from
  the caller.

**What could not be verified end-to-end, and why:** `herdr plugin link`
(and `herdr plugin install`) do not appear to be scoped by
`HERDR_SOCKET_PATH`/`HERDR_CONFIG_PATH` — an isolated `herdr server`
process still listed the real, live plugins from
`~/.config/herdr/plugins/` and `~/.config/herdr/local-plugins/` in `herdr
plugin list`. That means linking this plugin, even against an "isolated"
server, would register it in the same shared plugin directory the captain's
live session reads from. Per this repo's safety guidance, the full
`herdr plugin link` → `herdr plugin action invoke attach-trusted-peers
--plugin trillium.herdr-tailscale` path was therefore **not** exercised
against either the live session or the isolated one — only the underlying
`herdr tab create` / `herdr pane run` / `herdr tab list` mechanics were,
directly over the isolated socket. If live verification of the full
plugin-registration path is wanted before merge, that's a call for the
captain: it would need to happen against the real `~/.config/herdr`, however
briefly, since no additional isolation layer for plugin registration was
found.
