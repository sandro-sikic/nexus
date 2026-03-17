package main

import (
	"flag"
	"fmt"
	"os"

	"runner/config"
	"runner/ui"
)

func main() {
	cfgPath := flag.String("config", "runner.yaml", "path to runner config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "runner: %v\n", err)
		os.Exit(1)
	}

	if err := ui.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "runner: %v\n", err)
		os.Exit(1)
	}
}
