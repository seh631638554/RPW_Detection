package httpserver

import (
	"time"
)

// UploadService encapsulates upload business flow.
type UploadService struct {
	storage StorageService
	cfg     *ObjectStorageConfig
}

func NewUploadService(storage StorageService, cfg *ObjectStorageConfig) *UploadService {
	if cfg == nil {
		cfg = DefaultObjectStorageConfig()
	}
	return &UploadService{
		storage: storage,
		cfg:     cfg,
	}
}

func (s *UploadService) CreateUploadJob(req CreateUploadJobRequest) (*CreateUploadJobResponse, error) {
	jobID := GenerateJobID()
	storageKey := GenerateStorageKey(req.DeviceID, req.FileName)

	ttl := time.Duration(s.cfg.ExpireHours) * time.Hour
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	metadata := map[string]string{
		"device_id":   req.DeviceID,
		"job_id":      jobID,
		"file_type":   req.FileType,
		"description": req.Description,
		"upload_time": time.Now().Format(time.RFC3339),
	}

	uploadURL, err := s.storage.GeneratePresignedUploadURL(PresignedURLParams{
		Bucket:      s.cfg.Bucket,
		Key:         storageKey,
		Method:      "PUT",
		Expires:     ttl,
		ContentType: req.ContentType,
		Metadata:    metadata,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &CreateUploadJobResponse{
		JobID:          jobID,
		UploadURL:      uploadURL,
		Bucket:         s.cfg.Bucket,
		Key:            storageKey,
		TTL:            int64(ttl.Seconds()),
		ExpiresAt:      now.Add(ttl),
		ContentType:    req.ContentType,
		MaxFileSize:    req.FileSize,
		RequiredFields: []string{"file"},
		Status:         string(JobStatusPending),
		CreatedAt:      now,
	}, nil
}
