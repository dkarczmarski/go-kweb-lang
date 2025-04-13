package internal

const (
	CategoryPrCommits    = "pr-pr-commits"
	CategoryCommitFiles  = "pr-commit-files"
	CategoryFilePrsIndex = "pr-fileprs-index"
)

type PRCommits struct {
	UpdatedAt string
	CommitIds []string
}
