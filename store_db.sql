-- ============================================================
CREATE DATABASE IF NOT EXISTS `store_db` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `store_db`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `orders`;
DROP TABLE IF EXISTS `products`;
DROP TABLE IF EXISTS `customers`;
DROP TABLE IF EXISTS `categories`;

-- ------------------------------------------------------------
-- categories
-- ------------------------------------------------------------
CREATE TABLE `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) NOT NULL,
  `description` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_categories_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ------------------------------------------------------------
-- customers
-- ------------------------------------------------------------
CREATE TABLE `customers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) NOT NULL,
  `email` varchar(150) NOT NULL,
  `phone` varchar(30) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_customers_deleted_at` (`deleted_at`),
  KEY `idx_customers_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ------------------------------------------------------------
-- products  (N:1 categories)
-- ------------------------------------------------------------
CREATE TABLE `products` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(150) NOT NULL,
  `price` decimal(12,2) NOT NULL,
  `quantity` bigint NOT NULL,
  `category_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_products_deleted_at` (`deleted_at`),
  KEY `idx_products_name` (`name`),
  KEY `idx_products_category_id` (`category_id`),
  CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ------------------------------------------------------------
-- orders  (N:1 customers, N:1 products)
-- ------------------------------------------------------------
CREATE TABLE `orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `customer_id` bigint unsigned NOT NULL,
  `product_id` bigint unsigned NOT NULL,
  `quantity` bigint NOT NULL,
  `total_price` decimal(12,2) NOT NULL,
  `status` varchar(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_orders_deleted_at` (`deleted_at`),
  KEY `idx_orders_customer_id` (`customer_id`),
  KEY `idx_orders_product_id` (`product_id`),
  KEY `idx_orders_status` (`status`),
  CONSTRAINT `fk_orders_customer` FOREIGN KEY (`customer_id`) REFERENCES `customers` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_orders_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ------------------------------------------------------------
-- DATA
-- ------------------------------------------------------------
INSERT INTO `categories` (`id`,`created_at`,`updated_at`,`deleted_at`,`name`,`description`) VALUES
(1,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Laptops','โน้ตบุ๊กและแล็ปท็อป'),
(2,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Smartphones','สมาร์ทโฟน'),
(3,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Accessories','อุปกรณ์เสริมและอุปกรณ์ต่อพ่วง'),
(4,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Storage & Network','อุปกรณ์จัดเก็บข้อมูลและเครือข่าย');

INSERT INTO `customers` (`id`,`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`phone`) VALUES
(1,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Somchai Jaidee','somchai@example.com','0811111111'),
(2,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Suda Rakdee','suda@example.com','0822222222'),
(3,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Anan Boonmee','anan@example.com','0833333333'),
(4,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Malee Srisuk','malee@example.com','0844444444'),
(5,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Prasert Wongsawat','prasert@example.com','0855555555'),
(6,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Nattaya Charoen','nattaya@example.com','0866666666'),
(7,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Kittisak Thongdee','kittisak@example.com','0877777777'),
(8,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Pimchanok Saetang','pimchanok@example.com','0888888888');

-- quantity = สต็อกคงเหลือหลังหักคำสั่งซื้อที่ยังไม่ถูกยกเลิกแล้ว
INSERT INTO `products` (`id`,`created_at`,`updated_at`,`deleted_at`,`name`,`price`,`quantity`,`category_id`) VALUES
(1,'2026-09-01 09:00:00.000','2026-09-20 10:00:00.000',NULL,'MacBook Air M3',39900.00,7,1),
(2,'2026-09-01 09:00:00.000','2026-10-03 10:00:00.000',NULL,'ASUS Vivobook 15',19900.00,11,1),
(3,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Lenovo ThinkPad E14',24900.00,6,1),
(4,'2026-09-01 09:00:00.000','2026-10-15 10:00:00.000',NULL,'Acer Aspire 5',15900.00,2,1),
(5,'2026-09-01 09:00:00.000','2026-09-23 10:00:00.000',NULL,'iPhone 16',29900.00,14,2),
(6,'2026-09-01 09:00:00.000','2026-10-07 10:00:00.000',NULL,'Samsung Galaxy S24',27900.00,9,2),
(7,'2026-09-01 09:00:00.000','2026-10-13 10:00:00.000',NULL,'Xiaomi Redmi Note 13',7990.00,23,2),
(8,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'OPPO Reno 11',15990.00,4,2),
(9,'2026-09-01 09:00:00.000','2026-09-25 10:00:00.000',NULL,'Logitech Mouse MX Master 3S',3490.00,28,3),
(10,'2026-09-01 09:00:00.000','2026-09-27 10:00:00.000',NULL,'Mechanical Keyboard',2490.00,17,3),
(11,'2026-09-01 09:00:00.000','2026-09-30 10:00:00.000',NULL,'USB-C Hub 7-in-1',1290.00,37,3),
(12,'2026-09-01 09:00:00.000','2026-10-11 10:00:00.000',NULL,'Laptop Backpack',990.00,1,3),
(13,'2026-09-01 09:00:00.000','2026-10-09 10:00:00.000',NULL,'External SSD 1TB',2990.00,12,4),
(14,'2026-09-01 09:00:00.000','2026-10-12 10:00:00.000',NULL,'WiFi 6 Router',1890.00,7,4),
(15,'2026-09-01 09:00:00.000','2026-09-01 09:00:00.000',NULL,'Webcam 1080p',1190.00,0,4);

INSERT INTO `orders` (`id`,`created_at`,`updated_at`,`deleted_at`,`customer_id`,`product_id`,`quantity`,`total_price`,`status`) VALUES
(1,'2026-09-20 10:00:00.000','2026-09-20 10:00:00.000',NULL,1,1,1,39900.00,'Completed'),
(2,'2026-09-23 10:00:00.000','2026-09-23 10:00:00.000',NULL,2,5,1,29900.00,'Completed'),
(3,'2026-09-25 10:00:00.000','2026-09-25 10:00:00.000',NULL,3,9,2,6980.00,'Completed'),
(4,'2026-09-27 10:00:00.000','2026-09-27 10:00:00.000',NULL,4,10,1,2490.00,'Completed'),
(5,'2026-09-30 10:00:00.000','2026-09-30 10:00:00.000',NULL,5,11,3,3870.00,'Completed'),
(6,'2026-10-03 10:00:00.000','2026-10-03 10:00:00.000',NULL,6,2,1,19900.00,'Completed'),
(7,'2026-10-07 10:00:00.000','2026-10-07 10:00:00.000',NULL,7,6,1,27900.00,'Pending'),
(8,'2026-10-09 10:00:00.000','2026-10-09 10:00:00.000',NULL,8,13,2,5980.00,'Pending'),
(9,'2026-10-11 10:00:00.000','2026-10-11 10:00:00.000',NULL,1,12,1,990.00,'Pending'),
(10,'2026-10-12 10:00:00.000','2026-10-12 10:00:00.000',NULL,2,14,1,1890.00,'Cancelled'),
(11,'2026-10-13 10:00:00.000','2026-10-13 10:00:00.000',NULL,3,7,2,15980.00,'Pending'),
(12,'2026-10-15 10:00:00.000','2026-10-15 10:00:00.000',NULL,4,4,1,15900.00,'Pending');

SET FOREIGN_KEY_CHECKS = 1;
