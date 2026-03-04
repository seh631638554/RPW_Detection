package dao

import (
	models "RPW_Detection/Models"
)

func (r *Repo) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.First(&user, WithWhere("username = ?", username))
	if err != nil {
		return nil, err
	}
	return &user, nil
}
