package main

import (
	"errors"
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
		if !errors.Is(err, config.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "runner: %v\n", err)
			os.Exit(1)
		}

		// Config file missing — launch the setup wizard.
		result, wizErr := ui.RunWizard(*cfgPath)
		if wizErr != nil {
			fmt.Fprintf(os.Stderr, "runner: wizard error: %v\n", wizErr)
			os.Exit(1)
		}
		if result.Aborted {
			fmt.Fprintln(os.Stderr, "runner: setup aborted.")
			os.Exit(0)
		}
		if result.SaveErr != "" {
			fmt.Fprintf(os.Stderr, "runner: could not save config: %v\n", result.SaveErr)
			os.Exit(1)
		}

		// Use the freshly created config directly — no need to re-read disk.
		cfg = result.Config
	}

	if err := ui.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "runner: %v\n", err)
		os.Exit(1)
	}
}
