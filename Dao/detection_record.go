package dao

import (
	models "RPW_Detection/Models"
)

func (r *Repo) FindDetectionRecordByID(id uint) (*models.DetectionRecord, error) {
	var record models.DetectionRecord
	if err := r.GetByID(&record, id, WithPreload("User", "Park", "Tree", "Device")); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repo) FindDetectionRecordByJobID(jobID string) (*models.DetectionRecord, error) {
	var record models.DetectionRecord
	if err := r.First(
		&record,
		WithWhere("job_id = ?", jobID),
		WithPreload("User", "Park", "Tree", "Device"),
	); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repo) ListDetectionRecords(opts ...Option) ([]models.DetectionRecord, error) {
	var records []models.DetectionRecord
	if err := r.List(&records, opts...); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repo) ListDetectionRecordsByUserID(userID uint, opts ...Option) ([]models.DetectionRecord, error) {
	base := []Option{
		WithWhere("user_id = ?", userID),
		WithOrder("id DESC"),
	}
	return r.ListDetectionRecords(append(base, opts...)...)
}

func (r *Repo) UpdateDetectionRecordByJobID(jobID string, updates any) error {
	return r.Update(&models.DetectionRecord{}, updates, WithWhere("job_id = ?", jobID))
}
