package runner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"runner/config"
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

// buildCmd constructs an exec.Cmd from a config.Command.
func buildCmd(cmd config.Command) *exec.Cmd {
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd.Command)
	} else {
		c = exec.Command("sh", "-c", cmd.Command)
	}
	if cmd.Dir != "" {
		c.Dir = cmd.Dir
	}
	c.Env = os.Environ()
	return c
}

// Handoff replaces the current process with the command (or runs it attached).
// On Windows exec.Command is used with Stdin/Stdout/Stderr inherited.
func Handoff(cmd config.Command) error {
	c := buildCmd(cmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// Stream runs the command and sends output lines to the provided channel.
// It closes the channel when the process exits.
func Stream(cmd config.Command, lines chan<- LogLine) error {
	c := buildCmd(cmd)

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
		for scanner.Scan() {
			lines <- LogLine{Text: scanner.Text(), IsErr: isErr}
		}
	}

	wg.Add(2)
	go scanLines(stdout, false)
	go scanLines(stderr, true)

	go func() {
		wg.Wait()
		c.Wait() //nolint:errcheck
		close(lines)
	}()

	return nil
}

// RunBackground starts the command in the background and returns a BackgroundProc.
func RunBackground(cmd config.Command) (*BackgroundProc, error) {
	lines := make(chan LogLine, 256)
	done := make(chan struct{})

	c := buildCmd(cmd)

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

	bp := &BackgroundProc{
		Cmd:   c,
		Lines: lines,
		done:  done,
	}

	var wg sync.WaitGroup

	scan := func(r interface{ Read([]byte) (int, error) }, isErr bool) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			lines <- LogLine{Text: scanner.Text(), IsErr: isErr}
		}
	}

	wg.Add(2)
	go scan(stdout, false)
	go scan(stderr, true)

	go func() {
		wg.Wait()
		c.Wait() //nolint:errcheck
		close(lines)
		bp.once.Do(func() { close(done) })
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
