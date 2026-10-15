package main

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"gorm.io/gorm"
)

// CustomerUpdate ใช้แก้ไขลูกค้า
type CustomerUpdate struct {
	Name  *string
	Email *string
	Phone *string
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

// createCustomerDB เพิ่มลูกค้า
func createCustomerDB(db *gorm.DB, name, email, phone string) (*Customer, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" {
		return nil, errors.New("ชื่อลูกค้าต้องไม่ว่าง")
	}
	if !validEmail(email) {
		return nil, fmt.Errorf("รูปแบบอีเมลไม่ถูกต้อง: %q", email)
	}

	var count int64
	if err := db.Model(&Customer{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("อีเมล %s ถูกใช้งานแล้ว", email)
	}

	customer := Customer{Name: name, Email: email, Phone: strings.TrimSpace(phone)}
	if err := db.Create(&customer).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func getAllCustomersDB(db *gorm.DB) ([]Customer, error) {
	var customers []Customer
	err := db.Order("id").Find(&customers).Error
	return customers, err
}

// getCustomerByEmailDB
func getCustomerByEmailDB(db *gorm.DB, email string) (*Customer, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var customer Customer
	if err := db.Where("email = ?", email).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ไม่พบลูกค้าที่ใช้อีเมล %s", email)
		}
		return nil, err
	}
	return &customer, nil
}

// updateCustomerByIDDB
func updateCustomerByIDDB(db *gorm.DB, id uint, u CustomerUpdate) (*Customer, error) {
	var customer Customer
	if err := db.First(&customer, id).Error; err != nil {
		return nil, wrapNotFound(err, "ลูกค้า", id)
	}

	updates := map[string]interface{}{}
	if u.Name != nil {
		name := strings.TrimSpace(*u.Name)
		if name == "" {
			return nil, errors.New("ชื่อลูกค้าต้องไม่ว่าง")
		}
		updates["name"] = name
	}
	if u.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*u.Email))
		if !validEmail(email) {
			return nil, fmt.Errorf("รูปแบบอีเมลไม่ถูกต้อง: %q", email)
		}
		var count int64
		if err := db.Model(&Customer{}).Where("email = ? AND id <> ?", email, id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("อีเมล %s ถูกใช้งานแล้ว", email)
		}
		updates["email"] = email
	}
	if u.Phone != nil {
		updates["phone"] = strings.TrimSpace(*u.Phone)
	}
	if len(updates) == 0 {
		return nil, errors.New("ไม่มีข้อมูลให้แก้ไข")
	}

	if err := db.Model(&customer).Updates(updates).Error; err != nil {
		return nil, err
	}
	var updated Customer
	if err := db.First(&updated, id).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}

// deleteCustomerByIDDB ลบลูกค้า
func deleteCustomerByIDDB(db *gorm.DB, id uint) error {
	var customer Customer
	if err := db.First(&customer, id).Error; err != nil {
		return wrapNotFound(err, "ลูกค้า", id)
	}

	var count int64
	if err := db.Unscoped().Model(&Order{}).Where("customer_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("ลบลูกค้า %q ไม่ได้: มีประวัติคำสั่งซื้อ %d รายการ", customer.Name, count)
	}
	return db.Unscoped().Delete(&customer).Error
}
