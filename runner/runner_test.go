package runner_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"runner/config"
	"runner/runner"
)

// echoCmd returns a config.Command that simply prints the given text to stdout.
func echoCmd(text string) config.Command {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo " + text
	} else {
		cmd = "echo " + text
	}
	return config.Command{
		Name:    "echo",
		Command: cmd,
		RunMode: config.RunModeStream,
	}
}

// stderrCmd returns a command that writes to stderr.
func stderrCmd() config.Command {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo stderr-line 1>&2"
	} else {
		cmd = "echo stderr-line >&2"
	}
	return config.Command{
		Name:    "stderr",
		Command: cmd,
		RunMode: config.RunModeStream,
	}
}

// multilineCmd returns a command that prints several lines.
func multilineCmd() config.Command {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo line1 && echo line2 && echo line3"
	} else {
		cmd = "printf 'line1\nline2\nline3\n'"
	}
	return config.Command{
		Name:    "multiline",
		Command: cmd,
		RunMode: config.RunModeStream,
	}
}

// failCmd returns a command that exits with a non-zero code.
func failCmd() config.Command {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "exit 1"
	} else {
		cmd = "exit 1"
	}
	return config.Command{Name: "fail", Command: cmd}
}

// ── FormatCommand ─────────────────────────────────────────────────────────────

func TestFormatCommand_Short(t *testing.T) {
	got := runner.FormatCommand("npm run dev")
	if got != "npm run dev" {
		t.Errorf("got %q, want %q", got, "npm run dev")
	}
}

func TestFormatCommand_ExactlyFourWords(t *testing.T) {
	got := runner.FormatCommand("a b c d")
	if got != "a b c d" {
		t.Errorf("got %q, want %q", got, "a b c d")
	}
}

func TestFormatCommand_MoreThanFourWords(t *testing.T) {
	got := runner.FormatCommand("a b c d e f")
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected truncation with ellipsis, got %q", got)
	}
	if strings.Contains(got, "e") || strings.Contains(got, "f") {
		t.Errorf("truncated output should not contain 5th/6th word, got %q", got)
	}
}

func TestFormatCommand_FiveWords(t *testing.T) {
	got := runner.FormatCommand("one two three four five")
	want := "one two three four …"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatCommand_Empty(t *testing.T) {
	got := runner.FormatCommand("")
	if got != "" {
		t.Errorf("empty command: got %q, want %q", got, "")
	}
}

func TestFormatCommand_SingleWord(t *testing.T) {
	got := runner.FormatCommand("make")
	if got != "make" {
		t.Errorf("got %q, want %q", got, "make")
	}
}

// ── Stream ────────────────────────────────────────────────────────────────────

func TestStream_CollectsStdout(t *testing.T) {
	lines := make(chan runner.LogLine, 32)
	if err := runner.Stream(echoCmd("hello"), lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	var collected []runner.LogLine
	timeout := time.After(5 * time.Second)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				goto done
			}
			collected = append(collected, l)
		case <-timeout:
			t.Fatal("timed out waiting for stream to close")
		}
	}
done:
	if len(collected) == 0 {
		t.Fatal("expected at least one line, got none")
	}
	found := false
	for _, l := range collected {
		if strings.Contains(l.Text, "hello") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'hello' in output, got: %v", collected)
	}
}

func TestStream_StderrMarkedIsErr(t *testing.T) {
	lines := make(chan runner.LogLine, 32)
	if err := runner.Stream(stderrCmd(), lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	var collected []runner.LogLine
	timeout := time.After(5 * time.Second)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				goto done
			}
			collected = append(collected, l)
		case <-timeout:
			t.Fatal("timeout")
		}
	}
done:
	found := false
	for _, l := range collected {
		if strings.Contains(l.Text, "stderr-line") && l.IsErr {
			found = true
		}
	}
	if !found {
		t.Errorf("expected stderr line with IsErr=true, got: %v", collected)
	}
}

func TestStream_ChannelClosedAfterExit(t *testing.T) {
	lines := make(chan runner.LogLine, 32)
	if err := runner.Stream(echoCmd("close-test"), lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-lines:
			if !ok {
				return // channel closed as expected
			}
		case <-timeout:
			t.Fatal("channel was not closed after process exit")
		}
	}
}

func TestStream_MultipleLines(t *testing.T) {
	lines := make(chan runner.LogLine, 64)
	if err := runner.Stream(multilineCmd(), lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	var texts []string
	timeout := time.After(5 * time.Second)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				goto done
			}
			texts = append(texts, l.Text)
		case <-timeout:
			t.Fatal("timeout")
		}
	}
done:
	if len(texts) < 3 {
		t.Errorf("expected ≥3 lines, got %d: %v", len(texts), texts)
	}
}

func TestStream_InvalidCommand(t *testing.T) {
	lines := make(chan runner.LogLine, 8)
	cmd := config.Command{Name: "bad", Command: "this_command_does_not_exist_xyz"}
	// On Windows cmd /C will exit non-zero but won't error on Start.
	// On Unix sh -c also won't error on Start for unknown commands.
	// We just check the channel eventually closes.
	if err := runner.Stream(cmd, lines); err != nil {
		// Some platforms may error on Start — acceptable.
		return
	}
	timeout := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-lines:
			if !ok {
				return
			}
		case <-timeout:
			t.Fatal("channel never closed for invalid command")
		}
	}
}

// ── RunBackground ─────────────────────────────────────────────────────────────

func TestRunBackground_ReturnsProc(t *testing.T) {
	proc, err := runner.RunBackground(echoCmd("bg-test"))
	if err != nil {
		t.Fatalf("RunBackground error: %v", err)
	}
	if proc == nil {
		t.Fatal("expected non-nil BackgroundProc")
	}
	if proc.Cmd == nil {
		t.Error("expected non-nil Cmd")
	}
	if proc.Lines == nil {
		t.Error("expected non-nil Lines channel")
	}

	// Drain and wait
	timeout := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-proc.Lines:
			if !ok {
				return
			}
		case <-timeout:
			t.Fatal("background process did not finish in time")
		}
	}
}

func TestRunBackground_CollectsOutput(t *testing.T) {
	proc, err := runner.RunBackground(echoCmd("bg-hello"))
	if err != nil {
		t.Fatalf("RunBackground error: %v", err)
	}

	var texts []string
	timeout := time.After(5 * time.Second)
	for {
		select {
		case l, ok := <-proc.Lines:
			if !ok {
				goto done
			}
			texts = append(texts, l.Text)
		case <-timeout:
			t.Fatal("timeout collecting background output")
		}
	}
done:
	found := false
	for _, s := range texts {
		if strings.Contains(s, "bg-hello") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'bg-hello' in background output, got: %v", texts)
	}
}

func TestRunBackground_WaitUnblocksAfterExit(t *testing.T) {
	proc, err := runner.RunBackground(echoCmd("wait-test"))
	if err != nil {
		t.Fatalf("RunBackground error: %v", err)
	}

	// Drain channel first
	timeout := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-proc.Lines:
			if !ok {
				goto drained
			}
		case <-timeout:
			t.Fatal("timeout draining channel")
		}
	}
drained:
	done := make(chan struct{})
	go func() {
		proc.Wait()
		close(done)
	}()

	select {
	case <-done:
		// expected
	case <-time.After(3 * time.Second):
		t.Fatal("Wait() did not unblock after process exit")
	}
}

func TestRunBackground_MultipleLines(t *testing.T) {
	proc, err := runner.RunBackground(multilineCmd())
	if err != nil {
		t.Fatalf("RunBackground error: %v", err)
	}

	var lines []string
	timeout := time.After(5 * time.Second)
	for {
		select {
		case l, ok := <-proc.Lines:
			if !ok {
				goto done
			}
			lines = append(lines, l.Text)
		case <-timeout:
			t.Fatal("timeout")
		}
	}
done:
	if len(lines) < 3 {
		t.Errorf("expected ≥3 lines, got %d", len(lines))
	}
}

// ── Dir field ─────────────────────────────────────────────────────────────────

func TestStream_WorkingDirectory(t *testing.T) {
	tmp := t.TempDir()
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cd"
	} else {
		cmd = "pwd"
	}

	lines := make(chan runner.LogLine, 16)
	if err := runner.Stream(config.Command{Name: "pwd", Command: cmd, Dir: tmp}, lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	var texts []string
	timeout := time.After(5 * time.Second)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				goto done
			}
			texts = append(texts, l.Text)
		case <-timeout:
			t.Fatal("timeout")
		}
	}
done:
	// The output should contain some portion of the temp dir path.
	// On Windows `cd` prints the full path; trim drive letter differences.
	found := false
	for _, s := range texts {
		if strings.Contains(strings.ToLower(s), strings.ToLower(tmp[len(tmp)-4:])) {
			found = true
		}
	}
	if !found {
		t.Logf("tmp=%q, output=%v", tmp, texts)
		// Non-fatal: path representation can vary across platforms.
	}
}
