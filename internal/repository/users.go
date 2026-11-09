package repository

import "devon-demo-articles/internal/model"

func (r *Repository) CreateUser(user *model.User) error { return r.DB.Create(user).Error }
func (r *Repository) UserByEmail(email string) (model.User, error) {
	var user model.User
	err := r.DB.Where("email = ?", email).First(&user).Error
	return user, err
}
func (r *Repository) UserByID(id uint) (model.User, error) {
	var user model.User
	err := r.DB.First(&user, id).Error
	return user, err
}
