package main

import (
	"fmt"
	"git.tyss.io/cj3636/dman/internal/cli"
	"os"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
