package dao

import "gorm.io/gorm"

// Option customizes a gorm.DB query.
type Option func(*gorm.DB) *gorm.DB

// Repo is a lightweight DAO wrapper over gorm.DB.
type Repo struct {
	db *gorm.DB
}

// New creates a Repo with the provided gorm.DB.
func New(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func applyOptions(d *gorm.DB, opts ...Option) *gorm.DB {
	for _, opt := range opts {
		if opt != nil {
			d = opt(d)
		}
	}
	return d
}

// Create inserts a new record.
func (r *Repo) Create(model any) error {
	return r.db.Create(model).Error
}

// GetByID finds a record by primary key into dest.
func (r *Repo) GetByID(dest any, id any, opts ...Option) error {
	d := applyOptions(r.db, opts...)
	return d.First(dest, id).Error
}

// First finds the first record that matches options into dest.
func (r *Repo) First(dest any, opts ...Option) error {
	d := applyOptions(r.db, opts...)
	return d.First(dest).Error
}

// List finds records into dest.
func (r *Repo) List(dest any, opts ...Option) error {
	d := applyOptions(r.db, opts...)
	return d.Find(dest).Error
}

// Update updates fields for model with provided updates.
func (r *Repo) Update(model any, updates any, opts ...Option) error {
	d := applyOptions(r.db, opts...)
	return d.Model(model).Updates(updates).Error
}

// Delete deletes records that match options on model.
func (r *Repo) Delete(model any, opts ...Option) error {
	d := applyOptions(r.db, opts...)
	return d.Delete(model).Error
}

// DeleteByID deletes a record by primary key.
func (r *Repo) DeleteByID(model any, id any) error {
	return r.db.Delete(model, id).Error
}

// Count counts records for model with options.
func (r *Repo) Count(model any, opts ...Option) (int64, error) {
	d := applyOptions(r.db, opts...)
	var total int64
	if err := d.Model(model).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// Query helpers.
func WithWhere(query any, args ...any) Option {
	return func(d *gorm.DB) *gorm.DB {
		return d.Where(query, args...)
	}
}

func WithOrder(order string) Option {
	return func(d *gorm.DB) *gorm.DB {
		if order == "" {
			return d
		}
		return d.Order(order)
	}
}

func WithLimit(limit int) Option {
	return func(d *gorm.DB) *gorm.DB {
		if limit <= 0 {
			return d
		}
		return d.Limit(limit)
	}
}

func WithOffset(offset int) Option {
	return func(d *gorm.DB) *gorm.DB {
		if offset < 0 {
			return d
		}
		return d.Offset(offset)
	}
}

func WithSelect(query any, args ...any) Option {
	return func(d *gorm.DB) *gorm.DB {
		return d.Select(query, args...)
	}
}

func WithPreload(associations ...string) Option {
	return func(d *gorm.DB) *gorm.DB {
		for _, a := range associations {
			if a != "" {
				d = d.Preload(a)
			}
		}
		return d
	}
}

func WithJoins(query string, args ...any) Option {
	return func(d *gorm.DB) *gorm.DB {
		if query == "" {
			return d
		}
		return d.Joins(query, args...)
	}
}
