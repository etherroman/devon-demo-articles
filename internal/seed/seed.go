package seed

import (
	"devon-demo-articles/internal/model"
	"devon-demo-articles/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Run inserts a small editorial community. Existing seed rows are preserved.
func Run(r *repository.Repository) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("DemoPass123!"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return r.DB.Transaction(func(tx *gorm.DB) error {
		users := map[string]model.User{}
		for _, name := range []string{"hana", "min", "june"} {
			user := model.User{Username: name, Email: name + "@example.com", PasswordHash: string(hash)}
			if err := tx.Where("email = ?", user.Email).FirstOrCreate(&user).Error; err != nil {
				return err
			}
			users[name] = user
		}
		articles := map[string]model.Article{}
		for _, fixture := range articleFixtures {
			article := model.Article{Slug: fixture.Slug, Title: fixture.Title, Description: fixture.Description, Body: fixture.Body, AuthorID: users[fixture.Author].ID}
			result := tx.Where("slug = ?", fixture.Slug).Omit("Tags", "Author", "Comments", "Favorites").FirstOrCreate(&article)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				tags := []model.Tag{}
				for _, name := range fixture.Tags {
					tag := model.Tag{Name: name}
					if err := tx.Where("name = ?", name).FirstOrCreate(&tag).Error; err != nil {
						return err
					}
					tags = append(tags, tag)
				}
				if err := tx.Model(&article).Association("Tags").Replace(tags); err != nil {
					return err
				}
			}
			articles[fixture.Slug] = article
		}
		for _, fixture := range commentFixtures {
			comment := model.Comment{ArticleID: articles[fixture.Slug].ID, AuthorID: users[fixture.Author].ID, Body: fixture.Body}
			if err := tx.Where("article_id = ? AND author_id = ? AND body = ?", comment.ArticleID, comment.AuthorID, comment.Body).Omit("Author").FirstOrCreate(&comment).Error; err != nil {
				return err
			}
		}
		for _, fixture := range favoriteFixtures {
			favorite := model.Favorite{ArticleID: articles[fixture.Slug].ID, UserID: users[fixture.User].ID}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Omit("User").Create(&favorite).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
