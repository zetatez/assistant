package user

import (
	"assistant/internal/db"
	"assistant/internal/model"
)

type Repo struct{}

func NewRepo() *Repo {
	return &Repo{}
}

func (r *Repo) FindAll() ([]model.User, error) {
	var users []model.User
	result := db.DB.Find(&users)
	return users, result.Error
}

func (r *Repo) Create(name string) (model.User, error) {
	u := model.User{Name: name}
	result := db.DB.Create(&u)
	return u, result.Error
}
