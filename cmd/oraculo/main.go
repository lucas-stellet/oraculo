// cmd/oraculo/main.go
package main

import (
	"fmt"
	"os"

	"github.com/lucas/oraculo/src/cli"
)

var version = "dev"

func main() {
	cmd := cli.NewRoot(version)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
