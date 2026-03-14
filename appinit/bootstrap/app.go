package bootstrap

import "github.com/dkarczmarski/go-kweb-lang/appinit/config"

type App struct {
	Config   config.Config
	Services *Services
}

func New(cfg config.Config) (*App, error) {
	services, err := BuildServices(cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:   cfg,
		Services: services,
	}, nil
}
