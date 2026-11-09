package repository

import "devon-demo-articles/internal/model"

func (r *Repository) CreateComment(comment *model.Comment) error {
	return r.DB.Omit("Author").Create(comment).Error
}
func (r *Repository) Comments(articleID uint) ([]model.Comment, error) {
	comments := []model.Comment{}
	err := r.DB.Preload("Author").Where("article_id = ?", articleID).Order("id ASC").Find(&comments).Error
	return comments, err
}
func (r *Repository) Comment(articleID, id uint) (model.Comment, error) {
	var comment model.Comment
	err := r.DB.Preload("Author").Where("article_id = ? AND id = ?", articleID, id).First(&comment).Error
	return comment, err
}
func (r *Repository) DeleteComment(comment *model.Comment) error { return r.DB.Delete(comment).Error }
func (r *Repository) UpdateComment(comment *model.Comment, body string) error {
	return r.DB.Model(comment).Update("body", body).Error
}
