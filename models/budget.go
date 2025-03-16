package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

// Budget 预算模型
type Budget struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	CategoryID uint      `json:"category_id,omitempty"`
	Amount     float64   `json:"amount"`
	Month      time.Time `json:"month"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// 关联字段
	CategoryName string  `json:"category_name,omitempty"`
	CategoryIcon string  `json:"category_icon,omitempty"`
	UsedAmount   float64 `json:"used_amount"`
	Percentage   float64 `json:"percentage"`
}

// BudgetRequest 预算请求参数
type BudgetRequest struct {
	CategoryID uint    `json:"category_id"`
	Amount     float64 `json:"amount" valid:"Required"`
	Month      string  `json:"month" valid:"Required"`
}

// BudgetAlert 预算告警模型
type BudgetAlert struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	BudgetID  uint      `json:"budget_id"`
	Threshold int       `json:"threshold"` // 阈值百分比
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BudgetAlertRequest 预算告警请求参数
type BudgetAlertRequest struct {
	BudgetID  uint `json:"budget_id" valid:"Required"`
	Threshold int  `json:"threshold" valid:"Required;Range(1,100)"`
	IsActive  bool `json:"is_active"`
}

// CreateBudget 创建预算
func CreateBudget(userID uint, req *BudgetRequest) (*Budget, error) {
	// 解析月份
	month, err := time.Parse("2006-01", req.Month)
	if err != nil {
		logs.Error("Error parsing month: %v", err)
		return nil, errors.New("月份格式错误，正确格式为：YYYY-MM")
	}
	
	// 检查分类是否存在且属于该用户（如果指定了分类）
	if req.CategoryID > 0 {
		var exists bool
		var categoryType string
		err := DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM categories WHERE id = ? AND user_id = ?), type FROM categories WHERE id = ?",
			req.CategoryID, userID, req.CategoryID,
		).Scan(&exists, &categoryType)
		
		if err != nil {
			logs.Error("Error checking category: %v", err)
			return nil, err
		}
		
		if !exists {
			return nil, errors.New("分类不存在或不属于当前用户")
		}
		
		// 只能为支出分类设置预算
		if categoryType != "expense" {
			return nil, errors.New("只能为支出分类设置预算")
		}
		
		// 检查是否已有同月同分类的预算
		var count int
		err = DB.QueryRow(
			"SELECT COUNT(*) FROM budgets WHERE user_id = ? AND category_id = ? AND DATE_FORMAT(month, '%Y-%m') = ?",
			userID, req.CategoryID, req.Month,
		).Scan(&count)
		
		if err != nil {
			logs.Error("Error checking existing budget: %v", err)
			return nil, err
		}
		
		if count > 0 {
			return nil, errors.New("该分类在当月已有预算设置")
		}
	} else {
		// 检查是否已有同月的总预算
		var count int
		err = DB.QueryRow(
			"SELECT COUNT(*) FROM budgets WHERE user_id = ? AND category_id IS NULL AND DATE_FORMAT(month, '%Y-%m') = ?",
			userID, req.Month,
		).Scan(&count)
		
		if err != nil {
			logs.Error("Error checking existing total budget: %v", err)
			return nil, err
		}
		
		if count > 0 {
			return nil, errors.New("当月已有总预算设置")
		}
	}
	
	// 创建预算
	var result sql.Result
	if req.CategoryID > 0 {
		result, err = DB.Exec(
			"INSERT INTO budgets (user_id, category_id, amount, month) VALUES (?, ?, ?, ?)",
			userID, req.CategoryID, req.Amount, month,
		)
	} else {
		result, err = DB.Exec(
			"INSERT INTO budgets (user_id, amount, month) VALUES (?, ?, ?)",
			userID, req.Amount, month,
		)
	}
	
	if err != nil {
		logs.Error("Error creating budget: %v", err)
		return nil, err
	}
	
	// 获取预算ID
	budgetID, err := result.LastInsertId()
	if err != nil {
		logs.Error("Error getting budget ID: %v", err)
		return nil, err
	}
	
	// 获取完整的预算信息
	budget, err := GetBudget(uint(budgetID), userID)
	if err != nil {
		logs.Error("Error fetching new budget: %v", err)
		return nil, err
	}
	
	return budget, nil
}

// GetBudget 获取单个预算
func GetBudget(id, userID uint) (*Budget, error) {
	budget := &Budget{}
	var monthStr string
	var categoryID sql.NullInt64
	var categoryName, categoryIcon sql.NullString
	
	// 查询预算基本信息
	err := DB.QueryRow(`
		SELECT b.id, b.user_id, b.category_id, b.amount, DATE_FORMAT(b.month, '%Y-%m'), 
		       b.created_at, b.updated_at, c.name, c.icon
		FROM budgets b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.id = ? AND b.user_id = ?
	`, id, userID).Scan(
		&budget.ID,
		&budget.UserID,
		&categoryID,
		&budget.Amount,
		&monthStr,
		&budget.CreatedAt,
		&budget.UpdatedAt,
		&categoryName,
		&categoryIcon,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("预算不存在")
		}
		logs.Error("Error querying budget: %v", err)
		return nil, err
	}
	
	// 处理可空字段
	if categoryID.Valid {
		budget.CategoryID = uint(categoryID.Int64)
	}
	if categoryName.Valid {
		budget.CategoryName = categoryName.String
	}
	if categoryIcon.Valid {
		budget.CategoryIcon = categoryIcon.String
	}
	
	// 解析月份
	budget.Month, err = time.Parse("2006-01", monthStr)
	if err != nil {
		logs.Error("Error parsing month from database: %v", err)
		return nil, err
	}
	
	// 计算已使用金额和百分比
	startDate := budget.Month
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	
	var query string
	var args []interface{}
	
	if budget.CategoryID > 0 {
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM bills
			WHERE user_id = ? AND category_id = ? AND type = 'expense' AND date BETWEEN ? AND ?
		`
		args = []interface{}{userID, budget.CategoryID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")}
	} else {
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM bills
			WHERE user_id = ? AND type = 'expense' AND date BETWEEN ? AND ?
		`
		args = []interface{}{userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")}
	}
	
	err = DB.QueryRow(query, args...).Scan(&budget.UsedAmount)
	if err != nil {
		logs.Error("Error calculating used amount: %v", err)
		return nil, err
	}
	
	// 计算百分比
	if budget.Amount > 0 {
		budget.Percentage = (budget.UsedAmount / budget.Amount) * 100
	}
	
	return budget, nil
}

// GetBudgets 获取预算列表
func GetBudgets(userID uint, month string) ([]*Budget, error) {
	// 验证月份格式
	parsedMonth, err := time.Parse("2006-01", month)
	if err != nil {
		logs.Error("Error parsing month: %v", err)
		return nil, errors.New("月份格式错误，正确格式为：YYYY-MM")
	}
	
	// 构建查询
	startDate := parsedMonth
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	
	// 查询当月所有预算
	rows, err := DB.Query(`
		SELECT b.id, b.user_id, b.category_id, b.amount, DATE_FORMAT(b.month, '%Y-%m'), 
		       b.created_at, b.updated_at, c.name, c.icon
		FROM budgets b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.user_id = ? AND DATE_FORMAT(b.month, '%Y-%m') = ?
		ORDER BY b.category_id IS NULL DESC, c.name
	`, userID, month)
	
	if err != nil {
		logs.Error("Error querying budgets: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	// 处理查询结果
	budgets := make([]*Budget, 0)
	for rows.Next() {
		budget := &Budget{}
		var monthStr string
		var categoryID sql.NullInt64
		var categoryName, categoryIcon sql.NullString
		
		err := rows.Scan(
			&budget.ID,
			&budget.UserID,
			&categoryID,
			&budget.Amount,
			&monthStr,
			&budget.CreatedAt,
			&budget.UpdatedAt,
			&categoryName,
			&categoryIcon,
		)
		
		if err != nil {
			logs.Error("Error scanning budget row: %v", err)
			return nil, err
		}
		
		// 处理可空字段
		if categoryID.Valid {
			budget.CategoryID = uint(categoryID.Int64)
		}
		if categoryName.Valid {
			budget.CategoryName = categoryName.String
		}
		if categoryIcon.Valid {
			budget.CategoryIcon = categoryIcon.String
		}
		
		// 解析月份
		budget.Month, err = time.Parse("2006-01", monthStr)
		if err != nil {
			logs.Error("Error parsing month from database: %v", err)
			return nil, err
		}
		
		// 查询已使用金额
		var query string
		var args []interface{}
		
		if budget.CategoryID > 0 {
			query = `
				SELECT COALESCE(SUM(amount), 0)
				FROM bills
				WHERE user_id = ? AND category_id = ? AND type = 'expense' AND date BETWEEN ? AND ?
			`
			args = []interface{}{userID, budget.CategoryID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")}
		} else {
			query = `
				SELECT COALESCE(SUM(amount), 0)
				FROM bills
				WHERE user_id = ? AND type = 'expense' AND date BETWEEN ? AND ?
			`
			args = []interface{}{userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")}
		}
		
		err = DB.QueryRow(query, args...).Scan(&budget.UsedAmount)
		if err != nil {
			logs.Error("Error calculating used amount: %v", err)
			return nil, err
		}
		
		// 计算百分比
		if budget.Amount > 0 {
			budget.Percentage = (budget.UsedAmount / budget.Amount) * 100
		}
		
		budgets = append(budgets, budget)
	}
	
	if err = rows.Err(); err != nil {
		logs.Error("Error iterating budget rows: %v", err)
		return nil, err
	}
	
	return budgets, nil
}

// UpdateBudget 更新预算
func UpdateBudget(id, userID uint, req *BudgetRequest) (*Budget, error) {
	// 检查预算是否存在
	budget, err := GetBudget(id, userID)
	if err != nil {
		return nil, err
	}
	
	// 如果要修改分类
	if budget.CategoryID != req.CategoryID {
		// 如果指定了新分类
		if req.CategoryID > 0 {
			// 检查分类是否存在且属于该用户
			var exists bool
			var categoryType string
			err := DB.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM categories WHERE id = ? AND user_id = ?), type FROM categories WHERE id = ?",
				req.CategoryID, userID, req.CategoryID,
			).Scan(&exists, &categoryType)
			
			if err != nil {
				logs.Error("Error checking category: %v", err)
				return nil, err
			}
			
			if !exists {
				return nil, errors.New("分类不存在或不属于当前用户")
			}
			
			// 只能为支出分类设置预算
			if categoryType != "expense" {
				return nil, errors.New("只能为支出分类设置预算")
			}
			
			// 检查是否已有同月同分类的预算
			var count int
			err = DB.QueryRow(
				"SELECT COUNT(*) FROM budgets WHERE user_id = ? AND category_id = ? AND DATE_FORMAT(month, '%Y-%m') = ? AND id != ?",
				userID, req.CategoryID, req.Month, id,
			).Scan(&count)
			
			if err != nil {
				logs.Error("Error checking existing budget: %v", err)
				return nil, err
			}
			
			if count > 0 {
				return nil, errors.New("该分类在当月已有预算设置")
			}
		} else {
			// 检查是否已有同月的总预算
			var count int
			err = DB.QueryRow(
				"SELECT COUNT(*) FROM budgets WHERE user_id = ? AND category_id IS NULL AND DATE_FORMAT(month, '%Y-%m') = ? AND id != ?",
				userID, req.Month, id,
			).Scan(&count)
			
			if err != nil {
				logs.Error("Error checking existing total budget: %v", err)
				return nil, err
			}
			
			if count > 0 {
				return nil, errors.New("当月已有总预算设置")
			}
		}
	}
	
	// 解析月份
	month, err := time.Parse("2006-01", req.Month)
	if err != nil {
		logs.Error("Error parsing month: %v", err)
		return nil, errors.New("月份格式错误，正确格式为：YYYY-MM")
	}
	
	// 更新预算
	if req.CategoryID > 0 {
		_, err = DB.Exec(
			"UPDATE budgets SET category_id = ?, amount = ?, month = ? WHERE id = ? AND user_id = ?",
			req.CategoryID, req.Amount, month, id, userID,
		)
	} else {
		_, err = DB.Exec(
			"UPDATE budgets SET category_id = NULL, amount = ?, month = ? WHERE id = ? AND user_id = ?",
			req.Amount, month, id, userID,
		)
	}
	
	if err != nil {
		logs.Error("Error updating budget: %v", err)
		return nil, err
	}
	
	// 获取更新后的预算
	updatedBudget, err := GetBudget(id, userID)
	if err != nil {
		logs.Error("Error fetching updated budget: %v", err)
		return nil, err
	}
	
	return updatedBudget, nil
}

// DeleteBudget 删除预算
func DeleteBudget(id, userID uint) error {
	// 检查预算是否存在
	_, err := GetBudget(id, userID)
	if err != nil {
		return err
	}
	
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		logs.Error("Error starting transaction: %v", err)
		return err
	}
	
	// 删除关联的预算告警
	_, err = tx.Exec("DELETE FROM budget_alerts WHERE budget_id = ?", id)
	if err != nil {
		tx.Rollback()
		logs.Error("Error deleting budget alerts: %v", err)
		return err
	}
	
	// 删除预算
	_, err = tx.Exec("DELETE FROM budgets WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		tx.Rollback()
		logs.Error("Error deleting budget: %v", err)
		return err
	}
	
	// 提交事务
	if err = tx.Commit(); err != nil {
		logs.Error("Error committing transaction: %v", err)
		return err
	}
	
	return nil
}

// CreateBudgetAlert 创建预算告警
func CreateBudgetAlert(userID uint, req *BudgetAlertRequest) (*BudgetAlert, error) {
	// 检查预算是否存在且属于当前用户
	_, err := GetBudget(req.BudgetID, userID)
	if err != nil {
		return nil, err
	}
	
	// 检查阈值范围
	if req.Threshold < 1 || req.Threshold > 100 {
		return nil, errors.New("阈值必须在1-100之间")
	}
	
	// 检查是否已存在告警
	var count int
	err = DB.QueryRow(
		"SELECT COUNT(*) FROM budget_alerts WHERE budget_id = ? AND threshold = ?",
		req.BudgetID, req.Threshold,
	).Scan(&count)
	
	if err != nil {
		logs.Error("Error checking existing alert: %v", err)
		return nil, err
	}
	
	if count > 0 {
		return nil, fmt.Errorf("已存在相同阈值(%d%%)的告警", req.Threshold)
	}
	
	// 创建告警
	result, err := DB.Exec(
		"INSERT INTO budget_alerts (user_id, budget_id, threshold, is_active) VALUES (?, ?, ?, ?)",
		userID, req.BudgetID, req.Threshold, req.IsActive,
	)
	
	if err != nil {
		logs.Error("Error creating budget alert: %v", err)
		return nil, err
	}
	
	// 获取告警ID
	alertID, err := result.LastInsertId()
	if err != nil {
		logs.Error("Error getting alert ID: %v", err)
		return nil, err
	}
	
	// 获取完整的告警信息
	alert := &BudgetAlert{
		ID:        uint(alertID),
		UserID:    userID,
		BudgetID:  req.BudgetID,
		Threshold: req.Threshold,
		IsActive:  req.IsActive,
	}
	
	return alert, nil
}

// GetBudgetAlerts 获取预算告警列表
func GetBudgetAlerts(userID uint, budgetID uint) ([]*BudgetAlert, error) {
	// 构建查询
	query := `
		SELECT id, user_id, budget_id, threshold, is_active, created_at, updated_at
		FROM budget_alerts
		WHERE user_id = ?
	`
	args := []interface{}{userID}
	
	if budgetID > 0 {
		query += " AND budget_id = ?"
		args = append(args, budgetID)
	}
	
	query += " ORDER BY threshold"
	
	// 执行查询
	rows, err := DB.Query(query, args...)
	if err != nil {
		logs.Error("Error querying budget alerts: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	// 处理结果
	alerts := make([]*BudgetAlert, 0)
	for rows.Next() {
		alert := &BudgetAlert{}
		
		err := rows.Scan(
			&alert.ID,
			&alert.UserID,
			&alert.BudgetID,
			&alert.Threshold,
			&alert.IsActive,
			&alert.CreatedAt,
			&alert.UpdatedAt,
		)
		
		if err != nil {
			logs.Error("Error scanning alert row: %v", err)
			return nil, err
		}
		
		alerts = append(alerts, alert)
	}
	
	if err = rows.Err(); err != nil {
		logs.Error("Error iterating alert rows: %v", err)
		return nil, err
	}
	
	return alerts, nil
}

// UpdateBudgetAlert 更新预算告警
func UpdateBudgetAlert(id, userID uint, req *BudgetAlertRequest) (*BudgetAlert, error) {
	// 检查告警是否存在
	var exists bool
	err := DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM budget_alerts WHERE id = ? AND user_id = ?)",
		id, userID,
	).Scan(&exists)
	
	if err != nil {
		logs.Error("Error checking alert existence: %v", err)
		return nil, err
	}
	
	if !exists {
		return nil, errors.New("预算告警不存在")
	}
	
	// 检查预算是否存在且属于当前用户
	_, err = GetBudget(req.BudgetID, userID)
	if err != nil {
		return nil, err
	}
	
	// 检查阈值范围
	if req.Threshold < 1 || req.Threshold > 100 {
		return nil, errors.New("阈值必须在1-100之间")
	}
	
	// 检查是否与其他告警冲突
	var count int
	err = DB.QueryRow(
		"SELECT COUNT(*) FROM budget_alerts WHERE budget_id = ? AND threshold = ? AND id != ?",
		req.BudgetID, req.Threshold, id,
	).Scan(&count)
	
	if err != nil {
		logs.Error("Error checking alert conflict: %v", err)
		return nil, err
	}
	
	if count > 0 {
		return nil, fmt.Errorf("已存在相同阈值(%d%%)的告警", req.Threshold)
	}
	
	// 更新告警
	_, err = DB.Exec(
		"UPDATE budget_alerts SET budget_id = ?, threshold = ?, is_active = ? WHERE id = ? AND user_id = ?",
		req.BudgetID, req.Threshold, req.IsActive, id, userID,
	)
	
	if err != nil {
		logs.Error("Error updating budget alert: %v", err)
		return nil, err
	}
	
	// 获取更新后的告警信息
	alert := &BudgetAlert{}
	err = DB.QueryRow(
		"SELECT id, user_id, budget_id, threshold, is_active, created_at, updated_at FROM budget_alerts WHERE id = ?",
		id,
	).Scan(
		&alert.ID,
		&alert.UserID,
		&alert.BudgetID,
		&alert.Threshold,
		&alert.IsActive,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)
	
	if err != nil {
		logs.Error("Error fetching updated alert: %v", err)
		return nil, err
	}
	
	return alert, nil
}

// DeleteBudgetAlert 删除预算告警
func DeleteBudgetAlert(id, userID uint) error {
	// 检查告警是否存在
	var exists bool
	err := DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM budget_alerts WHERE id = ? AND user_id = ?)",
		id, userID,
	).Scan(&exists)
	
	if err != nil {
		logs.Error("Error checking alert existence: %v", err)
		return err
	}
	
	if !exists {
		return errors.New("预算告警不存在")
	}
	
	// 删除告警
	_, err = DB.Exec("DELETE FROM budget_alerts WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		logs.Error("Error deleting budget alert: %v", err)
		return err
	}
	
	return nil
}

// CheckBudgetAlerts 检查超出预算告警
func CheckBudgetAlerts(userID uint) ([]map[string]interface{}, error) {
	// 获取当前月份
	now := time.Now()
	currentMonth := now.Format("2006-01")
	
	// 获取当月的所有预算及其使用情况
	budgets, err := GetBudgets(userID, currentMonth)
	if err != nil {
		logs.Error("Error getting budgets: %v", err)
		return nil, err
	}
	
	// 获取所有激活的预算告警
	alerts, err := DB.Query(`
		SELECT ba.id, ba.budget_id, ba.threshold, b.amount, b.category_id, c.name
		FROM budget_alerts ba
		JOIN budgets b ON ba.budget_id = b.id
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE ba.user_id = ? AND ba.is_active = 1 AND DATE_FORMAT(b.month, '%Y-%m') = ?
	`, userID, currentMonth)
	
	if err != nil {
		logs.Error("Error querying active alerts: %v", err)
		return nil, err
	}
	defer alerts.Close()
	
	// 存储触发的告警
	triggeredAlerts := make([]map[string]interface{}, 0)
	
	// 检查每个告警是否触发
	for alerts.Next() {
		var alertID, budgetID uint
		var threshold int
		var budgetAmount float64
		var categoryID sql.NullInt64
		var categoryName sql.NullString
		
		err := alerts.Scan(&alertID, &budgetID, &threshold, &budgetAmount, &categoryID, &categoryName)
		if err != nil {
			logs.Error("Error scanning alert: %v", err)
			return nil, err
		}
		
		// 查找对应的预算
		var matchBudget *Budget
		for _, b := range budgets {
			if b.ID == budgetID {
				matchBudget = b
				break
			}
		}
		
		if matchBudget == nil {
			continue
		}
		
		// 计算使用百分比
		usedPercentage := matchBudget.Percentage
		
		// 检查是否超过阈值
		if usedPercentage >= float64(threshold) {
			alertInfo := map[string]interface{}{
				"alert_id":       alertID,
				"budget_id":      budgetID,
				"threshold":      threshold,
				"used_percent":   usedPercentage,
				"used_amount":    matchBudget.UsedAmount,
				"budget_amount":  budgetAmount,
			}
			
			if categoryID.Valid {
				alertInfo["category_id"] = categoryID.Int64
				if categoryName.Valid {
					alertInfo["category_name"] = categoryName.String
				}
				alertInfo["budget_type"] = "category"
			} else {
				alertInfo["budget_type"] = "total"
			}
			
			triggeredAlerts = append(triggeredAlerts, alertInfo)
		}
	}
	
	if err = alerts.Err(); err != nil {
		logs.Error("Error iterating alerts: %v", err)
		return nil, err
	}
	
	return triggeredAlerts, nil
} 