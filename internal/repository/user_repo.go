package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/zetatez/assistant/internal/models"
)

type UserRepository interface {
	Create(u *models.User) error
	GetByID(id uint) (*models.User, error)
	List(offset, limit int) ([]models.User, int64, error)
	Update(u *models.User) error
	Delete(id uint) error
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }

func (r *userRepository) Create(u *models.User) error {
	return r.db.Create(u).Error
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) List(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	r.db.Model(&models.User{}).Count(&total)
	if err := r.db.Offset(offset).Limit(limit).Order("id desc").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) Update(u *models.User) error {
	res := r.db.Model(&models.User{}).Where("id = ?", u.ID).Updates(map[string]interface{}{"name": u.Name, "email": u.Email})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *userRepository) Delete(id uint) error {
	res := r.db.Delete(&models.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
