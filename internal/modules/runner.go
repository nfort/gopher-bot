package modules

import (
	"context"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/nfort/gopher-bot/internal/models"
)

type (
	Func   func(ctx context.Context, workingDi string, r *git.Repository, hook *models.PRHook) error
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

func (c *Runner) Run() <-chan error {
	ch := make(chan error)
	var wg sync.WaitGroup

	baseCtx, baseCtxCancel := context.WithCancel(context.Background())

	wg.Add(len(c.handlers))
	for _, handler := range c.handlers {
		go func() {
			defer wg.Done()
			err := handler(baseCtx, c.workingDir, c.repo, c.hook)
			if err != nil {
				baseCtxCancel()
				ch <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		baseCtxCancel()
		close(ch)
	}()

	return ch
}
