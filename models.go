package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const dbName = "store_db"

// สถานะของคำสั่งซื้อ
const (
	StatusPending   = "Pending"
	StatusCompleted = "Completed"
	StatusCancelled = "Cancelled"
)

// Category
type Category struct {
	gorm.Model
	Name        string `gorm:"size:100;not null"`
	Description string `gorm:"size:255"`
}

// Product
type Product struct {
	gorm.Model
	Name       string   `gorm:"size:150;not null;index"`
	Price      float64  `gorm:"type:decimal(12,2);not null"`
	Quantity   int      `gorm:"not null"`
	CategoryID uint     `gorm:"not null;index"`
	Category   Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

// Customer
type Customer struct {
	gorm.Model
	Name  string `gorm:"size:100;not null"`
	Email string `gorm:"size:150;not null;index"`
	Phone string `gorm:"size:30"`
}

// Order
type Order struct {
	gorm.Model
	CustomerID uint     `gorm:"not null;index"`
	Customer   Customer `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ProductID  uint     `gorm:"not null;index"`
	Product    Product  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Quantity   int      `gorm:"not null"`
	TotalPrice float64  `gorm:"type:decimal(12,2);not null"`
	Status     string   `gorm:"size:20;not null;index"`
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// buildDSN
func buildDSN(database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "127.0.0.1"),
		getEnv("DB_PORT", "3306"),
		database,
	)
}

// createDatabaseIfNotExists
func createDatabaseIfNotExists() error {
	serverDB, err := gorm.Open(mysql.Open(buildDSN("")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("เชื่อมต่อ MySQL server ไม่สำเร็จ: %w", err)
	}
	sqlDB, err := serverDB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)
	if err := serverDB.Exec(query).Error; err != nil {
		return fmt.Errorf("สร้างฐานข้อมูล %s ไม่สำเร็จ: %w", dbName, err)
	}
	return nil
}

// connectDB
func connectDB() (*gorm.DB, error) {
	if err := createDatabaseIfNotExists(); err != nil {
		return nil, err
	}

	quietLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		},
	)

	db, err := gorm.Open(mysql.Open(buildDSN(dbName)), &gorm.Config{Logger: quietLogger})
	if err != nil {
		return nil, fmt.Errorf("เชื่อมต่อฐานข้อมูล %s ไม่สำเร็จ: %w", dbName, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping ฐานข้อมูลไม่สำเร็จ: %w", err)
	}

	if err := db.AutoMigrate(&Category{}, &Customer{}, &Product{}, &Order{}); err != nil {
		return nil, fmt.Errorf("AutoMigrate ไม่สำเร็จ: %w", err)
	}
	return db, nil
}

// wrapNotFound
func wrapNotFound(err error, what string, id uint) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("ไม่พบ%s ID %d", what, id)
	}
	return err
}
