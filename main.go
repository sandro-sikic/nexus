package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"nexus/config"
	"nexus/runner"
	"nexus/ui"
)

func main() {
	// Setup signal handling to cleanup processes on exit
	setupSignalHandling()

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

	// If a positional argument was given, fuzzy-match and run the first hit.
	if args := flag.Args(); len(args) > 0 {
		query := args[0]
		if err := ui.RunFirstMatch(cfg, query); err != nil {
			fmt.Fprintf(os.Stderr, "nexus: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := ui.Run(cfg, *cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "nexus: %v\n", err)
		os.Exit(1)
	}
}

// setupSignalHandling registers signal handlers to cleanup processes on exit.
func setupSignalHandling() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nnexus: received signal %v, cleaning up...\n", sig)

		// Kill all registered processes
		if err := runner.GetGlobalRegistry().KillAll(); err != nil {
			fmt.Fprintf(os.Stderr, "nexus: error killing processes: %v\n", err)
		}

		os.Exit(1)
	}()
}
