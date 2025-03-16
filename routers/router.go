package routers

import (
	"blog/controllers"
	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// 用户相关路由
	beego.Router("/api/user/register", &controllers.UserController{}, "post:Register")
	beego.Router("/api/user/login", &controllers.UserController{}, "post:Login")
	beego.Router("/api/user/profile", &controllers.UserController{}, "get:Profile;put:UpdateProfile")
	beego.Router("/api/user/password", &controllers.UserController{}, "put:ChangePassword")
	beego.Router("/api/user/forgot-password", &controllers.UserController{}, "post:ForgotPassword")

	// 分类相关路由
	beego.Router("/api/categories", &controllers.CategoryController{}, "get:List;post:Create")
	beego.Router("/api/categories/:id", &controllers.CategoryController{}, "get:Get;put:Update;delete:Delete")

	// 账单相关路由
	beego.Router("/api/bills", &controllers.BillController{}, "get:List;post:Create")
	beego.Router("/api/bills/:id", &controllers.BillController{}, "get:Get;put:Update;delete:Delete")
	beego.Router("/api/bills/stats/monthly", &controllers.BillController{}, "get:MonthlyStats")

	// 预算相关路由
	beego.Router("/api/budgets", &controllers.BudgetController{}, "get:List;post:Create")
	beego.Router("/api/budgets/:id", &controllers.BudgetController{}, "get:Get;put:Update;delete:Delete")

	// 预算告警相关路由
	beego.Router("/api/budget-alerts", &controllers.BudgetController{}, "get:ListAlerts;post:CreateAlert")
	beego.Router("/api/budget-alerts/:id", &controllers.BudgetController{}, "put:UpdateAlert;delete:DeleteAlert")
	beego.Router("/api/budget-alerts/check", &controllers.BudgetController{}, "get:CheckAlerts")
}
