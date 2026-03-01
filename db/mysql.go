package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// New creates a GORM DB instance.
// dsn example: "user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"
func New(dsn string, maxOpenConns, maxIdleConns int, connLifetime time.Duration) (*gorm.DB, error) {
	if dsn == "" {
		return nil, errors.New("db: empty dsn")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if maxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(maxOpenConns)
	}
	if maxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(maxIdleConns)
	}
	if connLifetime > 0 {
		sqlDB.SetConnMaxLifetime(connLifetime)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes the underlying sql.DB.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
