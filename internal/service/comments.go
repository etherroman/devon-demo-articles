package service

import (
	"devon-demo-articles/internal/model"
	"strings"
	"time"
)

type CommentInput struct {
	Body string `json:"body"`
}
type CommentView struct {
	ID        uint       `json:"id"`
	Body      string     `json:"body"`
	Author    AuthorView `json:"author"`
	CreatedAt time.Time  `json:"createdAt"`
}

func commentView(c model.Comment) CommentView {
	return CommentView{ID: c.ID, Body: c.Body, Author: AuthorView{c.Author.Username}, CreatedAt: c.CreatedAt}
}

func (s *Service) UpdateComment(slug string, id, userID uint, in CommentInput) (CommentView, error) {
	a, err := s.Repo.ArticleBySlug(slug)
	if err != nil {
		return CommentView{}, storageError(err)
	}
	c, err := s.Repo.Comment(a.ID, id)
	if err != nil {
		return CommentView{}, storageError(err)
	}
	if c.AuthorID != userID {
		return CommentView{}, Forbidden()
	}
	if !validText(in.Body, 5000) {
		return CommentView{}, Invalid(map[string]string{"body": "댓글은 1~5,000자입니다."})
	}
	if err := s.Repo.UpdateComment(&c, strings.TrimSpace(in.Body)); err != nil {
		return CommentView{}, storageError(err)
	}
	return commentView(c), nil
}

func (s *Service) CreateComment(slug string, userID uint, in CommentInput) (CommentView, error) {
	a, err := s.Repo.ArticleBySlug(slug)
	if err != nil {
		return CommentView{}, storageError(err)
	}
	if !validText(in.Body, 5000) {
		return CommentView{}, Invalid(map[string]string{"body": "댓글은 1~5,000자입니다."})
	}
	c := model.Comment{ArticleID: a.ID, AuthorID: userID, Body: strings.TrimSpace(in.Body)}
	if err := s.Repo.CreateComment(&c); err != nil {
		return CommentView{}, storageError(err)
	}
	c, err = s.Repo.Comment(a.ID, c.ID)
	if err != nil {
		return CommentView{}, storageError(err)
	}
	return commentView(c), nil
}
func (s *Service) Comments(slug string) ([]CommentView, error) {
	a, err := s.Repo.ArticleBySlug(slug)
	if err != nil {
		return nil, storageError(err)
	}
	comments, err := s.Repo.Comments(a.ID)
	if err != nil {
		return nil, storageError(err)
	}
	views := make([]CommentView, 0, len(comments))
	for _, comment := range comments {
		views = append(views, commentView(comment))
	}
	return views, nil
}
func (s *Service) DeleteComment(slug string, id, userID uint) error {
	a, err := s.Repo.ArticleBySlug(slug)
	if err != nil {
		return storageError(err)
	}
	c, err := s.Repo.Comment(a.ID, id)
	if err != nil {
		return storageError(err)
	}
	if c.AuthorID != userID {
		return Forbidden()
	}
	return storageError(s.Repo.DeleteComment(&c))
}
