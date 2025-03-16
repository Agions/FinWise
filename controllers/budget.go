package controllers

import (
	"blog/models"
	"net/http"
	"time"
)

// BudgetController 预算控制器
type BudgetController struct {
	BaseController
}

// List 获取预算列表
// @Title 获取预算列表
// @Description 获取指定月份的预算列表
// @Param month query string true "月份，格式：YYYY-MM"
// @Success 200 {array} models.Budget 预算列表
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/budgets [get]
func (c *BudgetController) List() {
	userID := c.GetUserID()
	
	month := c.Ctx.Input.Query("month")
	if month == "" {
		// 默认使用当前月份
		month = time.Now().Format("2006-01")
	}
	
	// 验证月份格式
	_, err := time.Parse("2006-01", month)
	if err != nil {
		c.Error(http.StatusBadRequest, "月份格式错误，正确格式为：YYYY-MM")
		return
	}
	
	budgets, err := models.GetBudgets(userID, month)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	c.Success(budgets)
}

// Create 创建预算
// @Title 创建预算
// @Description 创建新的预算
// @Param body body models.BudgetRequest true "预算信息"
// @Success 200 {object} models.Budget 创建的预算
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/budgets [post]
func (c *BudgetController) Create() {
	userID := c.GetUserID()
	
	var req models.BudgetRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	budget, err := models.CreateBudget(userID, &req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(budget)
}

// Get 获取单个预算
// @Title 获取预算详情
// @Description 获取单个预算的详细信息
// @Param id path int true "预算ID"
// @Success 200 {object} models.Budget 预算信息
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 预算不存在
// @Failure 500 服务器内部错误
// @Router /api/budgets/{id} [get]
func (c *BudgetController) Get() {
	userID := c.GetUserID()
	
	budgetID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "预算ID格式错误")
		return
	}
	
	budget, err := models.GetBudget(budgetID, userID)
	if err != nil {
		c.Error(http.StatusNotFound, err.Error())
		return
	}
	
	c.Success(budget)
}

// Update 更新预算
// @Title 更新预算
// @Description 更新预算信息
// @Param id path int true "预算ID"
// @Param body body models.BudgetRequest true "预算信息"
// @Success 200 {object} models.Budget 更新后的预算
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 预算不存在
// @Failure 500 服务器内部错误
// @Router /api/budgets/{id} [put]
func (c *BudgetController) Update() {
	userID := c.GetUserID()
	
	budgetID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "预算ID格式错误")
		return
	}
	
	var req models.BudgetRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	budget, err := models.UpdateBudget(budgetID, userID, &req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(budget)
}

// Delete 删除预算
// @Title 删除预算
// @Description 删除预算
// @Param id path int true "预算ID"
// @Success 200 {object} Response 删除成功
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 预算不存在
// @Failure 500 服务器内部错误
// @Router /api/budgets/{id} [delete]
func (c *BudgetController) Delete() {
	userID := c.GetUserID()
	
	budgetID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "预算ID格式错误")
		return
	}
	
	err = models.DeleteBudget(budgetID, userID)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(nil)
}

// CreateAlert 创建预算告警
// @Title 创建预算告警
// @Description 创建预算告警
// @Param body body models.BudgetAlertRequest true "预算告警信息"
// @Success 200 {object} models.BudgetAlert 创建的预算告警
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/budget-alerts [post]
func (c *BudgetController) CreateAlert() {
	userID := c.GetUserID()
	
	var req models.BudgetAlertRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	alert, err := models.CreateBudgetAlert(userID, &req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(alert)
}

// ListAlerts 获取预算告警列表
// @Title 获取预算告警列表
// @Description 获取预算告警列表
// @Param budget_id query int false "预算ID"
// @Success 200 {array} models.BudgetAlert 预算告警列表
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/budget-alerts [get]
func (c *BudgetController) ListAlerts() {
	userID := c.GetUserID()
	
	budgetIDStr := c.Ctx.Input.Query("budget_id")
	var budgetID uint = 0
	
	if budgetIDStr != "" {
		id, err := c.GetUintParam("budget_id")
		if err == nil {
			budgetID = id
		}
	}
	
	alerts, err := models.GetBudgetAlerts(userID, budgetID)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	c.Success(alerts)
}

// UpdateAlert 更新预算告警
// @Title 更新预算告警
// @Description 更新预算告警
// @Param id path int true "预算告警ID"
// @Param body body models.BudgetAlertRequest true "预算告警信息"
// @Success 200 {object} models.BudgetAlert 更新后的预算告警
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 预算告警不存在
// @Failure 500 服务器内部错误
// @Router /api/budget-alerts/{id} [put]
func (c *BudgetController) UpdateAlert() {
	userID := c.GetUserID()
	
	alertID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "预算告警ID格式错误")
		return
	}
	
	var req models.BudgetAlertRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	alert, err := models.UpdateBudgetAlert(alertID, userID, &req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(alert)
}

// DeleteAlert 删除预算告警
// @Title 删除预算告警
// @Description 删除预算告警
// @Param id path int true "预算告警ID"
// @Success 200 {object} Response 删除成功
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 预算告警不存在
// @Failure 500 服务器内部错误
// @Router /api/budget-alerts/{id} [delete]
func (c *BudgetController) DeleteAlert() {
	userID := c.GetUserID()
	
	alertID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "预算告警ID格式错误")
		return
	}
	
	err = models.DeleteBudgetAlert(alertID, userID)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(nil)
}

// CheckAlerts 检查预算告警
// @Title 检查预算告警
// @Description 检查当前触发的预算告警
// @Success 200 {array} object 触发的预算告警列表
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/budget-alerts/check [get]
func (c *BudgetController) CheckAlerts() {
	userID := c.GetUserID()
	
	alerts, err := models.CheckBudgetAlerts(userID)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	c.Success(alerts)
} 