package database

import (
	"database/sql"
	"fmt"
	"github.com/nootey/observv/pkg/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"sync"
	"time"
)

var (
	mysqlDB *gorm.DB
	once    sync.Once
)

func HashAndSaltPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func ConnectToDatabase(cfg *config.Config, disableLogging bool) (*gorm.DB, error) {
	var err error

	once.Do(func() {
		hosts := []string{cfg.DBHost, "localhost", "mysql"}
		logLevel := logger.Info
		if disableLogging {
			logLevel = logger.Silent
		}

		for _, host := range hosts {
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				cfg.DBUser, cfg.DBPassword, host, cfg.DBPort, cfg.Database)

			mysqlDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
				Logger: logger.Default.LogMode(logLevel),
			})
			if err != nil {
				log.Printf("Failed to connect to MySQL at %s: %v", host, err)
				continue
			}

			sqlDB, err := mysqlDB.DB()
			if err != nil {
				log.Printf("Failed to get mysql database instance: %v", err)
				continue
			}

			// Ping the database to check if the connection is alive
			sqlDB.SetConnMaxLifetime(time.Minute * 5)
			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetMaxOpenConns(100)

			if err := sqlDB.Ping(); err != nil {
				log.Printf("Could not ping MySQL at %s: %v", host, err)
				continue
			}

			log.Printf("Connected to MySQL at %s", host)
			return
		}

		err = fmt.Errorf("failed to connect to any MySQL host")
	})

	return mysqlDB, err
}

func DisconnectDatabase() error {
	if mysqlDB == nil {
		return nil
	}
	sqlDB, err := mysqlDB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func EnsureDatabaseExists(cfg *config.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	return nil
}

func DropDatabase(cfg *config.Config) error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/", cfg.DBUser, cfg.DBPassword, cfg.DBHost))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DROP DATABASE IF EXISTS " + cfg.Database)
	return err
}
