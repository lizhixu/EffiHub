package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"effihub/config"
)

// base64 超过此长度（字节）则自动上传到图床，避免数据库存储过大
const maxBase64Len = 32 * 1024 // 32KB

// 获取网站信息（图标 + 标题 + 描述）
func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "缺少url参数"})
		return
	}

	// 如果没有协议前缀，默认添加 https
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "无效的URL"})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}

	result := map[string]string{}

	// 获取网页HTML，解析 icon、title、description
	iconURL, title, desc := fetchSiteMeta(client, parsedURL)

	if title != "" {
		result["title"] = title
	}
	if desc != "" {
		result["description"] = desc
	}

	// 方法2: 回退到常见 favicon 路径
	if iconURL == "" {
		candidates := []string{
			"/favicon.ico",
			"/favicon.png",
			"/favicon.svg",
			"/apple-touch-icon.png",
			"/apple-touch-icon-precomposed.png",
		}
		for _, candidate := range candidates {
			testURL := parsedURL.Scheme + "://" + parsedURL.Host + candidate
			if probeIcon(client, testURL) {
				iconURL = testURL
				break
			}
		}
	}

	// 下载图标并转 base64；失败（过大且图床不可用）时降级为图标直链
	if iconURL != "" {
		dataURL := downloadAndEncode(client, iconURL)
		if dataURL != "" {
			result["icon"] = dataURL
		} else {
			result["icon"] = iconURL
		}
		result["icon_url"] = iconURL
	}

	if len(result) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "未获取到任何网站信息"})
		return
	}

	json.NewEncoder(w).Encode(result)
}

// 从 HTML 中解析网站元数据（图标、标题、描述）
func fetchSiteMeta(client *http.Client, pageURL *url.URL) (iconURL, title, desc string) {
	req, err := http.NewRequest("GET", pageURL.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; EffiHub/1.0; +https://effihub.dev)")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024)) // 最多读512KB
	if err != nil {
		return
	}

	htmlStr := string(body)

	// 提取标题: <title>...</title>
	titleRe := regexp.MustCompile(`(?is)<title[^>]*>\s*(.*?)\s*</title>`)
	if m := titleRe.FindStringSubmatch(htmlStr); len(m) > 1 {
		title = decodeHTMLEntities(strings.TrimSpace(m[1]))
	}

	// 提取描述: <meta name="description" content="...">
	descRe := regexp.MustCompile(`(?i)<meta[^>]*name=["']description["'][^>]*content=["']([^"']+)["']`)
	if m := descRe.FindStringSubmatch(htmlStr); len(m) > 1 {
		desc = decodeHTMLEntities(strings.TrimSpace(m[1]))
	}

	// 按优先级匹配: apple-touch-icon > icon > shortcut icon
	patterns := []string{
		`(?i)<link[^>]*rel=["'][^"']*apple-touch-icon[^"']*["'][^>]*href=["']([^"']+)["']`,
		`(?i)<link[^>]*href=["']([^"']+)["'][^>]*rel=["'][^"']*apple-touch-icon[^"']*["']`,
		`(?i)<link[^>]*rel=["'][^"']*(?:icon|shortcut\s*icon)[^"']*["'][^>]*href=["']([^"']+)["']`,
		`(?i)<link[^>]*href=["']([^"']+)["'][^>]*rel=["'][^"']*(?:icon|shortcut\s*icon)[^"']*["']`,
	}

	// 收集所有匹配，按 sizes 属性选最大的
	var candidates []string
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(htmlStr, -1)
		for _, match := range matches {
			if len(match) > 1 {
				href := match[1]
				resolved := resolveURL(href, pageURL)
				if resolved != "" && !strings.Contains(resolved, "data:") {
					candidates = append(candidates, resolved)
				}
			}
		}
	}

	if len(candidates) > 0 {
		sizePriority := []string{"192", "180", "128", "96", "64", "48", "32", "16"}
		for _, size := range sizePriority {
			for _, c := range candidates {
				if strings.Contains(c, size) || strings.Contains(c, size+"x"+size) {
					return c, title, desc
				}
			}
		}
		return candidates[0], title, desc
	}

	return "", title, desc
}

// 解码 HTML 实体
func decodeHTMLEntities(s string) string {
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&#39;", "'",
		"&apos;", "'",
		"&#x27;", "'",
		"&nbsp;", " ",
	)
	s = replacer.Replace(s)
	// 数字实体 &#12345;
	numRe := regexp.MustCompile(`&#(\d+);`)
	s = numRe.ReplaceAllStringFunc(s, func(m string) string {
		matches := numRe.FindStringSubmatch(m)
		if len(matches) > 1 {
			var code int
			fmt.Sscanf(matches[1], "%d", &code)
			return string(rune(code))
		}
		return m
	})
	return s
}

// 探测图标 URL 是否可访问
func probeIcon(client *http.Client, iconURL string) bool {
	req, err := http.NewRequest("HEAD", iconURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; EffiHub/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	return resp.StatusCode == http.StatusOK &&
		(strings.HasPrefix(contentType, "image/") || contentType == "" ||
			strings.Contains(contentType, "octet-stream"))
}

// 解析相对 URL
func resolveURL(href string, base *url.URL) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}

	// 跳过 data URI
	if strings.HasPrefix(href, "data:") {
		return ""
	}

	u, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(u)
	return resolved.String()
}

// 下载图标并转为 base64 data URL，超长时自动上传到图床返回 URL
func downloadAndEncode(client *http.Client, iconURL string) string {
	req, err := http.NewRequest("GET", iconURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; EffiHub/1.0)")
	req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;q=0.8,*/*;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 最多1MB
	if err != nil {
		return ""
	}

	// 至少100字节，太小不是有效图标
	if len(body) < 100 {
		return ""
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}

	// 确保是图片类型
	if !strings.HasPrefix(contentType, "image/") {
		return ""
	}

	base64Data := base64.StdEncoding.EncodeToString(body)
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64Data)

	// 未超长，直接返回 data URL
	if len(dataURL) <= maxBase64Len {
		return dataURL
	}

	// 超长则尝试上传到图床
	if uploaded := uploadToImageHost(body, contentType); uploaded != "" {
		return uploaded
	}

	// 图床失败不能返回超大 base64：超过 TEXT 列上限会导致入库失败（Error 1406）
	return ""
}

// UploadIconHandler 后台手动上传图标：接收 multipart 文件，转发到图床，
// 返回图床 URL。前端不再直连图床（图床无 CORS 头，且 token 不应暴露给浏览器）
func UploadIconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制 10MB，与后台使用场景匹配
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		jsonError(w, "文件过大或表单无效", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		jsonError(w, "缺少图片文件（字段名 image）", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		jsonError(w, "读取文件失败", http.StatusBadRequest)
		return
	}
	if len(data) == 0 {
		jsonError(w, "文件为空", http.StatusBadRequest)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(contentType, "image/") {
		jsonError(w, "仅支持图片文件", http.StatusBadRequest)
		return
	}

	url := uploadToImageHost(data, contentType)
	if url == "" {
		jsonError(w, "图床上传失败，请检查图床服务及 API Token 配置", http.StatusBadGateway)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"result": "success", "url": url})
}

// normalizeIcon 保证 icon 不超过数据库 TEXT 列上限（65535 字节）：
// 超长的 base64 data URL 自动转存图床换取 URL；转存失败或格式无效则返回错误，
// 避免入库时报 "Data too long for column 'icon'"（Error 1406）
func normalizeIcon(icon string) (string, error) {
	if len(icon) <= maxBase64Len {
		return icon, nil
	}
	if !strings.HasPrefix(icon, "data:") {
		return "", fmt.Errorf("图标数据过大（%dKB），请改为填写图片 URL 或重新上传", len(icon)/1024)
	}
	idx := strings.Index(icon, ";base64,")
	if idx < 0 {
		return "", fmt.Errorf("图标 base64 数据格式无效")
	}
	contentType := icon[len("data:"):idx]
	data, err := base64.StdEncoding.DecodeString(icon[idx+len(";base64,"):])
	if err != nil {
		return "", fmt.Errorf("图标 base64 数据无效")
	}
	uploaded := uploadToImageHost(data, contentType)
	if uploaded == "" {
		return "", fmt.Errorf("图标过大且转存图床失败，请直接填写图片 URL")
	}
	return uploaded, nil
}

// 上传图片到图床，成功返回图床 URL，失败返回空字符串
func uploadToImageHost(imageData []byte, contentType string) string {
	api := config.GetImageUploadAPI()
	token := config.GetImageUploadToken()
	if api == "" || token == "" {
		return ""
	}

	// 构造 multipart 表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// token 字段
	if err := writer.WriteField("token", token); err != nil {
		return ""
	}

	// image 字段（图床外部 Token 上传接口格式：POST /api/upload/token，字段 image + token）
	part, err := writer.CreateFormFile("image", "icon"+extFromContentType(contentType))
	if err != nil {
		return ""
	}
	if _, err := part.Write(imageData); err != nil {
		return ""
	}
	writer.Close()

	req, err := http.NewRequest("POST", api, &buf)
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return ""
	}

	var result struct {
		Result string `json:"result"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return ""
	}
	if result.Result == "success" && result.URL != "" {
		return result.URL
	}
	return ""
}

// 根据 ContentType 获取文件扩展名
func extFromContentType(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/x-icon":
		return ".ico"
	default:
		return ".bin"
	}
}
