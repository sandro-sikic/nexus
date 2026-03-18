package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"nexus/config"
	"nexus/ui"
)

func main() {
	cfgPath := flag.String("config", "nexus.yaml", "path to nexus config file")
	wizard := flag.Bool("wizard", false, "open the setup/edit wizard before launching")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		if !errors.Is(err, config.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "nexus: %v\n", err)
			os.Exit(1)
		}

		// Config file missing — launch the setup wizard (create mode).
		result, wizErr := ui.RunWizard(*cfgPath)
		if wizErr != nil {
			fmt.Fprintf(os.Stderr, "nexus: wizard error: %v\n", wizErr)
			os.Exit(1)
		}
		if result.Aborted {
			fmt.Fprintln(os.Stderr, "nexus: setup aborted.")
			os.Exit(0)
		}
		if result.SaveErr != "" {
			fmt.Fprintf(os.Stderr, "nexus: could not save config: %v\n", result.SaveErr)
			os.Exit(1)
		}

		// Use the freshly created config directly — no need to re-read disk.
		cfg = result.Config
	} else if *wizard {
		// Config exists and --wizard was requested — launch in edit mode.
		result, wizErr := ui.RunWizardEdit(*cfgPath, cfg)
		if wizErr != nil {
			fmt.Fprintf(os.Stderr, "nexus: wizard error: %v\n", wizErr)
			os.Exit(1)
		}
		if result.Aborted {
			fmt.Fprintln(os.Stderr, "nexus: edit aborted.")
			os.Exit(0)
		}
		if result.SaveErr != "" {
			fmt.Fprintf(os.Stderr, "nexus: could not save config: %v\n", result.SaveErr)
			os.Exit(1)
		}

		// Re-read the updated config from disk so per-command run_mode
		// inheritance is re-applied correctly.
		cfg, err = config.Load(*cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nexus: %v\n", err)
			os.Exit(1)
		}
	}

	if err := ui.Run(cfg, *cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "nexus: %v\n", err)
		os.Exit(1)
	}
}
