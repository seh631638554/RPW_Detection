package dao

import (
	models "RPW_Detection/Models"
)

func (r *Repo) FindDeviceByID(id uint) (*models.Device, error) {
	var device models.Device
	if err := r.GetByID(&device, id, WithPreload("Park", "Tree")); err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *Repo) FindDeviceByCode(code string) (*models.Device, error) {
	var device models.Device
	if err := r.First(
		&device,
		WithWhere("device_code = ?", code),
		WithPreload("Park", "Tree"),
	); err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *Repo) ListDevices(opts ...Option) ([]models.Device, error) {
	var devices []models.Device
	if err := r.List(&devices, opts...); err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *Repo) ListDevicesByParkID(parkID uint) ([]models.Device, error) {
	return r.ListDevices(
		WithWhere("park_id = ?", parkID),
		WithOrder("id DESC"),
	)
}

func (r *Repo) ListDevicesByTreeID(treeID uint) ([]models.Device, error) {
	return r.ListDevices(
		WithWhere("tree_id = ?", treeID),
		WithOrder("id DESC"),
	)
}

func (r *Repo) ListActiveDevices(opts ...Option) ([]models.Device, error) {
	opts = append([]Option{WithWhere("status = ?", 1), WithOrder("id DESC")}, opts...)
	return r.ListDevices(opts...)
}
