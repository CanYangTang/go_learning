package config

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DefaultDSN returns the DSN for the local development database started by
// deployments/docker-compose.yml.
func DefaultDSN() string {
	cfg := &DatabaseConfig{
		Host:         "127.0.0.1",
		Port:         3306,
		Username:     "root",
		Password:     "password",
		Database:     "go_learning",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}
	return cfg.DSN()
}

// ConnectGorm opens a GORM connection to MySQL.
// internal/config/database.go returns *sql.DB; the repository layer needs *gorm.DB.
//
// TranslateError is required for errors.Is(err, gorm.ErrDuplicatedKey) to work
// in internal/repository/user.go - without it, a duplicate-key insert returns
// the driver's raw *mysql.MySQLError instead of gorm.ErrDuplicatedKey.
func ConnectGorm(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, err
	}
	return db, nil
}
