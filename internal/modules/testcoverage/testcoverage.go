package testcoverage

import "fmt"

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

func (t *TestCoverage) IsUpCoverage() error {
	currentHash, err := t.cmd.GetCurrentCommitHash()
	if err != nil {
		return err
	}

	err = t.getAndPutCoverageProcentToRepo(currentHash)
	if err != nil {
		return err
	}

	previousHash, err := t.cmd.GetPreviousCommitHash()
	if err != nil {
		return err
	}

	err = t.getAndPutCoverageProcentToRepo(previousHash)
	if err != nil {
		return err
	}

	currentCoverageProcent, err := t.repo.GetCoverageProcent(t.projectName, currentHash)
	if err != nil {
		return err
	}

	previousCoverageProcent, err := t.repo.GetCoverageProcent(t.projectName, previousHash)
	if err != nil {
		return err
	}

	if currentCoverageProcent >= previousCoverageProcent {
		return nil
	}

	return fmt.Errorf("coverage went down from %f to %f", previousCoverageProcent, currentCoverageProcent)
}

func (t *TestCoverage) getAndPutCoverageProcentToRepo(hash string) error {
	ok, err := t.repo.HasCoverageProcent(t.projectName, hash)
	if err != nil {
		return err
	}
	if !ok {
		if err := t.cmd.CheckoutToCommitByHash(hash); err != nil {
			return err
		}
		coverageProcent, err := t.cmd.CoverageProcent()
		if err != nil {
			return err
		}
		err = t.repo.AddCoverageProcent(t.projectName, hash, coverageProcent)
		if err != nil {
			return err
		}
	}
	return nil
}
