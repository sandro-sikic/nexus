package runner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"nexus/config"
)

// LogLine is a single line of output from a background process.
type LogLine struct {
	Text  string
	IsErr bool
}

// BackgroundProc holds a running background process and its log stream.
type BackgroundProc struct {
	Cmd   *exec.Cmd
	Lines chan LogLine
	done  chan struct{}
	once  sync.Once
}

// Wait blocks until the background process exits.
func (b *BackgroundProc) Wait() {
	<-b.done
}

// buildCmdFromShell constructs an exec.Cmd from a raw shell string and an optional working directory.
func buildCmdFromShell(shellCmd string, dir string) *exec.Cmd {
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = newWindowsCmd(shellCmd)
	} else {
		c = exec.Command("sh", "-c", shellCmd)
	}
	if dir != "" {
		c.Dir = dir
	}
	c.Env = os.Environ()
	return c
}

// buildCmd constructs an exec.Cmd from a config.Command.
// Deprecated: use buildCmdFromShell directly when iterating over Steps().
func buildCmd(cmd config.Command) *exec.Cmd {
	return buildCmdFromShell(cmd.Command, cmd.Dir)
}

// Handoff runs each step of the command sequentially in the raw terminal,
// stopping on the first failure.
func Handoff(cmd config.Command) error {
	steps := cmd.AllSteps()
	if len(steps) == 0 {
		return nil
	}
	for i, step := range steps {
		c := buildCmdFromShell(step, cmd.Dir)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			if len(steps) > 1 {
				return fmt.Errorf("step %d/%d %q: %w", i+1, len(steps), step, err)
			}
			return err
		}
	}
	return nil
}

// Stream runs each step of the command sequentially and sends output lines to
// the provided channel. It closes the channel when all steps complete or a
// step fails.
func Stream(cmd config.Command, lines chan<- LogLine) error {
	steps := cmd.AllSteps()
	if len(steps) == 0 {
		close(lines)
		return nil
	}

	go func() {
		defer close(lines)
		for i, step := range steps {
			// Announce which step is running when there are multiple steps.
			if len(steps) > 1 {
				lines <- LogLine{
					Text:  fmt.Sprintf("[%d/%d] %s", i+1, len(steps), step),
					IsErr: false,
				}
			}
			if err := streamOne(buildCmdFromShell(step, cmd.Dir), lines); err != nil {
				lines <- LogLine{
					Text:  fmt.Sprintf("error: %v", err),
					IsErr: true,
				}
				return
			}
		}
	}()

	return nil
}

// streamOne runs a single exec.Cmd and pipes its stdout/stderr to lines.
// It blocks until the process exits.
func streamOne(c *exec.Cmd, lines chan<- LogLine) error {
	stdout, err := c.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := c.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	var wg sync.WaitGroup
	scanLines := func(r interface{ Read([]byte) (int, error) }, isErr bool) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB per line
		for scanner.Scan() {
			lines <- LogLine{Text: scanner.Text(), IsErr: isErr}
		}
	}

	wg.Add(2)
	go scanLines(stdout, false)
	go scanLines(stderr, true)

	wg.Wait()
	return c.Wait()
}

// RunBackground starts all steps of the command in the background, running
// them sequentially, and returns a BackgroundProc whose Lines channel receives
// all output.
func RunBackground(cmd config.Command) (*BackgroundProc, error) {
	steps := cmd.AllSteps()

	lines := make(chan LogLine, 256)
	done := make(chan struct{})

	bp := &BackgroundProc{
		Lines: lines,
		done:  done,
	}

	if len(steps) == 0 {
		go func() {
			close(lines)
			bp.once.Do(func() { close(done) })
		}()
		return bp, nil
	}

	// For BackgroundProc.Cmd we expose the first step's exec.Cmd (legacy field).
	// The goroutine below owns the actual execution of all steps.
	firstCmd := buildCmdFromShell(steps[0], cmd.Dir)
	bp.Cmd = firstCmd

	go func() {
		defer func() {
			close(lines)
			bp.once.Do(func() { close(done) })
		}()

		for i, step := range steps {
			if len(steps) > 1 {
				lines <- LogLine{
					Text:  fmt.Sprintf("[%d/%d] %s", i+1, len(steps), step),
					IsErr: false,
				}
			}

			var c *exec.Cmd
			if i == 0 {
				c = firstCmd
			} else {
				c = buildCmdFromShell(step, cmd.Dir)
			}

			stdout, err := c.StdoutPipe()
			if err != nil {
				lines <- LogLine{Text: fmt.Sprintf("error: stdout pipe: %v", err), IsErr: true}
				return
			}
			stderr, err := c.StderrPipe()
			if err != nil {
				lines <- LogLine{Text: fmt.Sprintf("error: stderr pipe: %v", err), IsErr: true}
				return
			}
			if err := c.Start(); err != nil {
				lines <- LogLine{Text: fmt.Sprintf("error: start: %v", err), IsErr: true}
				return
			}

			var wg sync.WaitGroup
			scan := func(r interface{ Read([]byte) (int, error) }, isErr bool) {
				defer wg.Done()
				scanner := bufio.NewScanner(r)
				scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB per line
				for scanner.Scan() {
					lines <- LogLine{Text: scanner.Text(), IsErr: isErr}
				}
			}
			wg.Add(2)
			go scan(stdout, false)
			go scan(stderr, true)
			wg.Wait()

			if err := c.Wait(); err != nil {
				lines <- LogLine{
					Text:  fmt.Sprintf("error: step %d/%d %q: %v", i+1, len(steps), step, err),
					IsErr: true,
				}
				return
			}
		}
	}()

	return bp, nil
}

// FormatCommand returns a short display string for a command.
func FormatCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) > 4 {
		return strings.Join(parts[:4], " ") + " …"
	}
	return cmd
}

// StreamWithBackground runs steps with support for background steps.
// Steps marked with Background=true are started in goroutines and continue running
// while foreground steps execute sequentially. All output is streamed to the lines channel.
func StreamWithBackground(cmd config.Command, lines chan LogLine) error {
	allSteps := cmd.Steps
	if len(allSteps) == 0 {
		close(lines)
		return nil
	}

	go func() {
		defer close(lines)

		// Track background processes so we can show their status
		var bgProcs []*BackgroundProc

		for i, step := range allSteps {
			if step.Background {
				// Start this step in the background
				lines <- LogLine{
					Text:  fmt.Sprintf("[%d/%d] Starting background: %s", i+1, len(allSteps), step.Command),
					IsErr: false,
				}

				bp, err := runBackgroundStep(step.Command, cmd.Dir, lines)
				if err != nil {
					lines <- LogLine{
						Text:  fmt.Sprintf("error: failed to start background step %d/%d: %v", i+1, len(allSteps), err),
						IsErr: true,
					}
					continue
				}
				bgProcs = append(bgProcs, bp)
			} else {
				// Run this step in the foreground (blocking)
				lines <- LogLine{
					Text:  fmt.Sprintf("[%d/%d] %s", i+1, len(allSteps), step.Command),
					IsErr: false,
				}

				if err := streamOne(buildCmdFromShell(step.Command, cmd.Dir), lines); err != nil {
					lines <- LogLine{
						Text:  fmt.Sprintf("error: step %d/%d %q: %v", i+1, len(allSteps), step.Command, err),
						IsErr: true,
					}
					return
				}
			}
		}

		// Wait for all background processes to finish
		if len(bgProcs) > 0 {
			lines <- LogLine{
				Text:  fmt.Sprintf("Waiting for %d background process(es) to complete...", len(bgProcs)),
				IsErr: false,
			}
			for _, bp := range bgProcs {
				bp.Wait()
			}
		}
	}()

	return nil
}

// runBackgroundStep starts a single command in the background and streams its output.
func runBackgroundStep(shellCmd, dir string, lines chan LogLine) (*BackgroundProc, error) {
	c := buildCmdFromShell(shellCmd, dir)

	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	done := make(chan struct{})
	bp := &BackgroundProc{
		Cmd:   c,
		Lines: lines,
		done:  done,
	}

	go func() {
		defer func() {
			bp.once.Do(func() { close(done) })
		}()

		var wg sync.WaitGroup
		scan := func(r interface{ Read([]byte) (int, error) }, isErr bool) {
			defer wg.Done()
			scanner := bufio.NewScanner(r)
			scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
			for scanner.Scan() {
				lines <- LogLine{Text: "[BG] " + scanner.Text(), IsErr: isErr}
			}
		}
		wg.Add(2)
		go scan(stdout, false)
		go scan(stderr, true)
		wg.Wait()

		if err := c.Wait(); err != nil {
			lines <- LogLine{
				Text:  fmt.Sprintf("[BG] error: %v", err),
				IsErr: true,
			}
		}
	}()

	return bp, nil
}
