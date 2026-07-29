# Project agent memory

This file is the project's committed home for project-intrinsic agent knowledge: build, test, release, architecture, and sharp-edge notes that should travel with the code.

- Add durable project-specific notes here as they are discovered through real work.

## Build / test

`go build ./...` (produces a stray `./herdr-tailscale` binary at repo root —
delete it or use `sh scripts/build.sh`, which builds to `./bin/herdr-tailscale`
as the manifest's `[[build]]` step does). `go test ./...` covers the peer
filter and idempotency logic without needing a running herdr instance.

## Testing against a real herdr server — DO NOT use the live default config

Do not run `herdr plugin link`, `herdr plugin install`, or `herdr plugin
action invoke` against the default `~/.config/herdr` — that is the captain's
live, in-daily-use session; any action fired there is real (real tabs, real
`--remote` SSH attaches). `HERDR_SOCKET_PATH` + `HERDR_CONFIG_PATH` env vars
isolate the RPC socket, letting you run a separate `herdr server` and drive
`herdr tab create` / `herdr pane run` / `herdr tab list` safely against a
throwaway workspace. But **plugin registration is not isolated**: an
"isolated" server's `herdr plugin list` still reads the real, shared
`~/.config/herdr/plugins/` — so `herdr plugin link`/`install` were never
exercised, even in isolation. See README.md's Verification section for what
was and wasn't verified end-to-end and why.

`HERDR_SOCKET_PATH` must be a short path — Unix socket paths are capped at
~104 bytes (`sun_path`); use `/tmp/...`, not a long scratch/session path.

## herdr CLI quirks discovered here

- `herdr tab create` has no way to run a startup command; use the
  `root_pane.pane_id` in its JSON response with `herdr pane run <pane_id>
  <command>` instead.
- `herdr pane run` prints **nothing** on success but a JSON error envelope
  on failure — asymmetric with every other `herdr <noun> <verb>` subcommand,
  which prints a JSON envelope either way. Don't reuse a generic
  parse-JSON-response helper for it without accounting for this.

## Maintaining this file

Keep this file for knowledge useful to almost every future agent session in this project.
Do not repeat what the codebase already shows; point to the authoritative file or command instead.
Prefer rewriting or pruning existing entries over appending new ones.
When updating this file, preserve this bar for all agents and keep entries concise.
