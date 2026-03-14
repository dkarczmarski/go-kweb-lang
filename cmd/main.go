package main

import (
	"context"
	"flag"
	"log"

	"github.com/dkarczmarski/go-kweb-lang/appinit"
)

//nolint:gochecknoglobals
var (
	flagRepoDir         = flag.String("repo-dir", "", "kubernetes website repository directory path")
	flagCacheDir        = flag.String("cache-dir", "", "cache directory path")
	flagLangCodes       = flag.String("lang-codes", "", "allowed lang codes")
	flagRunOnce         = flag.Bool("run-once", false, "run synchronization once at startup")
	flagRunInterval     = flag.Int("run-interval", 0, "run repeatedly with delay of N minutes between runs")
	flagGitHubToken     = flag.String("github-token", "", "github api access token")
	flagGitHubTokenFile = flag.String("github-token-file", "", "file path with github api access token")
	flagSkipGit         = flag.Bool("skip-git", false, "skip git repo checking")
	flagSkipPR          = flag.Bool("skip-pr", false, "skip pull request checking")
	flagNoWeb           = flag.Bool("no-web", false, "disable web server")
	flagWebHTTPAddr     = flag.String("web-http-addr", "", "TCP address for the server to listen on")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	app, err := appinit.NewApp(
		flagRepoDir,
		flagCacheDir,
		flagLangCodes,
		flagRunOnce,
		flagRunInterval,
		flagGitHubToken,
		flagGitHubTokenFile,
		flagSkipGit,
		flagSkipPR,
		flagNoWeb,
		flagWebHTTPAddr,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := appinit.Run(ctx, app); err != nil {
		log.Fatal(err)
	}
}
