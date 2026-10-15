package main

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductUpdate
type ProductUpdate struct {
	Name       *string
	Price      *float64
	Quantity   *int
	CategoryID *uint
}

// productQuery
func productQuery(db *gorm.DB) *gorm.DB {
	return db.Preload("Category")
}

// escapeLike
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

// ---------------------------------------------------------------- Create

// createProductDB
func createProductDB(db *gorm.DB, name string, price float64, quantity int, categoryID uint) (*Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("ชื่อสินค้าต้องไม่ว่าง")
	}
	if price < 0 {
		return nil, errors.New("ราคาสินค้าต้องไม่ติดลบ")
	}
	if quantity < 0 {
		return nil, errors.New("จำนวนสต็อกต้องไม่ติดลบ")
	}

	var category Category
	if err := db.First(&category, categoryID).Error; err != nil {
		return nil, wrapNotFound(err, "ประเภทสินค้า", categoryID)
	}

	product := Product{Name: name, Price: price, Quantity: quantity, CategoryID: categoryID}
	if err := db.Create(&product).Error; err != nil {
		return nil, err
	}
	product.Category = category
	return &product, nil
}

// ---------------------------------------------------------------- Read

func getAllProductsDB(db *gorm.DB) ([]Product, error) {
	var products []Product
	err := productQuery(db).Order("id").Find(&products).Error
	return products, err
}

func getProductByIDDB(db *gorm.DB, id uint) (*Product, error) {
	var product Product
	if err := productQuery(db).First(&product, id).Error; err != nil {
		return nil, wrapNotFound(err, "สินค้า", id)
	}
	return &product, nil
}

// getProductByNameDB
func getProductByNameDB(db *gorm.DB, keyword string) ([]Product, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("คำค้นหาต้องไม่ว่าง")
	}
	var products []Product
	err := productQuery(db).
		Where("name LIKE ?", "%"+escapeLike(keyword)+"%").
		Order("id").Find(&products).Error
	return products, err
}

func getProductsByCategoryDB(db *gorm.DB, categoryID uint) ([]Product, error) {
	var products []Product
	err := productQuery(db).Where("category_id = ?", categoryID).Order("id").Find(&products).Error
	return products, err
}

func getProductsByPriceRangeDB(db *gorm.DB, minPrice, maxPrice float64) ([]Product, error) {
	if minPrice > maxPrice {
		return nil, errors.New("ราคาต่ำสุดต้องไม่มากกว่าราคาสูงสุด")
	}
	var products []Product
	err := productQuery(db).
		Where("price BETWEEN ? AND ?", minPrice, maxPrice).
		Order("price").Find(&products).Error
	return products, err
}

// getProductsLowStockDB
func getProductsLowStockDB(db *gorm.DB, threshold int) ([]Product, error) {
	var products []Product
	err := productQuery(db).Where("quantity < ?", threshold).Order("quantity, id").Find(&products).Error
	return products, err
}

func getProductsByCategoryAndPriceDB(db *gorm.DB, categoryID uint, minPrice, maxPrice float64) ([]Product, error) {
	if minPrice > maxPrice {
		return nil, errors.New("ราคาต่ำสุดต้องไม่มากกว่าราคาสูงสุด")
	}
	var products []Product
	err := productQuery(db).
		Where("category_id = ? AND price BETWEEN ? AND ?", categoryID, minPrice, maxPrice).
		Order("price").Find(&products).Error
	return products, err
}

// getProductsSortedByPriceDB
func getProductsSortedByPriceDB(db *gorm.DB, direction string) ([]Product, error) {
	dir := strings.ToUpper(strings.TrimSpace(direction))
	if dir != "ASC" && dir != "DESC" {
		return nil, errors.New("direction ต้องเป็น ASC หรือ DESC")
	}
	var products []Product
	err := productQuery(db).Order("price " + dir + ", id").Find(&products).Error
	return products, err
}

// getTopExpensiveProductsDB
func getTopExpensiveProductsDB(db *gorm.DB, n int) ([]Product, error) {
	if n <= 0 {
		return nil, errors.New("N ต้องมากกว่า 0")
	}
	var products []Product
	err := productQuery(db).Order("price DESC, id").Limit(n).Find(&products).Error
	return products, err
}

// ---------------------------------------------------------------- Update

// updateProductByIDDB
func updateProductByIDDB(db *gorm.DB, id uint, u ProductUpdate) (*Product, error) {
	var product Product
	if err := db.First(&product, id).Error; err != nil {
		return nil, wrapNotFound(err, "สินค้า", id)
	}

	updates := map[string]interface{}{}
	if u.Name != nil {
		name := strings.TrimSpace(*u.Name)
		if name == "" {
			return nil, errors.New("ชื่อสินค้าต้องไม่ว่าง")
		}
		updates["name"] = name
	}
	if u.Price != nil {
		if *u.Price < 0 {
			return nil, errors.New("ราคาสินค้าต้องไม่ติดลบ")
		}
		updates["price"] = *u.Price
	}
	if u.Quantity != nil {
		if *u.Quantity < 0 {
			return nil, errors.New("จำนวนสต็อกต้องไม่ติดลบ")
		}
		updates["quantity"] = *u.Quantity
	}
	if u.CategoryID != nil {
		var category Category
		if err := db.First(&category, *u.CategoryID).Error; err != nil {
			return nil, wrapNotFound(err, "ประเภทสินค้า", *u.CategoryID)
		}
		updates["category_id"] = *u.CategoryID
	}
	if len(updates) == 0 {
		return nil, errors.New("ไม่มีข้อมูลให้แก้ไข")
	}

	if err := db.Model(&product).Updates(updates).Error; err != nil {
		return nil, err
	}
	return getProductByIDDB(db, id)
}

// updateProductPriceDB
func updateProductPriceDB(db *gorm.DB, id uint, price float64) (*Product, error) {
	if price < 0 {
		return nil, errors.New("ราคาสินค้าต้องไม่ติดลบ")
	}
	var product Product
	if err := db.First(&product, id).Error; err != nil {
		return nil, wrapNotFound(err, "สินค้า", id)
	}
	if err := db.Model(&product).Update("price", price).Error; err != nil {
		return nil, err
	}
	return getProductByIDDB(db, id)
}

// updateProductStockDB
func updateProductStockDB(db *gorm.DB, id uint, delta int) (*Product, error) {
	err := db.Transaction(func(tx *gorm.DB) error {
		var product Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, id).Error; err != nil {
			return wrapNotFound(err, "สินค้า", id)
		}
		newQty := product.Quantity + delta
		if newQty < 0 {
			return fmt.Errorf("สต็อกไม่พอ: %s เหลือ %d ชิ้น ไม่สามารถลด %d ชิ้นได้", product.Name, product.Quantity, -delta)
		}
		return tx.Model(&product).Update("quantity", newQty).Error
	})
	if err != nil {
		return nil, err
	}
	return getProductByIDDB(db, id)
}

// ---------------------------------------------------------------- Delete

// softDeleteProductDB Soft Delete
func softDeleteProductDB(db *gorm.DB, id uint) error {
	var product Product
	if err := db.First(&product, id).Error; err != nil {
		return wrapNotFound(err, "สินค้า", id)
	}
	return db.Delete(&product).Error
}

// hardDeleteProductDB
func hardDeleteProductDB(db *gorm.DB, id uint) error {
	var product Product
	if err := db.Unscoped().First(&product, id).Error; err != nil {
		return wrapNotFound(err, "สินค้า", id)
	}

	var count int64
	if err := db.Unscoped().Model(&Order{}).Where("product_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("ลบถาวร %q ไม่ได้: สินค้านี้อยู่ในประวัติคำสั่งซื้อ %d รายการ (ใช้ Soft Delete แทนได้)", product.Name, count)
	}
	return db.Unscoped().Delete(&product).Error
}

// getDeletedProductsDB
func getDeletedProductsDB(db *gorm.DB) ([]Product, error) {
	var products []Product
	err := productQuery(db).Unscoped().Where("deleted_at IS NOT NULL").Order("id").Find(&products).Error
	return products, err
}

// restoreProductDB
func restoreProductDB(db *gorm.DB, id uint) (*Product, error) {
	var product Product
	if err := db.Unscoped().Where("deleted_at IS NOT NULL").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ไม่พบสินค้า ID %d ในรายการที่ถูกลบ", id)
		}
		return nil, err
	}

	var category Category
	if err := db.First(&category, product.CategoryID).Error; err != nil {
		return nil, fmt.Errorf("กู้คืนไม่ได้: %w", wrapNotFound(err, "ประเภทสินค้า", product.CategoryID))
	}

	if err := db.Unscoped().Model(&Product{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}
	return getProductByIDDB(db, id)
}
