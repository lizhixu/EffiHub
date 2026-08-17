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
	case http.MethodGet:
		// 检查 token 有效性
		token := ExtractToken(r)
		isValidToken := token != "" && ValidateToken(token)

		// 分页/筛选参数（后台管理使用；前台导航不带这些参数，走全量返回）
		pageStr := r.URL.Query().Get("page")
		usePaging := pageStr != ""

		// 构造 WHERE 条件
		var where []string
		var args []interface{}
		if !isValidToken {
			where = append(where, "require_login = FALSE", "enabled = TRUE")
		}
		if search := r.URL.Query().Get("search"); search != "" {
			where = append(where, "(name LIKE ? OR slug LIKE ?)")
			args = append(args, "%"+search+"%", "%"+search+"%")
		}
		if enabled := r.URL.Query().Get("enabled"); enabled != "" && isValidToken {
			if enabled == "1" {
				where = append(where, "enabled = TRUE")
			} else if enabled == "0" {
				where = append(where, "enabled = FALSE")
			}
		}
		if reqLogin := r.URL.Query().Get("require_login"); reqLogin != "" && isValidToken {
			if reqLogin == "1" {
				where = append(where, "require_login = TRUE")
			} else if reqLogin == "0" {
				where = append(where, "require_login = FALSE")
			}
		}

		whereClause := ""
		if len(where) > 0 {
			whereClause = "WHERE " + strings.Join(where, " AND ")
		}

		// 不分页：原全量返回（前台导航用）
		if !usePaging {
			rows, err := config.DB.Query("SELECT id, name, icon, slug, sort, require_login, enabled FROM categories " + whereClause + " ORDER BY sort, created_at", args...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var categories []models.Category
			for rows.Next() {
				var c models.Category
				var enabled sql.NullBool
				if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.Slug, &c.Sort, &c.RequireLogin, &enabled); err != nil {
					continue
				}
				c.Enabled = enabled.Valid && enabled.Bool
				categories = append(categories, c)
			}
			json.NewEncoder(w).Encode(categories)
			return
		}

		// 分页：后台管理用
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize < 1 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100
		}

		// 查询总数
		var total int64
		countErr := config.DB.QueryRow("SELECT COUNT(*) FROM categories "+whereClause, args...).Scan(&total)
		if countErr != nil {
			http.Error(w, countErr.Error(), http.StatusInternalServerError)
			return
		}

		// 查询当前页数据
		offset := (page - 1) * pageSize
		queryArgs := append(args, pageSize, offset)
		rows, err := config.DB.Query("SELECT id, name, icon, slug, sort, require_login, enabled FROM categories "+whereClause+" ORDER BY sort, created_at LIMIT ? OFFSET ?", queryArgs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var categories []models.Category
		for rows.Next() {
			var c models.Category
			var enabled sql.NullBool
			if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.Slug, &c.Sort, &c.RequireLogin, &enabled); err != nil {
				continue
			}
			c.Enabled = enabled.Valid && enabled.Bool
			categories = append(categories, c)
		}
		totalPages := int(total) / pageSize
		if int(total)%pageSize != 0 {
			totalPages++
		}
		json.NewEncoder(w).Encode(models.PageResult{
			Items:      categories,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		})

	case http.MethodPost:
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
	case http.MethodGet:
		// 检查 token 有效性
		token := ExtractToken(r)
		isValidToken := token != "" && ValidateToken(token)

		categoryID := r.URL.Query().Get("category_id")

		// 分页/筛选参数（后台管理使用；前台导航不带这些参数，走全量返回）
		pageStr := r.URL.Query().Get("page")
		usePaging := pageStr != ""

		// 构造 WHERE 条件
		var where []string
		var args []interface{}
		if !isValidToken {
			where = append(where, "l.enabled = TRUE")
		}
		if categoryID != "" {
			where = append(where, "l.category_id = ?")
			args = append(args, categoryID)
		}
		if search := r.URL.Query().Get("search"); search != "" {
			where = append(where, "(l.name LIKE ? OR l.url LIKE ? OR l.description LIKE ?)")
			args = append(args, "%"+search+"%", "%"+search+"%", "%"+search+"%")
		}
		if enabled := r.URL.Query().Get("enabled"); enabled != "" && isValidToken {
			if enabled == "1" {
				where = append(where, "l.enabled = TRUE")
			} else if enabled == "0" {
				where = append(where, "l.enabled = FALSE")
			}
		}

		whereClause := ""
		if len(where) > 0 {
			whereClause = "WHERE " + strings.Join(where, " AND ")
		}

		// 不分页：原全量返回（前台导航用）
		if !usePaging {
			rows, err := config.DB.Query("SELECT l.id, l.category_id, l.name, l.url, l.icon, l.description, l.sort, l.enabled FROM links l "+whereClause+" ORDER BY l.category_id, l.sort, l.created_at", args...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var links []models.Link
			for rows.Next() {
				var l models.Link
				var enabled sql.NullBool
				if err := rows.Scan(&l.ID, &l.CategoryID, &l.Name, &l.URL, &l.Icon, &l.Desc, &l.Sort, &enabled); err != nil {
					continue
				}
				l.Enabled = enabled.Valid && enabled.Bool
				links = append(links, l)
			}
			json.NewEncoder(w).Encode(links)
			return
		}

		// 分页：后台管理用（JOIN categories 取分类名）
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize < 1 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100
		}

		// 查询总数
		var total int64
		countErr := config.DB.QueryRow("SELECT COUNT(*) FROM links l "+whereClause, args...).Scan(&total)
		if countErr != nil {
			http.Error(w, countErr.Error(), http.StatusInternalServerError)
			return
		}

		// 查询当前页数据（带分类名）
		offset := (page - 1) * pageSize
		queryArgs := append(args, pageSize, offset)
		rows, err := config.DB.Query(`SELECT l.id, l.category_id, l.name, l.url, l.icon, l.description, l.sort, l.enabled, c.name
			FROM links l LEFT JOIN categories c ON l.category_id = c.id
			`+whereClause+` ORDER BY l.category_id, l.sort, l.created_at LIMIT ? OFFSET ?`, queryArgs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		type LinkWithCategory struct {
			models.Link
			CategoryName string `json:"category_name"`
		}
		var links []LinkWithCategory
		for rows.Next() {
			var l LinkWithCategory
			var enabled sql.NullBool
			var catName sql.NullString
			if err := rows.Scan(&l.ID, &l.CategoryID, &l.Name, &l.URL, &l.Icon, &l.Desc, &l.Sort, &enabled, &catName); err != nil {
				continue
			}
			l.Enabled = enabled.Valid && enabled.Bool
			l.CategoryName = catName.String
			links = append(links, l)
		}
		totalPages := int(total) / pageSize
		if int(total)%pageSize != 0 {
			totalPages++
		}
		json.NewEncoder(w).Encode(models.PageResult{
			Items:      links,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		})

	case http.MethodPost:
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
