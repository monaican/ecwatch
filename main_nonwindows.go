//go:build !windows

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "ecwatch only supports Windows (build with GOOS=windows)")
	os.Exit(1)
}
