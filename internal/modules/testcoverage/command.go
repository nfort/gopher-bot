package testcoverage

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nfort/gopher-bot/internal/cmd"
)

type Command struct {
	workingDir string
	cmd        *cmd.Command
}

func NewCommand(workingDir string) *Command {
	return &Command{
		workingDir: workingDir,
		cmd:        cmd.NewCommand(workingDir),
	}
}

func (c *Command) CoverageProcent(ctx context.Context) (float64, error) {
	_, err := c.cmd.Run(ctx, "go", "test", "-coverprofile=coverage.out", "./...")
	if err != nil {
		return 0.0, err
	}

	defer os.RemoveAll(filepath.Join(c.workingDir, "coverage.out"))
	cmd := "go tool cover -func=coverage.out | tail -n 1 | awk '{print $3}' | tr -d '%'"
	coverage, err := c.cmd.Run(ctx, "bash", "-c", cmd)
	if err != nil {
		return 0.0, err
	}

	return strconv.ParseFloat(strings.TrimSuffix(coverage, "\n"), 64)
}

func (c *Command) GetCurrentCommitHash(ctx context.Context) (string, error) {
	hash, err := c.cmd.Run(ctx, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(hash, "\n"), nil
}

func (c *Command) GetPreviousCommitHash(ctx context.Context) (string, error) {
	hash, err := c.cmd.Run(ctx, "git", "rev-parse", "HEAD~1")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(hash, "\n"), nil
}

func (c *Command) CheckoutToCommitByHash(ctx context.Context, hash string) error {
	_, err := c.cmd.Run(ctx, "git", "checkout", hash)
	return err
}
