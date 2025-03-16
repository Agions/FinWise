package controllers

import (
	"blog/models"
	"net/http"
	"strconv"
)

// BillController 账单控制器
type BillController struct {
	BaseController
}

// List 获取账单列表
// @Title 获取账单列表
// @Description 获取账单列表，支持多种筛选条件
// @Param start_date query string false "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string false "结束日期，格式：YYYY-MM-DD"
// @Param type query string false "账单类型：income/expense"
// @Param category_id query int false "分类ID"
// @Param min_amount query number false "最小金额"
// @Param max_amount query number false "最大金额"
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页条数，默认10"
// @Success 200 {object} map[string]interface{} 账单列表和分页信息
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/bills [get]
func (c *BillController) List() {
	userID := c.GetUserID()
	page, pageSize := c.GetPagination()
	
	// 构建查询参数
	params := &models.BillQueryParams{
		StartDate:  c.Ctx.Input.Query("start_date"),
		EndDate:    c.Ctx.Input.Query("end_date"),
		Type:       c.Ctx.Input.Query("type"),
		Page:       page,
		PageSize:   pageSize,
	}
	
	// 处理数字类型的查询参数
	if categoryIDStr := c.Ctx.Input.Query("category_id"); categoryIDStr != "" {
		categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
		if err == nil {
			params.CategoryID = uint(categoryID)
		}
	}
	
	if minAmountStr := c.Ctx.Input.Query("min_amount"); minAmountStr != "" {
		minAmount, err := strconv.ParseFloat(minAmountStr, 64)
		if err == nil {
			params.MinAmount = minAmount
		}
	}
	
	if maxAmountStr := c.Ctx.Input.Query("max_amount"); maxAmountStr != "" {
		maxAmount, err := strconv.ParseFloat(maxAmountStr, 64)
		if err == nil {
			params.MaxAmount = maxAmount
		}
	}
	
	// 查询账单
	bills, total, err := models.GetBills(userID, params)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	// 计算总页数
	totalPages := (total + pageSize - 1) / pageSize
	
	// 构建分页信息
	pagination := Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
	
	c.SuccessWithPagination(bills, pagination)
}

// Create 创建账单
// @Title 创建账单
// @Description 创建新的账单记录
// @Param body body models.BillRequest true "账单信息"
// @Success 200 {object} models.Bill 创建的账单
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/bills [post]
func (c *BillController) Create() {
	userID := c.GetUserID()
	
	var req models.BillRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	bill, err := models.CreateBill(userID, &req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(bill)
}

// Get 获取单个账单
// @Title 获取账单详情
// @Description 获取单个账单的详细信息
// @Param id path int true "账单ID"
// @Success 200 {object} models.Bill 账单信息
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 账单不存在
// @Failure 500 服务器内部错误
// @Router /api/bills/{id} [get]
func (c *BillController) Get() {
	userID := c.GetUserID()
	
	billID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "账单ID格式错误")
		return
	}
	
	bill, err := models.GetBill(billID, userID)
	if err != nil {
		c.Error(http.StatusNotFound, err.Error())
		return
	}
	
	c.Success(bill)
}

// Update 更新账单
// @Title 更新账单
// @Description 更新账单信息
// @Param id path int true "账单ID"
// @Param body body models.BillRequest true "账单信息"
// @Success 200 {object} models.Bill 更新后的账单
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 账单不存在
// @Failure 500 服务器内部错误
// @Router /api/bills/{id} [put]
func (c *BillController) Update() {
	userID := c.GetUserID()
	
	billID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "账单ID格式错误")
		return
	}
	
	var req models.BillRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	bill, err := models.UpdateBill(billID, userID, &req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(bill)
}

// Delete 删除账单
// @Title 删除账单
// @Description 删除账单
// @Param id path int true "账单ID"
// @Success 200 {object} Response 删除成功
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 账单不存在
// @Failure 500 服务器内部错误
// @Router /api/bills/{id} [delete]
func (c *BillController) Delete() {
	userID := c.GetUserID()
	
	billID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "账单ID格式错误")
		return
	}
	
	err = models.DeleteBill(billID, userID)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(nil)
}

// MonthlyStats 获取月度统计
// @Title 获取月度统计
// @Description 获取指定月份的账单统计数据
// @Param year query int true "年份"
// @Param month query int true "月份 (1-12)"
// @Success 200 {object} map[string]interface{} 月度统计数据
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/bills/stats/monthly [get]
func (c *BillController) MonthlyStats() {
	userID := c.GetUserID()
	
	// 获取年月参数
	yearStr := c.Ctx.Input.Query("year")
	monthStr := c.Ctx.Input.Query("month")
	
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1900 || year > 2100 {
		c.Error(http.StatusBadRequest, "年份格式错误或超出范围")
		return
	}
	
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.Error(http.StatusBadRequest, "月份格式错误或超出范围")
		return
	}
	
	// 获取统计数据
	stats, err := models.GetMonthlyStats(userID, year, month)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	c.Success(stats)
} 