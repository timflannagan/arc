package main

import (
	"fmt"
	"os"

	"github.com/agentregistry-dev/ar/pkg/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
