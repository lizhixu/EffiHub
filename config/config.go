package config

import "os"

// 从环境变量读取配置，如果没有则使用默认值
func GetAdminPassword() string {
	if pwd := os.Getenv("ADMIN_PASSWORD"); pwd != "" {
		return pwd
	}
	return "1qaz@WSX#EDC" // 默认密码
}

func GetImageUploadToken() string {
	if token := os.Getenv("IMAGE_UPLOAD_TOKEN"); token != "" {
		return token
	}
	return "d5acdfb6421695ca250db908a0f3a100" // 默认token
}

func GetImageUploadAPI() string {
	if api := os.Getenv("IMAGE_UPLOAD_API"); api != "" {
		return api
	}
	return "https://img.lizhixu.cn/api/index.php"
}
