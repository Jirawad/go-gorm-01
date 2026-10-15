package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func normalizeStatus(s string) (string, bool) {
	s = strings.TrimSpace(s)
	for _, v := range []string{StatusPending, StatusCompleted, StatusCancelled} {
		if strings.EqualFold(s, v) {
			return v, true
		}
	}
	return "", false
}

// orderWithDetails Preload
func orderWithDetails(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Customer", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() }).
		Preload("Product", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() })
}

// ---------------------------------------------------------------- Create

// createOrderDB
func createOrderDB(db *gorm.DB, customerID, productID uint, quantity int) (*Order, error) {
	if quantity <= 0 {
		return nil, errors.New("จำนวนที่สั่งต้องมากกว่า 0")
	}

	var (
		order    Order
		customer Customer
		product  Product
	)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&customer, customerID).Error; err != nil {
			return wrapNotFound(err, "ลูกค้า", customerID)
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productID).Error; err != nil {
			return wrapNotFound(err, "สินค้า", productID)
		}
		if product.Quantity < quantity {
			return fmt.Errorf("สต็อกไม่พอ: %s เหลือ %d ชิ้น แต่สั่ง %d ชิ้น", product.Name, product.Quantity, quantity)
		}

		order = Order{
			CustomerID: customerID,
			ProductID:  productID,
			Quantity:   quantity,
			TotalPrice: round2(product.Price * float64(quantity)),
			Status:     StatusPending,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		return tx.Model(&product).Update("quantity", product.Quantity-quantity).Error
	})
	if err != nil {
		return nil, err
	}

	product.Quantity -= quantity
	order.Customer = customer
	order.Product = product
	return &order, nil
}

// ---------------------------------------------------------------- Read

func getOrdersByCustomerDB(db *gorm.DB, customerID uint) ([]Order, error) {
	var orders []Order
	err := orderWithDetails(db).Where("customer_id = ?", customerID).Order("created_at DESC, id DESC").Find(&orders).Error
	return orders, err
}

func getOrdersByStatusDB(db *gorm.DB, status string) ([]Order, error) {
	normalized, ok := normalizeStatus(status)
	if !ok {
		return nil, fmt.Errorf("สถานะ %q ไม่ถูกต้อง (ต้องเป็น Pending, Completed หรือ Cancelled)", status)
	}
	var orders []Order
	err := orderWithDetails(db).Where("status = ?", normalized).Order("id").Find(&orders).Error
	return orders, err
}

// getOrdersByDateRangeDB
func getOrdersByDateRangeDB(db *gorm.DB, start, end time.Time) ([]Order, error) {
	if end.Before(start) {
		return nil, errors.New("วันที่สิ้นสุดต้องไม่มาก่อนวันที่เริ่มต้น")
	}
	var orders []Order
	err := orderWithDetails(db).Where("created_at BETWEEN ? AND ?", start, end).Order("created_at, id").Find(&orders).Error
	return orders, err
}

// getOrdersWithDetailsDB
func getOrdersWithDetailsDB(db *gorm.DB) ([]Order, error) {
	var orders []Order
	err := orderWithDetails(db).Order("id").Find(&orders).Error
	return orders, err
}

// updateOrderStatusDB
func updateOrderStatusDB(db *gorm.DB, orderID uint, status string) (*Order, error) {
	newStatus, ok := normalizeStatus(status)
	if !ok {
		return nil, fmt.Errorf("สถานะ %q ไม่ถูกต้อง (ต้องเป็น Pending, Completed หรือ Cancelled)", status)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var order Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, orderID).Error; err != nil {
			return wrapNotFound(err, "คำสั่งซื้อ", orderID)
		}
		if order.Status == newStatus {
			return fmt.Errorf("คำสั่งซื้อ ID %d อยู่ในสถานะ %s อยู่แล้ว", orderID, newStatus)
		}
		if order.Status != StatusPending {
			return fmt.Errorf("เปลี่ยนสถานะไม่ได้: คำสั่งซื้อที่เป็น %s ถือว่าสิ้นสุดแล้ว", order.Status)
		}

		if newStatus == StatusCancelled {
			var product Product
			// สินค้าอาจถูก Soft Delete ไปแล้ว แต่ยังต้องคืนสต็อกได้
			if err := tx.Unscoped().Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, order.ProductID).Error; err != nil {
				return wrapNotFound(err, "สินค้า", order.ProductID)
			}
			if err := tx.Unscoped().Model(&product).Update("quantity", product.Quantity+order.Quantity).Error; err != nil {
				return err
			}
		}
		return tx.Model(&order).Update("status", newStatus).Error
	})
	if err != nil {
		return nil, err
	}

	var updated Order
	if err := orderWithDetails(db).First(&updated, orderID).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}
