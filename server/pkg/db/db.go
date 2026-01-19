package db

import (
	"database/sql"
	"fmt"
	"path"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/logging"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pressly/goose/v3"
	"github.com/spf13/viper"
)

var DB *gorm.DB

func DBconnection() error {
	var err error
	viper.AutomaticEnv()
	// here we are storing env variables to local var
	db_name := config.Config.Database.DBName
	db_username := config.Config.Database.DBUser
	db_password := config.Config.Database.DBPassword
	db_host := config.Config.Database.Host
	db_port := config.Config.Database.Port
	db_sslmode := config.Config.Database.SSLMode
	// string must be kept in sync with ./goose.sh

	DBurl := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		db_username,
		db_password,
		db_host,
		db_port,
		db_name,
		db_sslmode,
	)

	logger := logging.Logger.With("category", "gorm")

	DB, err = gorm.Open(constant.DB_DRIVER, DBurl)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	sqlDB := DB.DB()
	if sqlDB == nil {
		return fmt.Errorf("failed to get the underlying *sql.DB")
	}
	sqlDB.SetMaxOpenConns(config.Config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Config.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.Config.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.Config.Database.ConnMaxIdleTime)

	DB.LogMode(true)
	DB.SetLogger(&logging.GormLogger{
		Logger: logger,
	})

	return nil
}

func Automigrate(sqlDB *sql.DB, relativeRootPath string) error {
	err := goose.SetDialect(constant.DB_DRIVER)
	if err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// execute common scripts
	if err := goose.Up(sqlDB, path.Join(relativeRootPath, "pkg/db/scripts"), goose.WithAllowMissing()); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

func SeedDatabase(sqlDB *sql.DB, relativeRootPath string) error {
	err := goose.SetDialect(constant.DB_DRIVER)
	if err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// execute common scripts
	if err := goose.Up(sqlDB, relativeRootPath+"pkg/db/seeds", goose.WithNoVersioning()); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	return nil
}
