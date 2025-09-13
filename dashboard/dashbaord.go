package dashboard

import (
	"go-kweb-lang/gitseek"
)

type Dashboard struct {
	LangCode string
	Items    []Item
}

type Item struct {
	gitseek.FileInfo
	PRs []int
}
