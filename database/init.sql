-- cy-354 CampusMarket 初始化脚本：建表 + 种子数据（容器首次启动自动执行）
USE lpcampusmarket_db;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  phone VARCHAR(20) NOT NULL UNIQUE,
  password_hash VARCHAR(100) NOT NULL,
  nickname VARCHAR(32) NOT NULL,
  avatar VARCHAR(255) DEFAULT '',
  role VARCHAR(16) NOT NULL DEFAULT 'student',
  campus VARCHAR(64) DEFAULT '',
  credit_score INT NOT NULL DEFAULT 100,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS products (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  seller_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(64) NOT NULL,
  description TEXT,
  price DECIMAL(10,2) NOT NULL,
  category VARCHAR(24) NOT NULL,
  `condition` VARCHAR(16) DEFAULT '',
  campus VARCHAR(64) DEFAULT '',
  trade_location VARCHAR(128) DEFAULT '',
  images TEXT,
  status VARCHAR(16) NOT NULL DEFAULT 'on_sale',
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_products_seller (seller_id),
  INDEX idx_products_category (category),
  INDEX idx_products_campus (campus),
  INDEX idx_products_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS conversations (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  product_id BIGINT UNSIGNED NOT NULL,
  buyer_id BIGINT UNSIGNED NOT NULL,
  seller_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_conversations_product (product_id),
  INDEX idx_conversations_buyer (buyer_id),
  INDEX idx_conversations_seller (seller_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS messages (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  conversation_id BIGINT UNSIGNED NOT NULL,
  sender_id BIGINT UNSIGNED NOT NULL,
  content TEXT NOT NULL,
  `read` TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_messages_conversation (conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS trade_orders (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  product_id BIGINT UNSIGNED NOT NULL,
  buyer_id BIGINT UNSIGNED NOT NULL,
  seller_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'pending',
  buyer_confirmed_at DATETIME(3) NULL,
  seller_confirmed_at DATETIME(3) NULL,
  completed_at DATETIME(3) NULL,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_trade_orders_product (product_id),
  INDEX idx_trade_orders_buyer (buyer_id),
  INDEX idx_trade_orders_seller (seller_id),
  INDEX idx_trade_orders_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS reviews (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  trade_id BIGINT UNSIGNED NOT NULL,
  reviewer_id BIGINT UNSIGNED NOT NULL,
  reviewee_id BIGINT UNSIGNED NOT NULL,
  rating VARCHAR(16) NOT NULL,
  content TEXT,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_reviews_trade (trade_id),
  INDEX idx_reviews_reviewee (reviewee_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS book_exchanges (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  offer_book VARCHAR(64) NOT NULL,
  want_book VARCHAR(64) NOT NULL,
  description TEXT,
  status VARCHAR(16) NOT NULL DEFAULT 'open',
  matched_id BIGINT UNSIGNED NULL,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_book_exchanges_user (user_id),
  INDEX idx_book_exchanges_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 种子账号（bcrypt）
-- 13700000001/123456 小明同学；13700000002/123456 阿珍；13700000003/123456 二手达人；13800000001/admin123 平台管理员
INSERT INTO users (phone, password_hash, nickname, avatar, role, campus, credit_score) VALUES
('13700000001', '$2a$10$0ExtqEObqjU6pxp4dYe0G.7F06XTDb9WLu81fz7RQYDf3SZDQg1bW', '小明同学', '', 'student', '东校区', 100),
('13700000002', '$2a$10$0ExtqEObqjU6pxp4dYe0G.7F06XTDb9WLu81fz7RQYDf3SZDQg1bW', '阿珍', '', 'student', '西校区', 120),
('13700000003', '$2a$10$0ExtqEObqjU6pxp4dYe0G.7F06XTDb9WLu81fz7RQYDf3SZDQg1bW', '二手达人', '', 'student', '南校区', 90),
('13800000001', '$2a$10$hqsEfU3UaVc2Mkza/Tg2IuCRpaf4GvD65TAkEBkHFhRyZUJ9MQQ5G', '平台管理员', '', 'admin', '东校区', 300);

INSERT INTO products (seller_id, title, description, price, category, `condition`, campus, trade_location, images, status) VALUES
(1, '高等数学第六版', '九成新，有少量笔记', 15.00, 'books', '九成新', '东校区', '图书馆门口', '', 'on_sale'),
(2, 'iPad Air 5', '95新，带笔', 2800.00, 'electronics', '95新', '西校区', '三食堂', '', 'on_sale'),
(3, '宿舍小台灯', '暖光护眼', 20.00, 'daily', '全新', '南校区', '南门快递点', '', 'on_sale'),
(1, '毕业季正装一套', 'M码 黑色西服', 180.00, 'clothing', '九成新', '东校区', '东门', '', 'on_sale');

INSERT INTO book_exchanges (user_id, offer_book, want_book, description, status) VALUES
(1, '数据结构', '计算机网络', '希望交换', 'open'),
(2, '计算机网络', '数据结构', '同城交换', 'open');
