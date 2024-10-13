package modules

import (
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

	"github.com/nfort/gopher-bot/internal/models"
	"github.com/nfort/gopher-bot/internal/modules/config"
	"github.com/nfort/gopher-bot/pkg/cmd"

	"code.gitea.io/sdk/gitea"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type responseMsg struct {
	Msg string `json:"message"`
}

func HandlerWebHook(c *gin.Context) {
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
	defer os.RemoveAll(workingDir)

	r, err := git.PlainClone(workingDir, false, &git.CloneOptions{
		Auth:              config.Config.Token(instance).Git(),
		URL:               hook.Repository.CloneURL,
		Depth:             1,
		ReferenceName:     plumbing.NewBranchReferenceName(hook.PullRequest.Head.Ref),
		RecurseSubmodules: git.NoRecurseSubmodules,
	})
	if err != nil {
		log.Printf("PlainClone: %s", err)
		finishPr("PlainClone", err, hook)
		return
	}

	w, err := r.Worktree()
	if err != nil {
		log.Printf("Worktree: %s", err)
		finishPr("Worktree", err, hook)
		return
	}

	ref, err := r.Head()
	if err != nil {
		log.Printf("Head: %s", err)
		finishPr("Head", err, hook)
		return
	}

	err = w.Reset(&git.ResetOptions{
		Commit: ref.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		log.Printf("Reset: %s", err)
		finishPr("Reset", err, hook)
		return
	}

	user, _, err := c.GetUserInfo("gopher-bot")
	if err != nil {
		log.Printf("GetUserInfo: %s", err)
	}

	_, err = c.CreateReviewRequests(hook.PullRequest.Base.Repo.Owner.UserName, hook.PullRequest.Base.Repo.Name, hook.Number, gitea.PullReviewRequestOptions{
		Reviewers: []string{
			user.UserName,
		},
	})
	if err != nil {
		log.Printf("CreateReviewRequests: %v", err)
		finishPr("CreateReviewRequests", err, hook)
	}

	var cmderr error
	command := cmd.NewCommand(workingDir)
	makefile := filepath.Clean(filepath.Join(workingDir, "Makefile"))
	if _, err = os.Stat(makefile); errors.Is(err, os.ErrNotExist) {
		log.Printf("Makefile not found")
		cmderr = command.Run("go", "build")
	} else {
		cmderr = command.Run("make", "build")
	}

	// Почему то с этим не работает..., если пользак gopher-bot
	// c.SetSudo("gopher-bot")
	if cmderr != nil {
		_, _, err = c.CreatePullReview(hook.PullRequest.Base.Repo.Owner.UserName, hook.PullRequest.Base.Repo.Name, hook.Number, gitea.CreatePullReviewOptions{
			State: gitea.ReviewStateRequestChanges,
			Body:  "**Build error**\n ```\n" + cmderr.Error() + "\n```",
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
