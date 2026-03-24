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

// LogLine is a single line of output from a running process.
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

// Handoff runs each action sequentially in the raw terminal,
// stopping on the first failure.
// Note: Background actions in handoff mode still run in background.
func Handoff(task config.Task) error {
	if len(task.Actions) == 0 {
		return nil
	}

	// Track background processes
	var bgProcs []*BackgroundProc

	for i, action := range task.Actions {
		if action.Background {
			// Start in background but don't block
			bp, err := runBackgroundAction(action.Command, task.Dir, nil)
			if err != nil {
				return fmt.Errorf("action %d/%d background %q: %w", i+1, len(task.Actions), action.Command, err)
			}
			bgProcs = append(bgProcs, bp)
		} else {
			// Run in foreground
			c := buildCmdFromShell(action.Command, task.Dir)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("action %d/%d %q: %w", i+1, len(task.Actions), action.Command, err)
			}
		}
	}

	// Wait for all background processes to complete
	for _, bp := range bgProcs {
		bp.Wait()
	}

	return nil
}

// Stream runs actions sequentially and sends output lines to the provided channel.
// Background actions are started and continue while foreground actions execute.
// The channel is closed when all actions complete or a foreground action fails.
// If the task has handoff=true, the last action is NOT executed (should be handled by HandoffLastAction).
func Stream(task config.Task, lines chan LogLine) error {
	actions := task.Actions
	// If task has handoff, don't execute last action here
	if task.Handoff && len(actions) > 0 {
		actions = actions[:len(actions)-1]
	}

	if len(actions) == 0 {
		close(lines)
		return nil
	}

	go func() {
		defer close(lines)

		// Track background processes
		var bgProcs []*BackgroundProc

		for i, action := range actions {
			if action.Background {
				// Start background action and continue immediately
				lines <- LogLine{
					Text:  fmt.Sprintf("[%d/%d] [BG] Starting: %s", i+1, len(actions), action.Command),
					IsErr: false,
				}

				bp, err := runBackgroundAction(action.Command, task.Dir, lines)
				if err != nil {
					lines <- LogLine{
						Text:  fmt.Sprintf("error: failed to start background action %d/%d: %v", i+1, len(actions), err),
						IsErr: true,
					}
					return
				}
				bgProcs = append(bgProcs, bp)
			} else {
				// Run foreground action (blocking)
				if len(actions) > 1 {
					lines <- LogLine{
						Text:  fmt.Sprintf("[%d/%d] %s", i+1, len(actions), action.Command),
						IsErr: false,
					}
				}

				if err := streamOne(buildCmdFromShell(action.Command, task.Dir), lines); err != nil {
					lines <- LogLine{
						Text:  fmt.Sprintf("error: action %d/%d %q: %v", i+1, len(actions), action.Command, err),
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
			lines <- LogLine{
				Text:  "All background processes completed",
				IsErr: false,
			}
		}
	}()

	return nil
}

// HandoffLastAction runs the last action of the task in the raw terminal.
// This should be called after Stream() when the task has handoff=true.
// Returns nil if no handoff action exists.
func HandoffLastAction(task config.Task) error {
	if !task.Handoff || len(task.Actions) == 0 {
		return nil
	}

	lastAction := task.Actions[len(task.Actions)-1]

	// Run the last action in the foreground (raw terminal)
	c := buildCmdFromShell(lastAction.Command, task.Dir)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// streamOne runs a single exec.Cmd and pipes its stdout/stderr to lines.
// It blocks until the process exits.
func streamOne(c *exec.Cmd, lines chan LogLine) error {
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

// RunBackground starts all actions of the task in the background, running
// them sequentially, and returns a BackgroundProc whose Lines channel receives
// all output.
func RunBackground(task config.Task) (*BackgroundProc, error) {
	lines := make(chan LogLine, 256)
	done := make(chan struct{})

	bp := &BackgroundProc{
		Lines: lines,
		done:  done,
	}

	if len(task.Actions) == 0 {
		go func() {
			close(lines)
			bp.once.Do(func() { close(done) })
		}()
		return bp, nil
	}

	// For BackgroundProc.Cmd we expose the first action's exec.Cmd
	firstCmd := buildCmdFromShell(task.Actions[0].Command, task.Dir)
	bp.Cmd = firstCmd

	go func() {
		defer func() {
			close(lines)
			bp.once.Do(func() { close(done) })
		}()

		// Track background processes from actions
		var bgProcs []*BackgroundProc

		for i, action := range task.Actions {
			if len(task.Actions) > 1 {
				lines <- LogLine{
					Text:  fmt.Sprintf("[%d/%d] %s", i+1, len(task.Actions), action.Command),
					IsErr: false,
				}
			}

			if action.Background {
				// Start as background within background mode
				innerBp, err := runBackgroundAction(action.Command, task.Dir, lines)
				if err != nil {
					lines <- LogLine{
						Text:  fmt.Sprintf("error: failed to start background action %d/%d: %v", i+1, len(task.Actions), err),
						IsErr: true,
					}
					return
				}
				bgProcs = append(bgProcs, innerBp)
			} else {
				// Run foreground (blocking)
				var c *exec.Cmd
				if i == 0 {
					c = firstCmd
				} else {
					c = buildCmdFromShell(action.Command, task.Dir)
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
					scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
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
						Text:  fmt.Sprintf("error: action %d/%d %q: %v", i+1, len(task.Actions), action.Command, err),
						IsErr: true,
					}
					return
				}
			}
		}

		// Wait for any nested background processes
		for _, bgp := range bgProcs {
			bgp.Wait()
		}
	}()

	return bp, nil
}

// runBackgroundAction starts a single command in the background and streams its output.
// If lines is nil, output is discarded.
func runBackgroundAction(shellCmd, dir string, lines chan LogLine) (*BackgroundProc, error) {
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
				if lines != nil {
					lines <- LogLine{Text: "[BG] " + scanner.Text(), IsErr: isErr}
				}
			}
		}
		wg.Add(2)
		go scan(stdout, false)
		go scan(stderr, true)
		wg.Wait()

		if err := c.Wait(); err != nil {
			if lines != nil {
				lines <- LogLine{
					Text:  fmt.Sprintf("[BG] error: %v", err),
					IsErr: true,
				}
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
