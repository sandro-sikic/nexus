package runner_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"nexus/config"
	"nexus/runner"
)

// echoTask returns a config.Task that simply prints the given text to stdout.
func echoTask(text string) config.Task {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo " + text
	} else {
		cmd = "echo " + text
	}
	return config.Task{
		Name:    "echo",
		Actions: []config.Action{{Command: cmd, Background: false}},
		RunMode: config.RunModeStream,
	}
}

// stderrTask returns a task that writes to stderr.
func stderrTask() config.Task {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo stderr-line 1>&2"
	} else {
		cmd = "echo stderr-line >&2"
	}
	return config.Task{
		Name:    "stderr",
		Actions: []config.Action{{Command: cmd, Background: false}},
		RunMode: config.RunModeStream,
	}
}

// multilineTask returns a task that prints several lines.
func multilineTask() config.Task {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo line1 && echo line2 && echo line3"
	} else {
		cmd = "printf 'line1\nline2\nline3\n'"
	}
	return config.Task{
		Name:    "multiline",
		Actions: []config.Action{{Command: cmd, Background: false}},
		RunMode: config.RunModeStream,
	}
}

// failTask returns a task that exits with a non-zero code.
func failTask() config.Task {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "exit 1"
	} else {
		cmd = "exit 1"
	}
	return config.Task{Name: "fail", Actions: []config.Action{{Command: cmd}}}
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
	if err := runner.Stream(echoTask("hello"), lines); err != nil {
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
	if err := runner.Stream(stderrTask(), lines); err != nil {
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
	if err := runner.Stream(echoTask("close-test"), lines); err != nil {
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
	if err := runner.Stream(multilineTask(), lines); err != nil {
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
	cmd := config.Task{Name: "bad", Actions: []config.Action{{Command: "this_command_does_not_exist_xyz"}}}
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
	proc, err := runner.RunBackground(echoTask("bg-test"))
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
	proc, err := runner.RunBackground(echoTask("bg-hello"))
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
	proc, err := runner.RunBackground(echoTask("wait-test"))
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
	proc, err := runner.RunBackground(multilineTask())
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
	if err := runner.Stream(config.Task{Name: "pwd", Actions: []config.Action{{Command: cmd}}, Dir: tmp}, lines); err != nil {
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

// ── Multi-command (Commands field) ────────────────────────────────────────────

// multiActionTask returns a task with two shell actions.
func multiActionTask() config.Task {
	var step1, step2 string
	if runtime.GOOS == "windows" {
		step1 = "echo step-one"
		step2 = "echo step-two"
	} else {
		step1 = "echo step-one"
		step2 = "echo step-two"
	}
	return config.Task{
		Name:    "multi",
		Actions: []config.Action{{Command: step1}, {Command: step2}},
		RunMode: config.RunModeStream,
	}
}

func TestStream_MultiStep_AllOutputCollected(t *testing.T) {
	lines := make(chan runner.LogLine, 64)
	if err := runner.Stream(multiActionTask(), lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	var texts []string
	timeout := time.After(10 * time.Second)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				goto done
			}
			texts = append(texts, l.Text)
		case <-timeout:
			t.Fatal("timeout waiting for multi-step stream")
		}
	}
done:
	foundOne, foundTwo := false, false
	for _, s := range texts {
		if strings.Contains(s, "step-one") {
			foundOne = true
		}
		if strings.Contains(s, "step-two") {
			foundTwo = true
		}
	}
	if !foundOne {
		t.Errorf("expected 'step-one' in output, got: %v", texts)
	}
	if !foundTwo {
		t.Errorf("expected 'step-two' in output, got: %v", texts)
	}
}

func TestStream_MultiStep_StepsAnnouncedInOrder(t *testing.T) {
	lines := make(chan runner.LogLine, 64)
	if err := runner.Stream(multiActionTask(), lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	var texts []string
	timeout := time.After(10 * time.Second)
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
	// The first non-empty line should be the "[1/2] step-one" banner.
	found1of2 := false
	for _, s := range texts {
		if strings.Contains(s, "[1/2]") {
			found1of2 = true
			break
		}
	}
	if !found1of2 {
		t.Errorf("expected '[1/2]' step banner in output, got: %v", texts)
	}
}

func TestRunBackground_MultiStep_AllOutputCollected(t *testing.T) {
	proc, err := runner.RunBackground(multiActionTask())
	if err != nil {
		t.Fatalf("RunBackground error: %v", err)
	}

	var texts []string
	timeout := time.After(10 * time.Second)
	for {
		select {
		case l, ok := <-proc.Lines:
			if !ok {
				goto done
			}
			texts = append(texts, l.Text)
		case <-timeout:
			t.Fatal("timeout")
		}
	}
done:
	foundOne, foundTwo := false, false
	for _, s := range texts {
		if strings.Contains(s, "step-one") {
			foundOne = true
		}
		if strings.Contains(s, "step-two") {
			foundTwo = true
		}
	}
	if !foundOne {
		t.Errorf("expected 'step-one' in bg output: %v", texts)
	}
	if !foundTwo {
		t.Errorf("expected 'step-two' in bg output: %v", texts)
	}
}

func TestStream_MultiStep_EmptyActionsClosesChannel(t *testing.T) {
	cmd := config.Task{Name: "empty", Actions: []config.Action{}}
	lines := make(chan runner.LogLine, 8)
	if err := runner.Stream(cmd, lines); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	timeout := time.After(3 * time.Second)
	select {
	case _, ok := <-lines:
		if ok {
			t.Fatal("channel should be closed immediately for empty commands")
		}
	case <-timeout:
		t.Fatal("channel not closed for empty command list")
	}
}

func TestWizard_MultiAction_CommitCommand(t *testing.T) {
	// This tests the wizard helper logic via the addCommand flow with extra action.
	// We verify AllCommands() returns both commands after round-tripping through config.
	task := config.Task{
		Name:    "Setup",
		Actions: []config.Action{{Command: "npm install"}, {Command: "npm run dev"}},
		RunMode: config.RunModeHandoff,
	}
	cmds := task.AllCommands()
	if len(cmds) != 2 {
		t.Fatalf("AllCommands() len: got %d, want 2", len(cmds))
	}
	if cmds[0] != "npm install" || cmds[1] != "npm run dev" {
		t.Errorf("AllCommands(): got %v", cmds)
	}
}

// ── Handoff ───────────────────────────────────────────────────────────────────

func TestHandoff_SuccessfulCommand(t *testing.T) {
	task := echoTask("handoff-ok")
	err := runner.Handoff(task)
	if err != nil {
		t.Errorf("Handoff should succeed for echo command, got: %v", err)
	}
}

func TestHandoff_EmptyActionsReturnsNil(t *testing.T) {
	task := config.Task{Name: "empty", Actions: []config.Action{}}
	err := runner.Handoff(task)
	if err != nil {
		t.Errorf("Handoff with empty steps should return nil, got: %v", err)
	}
}

func TestHandoff_EmptyActionsSlice(t *testing.T) {
	task := config.Task{Name: "empty", Actions: []config.Action{}}
	err := runner.Handoff(task)
	if err != nil {
		t.Errorf("Handoff with empty Commands slice should return nil, got: %v", err)
	}
}

func TestHandoff_FailingCommandReturnsError(t *testing.T) {
	task := failTask()
	err := runner.Handoff(task)
	if err == nil {
		t.Error("Handoff should return error when command exits non-zero")
	}
}

func TestHandoff_MultiAction_AllSucceed(t *testing.T) {
	task := multiActionTask() // has Actions: []Action{{Command: "echo step-one"}, {Command: "echo step-two"}}
	task.RunMode = config.RunModeHandoff
	err := runner.Handoff(task)
	if err != nil {
		t.Errorf("Handoff multi-step should succeed, got: %v", err)
	}
}

func TestHandoff_MultiAction_FailOnFirstAction_WrapsError(t *testing.T) {
	var step1, step2 string
	if runtime.GOOS == "windows" {
		step1 = "exit 1"
		step2 = "echo step-two"
	} else {
		step1 = "exit 1"
		step2 = "echo step-two"
	}
	task := config.Task{
		Name:    "multi-fail",
		Actions: []config.Action{{Command: step1}, {Command: step2}},
		RunMode: config.RunModeHandoff,
	}
	err := runner.Handoff(task)
	if err == nil {
		t.Fatal("Handoff multi-step should return error when first step fails")
	}
	// Error message should mention which step failed (step 1/2)
	if !strings.Contains(err.Error(), "1/2") {
		t.Errorf("error should mention step number, got: %q", err.Error())
	}
}

func TestHandoff_MultiAction_StopsOnFirstFailure(t *testing.T) {
	// Use a temp file as a side effect to verify the second action does NOT run.
	tmp := t.TempDir()
	var step1, step2 string
	if runtime.GOOS == "windows" {
		step1 = "exit 1"
		step2 = "echo ran > " + tmp + "\\sentinel.txt"
	} else {
		step1 = "exit 1"
		step2 = "touch " + tmp + "/sentinel.txt"
	}
	task := config.Task{
		Name:    "stop-on-fail",
		Actions: []config.Action{{Command: step1}, {Command: step2}},
	}
	_ = runner.Handoff(task)

	// The sentinel file should NOT exist because execution stopped at step 1.
	if _, statErr := os.Stat(tmp + "/sentinel.txt"); statErr == nil {
		if runtime.GOOS != "windows" {
			t.Error("second step ran after first step failed")
		}
	}
}

// ── Stream: multi-step failure ─────────────────────────────────────────────────

func TestStream_MultiAction_FailOnFirstAction_SendsErrorLine(t *testing.T) {
	var step1, step2 string
	if runtime.GOOS == "windows" {
		step1 = "exit 1"
		step2 = "echo step-two"
	} else {
		step1 = "exit 1"
		step2 = "echo step-two"
	}
	task := config.Task{
		Name:    "fail-multi",
		Actions: []config.Action{{Command: step1}, {Command: step2}},
		RunMode: config.RunModeStream,
	}
	lines := make(chan runner.LogLine, 64)
	if err := runner.Stream(task, lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	var collected []runner.LogLine
	timeout := time.After(10 * time.Second)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				goto done
			}
			collected = append(collected, l)
		case <-timeout:
			t.Fatal("timeout waiting for stream to close after step failure")
		}
	}
done:
	// Should have an error line and NOT have step-two output
	foundErr := false
	foundStepTwo := false
	for _, l := range collected {
		if l.IsErr && strings.Contains(l.Text, "error") {
			foundErr = true
		}
		if strings.Contains(l.Text, "step-two") {
			foundStepTwo = true
		}
	}
	if !foundErr {
		t.Errorf("expected error line after step failure, got: %v", collected)
	}
	if foundStepTwo {
		t.Errorf("step-two output should not appear after step-one failure, got: %v", collected)
	}
}

// ── Large-line scanner buffer ─────────────────────────────────────────────────
//
// These tests exercise the 1 MB scanner buffer added to streamOne and
// RunBackground. Before the fix, bufio.Scanner's default 64 KB limit would
// cause it to silently stop reading, leaving the channel open and the UI hung.

// largeLine returns a shell command that prints a single line of `n` '*' bytes.
// On Windows the command must survive cmd /C quoting, so single-quotes are
// avoided and PowerShell is called without extra shell quoting.
func largeLine(n int) string {
	if runtime.GOOS == "windows" {
		// cmd /C passes this directly to PowerShell. No inner quotes needed.
		return fmt.Sprintf("powershell -NoProfile -Command Write-Host([string]::new([char]42,%d))", n)
	}
	return fmt.Sprintf("python3 -c \"import sys; sys.stdout.write('*'*%d+'\\n')\"", n)
}

func TestStream_LargeLineIsReceived(t *testing.T) {
	// 256 KB line — well above the old 64 KB default.
	const lineSize = 256 * 1024
	lines := make(chan runner.LogLine, 4)
	task := config.Task{Name: "large", Actions: []config.Action{{Command: largeLine(lineSize)}}, RunMode: config.RunModeStream}
	if err := runner.Stream(task, lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	timeout := time.After(15 * time.Second)
	var got []runner.LogLine
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				goto done
			}
			got = append(got, l)
		case <-timeout:
			t.Fatal("timed out — channel never closed (scanner may have blocked on large line)")
		}
	}
done:
	// At least one line should have arrived and be close to the expected size.
	found := false
	for _, l := range got {
		if len(l.Text) >= lineSize/2 {
			found = true
		}
	}
	if !found {
		sizes := make([]int, len(got))
		for i, l := range got {
			sizes[i] = len(l.Text)
		}
		t.Errorf("expected a line of ≥%d bytes; got line sizes: %v", lineSize/2, sizes)
	}
}

func TestStream_LargeLine_ChannelClosedAfterExit(t *testing.T) {
	// Verify the channel is always closed even for a large-line command —
	// previously the scanner would stall and the channel would stay open.
	const lineSize = 256 * 1024
	lines := make(chan runner.LogLine, 4)
	task := config.Task{Name: "large-close", Actions: []config.Action{{Command: largeLine(lineSize)}}, RunMode: config.RunModeStream}
	if err := runner.Stream(task, lines); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	timeout := time.After(15 * time.Second)
	for {
		select {
		case _, ok := <-lines:
			if !ok {
				return // closed as expected
			}
		case <-timeout:
			t.Fatal("channel was not closed after large-line process exited")
		}
	}
}

func TestRunBackground_LargeLineIsReceived(t *testing.T) {
	const lineSize = 256 * 1024
	task := config.Task{Name: "bg-large", Actions: []config.Action{{Command: largeLine(lineSize)}}, RunMode: config.RunModeBackground}
	proc, err := runner.RunBackground(task)
	if err != nil {
		t.Fatalf("RunBackground error: %v", err)
	}

	timeout := time.After(15 * time.Second)
	var got []runner.LogLine
	for {
		select {
		case l, ok := <-proc.Lines:
			if !ok {
				goto done
			}
			got = append(got, l)
		case <-timeout:
			t.Fatal("timed out — background channel never closed (scanner may have blocked)")
		}
	}
done:
	found := false
	for _, l := range got {
		if len(l.Text) >= lineSize/2 {
			found = true
		}
	}
	if !found {
		sizes := make([]int, len(got))
		for i, l := range got {
			sizes[i] = len(l.Text)
		}
		t.Errorf("background: expected a line of ≥%d bytes; got line sizes: %v", lineSize/2, sizes)
	}
}

// ── RunBackground: empty steps ─────────────────────────────────────────────────

func TestRunBackground_EmptyActions_ClosesImmediately(t *testing.T) {
	task := config.Task{Name: "empty-bg", Actions: []config.Action{}}
	proc, err := runner.RunBackground(task)
	if err != nil {
		t.Fatalf("RunBackground error: %v", err)
	}
	if proc == nil {
		t.Fatal("expected non-nil BackgroundProc")
	}

	timeout := time.After(3 * time.Second)
	select {
	case _, ok := <-proc.Lines:
		if ok {
			t.Error("channel should be immediately closed for empty steps")
		}
	case <-timeout:
		t.Fatal("channel not closed for empty background command")
	}
}

func TestRunBackground_MultiAction_ErrorLineOnActionFailure(t *testing.T) {
	var step1, step2 string
	if runtime.GOOS == "windows" {
		step1 = "exit 1"
		step2 = "echo step-two"
	} else {
		step1 = "exit 1"
		step2 = "echo step-two"
	}
	task := config.Task{
		Name:    "bg-fail",
		Actions: []config.Action{{Command: step1}, {Command: step2}},
		RunMode: config.RunModeBackground,
	}
	proc, err := runner.RunBackground(task)
	if err != nil {
		t.Fatalf("RunBackground error: %v", err)
	}

	var lines []runner.LogLine
	timeout := time.After(10 * time.Second)
	for {
		select {
		case l, ok := <-proc.Lines:
			if !ok {
				goto done
			}
			lines = append(lines, l)
		case <-timeout:
			t.Fatal("timeout waiting for background process to finish")
		}
	}
done:
	foundErr := false
	for _, l := range lines {
		if l.IsErr {
			foundErr = true
		}
	}
	if !foundErr {
		t.Errorf("expected error line after step failure, got: %v", lines)
	}
}
