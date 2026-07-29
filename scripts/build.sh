#!/bin/sh
#
# Build herdr-tailscale for `herdr plugin install`. herdr runs this as the
# manifest's [[build]] step after cloning the repo, in the plugin root. The
# result is ./bin/herdr-tailscale, which the manifest's action invokes.
#
# This is a fresh repo with no release pipeline yet, so unlike herdr-plus we
# don't fall back to downloading a prebuilt binary — a local Go toolchain is
# required for now.

set -eu

mkdir -p bin

if ! command -v go >/dev/null 2>&1; then
	echo "herdr-tailscale: no Go toolchain found; install Go (https://go.dev/dl/) and retry" >&2
	exit 1
fi

echo "herdr-tailscale: building from source (go build)…" >&2
exec go build -o bin/herdr-tailscale .
