package models

import "code.gitea.io/sdk/gitea"

type PRHook struct {
	Action      string            `json:"action"`
	Number      int64             `json:"number"`
	Title       string            `json:"title"`
	PullRequest *PullRequest      `json:"pull_request"`
	Repository  *gitea.Repository `json:"repository"`
}

type PullRequest struct {
	Number int64   `json:"number"`
	Title  string  `json:"title"`
	Base   *Branch `json:"base"`
	Head   *Branch `json:"head"`
}

type Branch struct {
	Ref  string            `json:"ref"`
	Repo *gitea.Repository `json:"repo"`
	SHA  string            `json:"sha"`
}

type LocalPR struct {
	ID        int64 `xorm:"pk autoincr"`
	Index     int64
	RepoID    int64
	CommentID int64
}
