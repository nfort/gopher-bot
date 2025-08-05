package testcoverage

import (
	"context"
	"fmt"
)

type TestCoverage struct {
	projectName string
	cmd         *Command
	repo        *Repo
}

func NewTestCoverage(projectName string, workingDir string, repo *Repo) *TestCoverage {
	return &TestCoverage{
		projectName: projectName,
		cmd:         NewCommand(workingDir),
		repo:        repo,
	}
}

func (t *TestCoverage) IsUpCoverage(ctx context.Context) error {
	currentHash, err := t.cmd.GetCurrentCommitHash(ctx)
	if err != nil {
		return err
	}

	err = t.getAndPutCoveragePercentToRepo(ctx, currentHash)
	if err != nil {
		return err
	}

	previousHash, err := t.cmd.GetPreviousCommitHash(ctx)
	if err != nil {
		return err
	}

	err = t.getAndPutCoveragePercentToRepo(ctx, previousHash)
	if err != nil {
		return err
	}

	currentCoveragePercent, err := t.repo.GetCoveragePercent(t.projectName, currentHash)
	if err != nil {
		return err
	}

	previousCoveragePercent, err := t.repo.GetCoveragePercent(t.projectName, previousHash)
	if err != nil {
		return err
	}

	if currentCoveragePercent >= previousCoveragePercent {
		return nil
	}

	return fmt.Errorf("coverage went down from %f to %f", previousCoveragePercent, currentCoveragePercent)
}

func (t *TestCoverage) getAndPutCoveragePercentToRepo(ctx context.Context, hash string) error {
	ok, err := t.repo.HasCoveragePercent(t.projectName, hash)
	if err != nil {
		return err
	}
	if !ok {
		if err := t.cmd.CheckoutToCommitByHash(ctx, hash); err != nil {
			return err
		}
		coveragePercent, err := t.cmd.CoveragePercent(ctx)
		if err != nil {
			return err
		}
		err = t.repo.AddCoveragePercent(t.projectName, hash, coveragePercent)
		if err != nil {
			return err
		}
	}
	return nil
}
