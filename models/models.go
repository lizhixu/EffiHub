package models

import (
	"effihub/config"
)

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
	Slug string `json:"slug"`
	Sort int    `json:"sort"`
}

type Link struct {
	ID         int    `json:"id"`
	CategoryID int    `json:"category_id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Icon       string `json:"icon"`
	Desc       string `json:"desc"`
	Sort       int    `json:"sort"`
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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
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

	return nil
}
