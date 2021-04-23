package main

import (
	"context"
	"fmt"

	"moul.io/motd"
)

func doVersion(ctx context.Context, args []string) error {
	fmt.Print(motd.Default())
	fmt.Println("version: n/a")
	return nil
}
