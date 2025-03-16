package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

// Bill 账单模型
type Bill struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	CategoryID  uint      `json:"category_id"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"` // income or expense
	Date        time.Time `json:"date"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// 关联字段
	CategoryName string `json:"category_name,omitempty"`
	CategoryIcon string `json:"category_icon,omitempty"`
}

// BillRequest 账单请求参数
type BillRequest struct {
	CategoryID  uint    `json:"category_id" valid:"Required"`
	Amount      float64 `json:"amount" valid:"Required"`
	Type        string  `json:"type" valid:"Required;Match(income|expense)"`
	Date        string  `json:"date" valid:"Required"`
	Description string  `json:"description,omitempty"`
}

// BillQueryParams 账单查询参数
type BillQueryParams struct {
	StartDate  string
	EndDate    string
	Type       string
	CategoryID uint
	MinAmount  float64
	MaxAmount  float64
	Page       int
	PageSize   int
}

// CreateBill 创建账单
func CreateBill(userID uint, req *BillRequest) (*Bill, error) {
	// 检查分类是否存在且属于该用户
	var categoryExists bool
	var categoryType string
	err := DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM categories WHERE id = ? AND user_id = ?), type FROM categories WHERE id = ?",
		req.CategoryID, userID, req.CategoryID,
	).Scan(&categoryExists, &categoryType)
	
	if err != nil {
		logs.Error("Error checking category: %v", err)
		return nil, err
	}
	
	if !categoryExists {
		return nil, errors.New("分类不存在或不属于当前用户")
	}
	
	// 确保账单类型与分类类型一致
	if categoryType != req.Type {
		return nil, errors.New("账单类型与分类类型不一致")
	}
	
	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		logs.Error("Error parsing date: %v", err)
		return nil, errors.New("日期格式错误，正确格式为：YYYY-MM-DD")
	}
	
	// 创建账单
	result, err := DB.Exec(
		"INSERT INTO bills (user_id, category_id, amount, type, date, description) VALUES (?, ?, ?, ?, ?, ?)",
		userID, req.CategoryID, req.Amount, req.Type, date, req.Description,
	)
	
	if err != nil {
		logs.Error("Error creating bill: %v", err)
		return nil, err
	}
	
	// 获取账单ID
	billID, err := result.LastInsertId()
	if err != nil {
		logs.Error("Error getting bill ID: %v", err)
		return nil, err
	}
	
	// 获取完整的账单信息
	bill, err := GetBill(uint(billID), userID)
	if err != nil {
		logs.Error("Error fetching new bill: %v", err)
		return nil, err
	}
	
	return bill, nil
}

// GetBill 获取单个账单
func GetBill(id, userID uint) (*Bill, error) {
	bill := &Bill{}
	var dateStr string
	
	err := DB.QueryRow(`
		SELECT b.id, b.user_id, b.category_id, b.amount, b.type, 
		       DATE_FORMAT(b.date, '%Y-%m-%d'), b.description, 
		       b.created_at, b.updated_at, c.name, c.icon 
		FROM bills b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.id = ? AND b.user_id = ?
	`, id, userID).Scan(
		&bill.ID,
		&bill.UserID,
		&bill.CategoryID,
		&bill.Amount,
		&bill.Type,
		&dateStr,
		&bill.Description,
		&bill.CreatedAt,
		&bill.UpdatedAt,
		&bill.CategoryName,
		&bill.CategoryIcon,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("账单不存在")
		}
		logs.Error("Error querying bill: %v", err)
		return nil, err
	}
	
	// 解析日期
	bill.Date, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		logs.Error("Error parsing date from database: %v", err)
		return nil, err
	}
	
	return bill, nil
}

// GetBills 获取账单列表
func GetBills(userID uint, params *BillQueryParams) ([]*Bill, int, error) {
	// 构建查询条件
	query := `
		SELECT b.id, b.user_id, b.category_id, b.amount, b.type, 
		       DATE_FORMAT(b.date, '%Y-%m-%d'), b.description, 
		       b.created_at, b.updated_at, c.name, c.icon 
		FROM bills b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.user_id = ?
	`
	countQuery := "SELECT COUNT(*) FROM bills b WHERE b.user_id = ?"
	args := []interface{}{userID}
	countArgs := []interface{}{userID}
	
	// 添加筛选条件
	if params.StartDate != "" {
		query += " AND b.date >= ?"
		countQuery += " AND b.date >= ?"
		args = append(args, params.StartDate)
		countArgs = append(countArgs, params.StartDate)
	}
	
	if params.EndDate != "" {
		query += " AND b.date <= ?"
		countQuery += " AND b.date <= ?"
		args = append(args, params.EndDate)
		countArgs = append(countArgs, params.EndDate)
	}
	
	if params.Type != "" {
		query += " AND b.type = ?"
		countQuery += " AND b.type = ?"
		args = append(args, params.Type)
		countArgs = append(countArgs, params.Type)
	}
	
	if params.CategoryID > 0 {
		query += " AND b.category_id = ?"
		countQuery += " AND b.category_id = ?"
		args = append(args, params.CategoryID)
		countArgs = append(countArgs, params.CategoryID)
	}
	
	if params.MinAmount > 0 {
		query += " AND b.amount >= ?"
		countQuery += " AND b.amount >= ?"
		args = append(args, params.MinAmount)
		countArgs = append(countArgs, params.MinAmount)
	}
	
	if params.MaxAmount > 0 {
		query += " AND b.amount <= ?"
		countQuery += " AND b.amount <= ?"
		args = append(args, params.MaxAmount)
		countArgs = append(countArgs, params.MaxAmount)
	}
	
	// 添加排序
	query += " ORDER BY b.date DESC, b.id DESC"
	
	// 添加分页
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query += " LIMIT ? OFFSET ?"
		args = append(args, params.PageSize, offset)
	}
	
	// 获取总数
	var total int
	err := DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		logs.Error("Error counting bills: %v", err)
		return nil, 0, err
	}
	
	// 执行查询
	rows, err := DB.Query(query, args...)
	if err != nil {
		logs.Error("Error querying bills: %v", err)
		return nil, 0, err
	}
	defer rows.Close()
	
	// 处理结果
	bills := make([]*Bill, 0)
	for rows.Next() {
		bill := &Bill{}
		var dateStr string
		
		err := rows.Scan(
			&bill.ID,
			&bill.UserID,
			&bill.CategoryID,
			&bill.Amount,
			&bill.Type,
			&dateStr,
			&bill.Description,
			&bill.CreatedAt,
			&bill.UpdatedAt,
			&bill.CategoryName,
			&bill.CategoryIcon,
		)
		
		if err != nil {
			logs.Error("Error scanning bill row: %v", err)
			return nil, 0, err
		}
		
		// 解析日期
		bill.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			logs.Error("Error parsing date: %v", err)
			return nil, 0, err
		}
		
		bills = append(bills, bill)
	}
	
	if err = rows.Err(); err != nil {
		logs.Error("Error iterating bill rows: %v", err)
		return nil, 0, err
	}
	
	return bills, total, nil
}

// UpdateBill 更新账单
func UpdateBill(id, userID uint, req *BillRequest) (*Bill, error) {
	// 检查账单是否存在
	_, err := GetBill(id, userID)
	if err != nil {
		return nil, err
	}
	
	// 检查分类是否存在且属于该用户
	var categoryExists bool
	var categoryType string
	err = DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM categories WHERE id = ? AND user_id = ?), type FROM categories WHERE id = ?",
		req.CategoryID, userID, req.CategoryID,
	).Scan(&categoryExists, &categoryType)
	
	if err != nil {
		logs.Error("Error checking category: %v", err)
		return nil, err
	}
	
	if !categoryExists {
		return nil, errors.New("分类不存在或不属于当前用户")
	}
	
	// 确保账单类型与分类类型一致
	if categoryType != req.Type {
		return nil, errors.New("账单类型与分类类型不一致")
	}
	
	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		logs.Error("Error parsing date: %v", err)
		return nil, errors.New("日期格式错误，正确格式为：YYYY-MM-DD")
	}
	
	// 更新账单
	_, err = DB.Exec(
		"UPDATE bills SET category_id = ?, amount = ?, type = ?, date = ?, description = ? WHERE id = ? AND user_id = ?",
		req.CategoryID, req.Amount, req.Type, date, req.Description, id, userID,
	)
	
	if err != nil {
		logs.Error("Error updating bill: %v", err)
		return nil, err
	}
	
	// 获取更新后的账单
	bill, err := GetBill(id, userID)
	if err != nil {
		logs.Error("Error fetching updated bill: %v", err)
		return nil, err
	}
	
	return bill, nil
}

// DeleteBill 删除账单
func DeleteBill(id, userID uint) error {
	// 检查账单是否存在
	_, err := GetBill(id, userID)
	if err != nil {
		return err
	}
	
	// 删除账单
	_, err = DB.Exec("DELETE FROM bills WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		logs.Error("Error deleting bill: %v", err)
		return err
	}
	
	return nil
}

// GetMonthlyStats 获取月度统计
func GetMonthlyStats(userID uint, year int, month int) (map[string]interface{}, error) {
	// 构建日期条件
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	
	// 获取总收入和总支出
	var totalIncome, totalExpense float64
	err := DB.QueryRow(
		"SELECT COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) FROM bills WHERE user_id = ? AND date BETWEEN ? AND ?",
		userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"),
	).Scan(&totalIncome, &totalExpense)
	
	if err != nil {
		logs.Error("Error calculating monthly totals: %v", err)
		return nil, err
	}
	
	// 获取分类统计
	rows, err := DB.Query(`
		SELECT c.id, c.name, c.type, c.icon, SUM(b.amount) as total
		FROM bills b
		JOIN categories c ON b.category_id = c.id
		WHERE b.user_id = ? AND b.date BETWEEN ? AND ?
		GROUP BY b.category_id
		ORDER BY total DESC
	`, userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	if err != nil {
		logs.Error("Error querying category stats: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	// 处理分类统计
	categoryStats := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id uint
		var name, catType, icon string
		var total float64
		
		err := rows.Scan(&id, &name, &catType, &icon, &total)
		if err != nil {
			logs.Error("Error scanning category stats row: %v", err)
			return nil, err
		}
		
		categoryStats = append(categoryStats, map[string]interface{}{
			"id":    id,
			"name":  name,
			"type":  catType,
			"icon":  icon,
			"total": total,
		})
	}
	
	if err = rows.Err(); err != nil {
		logs.Error("Error iterating category stats rows: %v", err)
		return nil, err
	}
	
	// 获取每日统计
	rows, err = DB.Query(`
		SELECT DATE_FORMAT(date, '%Y-%m-%d') as day,
		       SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income,
		       SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) as expense
		FROM bills
		WHERE user_id = ? AND date BETWEEN ? AND ?
		GROUP BY day
		ORDER BY day
	`, userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	if err != nil {
		logs.Error("Error querying daily stats: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	// 处理每日统计
	dailyStats := make([]map[string]interface{}, 0)
	for rows.Next() {
		var day string
		var income, expense float64
		
		err := rows.Scan(&day, &income, &expense)
		if err != nil {
			logs.Error("Error scanning daily stats row: %v", err)
			return nil, err
		}
		
		dailyStats = append(dailyStats, map[string]interface{}{
			"date":    day,
			"income":  income,
			"expense": expense,
			"balance": income - expense,
		})
	}
	
	if err = rows.Err(); err != nil {
		logs.Error("Error iterating daily stats rows: %v", err)
		return nil, err
	}
	
	// 返回统计结果
	return map[string]interface{}{
		"year":          year,
		"month":         month,
		"total_income":  totalIncome,
		"total_expense": totalExpense,
		"balance":       totalIncome - totalExpense,
		"categories":    categoryStats,
		"daily":         dailyStats,
	}, nil
} 