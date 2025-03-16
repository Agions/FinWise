package middleware

import (
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/dgrijalva/jwt-go"
)

// 定义JWT密钥
var JwtSecret = []byte("WalletWise_Secret_Key")

// Claims 自定义声明结构体
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

// 不需要验证的路径
var whitelist = map[string]bool{
	"/api/user/register": true,
	"/api/user/login":    true,
	"/api/user/forgot-password": true,
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour)

	claims := Claims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "walletwise",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(JwtSecret)

	return token, err
}

// ParseToken 解析JWT令牌
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}

// JwtFilter JWT中间件
func JwtFilter(ctx *context.Context) {
	// 检查是否在白名单中
	if whitelist[ctx.Request.URL.Path] {
		return
	}

	// 检查某些特定的URL前缀
	if strings.HasPrefix(ctx.Request.URL.Path, "/swagger") || strings.HasPrefix(ctx.Request.URL.Path, "/static") {
		return
	}

	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		ctx.Output.SetStatus(401)
		ctx.Output.JSON(map[string]interface{}{
			"code":    401,
			"message": "未授权，请登录",
		}, true, false)
		return
	}

	// Bearer Token格式验证
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		ctx.Output.SetStatus(401)
		ctx.Output.JSON(map[string]interface{}{
			"code":    401,
			"message": "认证格式有误",
		}, true, false)
		return
	}

	// 解析Token
	claims, err := ParseToken(parts[1])
	if err != nil || claims == nil {
		ctx.Output.SetStatus(401)
		ctx.Output.JSON(map[string]interface{}{
			"code":    401,
			"message": "无效的令牌",
		}, true, false)
		return
	}

	// 将用户ID存储在上下文中
	ctx.Input.SetData("user_id", claims.UserID)
} 