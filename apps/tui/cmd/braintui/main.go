package main

import (
	"fmt"
	"os"

	"github.com/reeinharrrd/brain/tui/internal/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	a, err := app.New()
	if err != nil {
		return err
	}
	return a.Run()
}
