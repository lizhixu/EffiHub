package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"effihub/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWT 密钥（延迟初始化）
var (
	jwtSecretOnce sync.Once
	jwtSecretVal   []byte
)

func getJWTSecret() []byte {
	jwtSecretOnce.Do(func() {
		jwtSecretVal = []byte(config.GetJWTSecret())
	})
	return jwtSecretVal
}

// 登录请求结构
type LoginRequest struct {
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

// 登录响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
}

// 生成 JWT token
func GenerateToken(rememberMe bool) (string, error) {
	// 设置过期时间：记住我 15 天，不记住 24 小时
	duration := 24 * time.Hour
	if rememberMe {
		duration = 15 * 24 * time.Hour
	}

	claims := jwt.MapClaims{
		"sub": "admin",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// 验证 JWT token
func ValidateToken(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil {
		return false
	}

	return token.Valid
}

// 从请求中提取 token
func ExtractToken(r *http.Request) string {
	// 从 Authorization header 提取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 从查询参数提取（用于某些场景）
	return r.URL.Query().Get("token")
}

// 登录验证
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Password == config.GetAdminPassword() {
		// 生成 token
		token, err := GenerateToken(req.RememberMe)
		if err != nil {
			// token 生成失败，返回失败以便前端提示
			log.Printf("Token 生成失败: %v", err)
			json.NewEncoder(w).Encode(LoginResponse{Success: false})
			return
		}

		json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
			Token:   token,
		})
	} else {
		json.NewEncoder(w).Encode(LoginResponse{Success: false})
	}
}

// 获取图片上传配置（token 已不再下发浏览器：前端上传统一走 /api/upload/icon 由后端转发）
func UploadConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"api": config.GetImageUploadAPI(),
	})
}
