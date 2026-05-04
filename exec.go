//go:build !windows

package main

import (
	"os"
)

func exec() error {
	_, err := os.Create("/tmp/live")
	if err != nil {
		panic(err)
	}
	defer os.Remove("/tmp/live")
	return command()
}
