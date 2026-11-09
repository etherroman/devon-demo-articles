package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devon-demo-articles/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct{ DB *gorm.DB }

func Open(path string) (*Repository, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			return nil, err
		}
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	for _, pragma := range []string{"PRAGMA foreign_keys = ON", "PRAGMA busy_timeout = 5000"} {
		if err := db.Exec(pragma).Error; err != nil {
			_ = sqlDB.Close()
			return nil, err
		}
	}
	return &Repository{DB: db}, nil
}

func (r *Repository) Migrate() error {
	return r.DB.AutoMigrate(&model.User{}, &model.Article{}, &model.Tag{}, &model.Comment{}, &model.Favorite{})
}

func (r *Repository) Close() error {
	db, err := r.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
func IsConflict(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey) || (err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed"))
}
