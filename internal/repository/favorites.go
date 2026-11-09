package repository

import (
	"devon-demo-articles/internal/model"
	"gorm.io/gorm/clause"
)

func (r *Repository) AddFavorite(articleID, userID uint) error {
	return r.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.Favorite{ArticleID: articleID, UserID: userID}).Error
}
func (r *Repository) RemoveFavorite(articleID, userID uint) error {
	return r.DB.Where("article_id = ? AND user_id = ?", articleID, userID).Delete(&model.Favorite{}).Error
}
