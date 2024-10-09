package modules

import (
	"code.gitea.io/sdk/gitea"
	"github.com/nfort/gopher-bot/internal/models"
	"github.com/nfort/gopher-bot/internal/modules/config"
)

func SetStatus(repo *gitea.Repository, commit string, status gitea.StatusState, desc string, isPR bool) {
	instance := models.RepositoryInstance(repo)
	c, err := gitea.NewClient(instance, gitea.SetToken(config.Config.Token(instance).Token), gitea.SetDebugMode())
	if err != nil {
		return
	}

	ctx := config.Config.Server.StatusContext
	if isPR {
		ctx = config.Config.Server.StatusContextPR
	}

	targetURL := ""
	if !isPR {
		targetURL = config.FullURL() + repo.Owner.UserName + "/" + repo.Name + "/" + commit + "?instance=" + instance
	}

	_, _, _ = c.CreateStatus(repo.Owner.UserName, repo.Name, commit, gitea.CreateStatusOption{
		State:       status,
		TargetURL:   targetURL,
		Description: desc,
		Context:     ctx,
	})
}
