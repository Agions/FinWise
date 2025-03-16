package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/server/web"
)

// BaseController 基础控制器，提供通用方法
type BaseController struct {
	web.Controller
}

// Response API统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// Success 成功响应
func (c *BaseController) Success(data interface{}) {
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = Response{
		Code:    200,
		Message: "成功",
		Data:    data,
	}
	c.ServeJSON()
}

// SuccessWithPagination 带分页的成功响应
func (c *BaseController) SuccessWithPagination(data interface{}, pagination Pagination) {
	result := map[string]interface{}{
		"items":      data,
		"pagination": pagination,
	}
	c.Success(result)
}

// Error 错误响应
func (c *BaseController) Error(code int, message string) {
	c.Ctx.Output.SetStatus(code)
	c.Data["json"] = Response{
		Code:    code,
		Message: message,
	}
	c.ServeJSON()
}

// ParseAndValidate 解析并验证JSON请求
func (c *BaseController) ParseAndValidate(v interface{}) error {
	err := json.Unmarshal(c.Ctx.Input.RequestBody, v)
	if err != nil {
		c.Error(http.StatusBadRequest, "请求参数格式错误")
		return err
	}
	return nil
}

// GetUserID 从上下文中获取当前用户ID
func (c *BaseController) GetUserID() uint {
	userID := c.Ctx.Input.GetData("user_id")
	if userID == nil {
		return 0
	}
	return userID.(uint)
}

// GetUintParam 获取并转换uint类型的URL参数
func (c *BaseController) GetUintParam(param string) (uint, error) {
	idStr := c.Ctx.Input.Param(":" + param)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// GetPagination 获取分页参数
func (c *BaseController) GetPagination() (page, pageSize int) {
	page, _ = strconv.Atoi(c.Ctx.Input.Query("page"))
	if page < 1 {
		page = 1
	}
	
	pageSize, _ = strconv.Atoi(c.Ctx.Input.Query("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	return page, pageSize
} 