package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() {
	var err error
	
	// 从配置文件获取数据库配置
	dbUser, _ := web.AppConfig.String("dbuser")
	dbPassword, _ := web.AppConfig.String("dbpassword")
	dbHost, _ := web.AppConfig.String("dbhost")
	dbPort, _ := web.AppConfig.String("dbport")
	dbName, _ := web.AppConfig.String("dbname")
	
	// 如果配置为空，使用默认值
	if dbUser == "" {
		dbUser = "root"
	}
	if dbPassword == "" {
		dbPassword = "root"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbName == "" {
		dbName = "walletwise"
	}
	
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", 
		dbUser, dbPassword, dbHost, dbPort, dbName)
	
	// 连接数据库
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		logs.Error("Failed to connect to database: %v", err)
		panic(err)
	}
	
	// 设置连接池
	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(time.Hour)
	
	// 测试连接
	err = DB.Ping()
	if err != nil {
		logs.Error("Failed to ping database: %v", err)
		panic(err)
	}
	
	logs.Info("Database connected successfully")
	
	// 初始化表结构
	initTables()
}

// 创建必要的表结构
func initTables() {
	// 用户表
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(100) NOT NULL,
			phone VARCHAR(20),
			avatar VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_email (email),
			INDEX idx_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		logs.Error("Failed to create users table: %v", err)
		panic(err)
	}
	
	// 分类表
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			name VARCHAR(50) NOT NULL,
			type ENUM('income', 'expense') NOT NULL,
			icon VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE KEY unique_category (user_id, name, type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		logs.Error("Failed to create categories table: %v", err)
		panic(err)
	}
	
	// 账单表
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS bills (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			category_id INT NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			type ENUM('income', 'expense') NOT NULL,
			date DATE NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
			INDEX idx_user_date (user_id, date),
			INDEX idx_category (category_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		logs.Error("Failed to create bills table: %v", err)
		panic(err)
	}
	
	// 预算表
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS budgets (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			category_id INT,
			amount DECIMAL(10,2) NOT NULL,
			month DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
			UNIQUE KEY unique_budget (user_id, category_id, month)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		logs.Error("Failed to create budgets table: %v", err)
		panic(err)
	}
	
	// 预算告警表
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS budget_alerts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			budget_id INT NOT NULL,
			threshold INT NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		logs.Error("Failed to create budget_alerts table: %v", err)
		panic(err)
	}
	
	logs.Info("Database tables created successfully")
} 