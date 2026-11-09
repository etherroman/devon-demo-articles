package repository

import (
	"devon-demo-articles/internal/model"
	"gorm.io/gorm"
)

type ArticleFilter struct {
	Tag, Author   string
	Limit, Offset int
}

func articleQuery(db *gorm.DB) *gorm.DB {
	return db.Preload("Author").Preload("Tags").Preload("Favorites")
}

func (r *Repository) ArticleBySlug(slug string) (model.Article, error) {
	var article model.Article
	err := articleQuery(r.DB).Where("slug = ?", slug).First(&article).Error
	return article, err
}

func (r *Repository) ListArticles(filter ArticleFilter) ([]model.Article, int64, error) {
	query := r.DB.Model(&model.Article{})
	if filter.Tag != "" {
		query = query.Where("articles.id IN (SELECT article_tags.article_id FROM article_tags JOIN tags ON tags.id = article_tags.tag_id WHERE tags.name = ?)", filter.Tag)
	}
	if filter.Author != "" {
		query = query.Where("articles.author_id IN (SELECT id FROM users WHERE username = ?)", filter.Author)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	articles := []model.Article{}
	err := articleQuery(query).Order("articles.id DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&articles).Error
	return articles, count, err
}

func ensureTags(tx *gorm.DB, names []string) ([]model.Tag, error) {
	tags := make([]model.Tag, 0, len(names))
	for _, name := range names {
		tag := model.Tag{Name: name}
		if err := tx.Where("name = ?", name).FirstOrCreate(&tag).Error; err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *Repository) CreateArticle(article *model.Article, tags []string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Tags", "Author", "Favorites", "Comments").Create(article).Error; err != nil {
			return err
		}
		storedTags, err := ensureTags(tx, tags)
		if err != nil {
			return err
		}
		return tx.Model(article).Association("Tags").Replace(storedTags)
	})
}

func (r *Repository) UpdateArticle(article *model.Article, changes map[string]any, tags *[]string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(article).Updates(changes).Error; err != nil {
			return err
		}
		if tags != nil {
			storedTags, err := ensureTags(tx, *tags)
			if err != nil {
				return err
			}
			return tx.Model(article).Association("Tags").Replace(storedTags)
		}
		return nil
	})
}

func (r *Repository) DeleteArticle(article *model.Article) error { return r.DB.Delete(article).Error }

func (r *Repository) Tags() ([]string, error) {
	tags := []string{}
	err := r.DB.Model(&model.Tag{}).Where("id IN (SELECT tag_id FROM article_tags)").Order("name ASC").Pluck("name", &tags).Error
	return tags, err
}
