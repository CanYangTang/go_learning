package config

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
}

// DSN returns the data source name for MySQL.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

// Connect establishes a connection to MySQL and returns *sql.DB.
func Connect(cfg *DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}