package modules

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nfort/gopher-bot/internal/cmd"
	"github.com/nfort/gopher-bot/internal/models"
	"github.com/nfort/gopher-bot/internal/modules/config"
	"github.com/nfort/gopher-bot/internal/modules/testcoverage"

	"code.gitea.io/sdk/gitea"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	database = "/var/gopher-bot/database/sqlite.db"
)

type responseMsg struct {
	Msg string `json:"message"`
}

func HandlerWebHook(c *gin.Context) {
	// Нахуа? Мб удалить
	if c.Request.Header["X-Gitea-Event"] == nil || len(c.Request.Header["X-Gitea-Event"]) != 1 {
		log.Println("missing header")
		c.JSON(http.StatusNotFound, responseMsg{"missing header"})
		return
	}

	secret := config.Config.Server.Secret
	if secret != "" {
		sigRaw := c.Request.Header["X-Gitea-Signature"]
		if len(sigRaw) != 1 {
			log.Println("bad secret header")
			c.JSON(http.StatusBadRequest, responseMsg{"bad secret header"})
			return
		}
		sig := sigRaw[0]

		var body []byte
		if cb, ok := c.Get(gin.BodyBytesKey); ok {
			if cbb, ok := cb.([]byte); ok {
				body = cbb
			}
		}
		if body == nil {
			var err error
			body, err = io.ReadAll(c.Request.Body)
			if err != nil {
				log.Printf("io.ReadAll: %s", err)
				c.JSON(http.StatusInternalServerError, responseMsg{err.Error()})
				return
			}
			c.Set(gin.BodyBytesKey, body)
		}

		sig256 := hmac.New(sha256.New, []byte(secret))
		_, err := io.Writer(sig256).Write(body)
		if err != nil {
			log.Printf("io.Writer: %s", err)
			c.JSON(http.StatusInternalServerError, responseMsg{err.Error()})
			return
		}

		sigExpected := hex.EncodeToString(sig256.Sum(nil))

		if sig != sigExpected {
			log.Println("bad secret")
			c.JSON(http.StatusUnauthorized, responseMsg{"bad secret"})
			return
		}

	}

	switch c.Request.Header["X-Gitea-Event"][0] {
	case "pull_request":
		if !config.Config.Server.AllowPR {
			log.Printf("pull_request: event not allowed")
			c.JSON(404, responseMsg{"event not supported"})
			return
		}
		var h models.PRHook

		if err := c.ShouldBindBodyWith(&h, binding.JSON); err != nil {
			log.Printf("ShouldBindBodyWith: %s", err)
			c.JSON(http.StatusBadRequest, responseMsg{err.Error()})
			return
		}

		if h.Action != "opened" && h.Action != "synchronized" && h.Action != "reopened" {
			log.Printf("action not supported: %s", h.Action)
			c.JSON(403, responseMsg{"action not supported"})
			return
		}
		if strings.HasPrefix(h.Title, "WIP:") {
			log.Printf("title not allowed: %s", h.Title)
			c.JSON(403, responseMsg{"title not allowed"})
			return
		}
		if config.Config.Server.Owner != "" && h.Repository.Owner.UserName != config.Config.Server.Owner {
			log.Printf("owner not allowed: %s", h.Repository.Owner.UserName)
			c.JSON(403, responseMsg{"owner not allowed"})
			return
		}
		if config.Config.Server.Repo != "" && h.Repository.Name != config.Config.Server.Repo {
			log.Printf("repo not allowed: %s", h.Repository.Name)
			c.JSON(403, responseMsg{"repo not allowed"})
			return
		}

		go startCheckPR(&h)
		c.JSON(http.StatusCreated, responseMsg{"created"})
	default:
		log.Printf("event not supported: %s", c.Request.Header["X-Gitea-Event"][0])
		c.JSON(404, responseMsg{"event not supported"})
	}
}

func startCheckPR(hook *models.PRHook) {
	time.Sleep(3 * time.Second) // sleep 3 seconds because Gitea takes a short time to update the remote repo
	runCheckPR(hook)
}

func runCheckPR(hook *models.PRHook) {
	instance := models.RepositoryInstance(hook.Repository)
	c, err := gitea.NewClient(instance, gitea.SetToken(config.Config.Token(instance).Token), gitea.SetDebugMode())
	if err != nil {
		log.Printf("NewClient: %s", err)
		finishPr("NewClient", err, hook)
		return
	}

	if strings.Contains(hook.PullRequest.Title, config.Config.Server.Skip) {
		return
	} else {
		var commit *gitea.Commit
		commit, _, err = c.GetSingleCommit(hook.Repository.Owner.UserName, hook.Repository.Name, hook.PullRequest.Head.SHA)
		if err != nil {
			log.Printf("GetSingleCommit: %s", err)
			finishPr("GetSingleCommit", err, hook)
			return
		}
		if strings.Contains(commit.RepoCommit.Message, config.Config.Server.Skip) {
			return
		}
	}

	workingDir, err := os.MkdirTemp("", "gopher-bot-*")
	if err != nil {
		log.Printf("TempDir: %s", err)
		finishPr("TempDir", err, hook)
		return
	}

	r, err := git.PlainClone(workingDir, false, &git.CloneOptions{
		Auth:              config.Config.Token(instance).Git(),
		URL:               hook.Repository.CloneURL,
		ReferenceName:     plumbing.NewBranchReferenceName(hook.PullRequest.Head.Ref),
		RecurseSubmodules: git.NoRecurseSubmodules,
	})
	if err != nil {
		log.Printf("PlainClone: %s", err)
		finishPr("PlainClone", err, hook)
		return
	}

	_, err = c.CreateReviewRequests(hook.PullRequest.Base.Repo.Owner.UserName, hook.PullRequest.Base.Repo.Name, hook.Number, gitea.PullReviewRequestOptions{
		Reviewers: []string{
			"gopher-bot",
		},
	})
	if err != nil {
		log.Printf("CreateReviewRequests: %v", err)
		finishPr("CreateReviewRequests", err, hook)
	}

	runner := NewRunner(workingDir, r, hook,
		func(ctx context.Context, workingDir string, r *git.Repository, hook *models.PRHook) error {
			var cmderr error

			command := cmd.NewCommand(workingDir)
			makefile := filepath.Clean(filepath.Join(workingDir, "Makefile"))
			if _, err = os.Stat(makefile); errors.Is(err, os.ErrNotExist) {
				log.Printf("Makefile not found")
				_, cmderr = command.Run(ctx, "go", "build")
			} else {
				_, cmderr = command.Run(ctx, "make", "build")
			}

			if cmderr != nil {
				return fmt.Errorf("**Build error**\n ```\n%s\n```", cmderr.Error())
			}

			return nil
		},
		func(ctx context.Context, workingDir string, r *git.Repository, hook *models.PRHook) error {
			var cmdStdout string
			var cmdErr error

			command := cmd.NewCommand(workingDir)
			makefile := filepath.Clean(filepath.Join(workingDir, "Makefile"))
			if _, err = os.Stat(makefile); errors.Is(err, os.ErrNotExist) {
				log.Printf("Makefile not found")
				cmdStdout, cmdErr = command.Run(ctx, "golangci-lint", "run", "-v")
			} else {
				cmdStdout, cmdErr = command.Run(ctx, "make", "lint")
			}

			if cmdErr != nil {
				if len(cmdStdout) == 0 {
					return fmt.Errorf("**Golangci-lint error**\n ```\n%s\n```", cmdErr.Error())
				}
				return fmt.Errorf("**Golangci-lint error**\n ```\n%s\n```", cmdStdout)
			}

			return nil
		},
		func(ctx context.Context, workingDir string, r *git.Repository, hook *models.PRHook) error {
			var cmdStdout string
			var cmdErr error

			command := cmd.NewCommand(workingDir)
			makefile := filepath.Clean(filepath.Join(workingDir, "Makefile"))
			if _, err = os.Stat(makefile); errors.Is(err, os.ErrNotExist) {
				log.Printf("Makefile not found")
				cmdStdout, cmdErr = command.Run(ctx, "go", "test", "./...")
			} else {
				cmdStdout, cmdErr = command.Run(ctx, "make", "test")
			}

			if cmdErr != nil {
				if len(cmdStdout) == 0 {
					return fmt.Errorf("**Test error**\n ```\n%s\n```", cmdErr.Error())
				}
				return fmt.Errorf("**Test error**\n ```\n%s\n\n\n\n%s\n```", cmdStdout, cmdErr.Error())
			}

			return nil
		},
	)

	err = <-runner.Run()
	if err != nil {
		log.Printf("Runner.Run: %s", err)
	}

	go func() {
		defer os.RemoveAll(workingDir)
		repo := testcoverage.NewRepo(database)
		tc := testcoverage.NewTestCoverage(hook.Repository.Name, workingDir, repo)
		errCover := tc.IsUpCoverage(context.Background())
		if errCover != nil {
			_, _, createIssueCommentErr := c.CreateIssueComment(hook.Repository.Owner.UserName, hook.Repository.Name, hook.Number, gitea.CreateIssueCommentOption{
				Body: fmt.Sprintf("**Warning: Test coverage error**\n ```\n%s\n```", errCover.Error()),
			})

			if createIssueCommentErr != nil {
				log.Printf("CreateIssueComment: %s", errCover)
			}
		}
	}()

	if err != nil {
		_, _, err = c.CreatePullReview(hook.PullRequest.Base.Repo.Owner.UserName, hook.PullRequest.Base.Repo.Name, hook.Number, gitea.CreatePullReviewOptions{
			State: gitea.ReviewStateRequestChanges,
			Body:  err.Error(),
		})
		if err != nil {
			log.Printf("CreatePullReview: %v", err)
		}
		finishPr("run go build", errors.New("build error"), hook)
		return
	} else {
		_, _, err = c.CreatePullReview(hook.PullRequest.Base.Repo.Owner.UserName, hook.PullRequest.Base.Repo.Name, hook.Number, gitea.CreatePullReviewOptions{
			State: gitea.ReviewStateApproved,
		})
		if err != nil {
			finishPr("CreatePullReview", err, hook)
			return
		}
	}

	finishPr("", nil, hook)
}

func finishPr(tag string, err error, hook *models.PRHook) {
	if err != nil {
		SetStatus(hook.Repository, hook.PullRequest.Head.SHA, gitea.StatusError, fmt.Sprintf("%s: %v\n", tag, err), true)
	} else {
		SetStatus(hook.Repository, hook.PullRequest.Head.SHA, gitea.StatusSuccess, "", true)
	}
}
