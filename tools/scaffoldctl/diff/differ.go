package diff

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func DiffFile(targetPath, generated string) error {
	return DiffFileWithOptions(targetPath, generated, Options{})
}

type Options struct {
	NoPager bool
}

func DiffFileWithOptions(targetPath, generated string, opts Options) error {
	tmp, err := os.CreateTemp("", "scaffold-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, writeErr := tmp.WriteString(generated); writeErr != nil {
		_ = tmp.Close()
		return writeErr
	}
	if closeErr := tmp.Close(); closeErr != nil {
		return closeErr
	}

	old := targetPath
	if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
		old = "/dev/null"
	}

	diffCmd := exec.CommandContext(context.Background(), "diff", "-u", old, tmp.Name()) // #nosec G204

	if opts.NoPager {
		return runPlainDiff(diffCmd)
	}

	if _, lookupErr := exec.LookPath("delta"); lookupErr != nil {
		return runPlainDiff(diffCmd)
	}

	// pipe into delta
	deltaCmd := exec.CommandContext(context.Background(), "delta")

	pipe, err := diffCmd.StdoutPipe()
	if err != nil {
		return err
	}

	deltaCmd.Stdin = pipe
	deltaCmd.Stdout = os.Stdout
	deltaCmd.Stderr = os.Stderr
	deltaCmd.Stdin = pipe

	if err := deltaCmd.Start(); err != nil {
		return err
	}

	if err := diffCmd.Start(); err != nil {
		return err
	}

	diffErr := diffCmd.Wait()
	_ = pipe.Close()

	deltaErr := deltaCmd.Wait()

	if diffErr != nil {
		var exitErr *exec.ExitError
		if errors.As(diffErr, &exitErr) {
			if exitErr.ExitCode() != 1 {
				return fmt.Errorf("diff failed: %w", diffErr)
			}
		} else {
			return diffErr
		}
	}

	if deltaErr != nil {
		return fmt.Errorf("delta failed: %w", deltaErr)
	}
	return nil
}

func runPlainDiff(diffCmd *exec.Cmd) error {
	diffCmd.Stdout = os.Stdout
	diffCmd.Stderr = os.Stderr
	if err := diffCmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return nil
			}
		}
		return fmt.Errorf("diff failed: %w", err)
	}
	return nil
}
