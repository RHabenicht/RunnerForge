//go:build !windows

package main

import "fmt"

func main() {
	fmt.Println("RunnerForge desktop GUI runs on Windows; run `go test ./...` to validate the core.")
}
