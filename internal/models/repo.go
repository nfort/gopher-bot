package models

import (
	"strings"

	"code.gitea.io/sdk/gitea"
)

type RepoHook struct {
	Ref        string            `json:"ref"`
	HeadCommit *Commit           `json:"head_commit"`
	Repository *gitea.Repository `json:"repository"`
}

func LocalRepository(r *gitea.Repository) *LocalRepo {
	return &LocalRepo{
		Instance: RepositoryInstance(r),
		Name:     r.Name,
		Owner:    r.Owner.UserName,
	}
}

func RepositoryInstance(r *gitea.Repository) string {
	url := strings.Split(r.CloneURL, "/")
	url = url[:len(url)-2]
	return strings.Join(url, "/")
}

type User struct {
	Login string `json:"login"`
}

type Commit struct {
	SHA     string `json:"id"`
	Message string `json:"message"`
}

type LocalRepo struct {
	ID       int64 `xorm:"pk autoincr"`
	Instance string
	Name     string
	Owner    string
}
