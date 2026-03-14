package config

type Config struct {
	RepoDir         string
	CacheDir        string
	LangCodes       []string
	RunOnce         bool
	RunInterval     int
	GitHubToken     string
	GitHubUserAgent string
	GitHubTokenFile string
	SkipGitChecking bool
	SkipPRChecking  bool
	NoWeb           bool
	WebHTTPAddr     string
}

func Default() Config {
	//nolint:exhaustruct
	return Config{
		RepoDir:         "./.appdata/kubernetes-website",
		CacheDir:        "./.appdata/cache",
		GitHubTokenFile: ".github-token.txt",
		WebHTTPAddr:     ":8080",
	}
}
