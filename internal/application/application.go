package application

import (
	"devon-demo-articles/internal/httpapi"
	"devon-demo-articles/internal/repository"
	"devon-demo-articles/internal/security"
	"devon-demo-articles/internal/service"
	"net/http"
)

type Application struct {
	Handler http.Handler
	Repo    *repository.Repository
}

func New(databasePath, secret string) (*Application, error) {
	tokens, err := security.New(secret)
	if err != nil {
		return nil, err
	}
	repo, err := repository.Open(databasePath)
	if err != nil {
		return nil, err
	}
	if err := repo.Migrate(); err != nil {
		_ = repo.Close()
		return nil, err
	}
	return &Application{Handler: httpapi.New(service.New(repo, tokens)), Repo: repo}, nil
}

func (a *Application) Close() error { return a.Repo.Close() }
