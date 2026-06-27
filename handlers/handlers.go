package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"effihub/config"
	"effihub/models"
)

// 获取所有分类
func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// 检查 token 有效性
		token := ExtractToken(r)
		isValidToken := token != "" && ValidateToken(token)

		var query string
		if isValidToken {
			// 有效 token：返回所有分类（包括禁用的）
			query = "SELECT id, name, icon, slug, sort, require_login, enabled FROM categories ORDER BY sort, created_at"
		} else {
			// 无 token 或无效 token：只返回已启用且不需要登录的分类
			query = "SELECT id, name, icon, slug, sort, require_login, enabled FROM categories WHERE require_login = FALSE AND enabled = TRUE ORDER BY sort, created_at"
		}

		rows, err := config.DB.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var categories []models.Category
		for rows.Next() {
			var c models.Category
			var enabled sql.NullBool
			err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.Slug, &c.Sort, &c.RequireLogin, &enabled)
			if err != nil {
				continue
			}
			c.Enabled = enabled.Valid && enabled.Bool
			categories = append(categories, c)
		}
		json.NewEncoder(w).Encode(categories)

	case "POST":
		var c models.Category
		json.NewDecoder(r.Body).Decode(&c)
		result, err := config.DB.Exec("INSERT INTO categories (name, icon, slug, sort, require_login, enabled) VALUES (?, ?, ?, ?, ?, ?)",
			c.Name, c.Icon, c.Slug, c.Sort, c.RequireLogin, c.Enabled)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		c.ID = int(id)
		json.NewEncoder(w).Encode(c)
	}
}

// 获取所有链接
func LinksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// 检查 token 有效性
		token := ExtractToken(r)
		isValidToken := token != "" && ValidateToken(token)

		categoryID := r.URL.Query().Get("category_id")
		var rows *sql.Rows
		var err error

		if isValidToken {
			// 有效 token：返回所有链接（包括禁用的）
			if categoryID != "" {
				rows, err = config.DB.Query(
					"SELECT id, category_id, name, url, icon, description, sort, enabled FROM links WHERE category_id = ? ORDER BY sort, created_at",
					categoryID)
			} else {
				rows, err = config.DB.Query(
					"SELECT id, category_id, name, url, icon, description, sort, enabled FROM links ORDER BY category_id, sort, created_at")
			}
		} else {
			// 无 token 或无效 token：只返回已启用的链接
			if categoryID != "" {
				rows, err = config.DB.Query(
					"SELECT id, category_id, name, url, icon, description, sort, enabled FROM links WHERE category_id = ? AND enabled = TRUE ORDER BY sort, created_at",
					categoryID)
			} else {
				rows, err = config.DB.Query(
					"SELECT id, category_id, name, url, icon, description, sort, enabled FROM links WHERE enabled = TRUE ORDER BY category_id, sort, created_at")
			}
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var links []models.Link
		for rows.Next() {
			var l models.Link
			var enabled sql.NullBool
			err := rows.Scan(&l.ID, &l.CategoryID, &l.Name, &l.URL, &l.Icon, &l.Desc, &l.Sort, &enabled)
			if err != nil {
				continue
			}
			l.Enabled = enabled.Valid && enabled.Bool
			links = append(links, l)
		}
		json.NewEncoder(w).Encode(links)

	case "POST":
		var l models.Link
		json.NewDecoder(r.Body).Decode(&l)
		result, err := config.DB.Exec(
			"INSERT INTO links (category_id, name, url, icon, description, sort, enabled) VALUES (?, ?, ?, ?, ?, ?, ?)",
			l.CategoryID, l.Name, l.URL, l.Icon, l.Desc, l.Sort, l.Enabled)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		l.ID = int(id)
		json.NewEncoder(w).Encode(l)
	}
}

// 单个链接操作
func LinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/links/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		var l models.Link
		var enabled sql.NullBool
		err = config.DB.QueryRow(
			"SELECT id, category_id, name, url, icon, description, sort, enabled FROM links WHERE id = ?", id).
			Scan(&l.ID, &l.CategoryID, &l.Name, &l.URL, &l.Icon, &l.Desc, &l.Sort, &enabled)
		if err != nil {
			http.Error(w, "链接不存在", http.StatusNotFound)
			return
		}
		l.Enabled = enabled.Valid && enabled.Bool
		json.NewEncoder(w).Encode(l)

	case "PUT":
		var l models.Link
		json.NewDecoder(r.Body).Decode(&l)
		_, err := config.DB.Exec(
			"UPDATE links SET category_id=?, name=?, url=?, icon=?, description=?, sort=?, enabled=? WHERE id=?",
			l.CategoryID, l.Name, l.URL, l.Icon, l.Desc, l.Sort, l.Enabled, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		l.ID = id
		json.NewEncoder(w).Encode(l)

	case "PATCH":
		var body struct {
			Enabled *bool `json:"enabled"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.Enabled == nil {
			http.Error(w, "enabled 字段必填", http.StatusBadRequest)
			return
		}
		_, err := config.DB.Exec("UPDATE links SET enabled=? WHERE id=?", *body.Enabled, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "enabled": *body.Enabled})

	case "DELETE":
		_, err := config.DB.Exec("DELETE FROM links WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "删除成功"})
	}
}

// 健康检查
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// 单个分类操作
func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		var c models.Category
		var enabled sql.NullBool
		err = config.DB.QueryRow(
			"SELECT id, name, icon, slug, sort, require_login, enabled FROM categories WHERE id = ?", id).
			Scan(&c.ID, &c.Name, &c.Icon, &c.Slug, &c.Sort, &c.RequireLogin, &enabled)
		if err != nil {
			http.Error(w, "分类不存在", http.StatusNotFound)
			return
		}
		c.Enabled = enabled.Valid && enabled.Bool
		json.NewEncoder(w).Encode(c)

	case "PUT":
		var c models.Category
		json.NewDecoder(r.Body).Decode(&c)
		_, err := config.DB.Exec(
			"UPDATE categories SET name=?, icon=?, slug=?, sort=?, require_login=?, enabled=? WHERE id=?",
			c.Name, c.Icon, c.Slug, c.Sort, c.RequireLogin, c.Enabled, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c.ID = id
		json.NewEncoder(w).Encode(c)

	case "PATCH":
		var body struct {
			Enabled *bool `json:"enabled"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.Enabled == nil {
			http.Error(w, "enabled 字段必填", http.StatusBadRequest)
			return
		}
		_, err := config.DB.Exec("UPDATE categories SET enabled=? WHERE id=?", *body.Enabled, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "enabled": *body.Enabled})

	case "DELETE":
		_, err := config.DB.Exec("DELETE FROM categories WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "删除成功"})
	}
}
