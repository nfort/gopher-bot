package modules

import (
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/nfort/gopher-bot/internal/models"
)

type (
	Func   func(workingDi string, r *git.Repository, hook *models.PRHook) error
	Runner struct {
		handlers   []Func
		repo       *git.Repository
		hook       *models.PRHook
		workingDir string
	}
)

func NewRunner(workingDir string, repo *git.Repository, hook *models.PRHook, funcs ...Func) *Runner {
	return &Runner{
		handlers:   funcs,
		repo:       repo,
		hook:       hook,
		workingDir: workingDir,
	}
}

func (c *Runner) Run() error {
	var err error
	for _, handler := range c.handlers {
		err = c.reset()
		if err != nil {
			return err
		}
		err = handler(c.workingDir, c.repo, c.hook)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Runner) reset() error {
	w, err := c.repo.Worktree()
	if err != nil {
		log.Printf("Worktree: %s", err)
		return err
	}

	ref, err := c.repo.Head()
	if err != nil {
		log.Printf("Head: %s", err)
		return err
	}

	err = w.Reset(&git.ResetOptions{
		Commit: ref.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		log.Printf("Reset: %s", err)
		return err
	}

	return nil
}
