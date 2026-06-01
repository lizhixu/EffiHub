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
		rows, err := config.DB.Query("SELECT id, name, icon, slug, sort FROM categories ORDER BY sort")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var categories []models.Category
		for rows.Next() {
			var c models.Category
			rows.Scan(&c.ID, &c.Name, &c.Icon, &c.Slug, &c.Sort)
			categories = append(categories, c)
		}
		json.NewEncoder(w).Encode(categories)

	case "POST":
		var c models.Category
		json.NewDecoder(r.Body).Decode(&c)
		result, err := config.DB.Exec("INSERT INTO categories (name, icon, slug, sort) VALUES (?, ?, ?, ?)",
			c.Name, c.Icon, c.Slug, c.Sort)
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
		categoryID := r.URL.Query().Get("category_id")
		var rows *sql.Rows
		var err error

		if categoryID != "" {
			rows, err = config.DB.Query(
				"SELECT id, category_id, name, url, icon, description, sort FROM links WHERE category_id = ? ORDER BY sort",
				categoryID)
		} else {
			rows, err = config.DB.Query(
				"SELECT id, category_id, name, url, icon, description, sort FROM links ORDER BY category_id, sort")
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var links []models.Link
		for rows.Next() {
			var l models.Link
			rows.Scan(&l.ID, &l.CategoryID, &l.Name, &l.URL, &l.Icon, &l.Desc, &l.Sort)
			links = append(links, l)
		}
		json.NewEncoder(w).Encode(links)

	case "POST":
		var l models.Link
		json.NewDecoder(r.Body).Decode(&l)
		result, err := config.DB.Exec(
			"INSERT INTO links (category_id, name, url, icon, description, sort) VALUES (?, ?, ?, ?, ?, ?)",
			l.CategoryID, l.Name, l.URL, l.Icon, l.Desc, l.Sort)
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
		err := config.DB.QueryRow(
			"SELECT id, category_id, name, url, icon, description, sort FROM links WHERE id = ?", id).
			Scan(&l.ID, &l.CategoryID, &l.Name, &l.URL, &l.Icon, &l.Desc, &l.Sort)
		if err != nil {
			http.Error(w, "链接不存在", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(l)

	case "PUT":
		var l models.Link
		json.NewDecoder(r.Body).Decode(&l)
		_, err := config.DB.Exec(
			"UPDATE links SET category_id=?, name=?, url=?, icon=?, description=?, sort=? WHERE id=?",
			l.CategoryID, l.Name, l.URL, l.Icon, l.Desc, l.Sort, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		l.ID = id
		json.NewEncoder(w).Encode(l)

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
		err := config.DB.QueryRow(
			"SELECT id, name, icon, slug, sort FROM categories WHERE id = ?", id).
			Scan(&c.ID, &c.Name, &c.Icon, &c.Slug, &c.Sort)
		if err != nil {
			http.Error(w, "分类不存在", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(c)

	case "PUT":
		var c models.Category
		json.NewDecoder(r.Body).Decode(&c)
		_, err := config.DB.Exec(
			"UPDATE categories SET name=?, icon=?, slug=?, sort=? WHERE id=?",
			c.Name, c.Icon, c.Slug, c.Sort, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c.ID = id
		json.NewEncoder(w).Encode(c)

	case "DELETE":
		_, err := config.DB.Exec("DELETE FROM categories WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "删除成功"})
	}
}
