package controllers

import (
	"blog/models"
	"net/http"
)

// CategoryController 分类控制器
type CategoryController struct {
	BaseController
}

// List 获取分类列表
// @Title 获取分类列表
// @Description 获取当前用户的所有分类
// @Param type query string false "分类类型: income/expense"
// @Success 200 {array} models.Category 分类列表
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/categories [get]
func (c *CategoryController) List() {
	userID := c.GetUserID()
	categoryType := c.Ctx.Input.Query("type")
	
	categories, err := models.GetCategories(userID, categoryType)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	c.Success(categories)
}

// Create 创建分类
// @Title 创建分类
// @Description 创建新的分类
// @Param body body models.CategoryRequest true "分类信息"
// @Success 200 {object} models.Category 创建的分类
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/categories [post]
func (c *CategoryController) Create() {
	userID := c.GetUserID()
	
	var req models.CategoryRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	category, err := models.CreateCategory(userID, &req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(category)
}

// Get 获取单个分类
// @Title 获取分类详情
// @Description 获取单个分类的详细信息
// @Param id path int true "分类ID"
// @Success 200 {object} models.Category 分类信息
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 分类不存在
// @Failure 500 服务器内部错误
// @Router /api/categories/{id} [get]
func (c *CategoryController) Get() {
	userID := c.GetUserID()
	
	categoryID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "分类ID格式错误")
		return
	}
	
	category, err := models.GetCategory(categoryID, userID)
	if err != nil {
		c.Error(http.StatusNotFound, err.Error())
		return
	}
	
	c.Success(category)
}

// Update 更新分类
// @Title 更新分类
// @Description 更新分类信息
// @Param id path int true "分类ID"
// @Param body body models.CategoryRequest true "分类信息"
// @Success 200 {object} models.Category 更新后的分类
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 分类不存在
// @Failure 500 服务器内部错误
// @Router /api/categories/{id} [put]
func (c *CategoryController) Update() {
	userID := c.GetUserID()
	
	categoryID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "分类ID格式错误")
		return
	}
	
	var req models.CategoryRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	category, err := models.UpdateCategory(categoryID, userID, &req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(category)
}

// Delete 删除分类
// @Title 删除分类
// @Description 删除分类
// @Param id path int true "分类ID"
// @Success 200 {object} Response 删除成功
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 404 分类不存在
// @Failure 500 服务器内部错误
// @Router /api/categories/{id} [delete]
func (c *CategoryController) Delete() {
	userID := c.GetUserID()
	
	categoryID, err := c.GetUintParam("id")
	if err != nil {
		c.Error(http.StatusBadRequest, "分类ID格式错误")
		return
	}
	
	err = models.DeleteCategory(categoryID, userID)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(nil)
} 