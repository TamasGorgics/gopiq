// Command gopiq runs the gopiq image-processing CLI.
package main

import (
	"os"

	"github.com/TamasGorgics/gopiq/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
