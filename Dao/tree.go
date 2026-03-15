package dao

import (
	models "RPW_Detection/Models"
)

func (r *Repo) FindTreeByID(id uint) (*models.Tree, error) {
	var tree models.Tree
	if err := r.GetByID(&tree, id, WithPreload("Park")); err != nil {
		return nil, err
	}
	return &tree, nil
}

func (r *Repo) FindTreeByParkAndCode(parkID uint, treeCode string) (*models.Tree, error) {
	var tree models.Tree
	if err := r.First(
		&tree,
		WithWhere("park_id = ? AND tree_code = ?", parkID, treeCode),
		WithPreload("Park"),
	); err != nil {
		return nil, err
	}
	return &tree, nil
}

func (r *Repo) ListTrees(opts ...Option) ([]models.Tree, error) {
	var trees []models.Tree
	if err := r.List(&trees, opts...); err != nil {
		return nil, err
	}
	return trees, nil
}

func (r *Repo) ListTreesByParkID(parkID uint) ([]models.Tree, error) {
	return r.ListTrees(
		WithWhere("park_id = ?", parkID),
		WithOrder("id DESC"),
	)
}

func (r *Repo) ListActiveTreesByParkID(parkID uint) ([]models.Tree, error) {
	return r.ListTrees(
		WithWhere("park_id = ? AND status = ?", parkID, 1),
		WithOrder("id DESC"),
	)
}
