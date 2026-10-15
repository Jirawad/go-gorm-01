package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatalf("✗ %v\n  (ตรวจสอบว่า MySQL เปิดอยู่ และตั้งค่า DB_USER / DB_PASSWORD / DB_HOST / DB_PORT ให้ถูกต้อง)", err)
	}
	fmt.Printf("✓ เชื่อมต่อฐานข้อมูล %s และสร้างตารางเรียบร้อย\n", dbName)

	if err := seedData(db); err != nil {
		log.Fatalf("✗ เพิ่มข้อมูลทดสอบไม่สำเร็จ: %v", err)
	}

	runMenu(db)
}

// ---------------------------------------------------------------- เมนู

func runMenu(db *gorm.DB) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n========== Store Management System ==========")
		fmt.Println(" 1) ทดสอบ Read / Conditional Queries")
		fmt.Println(" 2) ทดสอบ Create / Update / Delete (เพิ่มคำสั่งซื้อทดสอบที่ถูกยกเลิก 1 รายการต่อครั้ง)")
		fmt.Println(" 3) รันทั้งหมด")
		fmt.Println(" 0) ออกจากโปรแกรม")
		fmt.Print("เลือกเมนู: ")

		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			fmt.Println()
			return
		}

		switch strings.TrimSpace(line) {
		case "1":
			runReadDemo(db)
		case "2":
			runWriteDemo(db)
		case "3":
			runReadDemo(db)
			runWriteDemo(db)
		case "0":
			fmt.Println("ลาก่อน")
			return
		default:
			fmt.Println("กรุณาเลือก 0-3")
		}
	}
}

// ---------------------------------------------------------------- ฟังก์ชันช่วยแสดงผล

func report(label string, err error) {
	if err != nil {
		fmt.Printf("  ✗ %s → %v\n", label, err)
		return
	}
	fmt.Printf("  ✓ %s\n", label)
}

func printProducts(title string, products []Product, err error) {
	fmt.Printf("\n--- %s ---\n", title)
	if err != nil {
		fmt.Println("  ข้อผิดพลาด:", err)
		return
	}
	if len(products) == 0 {
		fmt.Println("  (ไม่มีข้อมูล)")
		return
	}
	fmt.Printf("  %-4s %-32s %12s %6s  %s\n", "ID", "Name", "Price", "Qty", "Category")
	for _, p := range products {
		fmt.Printf("  %-4d %-32s %12.2f %6d  %s\n", p.ID, p.Name, p.Price, p.Quantity, p.Category.Name)
	}
}

func printOrders(title string, orders []Order, err error) {
	fmt.Printf("\n--- %s ---\n", title)
	if err != nil {
		fmt.Println("  ข้อผิดพลาด:", err)
		return
	}
	if len(orders) == 0 {
		fmt.Println("  (ไม่มีข้อมูล)")
		return
	}
	fmt.Printf("  %-4s %-20s %-30s %4s %12s  %-10s %s\n", "ID", "Customer", "Product", "Qty", "Total", "Status", "Created")
	for _, o := range orders {
		fmt.Printf("  %-4d %-20s %-30s %4d %12.2f  %-10s %s\n",
			o.ID, o.Customer.Name, o.Product.Name, o.Quantity, o.TotalPrice, o.Status, o.CreatedAt.Format("2006-01-02"))
	}
}

func printCustomer(title string, c *Customer, err error) {
	fmt.Printf("\n--- %s ---\n", title)
	if err != nil {
		fmt.Println("  ข้อผิดพลาด:", err)
		return
	}
	fmt.Printf("  ID %d | %s | %s | %s\n", c.ID, c.Name, c.Email, c.Phone)
}

func categoryIDByName(db *gorm.DB, name string) uint {
	var c Category
	if err := db.Where("name = ?", name).First(&c).Error; err != nil {
		return 0
	}
	return c.ID
}

func stockOf(db *gorm.DB, productID uint) int {
	p, err := getProductByIDDB(db, productID)
	if err != nil {
		return -1
	}
	return p.Quantity
}

// ---------------------------------------------------------------- Demo: Read

func runReadDemo(db *gorm.DB) {
	fmt.Println("\n================ READ / CONDITIONAL QUERIES ================")

	products, err := getAllProductsDB(db)
	printProducts("getAllProductsDB()", products, err)

	if p, err := getProductByIDDB(db, 1); err != nil {
		fmt.Println("\ngetProductByIDDB(1):", err)
	} else {
		printProducts("getProductByIDDB(1)", []Product{*p}, nil)
	}
	if _, err := getProductByIDDB(db, 99999); err != nil {
		fmt.Println("  (ทดสอบ ID ที่ไม่มีอยู่)", err)
	}

	products, err = getProductByNameDB(db, "mac")
	printProducts(`getProductByNameDB("mac")`, products, err)

	accessoriesID := categoryIDByName(db, "Accessories")
	products, err = getProductsByCategoryDB(db, accessoriesID)
	printProducts(fmt.Sprintf("getProductsByCategoryDB(%d)  [Accessories]", accessoriesID), products, err)

	products, err = getProductsByPriceRangeDB(db, 1000, 10000)
	printProducts("getProductsByPriceRangeDB(1000, 10000)", products, err)

	products, err = getProductsLowStockDB(db, 5)
	printProducts("getProductsLowStockDB(5)", products, err)

	products, err = getProductsByCategoryAndPriceDB(db, accessoriesID, 500, 5000)
	printProducts(fmt.Sprintf("getProductsByCategoryAndPriceDB(%d, 500, 5000)", accessoriesID), products, err)

	products, err = getProductsSortedByPriceDB(db, "ASC")
	printProducts(`getProductsSortedByPriceDB("ASC")`, products, err)

	products, err = getProductsSortedByPriceDB(db, "DESC")
	printProducts(`getProductsSortedByPriceDB("DESC")`, products, err)

	products, err = getTopExpensiveProductsDB(db, 3)
	printProducts("getTopExpensiveProductsDB(3)", products, err)

	customer, err := getCustomerByEmailDB(db, "somchai@example.com")
	printCustomer(`getCustomerByEmailDB("somchai@example.com")`, customer, err)

	if customer != nil {
		orders, err := getOrdersByCustomerDB(db, customer.ID)
		printOrders(fmt.Sprintf("getOrdersByCustomerDB(%d)", customer.ID), orders, err)
	}

	orders, err := getOrdersByStatusDB(db, "Pending")
	printOrders(`getOrdersByStatusDB("Pending")`, orders, err)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -14)
	orders, err = getOrdersByDateRangeDB(db, start, now)
	printOrders(fmt.Sprintf("getOrdersByDateRangeDB(%s → %s)", start.Format("2006-01-02"), now.Format("2006-01-02")), orders, err)

	orders, err = getOrdersWithDetailsDB(db)
	printOrders("getOrdersWithDetailsDB()", orders, err)
}

// ---------------------------------------------------------------- Demo: Create / Update / Delete

func runWriteDemo(db *gorm.DB) {
	fmt.Println("\n================ CREATE / UPDATE / DELETE ================")
	suffix := time.Now().Format("150405")

	// ---------- Category + Product
	fmt.Println("\n[Category & Product]")
	cat, err := createCategoryDB(db, "Demo Category "+suffix, "ประเภทสำหรับทดสอบ")
	report("createCategoryDB", err)
	if err != nil {
		return
	}
	_, err = createCategoryDB(db, cat.Name, "ชื่อซ้ำ")
	report("createCategoryDB (ชื่อซ้ำ → ควรล้มเหลว)", err)

	prod, err := createProductDB(db, "Demo Product "+suffix, 999, 10, cat.ID)
	report("createProductDB", err)
	if err != nil {
		_ = deleteCategoryByIDDB(db, cat.ID)
		return
	}
	_, err = createProductDB(db, "Bad Product", -1, 1, cat.ID)
	report("createProductDB (ราคาติดลบ → ควรล้มเหลว)", err)
	_, err = createProductDB(db, "Bad Product", 1, 1, 99999)
	report("createProductDB (ไม่มีประเภทนี้ → ควรล้มเหลว)", err)

	_, err = updateProductPriceDB(db, prod.ID, 1099)
	report("updateProductPriceDB(→ 1099)", err)
	_, err = updateProductPriceDB(db, prod.ID, -5)
	report("updateProductPriceDB(→ -5, ควรล้มเหลว)", err)

	_, err = updateProductStockDB(db, prod.ID, +5)
	report(fmt.Sprintf("updateProductStockDB(+5) สต็อกเป็น %d", stockOf(db, prod.ID)), err)
	_, err = updateProductStockDB(db, prod.ID, -100)
	report("updateProductStockDB(-100, ควรล้มเหลว)", err)

	newName := "Demo Product (renamed) " + suffix
	newQty := 20
	_, err = updateProductByIDDB(db, prod.ID, ProductUpdate{Name: &newName, Quantity: &newQty})
	report("updateProductByIDDB(name, quantity)", err)

	err = deleteCategoryByIDDB(db, cat.ID)
	report("deleteCategoryByIDDB (ยังมีสินค้า → ควรล้มเหลว)", err)

	report("softDeleteProductDB", softDeleteProductDB(db, prod.ID))
	deleted, err := getDeletedProductsDB(db)
	printProducts("getDeletedProductsDB()", deleted, err)
	_, err = restoreProductDB(db, prod.ID)
	report("restoreProductDB", err)
	report("softDeleteProductDB (อีกครั้ง)", softDeleteProductDB(db, prod.ID))
	report("hardDeleteProductDB", hardDeleteProductDB(db, prod.ID))
	report("deleteCategoryByIDDB (ไม่มีสินค้าแล้ว)", deleteCategoryByIDDB(db, cat.ID))

	// ---------- Customer
	fmt.Println("\n[Customer]")
	cust, err := createCustomerDB(db, "Demo Customer", "demo"+suffix+"@example.com", "0800000000")
	report("createCustomerDB", err)
	if err == nil {
		_, err = createCustomerDB(db, "Duplicate", cust.Email, "")
		report("createCustomerDB (อีเมลซ้ำ → ควรล้มเหลว)", err)
		newPhone := "0899999999"
		_, err = updateCustomerByIDDB(db, cust.ID, CustomerUpdate{Phone: &newPhone})
		report("updateCustomerByIDDB(phone)", err)
		report("deleteCustomerByIDDB (ไม่มีประวัติสั่งซื้อ)", deleteCustomerByIDDB(db, cust.ID))
	}
	_, err = createCustomerDB(db, "Bad Email", "not-an-email", "")
	report("createCustomerDB (อีเมลผิดรูปแบบ → ควรล้มเหลว)", err)

	// ---------- Order
	fmt.Println("\n[Order]")
	mice, errMouse := getProductByNameDB(db, "Logitech Mouse")
	somchai, errCust := getCustomerByEmailDB(db, "somchai@example.com")
	if errMouse != nil || errCust != nil || len(mice) == 0 {
		fmt.Println("  ไม่พบข้อมูลตัวอย่าง (Logitech Mouse / somchai@example.com) ข้ามการทดสอบคำสั่งซื้อ")
		return
	}
	mouse := mice[0]

	_, err = createOrderDB(db, somchai.ID, mouse.ID, 100000)
	report("createOrderDB (สต็อกไม่พอ → ควรล้มเหลว)", err)
	_, err = createOrderDB(db, 99999, mouse.ID, 1)
	report("createOrderDB (ไม่มีลูกค้า → ควรล้มเหลว)", err)

	before := stockOf(db, mouse.ID)
	order, err := createOrderDB(db, somchai.ID, mouse.ID, 2)
	report(fmt.Sprintf("createOrderDB (สั่ง 2 ชิ้น) สต็อก %d → %d", before, stockOf(db, mouse.ID)), err)
	if err != nil {
		return
	}
	fmt.Printf("    Order ID %d | TotalPrice = %.2f × 2 = %.2f | Status = %s\n", order.ID, mouse.Price, order.TotalPrice, order.Status)

	_, err = updateOrderStatusDB(db, order.ID, "Shipped")
	report(`updateOrderStatusDB("Shipped", ควรล้มเหลว)`, err)
	_, err = updateOrderStatusDB(db, order.ID, StatusCancelled)
	report(fmt.Sprintf("updateOrderStatusDB(Cancelled) คืนสต็อกเป็น %d", stockOf(db, mouse.ID)), err)
	_, err = updateOrderStatusDB(db, order.ID, StatusCompleted)
	report("updateOrderStatusDB(Cancelled → Completed, ควรล้มเหลว)", err)

	// ---------- กติกาการลบที่เกี่ยวกับประวัติคำสั่งซื้อ
	fmt.Println("\n[Delete rules กับประวัติคำสั่งซื้อ]")
	report("deleteCustomerByIDDB (ลูกค้ามีประวัติ → ควรล้มเหลว)", deleteCustomerByIDDB(db, somchai.ID))
	report("hardDeleteProductDB (สินค้าอยู่ในประวัติ → ควรล้มเหลว)", hardDeleteProductDB(db, mouse.ID))
	fmt.Println("\nเสร็จสิ้น")
}

// ---------------------------------------------------------------- Seed data

func seedData(db *gorm.DB) error {
	var n int64
	if err := db.Model(&Category{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		fmt.Println("พบข้อมูลในฐานข้อมูลแล้ว ข้ามการเพิ่มข้อมูลทดสอบ")
		return nil
	}
	fmt.Println("กำลังเพิ่มข้อมูลทดสอบ (4 ประเภท, 15 สินค้า, 8 ลูกค้า, 12 คำสั่งซื้อ)...")

	categoryData := []struct{ Name, Desc string }{
		{"Laptops", "โน้ตบุ๊กและแล็ปท็อป"},
		{"Smartphones", "สมาร์ทโฟน"},
		{"Accessories", "อุปกรณ์เสริมและอุปกรณ์ต่อพ่วง"},
		{"Storage & Network", "อุปกรณ์จัดเก็บข้อมูลและเครือข่าย"},
	}
	cats := make([]*Category, 0, len(categoryData))
	for _, c := range categoryData {
		cat, err := createCategoryDB(db, c.Name, c.Desc)
		if err != nil {
			return err
		}
		cats = append(cats, cat)
	}

	productData := []struct {
		Name  string
		Price float64
		Qty   int
		Cat   int // index ใน cats
	}{
		{"MacBook Air M3", 39900, 8, 0},
		{"ASUS Vivobook 15", 19900, 12, 0},
		{"Lenovo ThinkPad E14", 24900, 6, 0},
		{"Acer Aspire 5", 15900, 3, 0},
		{"iPhone 16", 29900, 15, 1},
		{"Samsung Galaxy S24", 27900, 10, 1},
		{"Xiaomi Redmi Note 13", 7990, 25, 1},
		{"OPPO Reno 11", 15990, 4, 1},
		{"Logitech Mouse MX Master 3S", 3490, 30, 2},
		{"Mechanical Keyboard", 2490, 18, 2},
		{"USB-C Hub 7-in-1", 1290, 40, 2},
		{"Laptop Backpack", 990, 2, 2},
		{"External SSD 1TB", 2990, 14, 3},
		{"WiFi 6 Router", 1890, 7, 3},
		{"Webcam 1080p", 1190, 0, 3},
	}
	prods := make([]*Product, 0, len(productData))
	for _, p := range productData {
		prod, err := createProductDB(db, p.Name, p.Price, p.Qty, cats[p.Cat].ID)
		if err != nil {
			return err
		}
		prods = append(prods, prod)
	}

	customerData := []struct{ Name, Email, Phone string }{
		{"Somchai Jaidee", "somchai@example.com", "0811111111"},
		{"Suda Rakdee", "suda@example.com", "0822222222"},
		{"Anan Boonmee", "anan@example.com", "0833333333"},
		{"Malee Srisuk", "malee@example.com", "0844444444"},
		{"Prasert Wongsawat", "prasert@example.com", "0855555555"},
		{"Nattaya Charoen", "nattaya@example.com", "0866666666"},
		{"Kittisak Thongdee", "kittisak@example.com", "0877777777"},
		{"Pimchanok Saetang", "pimchanok@example.com", "0888888888"},
	}
	custs := make([]*Customer, 0, len(customerData))
	for _, c := range customerData {
		cust, err := createCustomerDB(db, c.Name, c.Email, c.Phone)
		if err != nil {
			return err
		}
		custs = append(custs, cust)
	}

	orderData := []struct {
		Cust, Prod, Qty int // index ใน custs / prods
		Status          string
		DaysAgo         int
	}{
		{0, 0, 1, StatusCompleted, 25},
		{1, 4, 1, StatusCompleted, 22},
		{2, 8, 2, StatusCompleted, 20},
		{3, 9, 1, StatusCompleted, 18},
		{4, 10, 3, StatusCompleted, 15},
		{5, 1, 1, StatusCompleted, 12},
		{6, 5, 1, StatusPending, 8},
		{7, 12, 2, StatusPending, 6},
		{0, 11, 1, StatusPending, 4},
		{1, 13, 1, StatusCancelled, 3},
		{2, 6, 2, StatusPending, 2},
		{3, 3, 1, StatusPending, 0},
	}
	for _, o := range orderData {
		order, err := createOrderDB(db, custs[o.Cust].ID, prods[o.Prod].ID, o.Qty)
		if err != nil {
			return err
		}
		if o.Status != StatusPending {
			if _, err := updateOrderStatusDB(db, order.ID, o.Status); err != nil {
				return err
			}
		}
		// ย้อนวันที่สร้างคำสั่งซื้อ เพื่อให้ทดสอบค้นหาตามช่วงวันที่ได้
		createdAt := time.Now().AddDate(0, 0, -o.DaysAgo)
		if err := db.Model(&Order{}).Where("id = ?", order.ID).UpdateColumn("created_at", createdAt).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ เพิ่มข้อมูลทดสอบเรียบร้อย")
	return nil
}
