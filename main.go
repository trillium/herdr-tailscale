// Command herdr-tailscale is the entry point herdr-plugin.toml's
// attach-trusted-peers action invokes.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: herdr-tailscale attach-trusted-peers")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "attach-trusted-peers":
		if err := runAttachTrustedPeers(); err != nil {
			fmt.Fprintln(os.Stderr, "herdr-tailscale:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "herdr-tailscale: unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
}
