package appinit

import (
	"context"
	"fmt"

	"github.com/dkarczmarski/go-kweb-lang/appinit/bootstrap"
	"github.com/dkarczmarski/go-kweb-lang/appinit/config"
	"github.com/dkarczmarski/go-kweb-lang/appinit/runtime"
)

func NewApp(
	flagRepoDir *string,
	flagCacheDir *string,
	flagLangCodes *string,
	flagOnce *bool,
	flagInterval *int,
	flagGitHubToken *string,
	flagGitHubTokenFile *string,
	flagSkipGit *bool,
	flagSkipPR *bool,
	flagNoWeb *bool,
	flagWebHTTPAddr *string,
) (*bootstrap.App, error) {
	cfg := config.Default()

	if err := config.FromEnv(&cfg); err != nil {
		return nil, fmt.Errorf("load config from env: %w", err)
	}

	config.ApplyFlags(
		&cfg,
		flagRepoDir,
		flagCacheDir,
		flagLangCodes,
		flagOnce,
		flagInterval,
		flagGitHubToken,
		flagGitHubTokenFile,
		flagSkipGit,
		flagSkipPR,
		flagNoWeb,
		flagWebHTTPAddr,
	)

	config.Show(cfg, true)

	if err := config.ReadGitHubTokenFile(&cfg, true, true); err != nil {
		return nil, fmt.Errorf("read github token file: %w", err)
	}

	app, err := bootstrap.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap app: %w", err)
	}

	return app, nil
}

func Run(ctx context.Context, app *bootstrap.App) error {
	if err := runtime.Run(ctx, app); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}
