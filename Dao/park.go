package dao

import (
	models "RPW_Detection/Models"
)

func (r *Repo) FindParkByID(id uint) (*models.Park, error) {
	var park models.Park
	if err := r.GetByID(&park, id); err != nil {
		return nil, err
	}
	return &park, nil
}

func (r *Repo) FindParkByCode(code string) (*models.Park, error) {
	var park models.Park
	if err := r.First(&park, WithWhere("code = ?", code)); err != nil {
		return nil, err
	}
	return &park, nil
}

func (r *Repo) FindParkByName(name string) (*models.Park, error) {
	var park models.Park
	if err := r.First(&park, WithWhere("name = ?", name)); err != nil {
		return nil, err
	}
	return &park, nil
}

func (r *Repo) ListParks(opts ...Option) ([]models.Park, error) {
	var parks []models.Park
	if err := r.List(&parks, opts...); err != nil {
		return nil, err
	}
	return parks, nil
}

func (r *Repo) ListActiveParks() ([]models.Park, error) {
	return r.ListParks(
		WithWhere("status = ?", 1),
		WithOrder("id DESC"),
	)
}
