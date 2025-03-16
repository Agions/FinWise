package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"golang.org/x/crypto/bcrypt"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // 不在JSON中显示密码
	Phone     string    `json:"phone,omitempty"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" valid:"Required;MinSize(3);MaxSize(50)"`
	Email    string `json:"email" valid:"Required;Email"`
	Password string `json:"password" valid:"Required;MinSize(6)"`
	Phone    string `json:"phone,omitempty"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username"` // 可以是用户名或邮箱
	Password string `json:"password" valid:"Required"`
}

// UserProfileResponse 用户资料响应
type UserProfileResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUser 创建新用户
func CreateUser(req *RegisterRequest) (*User, error) {
	// 检查用户名是否已存在
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", req.Username).Scan(&exists)
	if err != nil {
		logs.Error("Error checking username existence: %v", err)
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", req.Email).Scan(&exists)
	if err != nil {
		logs.Error("Error checking email existence: %v", err)
		return nil, err
	}
	if exists {
		return nil, errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logs.Error("Error hashing password: %v", err)
		return nil, err
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		logs.Error("Error starting transaction: %v", err)
		return nil, err
	}

	// 创建用户
	result, err := tx.Exec(
		"INSERT INTO users (username, email, password, phone) VALUES (?, ?, ?, ?)",
		req.Username, req.Email, hashedPassword, req.Phone,
	)
	if err != nil {
		tx.Rollback()
		logs.Error("Error inserting user: %v", err)
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		logs.Error("Error getting last insert ID: %v", err)
		return nil, err
	}

	// 创建默认分类
	defaultCategories := []struct {
		Name string
		Type string
		Icon string
	}{
		{"餐饮", "expense", "food"},
		{"购物", "expense", "shopping"},
		{"交通", "expense", "transport"},
		{"住房", "expense", "home"},
		{"工资", "income", "salary"},
		{"奖金", "income", "bonus"},
		{"投资", "income", "investment"},
	}

	for _, category := range defaultCategories {
		_, err = tx.Exec(
			"INSERT INTO categories (user_id, name, type, icon) VALUES (?, ?, ?, ?)",
			userID, category.Name, category.Type, category.Icon,
		)
		if err != nil {
			tx.Rollback()
			logs.Error("Error creating default categories: %v", err)
			return nil, err
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		logs.Error("Error committing transaction: %v", err)
		return nil, err
	}

	// 返回用户对象
	user := &User{
		ID:       uint(userID),
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
	}

	return user, nil
}

// GetUserByID 通过ID获取用户
func GetUserByID(id uint) (*User, error) {
	user := &User{}
	err := DB.QueryRow(
		"SELECT id, username, email, phone, avatar, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &user.Avatar, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		logs.Error("Error querying user by ID: %v", err)
		return nil, err
	}

	return user, nil
}

// AuthenticateUser 验证用户凭据
func AuthenticateUser(login *LoginRequest) (*User, error) {
	user := &User{}
	var hashedPassword string

	// 支持用户名或邮箱登录
	err := DB.QueryRow(
		"SELECT id, username, email, password, phone, avatar, created_at, updated_at FROM users WHERE username = ? OR email = ?",
		login.Username, login.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword, &user.Phone, &user.Avatar, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		logs.Error("Error querying user for authentication: %v", err)
		return nil, err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(login.Password))
	if err != nil {
		return nil, errors.New("密码错误")
	}

	return user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(id uint, username, email, phone, avatar string) error {
	_, err := DB.Exec(
		"UPDATE users SET username = ?, email = ?, phone = ?, avatar = ? WHERE id = ?",
		username, email, phone, avatar, id,
	)
	
	if err != nil {
		logs.Error("Error updating user: %v", err)
		return err
	}
	
	return nil
}

// UpdatePassword 更新用户密码
func UpdatePassword(id uint, oldPassword, newPassword string) error {
	var hashedPassword string
	
	// 获取当前密码
	err := DB.QueryRow("SELECT password FROM users WHERE id = ?", id).Scan(&hashedPassword)
	if err != nil {
		logs.Error("Error getting current password: %v", err)
		return err
	}
	
	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(oldPassword))
	if err != nil {
		return errors.New("原密码错误")
	}
	
	// 加密新密码
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logs.Error("Error hashing new password: %v", err)
		return err
	}
	
	// 更新密码
	_, err = DB.Exec("UPDATE users SET password = ? WHERE id = ?", newHashedPassword, id)
	if err != nil {
		logs.Error("Error updating password: %v", err)
		return err
	}
	
	return nil
}

// ResetPassword 重置密码（忘记密码功能）
func ResetPassword(email, newPassword string) error {
	// 检查邮箱是否存在
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		logs.Error("Error checking email existence: %v", err)
		return err
	}
	if !exists {
		return errors.New("邮箱不存在")
	}
	
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logs.Error("Error hashing password: %v", err)
		return err
	}
	
	// 更新密码
	_, err = DB.Exec("UPDATE users SET password = ? WHERE email = ?", hashedPassword, email)
	if err != nil {
		logs.Error("Error resetting password: %v", err)
		return err
	}
	
	return nil
} 