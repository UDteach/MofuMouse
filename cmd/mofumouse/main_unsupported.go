//go:build !windows

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "MofuMouse MVP is Windows-only.")
	os.Exit(1)
}
