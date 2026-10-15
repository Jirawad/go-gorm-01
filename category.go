package main

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func createCategoryDB(db *gorm.DB, name, description string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("ชื่อประเภทสินค้าต้องไม่ว่าง")
	}

	var count int64
	if err := db.Model(&Category{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("มีประเภทสินค้าชื่อ %q อยู่แล้ว", name)
	}

	category := Category{Name: name, Description: strings.TrimSpace(description)}
	if err := db.Create(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// getAllCategoriesDB ดึงประเภทสินค้าทั้งหมด
func getAllCategoriesDB(db *gorm.DB) ([]Category, error) {
	var categories []Category
	err := db.Order("id").Find(&categories).Error
	return categories, err
}

// deleteCategoryByIDDB ลบประเภทสินค้า
func deleteCategoryByIDDB(db *gorm.DB, id uint) error {
	var category Category
	if err := db.First(&category, id).Error; err != nil {
		return wrapNotFound(err, "ประเภทสินค้า", id)
	}

	var count int64
	if err := db.Unscoped().Model(&Product{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("ลบประเภท %q ไม่ได้: ยังมีสินค้า %d รายการอยู่ในประเภทนี้", category.Name, count)
	}

	return db.Unscoped().Delete(&category).Error
}
