package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// JWT配置
type JWTConfig struct {
	SecretKey  string
	ExpireTime time.Duration
}

// JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 默认JWT配置
func NewJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:  "your-secret-key-here", // 在生产环境中应该从环境变量读取
		ExpireTime: 24 * time.Hour,         // 24小时过期
	}
}

// 生成JWT Token
func GenerateJWT(userID, username string, config *JWTConfig) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.ExpireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// 验证JWT Token
func ValidateJWT(tokenString string, config *JWTConfig) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// JWT认证中间件
func JWTAuthMiddleware(config *JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			//缺少token
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证token",
				"time":    time.Now().Format("2006-01-02 15:04:05"),
			})
			c.Abort()
			return
		}

		// 检查token格式
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token格式错误",
				"time":    time.Now().Format("2006-01-02 15:04:05"),
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 验证token
		claims, err := ValidateJWT(tokenString, config)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效或已过期",
				"time":    time.Now().Format("2006-01-02 15:04:05"),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// 请求日志中间件
func RequestLoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %s %d %s %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}

// 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误: " + err,
				"time":    time.Now().Format("2006-01-02 15:04:05"),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误",
				"time":    time.Now().Format("2006-01-02 15:04:05"),
			})
		}
		c.Abort()
	})
}

// 限流中间件（简单实现）
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	// 这里应该使用Redis或其他存储来实现真正的限流
	// 当前实现仅作为示例
	return func(c *gin.Context) {
		// TODO: 实现实际的限流逻辑
		// 1. 检查客户端IP的请求频率
		// 2. 如果超过限制，返回429状态码
		// 3. 否则继续处理请求

		c.Next()
	}
}

// CORS中间件配置
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// 生成请求ID
func generateRequestID() string {
	return "req_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// 性能监控中间件
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)

		// 记录慢请求
		if latency > 1*time.Second {
			log.Printf("慢请求警告: %s %s 耗时: %v", c.Request.Method, c.Request.URL.Path, latency)
		}

		// 添加响应头
		c.Header("X-Response-Time", latency.String())
	}
}

// func CheckLegal() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 	}

// }
