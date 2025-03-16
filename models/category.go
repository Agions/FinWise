package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

// Category 分类模型
type Category struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // income or expense
	Icon      string    `json:"icon,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CategoryRequest 分类请求参数
type CategoryRequest struct {
	Name string `json:"name" valid:"Required;MinSize(1);MaxSize(50)"`
	Type string `json:"type" valid:"Required;Match(income|expense)"`
	Icon string `json:"icon,omitempty"`
}

// GetCategories 获取用户的所有分类
func GetCategories(userID uint, categoryType string) ([]*Category, error) {
	var rows *sql.Rows
	var err error
	
	if categoryType != "" {
		rows, err = DB.Query(
			"SELECT id, user_id, name, type, icon, created_at, updated_at FROM categories WHERE user_id = ? AND type = ? ORDER BY name",
			userID, categoryType,
		)
	} else {
		rows, err = DB.Query(
			"SELECT id, user_id, name, type, icon, created_at, updated_at FROM categories WHERE user_id = ? ORDER BY type, name",
			userID,
		)
	}
	
	if err != nil {
		logs.Error("Error querying categories: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	categories := make([]*Category, 0)
	for rows.Next() {
		category := &Category{}
		err := rows.Scan(
			&category.ID,
			&category.UserID,
			&category.Name,
			&category.Type,
			&category.Icon,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			logs.Error("Error scanning category row: %v", err)
			return nil, err
		}
		categories = append(categories, category)
	}
	
	if err = rows.Err(); err != nil {
		logs.Error("Error iterating category rows: %v", err)
		return nil, err
	}
	
	return categories, nil
}

// GetCategory 获取单个分类
func GetCategory(id, userID uint) (*Category, error) {
	category := &Category{}
	err := DB.QueryRow(
		"SELECT id, user_id, name, type, icon, created_at, updated_at FROM categories WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.Type,
		&category.Icon,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("分类不存在")
		}
		logs.Error("Error querying category: %v", err)
		return nil, err
	}
	
	return category, nil
}

// CreateCategory 创建新分类
func CreateCategory(userID uint, req *CategoryRequest) (*Category, error) {
	// 检查分类名是否已存在
	var exists bool
	err := DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM categories WHERE user_id = ? AND name = ? AND type = ?)",
		userID, req.Name, req.Type,
	).Scan(&exists)
	
	if err != nil {
		logs.Error("Error checking category existence: %v", err)
		return nil, err
	}
	
	if exists {
		return nil, errors.New("分类名已存在")
	}
	
	// 创建分类
	result, err := DB.Exec(
		"INSERT INTO categories (user_id, name, type, icon) VALUES (?, ?, ?, ?)",
		userID, req.Name, req.Type, req.Icon,
	)
	
	if err != nil {
		logs.Error("Error creating category: %v", err)
		return nil, err
	}
	
	// 获取分类ID
	categoryID, err := result.LastInsertId()
	if err != nil {
		logs.Error("Error getting category ID: %v", err)
		return nil, err
	}
	
	// 查询完整的分类信息
	category, err := GetCategory(uint(categoryID), userID)
	if err != nil {
		logs.Error("Error fetching new category: %v", err)
		return nil, err
	}
	
	return category, nil
}

// UpdateCategory 更新分类
func UpdateCategory(id, userID uint, req *CategoryRequest) (*Category, error) {
	// 检查分类是否存在
	_, err := GetCategory(id, userID)
	if err != nil {
		return nil, err
	}
	
	// 检查修改后的名称是否与其他分类冲突
	var exists bool
	err = DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM categories WHERE user_id = ? AND name = ? AND type = ? AND id != ?)",
		userID, req.Name, req.Type, id,
	).Scan(&exists)
	
	if err != nil {
		logs.Error("Error checking category name conflict: %v", err)
		return nil, err
	}
	
	if exists {
		return nil, errors.New("已存在同名同类型的分类")
	}
	
	// 更新分类
	_, err = DB.Exec(
		"UPDATE categories SET name = ?, type = ?, icon = ? WHERE id = ? AND user_id = ?",
		req.Name, req.Type, req.Icon, id, userID,
	)
	
	if err != nil {
		logs.Error("Error updating category: %v", err)
		return nil, err
	}
	
	// 返回更新后的分类
	category, err := GetCategory(id, userID)
	if err != nil {
		logs.Error("Error fetching updated category: %v", err)
		return nil, err
	}
	
	return category, nil
}

// DeleteCategory 删除分类
func DeleteCategory(id, userID uint) error {
	// 检查分类是否存在
	_, err := GetCategory(id, userID)
	if err != nil {
		return err
	}
	
	// 检查分类是否被账单使用
	var billsCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM bills WHERE category_id = ?", id).Scan(&billsCount)
	if err != nil {
		logs.Error("Error checking if category is used in bills: %v", err)
		return err
	}
	
	if billsCount > 0 {
		return errors.New("该分类已被使用，无法删除")
	}
	
	// 检查分类是否被预算使用
	var budgetsCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM budgets WHERE category_id = ?", id).Scan(&budgetsCount)
	if err != nil {
		logs.Error("Error checking if category is used in budgets: %v", err)
		return err
	}
	
	if budgetsCount > 0 {
		return errors.New("该分类已设置预算，无法删除")
	}
	
	// 删除分类
	_, err = DB.Exec("DELETE FROM categories WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		logs.Error("Error deleting category: %v", err)
		return err
	}
	
	return nil
} 