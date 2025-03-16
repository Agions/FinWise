package controllers

import (
	"blog/middleware"
	"blog/models"
	"net/http"
)

// UserController 用户控制器
type UserController struct {
	BaseController
}

// Register 用户注册
// @Title 用户注册
// @Description 创建新用户
// @Param body body models.RegisterRequest true "用户注册信息"
// @Success 200 {object} models.User 注册成功
// @Failure 400 参数错误
// @Failure 500 服务器内部错误
// @Router /api/user/register [post]
func (c *UserController) Register() {
	var req models.RegisterRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	user, err := models.CreateUser(&req)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.Error(http.StatusInternalServerError, "生成令牌失败")
		return
	}
	
	c.Success(map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// Login 用户登录
// @Title 用户登录
// @Description 用户登录并返回JWT令牌
// @Param body body models.LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} 登录成功
// @Failure 400 参数错误
// @Failure 401 认证失败
// @Failure 500 服务器内部错误
// @Router /api/user/login [post]
func (c *UserController) Login() {
	var req models.LoginRequest
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	user, err := models.AuthenticateUser(&req)
	if err != nil {
		c.Error(http.StatusUnauthorized, err.Error())
		return
	}
	
	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.Error(http.StatusInternalServerError, "生成令牌失败")
		return
	}
	
	c.Success(map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// Profile 获取当前用户信息
// @Title 获取用户信息
// @Description 获取当前登录用户信息
// @Success 200 {object} models.UserProfileResponse 用户信息
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/user/profile [get]
func (c *UserController) Profile() {
	userID := c.GetUserID()
	
	user, err := models.GetUserByID(userID)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	c.Success(user)
}

// UpdateProfile 更新用户信息
// @Title 更新用户信息
// @Description 更新当前登录用户信息
// @Param body body models.UserProfileResponse true "用户信息"
// @Success 200 {object} models.UserProfileResponse 更新后的用户信息
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/user/profile [put]
func (c *UserController) UpdateProfile() {
	userID := c.GetUserID()
	
	var profile models.UserProfileResponse
	if err := c.ParseAndValidate(&profile); err != nil {
		return
	}
	
	// 确保只能更新当前用户
	profile.ID = userID
	
	err := models.UpdateUser(userID, profile.Username, profile.Email, profile.Phone, profile.Avatar)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	// 获取更新后的用户信息
	user, err := models.GetUserByID(userID)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}
	
	c.Success(user)
}

// ChangePassword 修改密码
// @Title 修改密码
// @Description 修改当前登录用户密码
// @Param body body object true "密码信息"
// @Success 200 {object} Response 修改成功
// @Failure 400 参数错误
// @Failure 401 未授权
// @Failure 500 服务器内部错误
// @Router /api/user/password [put]
func (c *UserController) ChangePassword() {
	userID := c.GetUserID()
	
	var req struct {
		OldPassword string `json:"old_password" valid:"Required"`
		NewPassword string `json:"new_password" valid:"Required;MinSize(6)"`
	}
	
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	err := models.UpdatePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(nil)
}

// ForgotPassword 忘记密码
// @Title 忘记密码
// @Description 重置密码（需要进一步扩展为邮件验证等安全方式）
// @Param body body object true "邮箱和新密码"
// @Success 200 {object} Response 重置成功
// @Failure 400 参数错误
// @Failure 500 服务器内部错误
// @Router /api/user/forgot-password [post]
func (c *UserController) ForgotPassword() {
	var req struct {
		Email       string `json:"email" valid:"Required;Email"`
		NewPassword string `json:"new_password" valid:"Required;MinSize(6)"`
	}
	
	if err := c.ParseAndValidate(&req); err != nil {
		return
	}
	
	// 注意：实际应用中应该发送验证码到邮箱，用户验证后才能重置密码
	// 这里简化处理，直接通过邮箱重置密码
	err := models.ResetPassword(req.Email, req.NewPassword)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	
	c.Success(nil)
} 