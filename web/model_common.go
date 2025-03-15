package web

type LinkModel struct {
	Text string
	Url  string
}

func toLinkModel(commitId string) LinkModel {
	return LinkModel{
		Text: commitId,
		Url:  "https://github.com/kubernetes/website/commit/" + commitId,
	}
}
