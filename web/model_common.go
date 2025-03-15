package web

type LinkModel struct {
	Text string
	URL  string
}

func toLinkModel(commitID string) LinkModel {
	return LinkModel{
		Text: commitID,
		URL:  "https://github.com/kubernetes/website/commit/" + commitID,
	}
}
