//go:build !windows

package main

import "fmt"

func main() { fmt.Println("iMirror GUI is Windows-only: GOOS=windows go build") }
