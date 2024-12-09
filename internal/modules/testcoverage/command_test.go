package testcoverage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nfort/gopher-bot/internal/cmd"
)

func TestCoverageProcent(t *testing.T) {
	repo, _ := filepath.Abs("git-repo.tar.gz")
	testRepoFolder, _ := os.MkdirTemp("", "gopher-bot-test-repo-*")

	defer os.RemoveAll(testRepoFolder)

	cmd := cmd.NewCommand(testRepoFolder)
	_, err := cmd.Run("tar", "-zxf", repo, "-C", testRepoFolder)
	if err != nil {
		t.Fatal(err)
	}

	c := NewCommand(testRepoFolder)

	coverage, err := c.CoverageProcent()
	if err != nil {
		t.Fatal(err)
	}

	if coverage != 33.300000 {
		t.Fatalf("get invalid coverage: %f", coverage)
	}
}

func TestRepository(t *testing.T) {
	repo, _ := filepath.Abs("git-repo.tar.gz")
	projectFolder, _ := os.MkdirTemp("", "gopher-bot-test-repo-*")

	defer os.RemoveAll(projectFolder)

	cmd := cmd.NewCommand(projectFolder)
	_, err := cmd.Run("tar", "-zxf", repo, "-C", projectFolder)
	if err != nil {
		t.Fatal(err)
	}

	c := NewCommand(projectFolder)
	hash, err := c.GetCurrentCommitHash()
	if err != nil {
		t.Fatal(err)
	}

	if hash != "935377b18bdbce571b4ec7afa97b8dbbbfcdcf5b" {
		t.Fatalf("get invalid hash: %q", hash)
	}

	hash, err = c.GetPreviousCommitHash()
	if err != nil {
		t.Fatal(err)
	}

	if hash != "681fc9102edd7b37d5775fcc8115d210a1471fd1" {
		t.Fatalf("get invalid hash: %q", hash)
	}

	err = c.CheckoutToCommitByHash(hash)
	if err != nil {
		t.Fatal(err)
	}

	hash, err = c.GetCurrentCommitHash()
	if err != nil {
		t.Fatal(err)
	}

	if hash != "681fc9102edd7b37d5775fcc8115d210a1471fd1" {
		t.Fatalf("get invalid hash: %q", hash)
	}
}
