package main

import (
	"os"

	"github.com/canpok1/ai-feed/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
