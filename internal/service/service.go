package service

import (
	"devon-demo-articles/internal/repository"
	"devon-demo-articles/internal/security"
)

type Service struct {
	Repo   *repository.Repository
	Tokens *security.Tokens
}

func New(repo *repository.Repository, tokens *security.Tokens) *Service {
	return &Service{Repo: repo, Tokens: tokens}
}
