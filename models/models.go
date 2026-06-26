package models

import (
	"effihub/config"
)

type Category struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	Slug         string `json:"slug"`
	Sort         int    `json:"sort"`
	RequireLogin bool   `json:"require_login"`
	Enabled      bool   `json:"enabled"`
}

type Link struct {
	ID         int    `json:"id"`
	CategoryID int    `json:"category_id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Icon       string `json:"icon"`
	Desc       string `json:"desc"`
	Sort       int    `json:"sort"`
	Enabled    bool   `json:"enabled"`
}

func InitTables() error {
	// 创建分类表
	_, err := config.DB.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			icon TEXT,
			slug VARCHAR(50) UNIQUE,
			sort INT DEFAULT 0,
			require_login BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 自动迁移：添加 require_login 字段（如果不存在）
	var colExists bool
	config.DB.QueryRow(`SELECT COUNT(*) > 0 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'categories' AND COLUMN_NAME = 'require_login'`).Scan(&colExists)
	if !colExists {
		config.DB.Exec(`ALTER TABLE categories ADD COLUMN require_login BOOLEAN DEFAULT FALSE`)
	}

	// 自动迁移：添加 enabled 字段到 categories 表（如果不存在）
	config.DB.QueryRow(`SELECT COUNT(*) > 0 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'categories' AND COLUMN_NAME = 'enabled'`).Scan(&colExists)
	if !colExists {
		config.DB.Exec(`ALTER TABLE categories ADD COLUMN enabled BOOLEAN DEFAULT TRUE`)
	}

	// 创建链接表
	_, err = config.DB.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id INT AUTO_INCREMENT PRIMARY KEY,
			category_id INT NOT NULL,
			name VARCHAR(100) NOT NULL,
			url TEXT NOT NULL,
			icon TEXT,
			description VARCHAR(500),
			sort INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// 自动迁移：添加 enabled 字段到 links 表（如果不存在）
	config.DB.QueryRow(`SELECT COUNT(*) > 0 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'links' AND COLUMN_NAME = 'enabled'`).Scan(&colExists)
	if !colExists {
		config.DB.Exec(`ALTER TABLE links ADD COLUMN enabled BOOLEAN DEFAULT TRUE`)
	}

	return nil
}
