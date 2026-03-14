package web

import "strconv"

const (
	githubWebsiteBlobBaseURL   = "https://github.com/kubernetes/website/blob/main/"
	githubWebsiteCommitBaseURL = "https://github.com/kubernetes/website/commit/"
	githubWebsitePullBaseURL   = "https://github.com/kubernetes/website/pull/"
)

type GitHubLinks struct{}

func (g GitHubLinks) File(path string) string {
	return githubWebsiteBlobBaseURL + path
}

func (g GitHubLinks) Commit(id string) string {
	return githubWebsiteCommitBaseURL + id
}

func (g GitHubLinks) PR(number int) string {
	return githubWebsitePullBaseURL + strconv.Itoa(number)
}
