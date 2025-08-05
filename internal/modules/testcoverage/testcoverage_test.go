package testcoverage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nfort/gopher-bot/internal/cmd"
)

func TestIsUpCoverage(t *testing.T) {
	os.RemoveAll("sqlite.db")

	gitRepoTarGz, _ := filepath.Abs("git-repo.tar.gz")
	workingDir, _ := os.MkdirTemp("", "gopher-bot-test-repo-*")

	defer os.RemoveAll(workingDir)

	cmd := cmd.NewCommand(workingDir)
	_, err := cmd.Run(context.Background(), "tar", "-zxf", gitRepoTarGz, "-C", workingDir)
	if err != nil {
		t.Fatal(err)
	}

	r := NewRepo("sqlite.db")

	tc := NewTestCoverage("go-test", workingDir, r)
	err = tc.cmd.CheckoutToCommitByHash(context.Background(), "681fc9102edd7b37d5775fcc8115d210a1471fd1")
	if err != nil {
		t.Fatal(err)
	}

	err = tc.IsUpCoverage(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	currentCoveragePercent, _ := r.GetCoveragePercent("go-test", "681fc9102edd7b37d5775fcc8115d210a1471fd1")
	prevCoveragePercent, _ := r.GetCoveragePercent("go-test", "b3dc50b69c174aacdc0be4d0d25ca8985490cfa3")
	if 50.00000 != currentCoveragePercent {
		t.Fatalf("invalid coverage procent: got: %f, expected: %f", currentCoveragePercent, 50.00000)
	}
	if 0.00000 != prevCoveragePercent {
		t.Fatalf("invalid coverage procent: got: %f, expected: %f", prevCoveragePercent, 0.00000)
	}
}

func TestIsUpCoverageFailed(t *testing.T) {
	os.RemoveAll("sqlite.db")

	gitRepoTarGz, _ := filepath.Abs("git-repo.tar.gz")
	workingDir, _ := os.MkdirTemp("", "gopher-bot-test-repo-*")

	defer os.RemoveAll(workingDir)

	cmd := cmd.NewCommand(workingDir)
	_, err := cmd.Run(context.Background(), "tar", "-zxf", gitRepoTarGz, "-C", workingDir)
	if err != nil {
		t.Fatal(err)
	}

	r := NewRepo("sqlite.db")

	tc := NewTestCoverage("go-test", workingDir, r)

	err = tc.IsUpCoverage(context.Background())
	if err == nil {
		t.Fatal(err)
	}

	currentCoveragePercent, _ := r.GetCoveragePercent("go-test", "935377b18bdbce571b4ec7afa97b8dbbbfcdcf5b")
	prevCoveragePercent, _ := r.GetCoveragePercent("go-test", "681fc9102edd7b37d5775fcc8115d210a1471fd1")
	if 33.30000 != currentCoveragePercent {
		t.Fatalf("invalid coverage procent: got: %f, expected: %f", currentCoveragePercent, 33.300000)
	}
	if 50.00000 != prevCoveragePercent {
		t.Fatalf("invalid coverage procent: got: %f, expected: %f", prevCoveragePercent, 50.00000)
	}
}
