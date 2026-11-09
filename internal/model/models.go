// Package model defines persisted domain relationships, without HTTP concerns.
package model

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"size:32;uniqueIndex;not null"`
	Email        string `gorm:"size:254;uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    time.Time
}

type Article struct {
	ID          uint       `gorm:"primaryKey"`
	Slug        string     `gorm:"size:180;uniqueIndex;not null"`
	Title       string     `gorm:"size:200;not null"`
	Description string     `gorm:"size:500;not null"`
	Body        string     `gorm:"not null"`
	AuthorID    uint       `gorm:"index;not null"`
	Author      User       `gorm:"constraint:OnDelete:RESTRICT"`
	Tags        []Tag      `gorm:"many2many:article_tags;constraint:OnDelete:CASCADE"`
	Comments    []Comment  `gorm:"constraint:OnDelete:CASCADE"`
	Favorites   []Favorite `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:32;uniqueIndex;not null"`
}

type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	ArticleID uint   `gorm:"index;not null"`
	AuthorID  uint   `gorm:"index;not null"`
	Author    User   `gorm:"constraint:OnDelete:RESTRICT"`
	Body      string `gorm:"not null"`
	CreatedAt time.Time
}

// The composite primary key makes repeated favorites safe at the database layer.
type Favorite struct {
	ArticleID uint `gorm:"primaryKey"`
	UserID    uint `gorm:"primaryKey"`
	User      User `gorm:"constraint:OnDelete:CASCADE"`
}
