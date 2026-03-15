package models

import "time"

type DetectionRecord struct {
	ID            uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobID         string    `gorm:"column:job_id;size:64;uniqueIndex;not null" json:"job_id"`
	UserID        uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	ParkID        uint      `gorm:"column:park_id;not null;index" json:"park_id"`
	TreeID        uint      `gorm:"column:tree_id;not null;index" json:"tree_id"`
	DeviceID      uint      `gorm:"column:device_id;not null;index" json:"device_id"`
	AudioBucket   string    `gorm:"column:audio_bucket;size:128;not null" json:"audio_bucket"`
	AudioKey      string    `gorm:"column:audio_key;size:512;not null" json:"audio_key"`
	FeatureBucket string    `gorm:"column:feature_bucket;size:128;default:''" json:"feature_bucket"`
	FeatureKey    string    `gorm:"column:feature_key;size:512;default:''" json:"feature_key"`
	Status        string    `gorm:"column:status;size:32;not null;default:uploaded;index" json:"status"`
	ResultLabel   string    `gorm:"column:result_label;size:32;default:''" json:"result_label"`
	ResultScore   *float64  `gorm:"column:result_score;type:decimal(8,6)" json:"result_score,omitempty"`
	ErrorMessage  string    `gorm:"column:error_message;size:255;default:''" json:"error_message"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	User   User   `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Park   Park   `gorm:"foreignKey:ParkID;references:ID" json:"park,omitempty"`
	Tree   Tree   `gorm:"foreignKey:TreeID;references:ID" json:"tree,omitempty"`
	Device Device `gorm:"foreignKey:DeviceID;references:ID" json:"device,omitempty"`
}

func (DetectionRecord) TableName() string {
	return "detection_records"
}
