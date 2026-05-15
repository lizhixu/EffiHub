package config

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	_ "embed"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

//go:embed ca.pem
var caCert []byte

var DB *sql.DB

func InitDB() error {
	// 配置 TLS
	rootCertPool := x509.NewCertPool()
	if ok := rootCertPool.AppendCertsFromPEM(caCert); !ok {
		return fmt.Errorf("failed to append CA cert")
	}

	mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs:    rootCertPool,
		MinVersion: tls.VersionTLS12,
	})

	dbUser := getEnvOrDefault("DB_USER", "root")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "")
	dbHost := getEnvOrDefault("DB_HOST", "127.0.0.1")
	dbPort := getEnvOrDefault("DB_PORT", "3306")
	dbName := getEnvOrDefault("DB_NAME", "effihub")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=custom&parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	return nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
