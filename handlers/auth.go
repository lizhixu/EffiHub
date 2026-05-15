package handlers

import (
	"encoding/json"
	"net/http"

	"effihub/config"
)

// 登录验证
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Password == config.GetAdminPassword() {
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	} else {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	}
}

// 获取图片上传配置
func UploadConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"api":   config.GetImageUploadAPI(),
		"token": config.GetImageUploadToken(),
	})
}
